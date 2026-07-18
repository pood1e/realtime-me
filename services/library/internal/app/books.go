package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

// BookService manages the private bookshelf without routing through the drive.
type BookService struct {
	store    domain.BookStore
	contents domain.ContentStore
	content  *ContentService
	files    ContentReader
}

// NewBookService constructs the bookshelf application service.
func NewBookService(store domain.BookStore, contents domain.ContentStore, content *ContentService, files ContentReader) *BookService {
	return &BookService{store: store, contents: contents, content: content, files: files}
}

func (s *BookService) Get(ctx context.Context, uid string) (domain.Book, error) {
	return s.store.GetBook(ctx, uid, false)
}

func (s *BookService) List(ctx context.Context, query, shelfUID string, format domain.BookFormat, trashed bool, pageSize int, pageToken string) (domain.BookPage, error) {
	return s.store.ListBooks(ctx, domain.BookListQuery{
		Query: strings.TrimSpace(query), ShelfUID: strings.TrimSpace(shelfUID),
		Format: format, Trashed: trashed, PageSize: pageSize, PageToken: pageToken,
	})
}

func (s *BookService) Import(ctx context.Context, uploadUID string) (domain.Book, error) {
	upload, sealed, unlock, err := s.content.SealForClaim(ctx, uploadUID)
	if err != nil {
		return domain.Book{}, err
	}
	defer unlock()
	book, err := s.store.ImportBook(ctx, uploadUID, sealed)
	if err != nil {
		return domain.Book{}, err
	}
	if upload.Status != domain.UploadStatusClaimed {
		_ = s.content.FinishClaim(ctx, uploadUID)
	}
	return book, nil
}

func (s *BookService) Update(ctx context.Context, book domain.Book) (domain.Book, error) {
	if err := validateDisplayName(book.Title); err != nil {
		return domain.Book{}, err
	}
	book.Authors = normalizedStrings(book.Authors, 20, 255)
	if len(book.Series) > 255 || len(book.SeriesNumber) > 32 || len(book.Description) > 16<<10 {
		return domain.Book{}, fmt.Errorf("%w: book metadata is too long", domain.ErrInvalidArgument)
	}
	return s.store.UpdateBook(ctx, book)
}

func (s *BookService) Trash(ctx context.Context, uid string) (domain.Book, error) {
	return s.store.TrashBook(ctx, uid)
}

func (s *BookService) Restore(ctx context.Context, uid string) (domain.Book, error) {
	return s.store.RestoreBook(ctx, uid)
}

func (s *BookService) Purge(ctx context.Context, uid string) error {
	if err := s.store.PurgeBook(ctx, uid); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *BookService) EmptyTrash(ctx context.Context) error {
	if err := s.store.EmptyBookTrash(ctx); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *BookService) RetryProcessing(ctx context.Context, uid string) (domain.Book, error) {
	return s.store.QueueBookProcessing(ctx, uid)
}

func (s *BookService) GetProgress(ctx context.Context, uid string) (domain.ReadingProgress, error) {
	return s.store.GetReadingProgress(ctx, uid)
}

func (s *BookService) UpdateProgress(ctx context.Context, progress domain.ReadingProgress) (domain.ReadingProgress, error) {
	if progress.ProgressPercent < 0 || progress.ProgressPercent > 1 {
		return domain.ReadingProgress{}, fmt.Errorf("%w: progress must be between zero and one", domain.ErrInvalidArgument)
	}
	book, err := s.store.GetBook(ctx, progress.BookUID, false)
	if err != nil {
		return domain.ReadingProgress{}, err
	}
	switch book.Format {
	case domain.BookFormatPDF:
		if progress.LocationKind != "pdf" || progress.PDFPageNumber < 1 || progress.PDFPageCount < progress.PDFPageNumber {
			return domain.ReadingProgress{}, fmt.Errorf("%w: invalid PDF location", domain.ErrInvalidArgument)
		}
		progress.EPUBCFI = ""
	case domain.BookFormatEPUB:
		if progress.LocationKind != "epub" || strings.TrimSpace(progress.EPUBCFI) == "" || len(progress.EPUBCFI) > 4096 {
			return domain.ReadingProgress{}, fmt.Errorf("%w: invalid EPUB location", domain.ErrInvalidArgument)
		}
		progress.PDFPageNumber, progress.PDFPageCount = 0, 0
	default:
		return domain.ReadingProgress{}, fmt.Errorf("%w: unsupported book format", domain.ErrConflict)
	}
	return s.store.UpsertReadingProgress(ctx, progress)
}

func (s *BookService) ListShelves(ctx context.Context) ([]domain.Shelf, error) {
	return s.store.ListShelves(ctx)
}

func (s *BookService) CreateShelf(ctx context.Context, displayName string) (domain.Shelf, error) {
	if err := validateDisplayName(displayName); err != nil {
		return domain.Shelf{}, err
	}
	return s.store.CreateShelf(ctx, domain.Shelf{UID: uuid.NewString(), DisplayName: strings.TrimSpace(displayName)})
}

func (s *BookService) DeleteShelf(ctx context.Context, uid string) error {
	return s.store.DeleteShelf(ctx, uid)
}

func (s *BookService) AddToShelf(ctx context.Context, shelfUID, bookUID string) error {
	return s.store.AddBookToShelf(ctx, shelfUID, bookUID)
}

func (s *BookService) RemoveFromShelf(ctx context.Context, shelfUID, bookUID string) error {
	return s.store.RemoveBookFromShelf(ctx, shelfUID, bookUID)
}

func (s *BookService) OpenContent(ctx context.Context, uid string) (*os.File, domain.Book, error) {
	book, err := s.store.GetBook(ctx, uid, false)
	if err != nil {
		return nil, domain.Book{}, err
	}
	content, err := s.contents.GetContent(ctx, book.ContentUID)
	if err != nil {
		return nil, domain.Book{}, err
	}
	file, err := s.files.Open(ctx, content.StorageKey)
	return file, book, err
}

func (s *BookService) OpenCover(ctx context.Context, uid string) (*os.File, domain.Book, error) {
	book, err := s.store.GetBook(ctx, uid, false)
	if err != nil {
		return nil, domain.Book{}, err
	}
	if book.CoverStorageKey == "" {
		return nil, domain.Book{}, fmt.Errorf("%w: book cover", domain.ErrNotFound)
	}
	file, err := s.files.Open(ctx, book.CoverStorageKey)
	return file, book, err
}

func normalizedStrings(values []string, maximumCount, maximumLength int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > maximumLength {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if len(result) == maximumCount {
			break
		}
	}
	return result
}
