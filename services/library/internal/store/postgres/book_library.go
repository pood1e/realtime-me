package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func (s *Store) GetReadingProgress(ctx context.Context, bookUID string) (domain.ReadingProgress, error) {
	var progress domain.ReadingProgress
	err := s.pool.QueryRow(ctx, `SELECT book_uid, progress_percent, location_kind, pdf_page_number,
		pdf_page_count, epub_cfi, update_time FROM reading_progress WHERE book_uid = $1`, bookUID).Scan(
		&progress.BookUID, &progress.ProgressPercent, &progress.LocationKind, &progress.PDFPageNumber,
		&progress.PDFPageCount, &progress.EPUBCFI, &progress.UpdateTime)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, bookErr := s.GetBook(ctx, bookUID, false); bookErr != nil {
			return domain.ReadingProgress{}, bookErr
		}
		return domain.ReadingProgress{BookUID: bookUID}, nil
	}
	if err != nil {
		return domain.ReadingProgress{}, fmt.Errorf("get reading progress: %w", err)
	}
	return progress, nil
}

// UpsertReadingProgress stores the latest position.
func (s *Store) UpsertReadingProgress(ctx context.Context, progress domain.ReadingProgress) (domain.ReadingProgress, error) {
	if _, err := s.GetBook(ctx, progress.BookUID, false); err != nil {
		return domain.ReadingProgress{}, err
	}
	err := s.pool.QueryRow(ctx, `INSERT INTO reading_progress
		(book_uid, progress_percent, location_kind, pdf_page_number, pdf_page_count, epub_cfi)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (book_uid) DO UPDATE SET progress_percent = EXCLUDED.progress_percent,
		location_kind = EXCLUDED.location_kind, pdf_page_number = EXCLUDED.pdf_page_number,
		pdf_page_count = EXCLUDED.pdf_page_count, epub_cfi = EXCLUDED.epub_cfi, update_time = now()
		RETURNING update_time`, progress.BookUID, progress.ProgressPercent, progress.LocationKind,
		progress.PDFPageNumber, progress.PDFPageCount, progress.EPUBCFI).Scan(&progress.UpdateTime)
	if err != nil {
		return domain.ReadingProgress{}, fmt.Errorf("save reading progress: %w", err)
	}
	return progress, nil
}

// ListShelves lists collections and visible membership counts.
func (s *Store) ListShelves(ctx context.Context) ([]domain.Shelf, error) {
	rows, err := s.pool.Query(ctx, `SELECT shelf.uid, shelf.display_name,
		COUNT(book.uid) FILTER (WHERE book.delete_time IS NULL), shelf.create_time
		FROM shelves shelf LEFT JOIN shelf_books membership ON membership.shelf_uid = shelf.uid
		LEFT JOIN books book ON book.uid = membership.book_uid
		GROUP BY shelf.uid ORDER BY shelf.display_name, shelf.uid`)
	if err != nil {
		return nil, fmt.Errorf("list shelves: %w", err)
	}
	defer rows.Close()
	var shelves []domain.Shelf
	for rows.Next() {
		var shelf domain.Shelf
		if err := rows.Scan(&shelf.UID, &shelf.DisplayName, &shelf.BookCount, &shelf.CreateTime); err != nil {
			return nil, fmt.Errorf("scan shelf: %w", err)
		}
		shelves = append(shelves, shelf)
	}
	return shelves, rows.Err()
}

// CreateShelf creates a collection.
func (s *Store) CreateShelf(ctx context.Context, shelf domain.Shelf) (domain.Shelf, error) {
	err := s.pool.QueryRow(ctx, `INSERT INTO shelves (uid, display_name) VALUES ($1, $2)
		RETURNING create_time`, shelf.UID, shelf.DisplayName).Scan(&shelf.CreateTime)
	if err != nil {
		return domain.Shelf{}, fmt.Errorf("create shelf: %w", err)
	}
	return shelf, nil
}

// DeleteShelf deletes a collection without deleting books.
func (s *Store) DeleteShelf(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM shelves WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("delete shelf: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: shelf", domain.ErrNotFound)
	}
	return nil
}

// AddBookToShelf adds one membership idempotently.
func (s *Store) AddBookToShelf(ctx context.Context, shelfUID, bookUID string) error {
	if _, err := s.GetBook(ctx, bookUID, false); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO shelf_books (shelf_uid, book_uid) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, shelfUID, bookUID)
	return wrapDatabaseError("add book to shelf", err)
}

// RemoveBookFromShelf removes one membership.
func (s *Store) RemoveBookFromShelf(ctx context.Context, shelfUID, bookUID string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM shelf_books WHERE shelf_uid = $1 AND book_uid = $2", shelfUID, bookUID)
	return wrapDatabaseError("remove book from shelf", err)
}

// AdoptDriveBooks creates book entries for existing visible PDF and EPUB drive files.
