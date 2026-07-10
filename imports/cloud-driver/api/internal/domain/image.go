package domain

import "time"

// Image is one private image asset.
type Image struct {
	UID               string
	ContentUID        string
	AlbumUID          *string
	DisplayName       string
	OriginalFileName  string
	ContentType       string
	SizeBytes         int64
	Width             int
	Height            int
	PreviewStorageKey string
	ProcessingStatus  ProcessingStatus
	CreateTime        time.Time
	UpdateTime        time.Time
	DeleteTime        *time.Time
}

// ImagePage is one cursor page of images.
type ImagePage struct {
	Images        []Image
	NextPageToken string
}

// ImageAlbum is one private image collection.
type ImageAlbum struct {
	UID         string
	DisplayName string
	ImageCount  int
	CreateTime  time.Time
}

// ImageLink is one stable anonymous raw-image link.
type ImageLink struct {
	UID        string
	ImageUID   string
	CreateTime time.Time
	RevokeTime *time.Time
}

// Wallpaper is one manually published image.
type Wallpaper struct {
	UID           string
	ImageUID      string
	Title         string
	Tags          []string
	DominantColor string
	Width         int
	Height        int
	ContentType   string
	StorageKey    string
	PublishTime   time.Time
	UpdateTime    time.Time
	Variants      []Artifact
}

// WallpaperPage is one cursor page of public wallpapers.
type WallpaperPage struct {
	Wallpapers    []Wallpaper
	NextPageToken string
}
