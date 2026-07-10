package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	bookColumns = `book.uid, book.content_uid, book.title, book.authors, book.series, book.series_number,
		book.description, book.format, book.original_file_name, content.size_bytes, book.page_count,
		COALESCE(cover.storage_key, ''), book.processing_status, book.create_time, book.update_time, book.delete_time`
	bookFrom = `books book JOIN content_objects content ON content.uid = book.content_uid
		LEFT JOIN content_artifacts cover ON cover.content_uid = book.content_uid
		AND cover.kind = 'book_cover' AND cover.variant = 'default'`
)

// GetBook returns one catalog entry.
func (s *Store) GetBook(ctx context.Context, uid string, includeTrashed bool) (domain.Book, error) {
	query := "SELECT " + bookColumns + " FROM " + bookFrom + " WHERE book.uid = $1"
	if !includeTrashed {
		query += " AND book.delete_time IS NULL"
	}
	book, err := scanBook(s.pool.QueryRow(ctx, query, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Book{}, fmt.Errorf("%w: book", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Book{}, fmt.Errorf("get book: %w", err)
	}
	return book, nil
}

// ListBooks lists catalog entries with optional search and collection filters.
func (s *Store) ListBooks(ctx context.Context, queryText, shelfUID string, format domain.BookFormat, trashed bool, pageSize int, pageToken string) (domain.BookPage, error) {
	cursor, err := decodeCursor(pageToken)
	if err != nil {
		return domain.BookPage{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := "SELECT " + bookColumns + " FROM " + bookFrom
	arguments := []any{}
	conditions := []string{"book.delete_time IS " + map[bool]string{true: "NOT NULL", false: "NULL"}[trashed]}
	if queryText != "" {
		arguments = append(arguments, queryText)
		conditions = append(conditions, fmt.Sprintf("(book.title ILIKE '%%' || $%d || '%%' OR array_to_string(book.authors, ' ') ILIKE '%%' || $%d || '%%')", len(arguments), len(arguments)))
	}
	if shelfUID != "" {
		arguments = append(arguments, shelfUID)
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM shelf_books membership WHERE membership.book_uid = book.uid AND membership.shelf_uid = $%d)", len(arguments)))
	}
	if format != "" {
		arguments = append(arguments, string(format))
		conditions = append(conditions, fmt.Sprintf("book.format = $%d", len(arguments)))
	}
	if cursor != nil {
		arguments = append(arguments, cursor.name, cursor.uid)
		conditions = append(conditions, fmt.Sprintf("(book.title, book.uid) > ($%d, $%d)", len(arguments)-1, len(arguments)))
	}
	query += " WHERE " + strings.Join(conditions, " AND ")
	arguments = append(arguments, pageSize+1)
	query += fmt.Sprintf(" ORDER BY book.title, book.uid LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.BookPage{}, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()
	page := domain.BookPage{}
	for rows.Next() {
		book, err := scanBook(rows)
		if err != nil {
			return domain.BookPage{}, fmt.Errorf("scan book: %w", err)
		}
		page.Books = append(page.Books, book)
	}
	if len(page.Books) > pageSize {
		last := page.Books[pageSize-1]
		page.Books = page.Books[:pageSize]
		page.NextPageToken = encodeCursor(last.Title, last.UID)
	}
	return page, rows.Err()
}

// ImportBook claims an upload as a PDF or EPUB publication.
func (s *Store) ImportBook(ctx context.Context, uploadUID string, sealed domain.SealedContent) (domain.Book, error) {
	format := bookFormat(sealed.ContentType)
	if format == "" {
		return domain.Book{}, fmt.Errorf("%w: upload is not PDF or EPUB", domain.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Book{}, fmt.Errorf("begin book import: %w", err)
	}
	defer tx.Rollback(ctx)
	upload, err := lockCompleteUpload(ctx, tx, uploadUID)
	if err != nil {
		return domain.Book{}, err
	}
	if upload.Status == domain.UploadStatusClaimed {
		if err := tx.Commit(ctx); err != nil {
			return domain.Book{}, fmt.Errorf("commit repeated book import: %w", err)
		}
		return s.GetBook(ctx, upload.ClaimedUID, false)
	}
	content, err := upsertContent(ctx, tx, sealed)
	if err != nil {
		return domain.Book{}, err
	}
	var existingUID string
	err = tx.QueryRow(ctx, "SELECT uid FROM books WHERE content_uid = $1", content.UID).Scan(&existingUID)
	if err == nil {
		if err := markUploadClaimed(ctx, tx, uploadUID, existingUID); err != nil {
			return domain.Book{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return domain.Book{}, fmt.Errorf("commit deduplicated book import: %w", err)
		}
		return s.GetBook(ctx, existingUID, true)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Book{}, fmt.Errorf("find imported book: %w", err)
	}
	bookUID := uuid.NewString()
	title := displayName(upload.FileName)
	if _, err := tx.Exec(ctx, `INSERT INTO books
		(uid, content_uid, title, format, original_file_name, processing_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')`, bookUID, content.UID, title, format, upload.FileName); err != nil {
		return domain.Book{}, fmt.Errorf("create book: %w", err)
	}
	if err := enqueueJob(ctx, tx, "book", bookUID); err != nil {
		return domain.Book{}, err
	}
	if err := markUploadClaimed(ctx, tx, uploadUID, bookUID); err != nil {
		return domain.Book{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Book{}, fmt.Errorf("commit book import: %w", err)
	}
	return s.GetBook(ctx, bookUID, false)
}

// UpdateBook replaces owner-editable metadata.
func (s *Store) UpdateBook(ctx context.Context, book domain.Book) (domain.Book, error) {
	command, err := s.pool.Exec(ctx, `UPDATE books SET title = $2, authors = $3, series = $4,
		series_number = $5, description = $6, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, book.UID, book.Title, book.Authors, book.Series, book.SeriesNumber, book.Description)
	if err != nil {
		return domain.Book{}, fmt.Errorf("update book: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Book{}, fmt.Errorf("%w: book", domain.ErrNotFound)
	}
	return s.GetBook(ctx, book.UID, false)
}

// TrashBook moves a book into its catalog trash.
func (s *Store) TrashBook(ctx context.Context, uid string) (domain.Book, error) {
	command, err := s.pool.Exec(ctx, `UPDATE books SET delete_time = now(), update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid)
	if err != nil {
		return domain.Book{}, fmt.Errorf("trash book: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Book{}, fmt.Errorf("%w: book", domain.ErrNotFound)
	}
	return s.GetBook(ctx, uid, true)
}

// RestoreBook restores a trashed book.
func (s *Store) RestoreBook(ctx context.Context, uid string) (domain.Book, error) {
	command, err := s.pool.Exec(ctx, `UPDATE books SET delete_time = NULL, update_time = now()
		WHERE uid = $1 AND delete_time IS NOT NULL`, uid)
	if err != nil {
		return domain.Book{}, fmt.Errorf("restore book: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Book{}, fmt.Errorf("%w: trashed book", domain.ErrNotFound)
	}
	return s.GetBook(ctx, uid, false)
}

// PurgeBook permanently deletes one trashed book.
func (s *Store) PurgeBook(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM books WHERE uid = $1 AND delete_time IS NOT NULL", uid)
	if err != nil {
		return fmt.Errorf("purge book: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: trashed book", domain.ErrNotFound)
	}
	return nil
}

// EmptyBookTrash permanently deletes every trashed book.
func (s *Store) EmptyBookTrash(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM books WHERE delete_time IS NOT NULL")
	return wrapDatabaseError("empty book trash", err)
}

// PurgeTrashedBooks deletes books past retention.
func (s *Store) PurgeTrashedBooks(ctx context.Context, cutoff time.Time) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM books WHERE delete_time <= $1", cutoff)
	return wrapDatabaseError("purge expired books", err)
}

// QueueBookProcessing explicitly retries metadata extraction.
func (s *Store) QueueBookProcessing(ctx context.Context, uid string) (domain.Book, error) {
	return s.queueProcessing(ctx, "books", "book", uid, func() (domain.Book, error) { return s.GetBook(ctx, uid, false) })
}

// GetReadingProgress returns the latest stored reader position.
