package domain

import "time"

// BookFormat identifies a supported publication container.
type BookFormat string

const (
	BookFormatPDF  BookFormat = "pdf"
	BookFormatEPUB BookFormat = "epub"
)

// Book is one private catalog entry.
type Book struct {
	UID              string
	ContentUID       string
	Title            string
	Authors          []string
	Series           string
	SeriesNumber     string
	Description      string
	Format           BookFormat
	OriginalFileName string
	SizeBytes        int64
	PageCount        int
	CoverStorageKey  string
	ProcessingStatus ProcessingStatus
	CreateTime       time.Time
	UpdateTime       time.Time
	DeleteTime       *time.Time
}

// BookPage is one cursor page of books.
type BookPage struct {
	Books         []Book
	NextPageToken string
}

// BookListQuery contains the complete cursor filter for the book catalog.
type BookListQuery struct {
	Query     string
	ShelfUID  string
	Format    BookFormat
	Trashed   bool
	PageSize  int
	PageToken string
}

// Shelf is one named collection.
type Shelf struct {
	UID         string
	DisplayName string
	BookCount   int
	CreateTime  time.Time
}

// ReadingProgress stores one format-specific reading position.
type ReadingProgress struct {
	BookUID         string
	ProgressPercent float32
	LocationKind    string
	PDFPageNumber   int
	PDFPageCount    int
	EPUBCFI         string
	UpdateTime      time.Time
}
