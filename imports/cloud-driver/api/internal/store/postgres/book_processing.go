package postgres

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *Store) AdoptDriveBooks(ctx context.Context) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin drive book adoption: %w", err)
	}
	defer tx.Rollback(ctx)
	command, err := tx.Exec(ctx, `INSERT INTO books
		(uid, content_uid, title, format, original_file_name, processing_status)
	SELECT gen_random_uuid()::text, source.content_uid,
		regexp_replace(source.name, '\.(pdf|epub)$', '', 'i'),
		CASE WHEN lower(source.name) LIKE '%.epub' OR source.content_type = 'application/epub+zip' THEN 'epub' ELSE 'pdf' END,
		source.name, 'pending'
	FROM (
		SELECT DISTINCT ON (item.content_uid) item.content_uid, item.name, content.content_type
		FROM drive_items item JOIN content_objects content ON content.uid = item.content_uid
		WHERE item.kind = 'file' AND item.delete_time IS NULL
		AND (lower(item.name) LIKE '%.pdf' OR lower(item.name) LIKE '%.epub'
			OR content.content_type IN ('application/pdf', 'application/epub+zip'))
		ORDER BY item.content_uid, item.name
	) source
	ON CONFLICT (content_uid) DO NOTHING`)
	if err != nil {
		return 0, fmt.Errorf("adopt drive books: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO processing_jobs (uid, kind, resource_uid, status)
		SELECT gen_random_uuid()::text, 'book', book.uid, 'pending' FROM books book
		LEFT JOIN processing_jobs job ON job.kind = 'book' AND job.resource_uid = book.uid
		WHERE book.processing_status = 'pending' AND job.uid IS NULL`); err != nil {
		return 0, fmt.Errorf("queue adopted books: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit drive book adoption: %w", err)
	}
	return command.RowsAffected(), nil
}

// GetBookForProcessing returns the source metadata required by the worker.
func (s *Store) GetBookForProcessing(ctx context.Context, uid string) (domain.Book, domain.ContentObject, error) {
	book, err := s.GetBook(ctx, uid, true)
	if err != nil {
		return domain.Book{}, domain.ContentObject{}, err
	}
	content, err := s.GetContent(ctx, book.ContentUID)
	return book, content, err
}

// CompleteBookProcessing persists extracted metadata and optional cover.
func (s *Store) CompleteBookProcessing(ctx context.Context, job domain.ProcessingJob, title string, authors []string, pageCount int, cover *domain.Artifact) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin book processing completion: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := lockProcessingJobLease(ctx, tx, job); err != nil {
		return err
	}
	if cover != nil {
		if err := upsertArtifact(ctx, tx, *cover); err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `UPDATE books SET
		title = CASE WHEN title = regexp_replace(original_file_name, '\.[^.]+$', '') AND $2 <> '' THEN $2 ELSE title END,
		authors = CASE WHEN cardinality(authors) = 0 AND cardinality($3::text[]) > 0 THEN $3 ELSE authors END,
		page_count = $4, processing_status = 'ready', update_time = now() WHERE uid = $1`, job.ResourceUID, title, authors, pageCount)
	if err != nil {
		return fmt.Errorf("complete book processing: %w", err)
	}
	return tx.Commit(ctx)
}

func scanBook(row rowScanner) (domain.Book, error) {
	var book domain.Book
	var format, status string
	var deleteTime pgtype.Timestamptz
	if err := row.Scan(&book.UID, &book.ContentUID, &book.Title, &book.Authors, &book.Series, &book.SeriesNumber,
		&book.Description, &format, &book.OriginalFileName, &book.SizeBytes, &book.PageCount, &book.CoverStorageKey,
		&status, &book.CreateTime, &book.UpdateTime, &deleteTime); err != nil {
		return domain.Book{}, err
	}
	book.Format = domain.BookFormat(format)
	book.ProcessingStatus = domain.ProcessingStatus(status)
	if deleteTime.Valid {
		value := deleteTime.Time.UTC()
		book.DeleteTime = &value
	}
	return book, nil
}

func bookFormat(contentType string) domain.BookFormat {
	switch contentType {
	case "application/pdf":
		return domain.BookFormatPDF
	case "application/epub+zip":
		return domain.BookFormatEPUB
	default:
		return ""
	}
}

func displayName(fileName string) string {
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if strings.TrimSpace(name) == "" {
		return fileName
	}
	return name
}
