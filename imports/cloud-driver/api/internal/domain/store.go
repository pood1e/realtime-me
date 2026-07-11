package domain

import (
	"context"
	"time"
)

// HealthStore verifies persistence availability without exposing feature operations.
type HealthStore interface {
	Ping(context.Context) error
}

// DriveStore persists the generic drive hierarchy and its share links.
type DriveStore interface {
	GetItem(context.Context, string, bool) (Item, error)
	ListItems(context.Context, DriveListQuery) (Page, error)
	ListTrashedItems(context.Context, int, string) (Page, error)
	SearchItems(context.Context, string, int, string) (Page, error)
	CreateDirectory(context.Context, *string, string) (Item, error)
	RenameItem(context.Context, string, string) (Item, error)
	MoveItem(context.Context, string, *string) (Item, error)
	TrashItem(context.Context, string) (Item, error)
	RestoreItem(context.Context, string) (Item, error)
	PurgeTrashedItem(context.Context, string) error
	EmptyTrash(context.Context) error
	PurgeTrashedItems(context.Context, time.Time) error
	ImportDriveFile(context.Context, string, *string, string, SealedContent) (Item, error)
	CreateShare(context.Context, ShareLink, []byte) (ShareLink, error)
	ListShareLinks(context.Context, string, int, string) (SharePage, error)
	GetShareByTokenHash(context.Context, []byte) (ShareLink, error)
	RevokeShare(context.Context, string) (ShareLink, error)
	ListSharedItems(context.Context, string, *string, int, string) (Page, error)
	CanReadSharedItem(context.Context, string, string) (Item, error)
}

// UploadStore persists application-neutral upload sessions.
type UploadStore interface {
	CreateUpload(context.Context, Upload) (Upload, error)
	ReservedUploadBytes(context.Context) (int64, error)
	GetUpload(context.Context, string) (Upload, error)
	RecordUploadChunk(context.Context, string, UploadChunk) (Upload, error)
	BeginUploadFinalization(context.Context, string) (Upload, error)
	DeleteUpload(context.Context, string) error
	DeleteRetainedUpload(context.Context, string) error
	ListDiscardableUploads(context.Context, time.Time, time.Time) ([]Upload, error)
}

// ContentStore persists immutable content metadata and migration state.
type ContentStore interface {
	ListUnhashedContent(context.Context, int) ([]ContentObject, error)
	CommitContentMigration(context.Context, string, SealedContent) (string, error)
	FinalizeContentMigration(context.Context) error
	GetContent(context.Context, string) (ContentObject, error)
	ListUnreferencedContent(context.Context, int) ([]ContentObject, error)
	TombstoneContent(context.Context, string, time.Time) error
	ListExpiredContentTombstones(context.Context, time.Time, int) ([]ContentTombstone, error)
	DeleteContentTombstone(context.Context, string) error
	ReferencedStorageKeys(context.Context, []string) ([]string, error)
}

// BookStore persists the private book catalog and reading state.
type BookStore interface {
	GetBook(context.Context, string, bool) (Book, error)
	ListBooks(context.Context, BookListQuery) (BookPage, error)
	ImportBook(context.Context, string, SealedContent) (Book, error)
	UpdateBook(context.Context, Book) (Book, error)
	TrashBook(context.Context, string) (Book, error)
	RestoreBook(context.Context, string) (Book, error)
	PurgeBook(context.Context, string) error
	EmptyBookTrash(context.Context) error
	PurgeTrashedBooks(context.Context, time.Time) error
	QueueBookProcessing(context.Context, string) (Book, error)
	GetReadingProgress(context.Context, string) (ReadingProgress, error)
	UpsertReadingProgress(context.Context, ReadingProgress) (ReadingProgress, error)
	ListShelves(context.Context) ([]Shelf, error)
	CreateShelf(context.Context, Shelf) (Shelf, error)
	DeleteShelf(context.Context, string) error
	AddBookToShelf(context.Context, string, string) error
	RemoveBookFromShelf(context.Context, string, string) error
	AdoptDriveBooks(context.Context) (int64, error)
}

// MusicLibraryStore persists the private audio catalog and playback state.
type MusicLibraryStore interface {
	GetTrack(context.Context, string, bool) (Track, error)
	GetTrackBySource(context.Context, MusicProvider, string) (Track, error)
	ListTracks(context.Context, TrackListQuery) (TrackPage, error)
	ImportTrack(context.Context, string, SealedContent) (Track, error)
	SetTrackFavorite(context.Context, string, bool) (Track, error)
	TrashTrack(context.Context, string) (Track, error)
	RestoreTrack(context.Context, string) (Track, error)
	PurgeTrack(context.Context, string) error
	EmptyTrackTrash(context.Context) error
	PurgeTrashedTracks(context.Context, time.Time) error
	QueueTrackProcessing(context.Context, string) (Track, error)
	RecordPlayback(context.Context, PlaybackEntry) (PlaybackEntry, error)
	ListPlaybackHistory(context.Context, int, string) (PlaybackPage, error)
}

// MusicPlaylistStore persists imported playlist operations and snapshots.
type MusicPlaylistStore interface {
	QueuePlaylistImport(context.Context, PlaylistImport) (PlaylistImport, error)
	GetPlaylistImport(context.Context, string) (PlaylistImport, error)
	GetPlaylist(context.Context, string) (Playlist, error)
	ListPlaylists(context.Context, int, string) (PlaylistPage, error)
	ListPlaylistTracks(context.Context, string, int, string) (PlaylistTrackPage, error)
	QueuePlaylistDownload(context.Context, string) (Playlist, error)
	DeletePlaylist(context.Context, string) error
}

// MusicStore is the wiring-time aggregate implemented by the PostgreSQL adapter.
type MusicStore interface {
	MusicLibraryStore
	MusicPlaylistStore
}

// MusicProviderStore persists encrypted external accounts and login attempts.
type MusicProviderStore interface {
	ListProviderConnections(context.Context) ([]ProviderConnection, error)
	GetProviderConnection(context.Context, MusicProvider) (ProviderConnection, error)
	UpsertProviderConnection(context.Context, ProviderConnection) (ProviderConnection, error)
	DeleteProviderConnection(context.Context, MusicProvider) error
	CreateProviderConnectionAttempt(context.Context, ProviderConnectionAttempt) (ProviderConnectionAttempt, error)
	GetProviderConnectionAttempt(context.Context, string) (ProviderConnectionAttempt, error)
	GetProviderConnectionAttemptByStateHash(context.Context, []byte) (ProviderConnectionAttempt, error)
	UpdateProviderConnectionAttempt(context.Context, ProviderConnectionAttempt) (ProviderConnectionAttempt, error)
	ConsumeProviderConnectionAttempt(context.Context, string, time.Time) error
	PurgeExpiredProviderConnectionAttempts(context.Context, time.Time) error
}

// ImageStore persists private images and anonymous links.
type ImageStore interface {
	GetImage(context.Context, string, bool) (Image, error)
	ListImages(context.Context, ImageListQuery) (ImagePage, error)
	ImportImage(context.Context, string, *string, SealedContent) (Image, error)
	UpdateImage(context.Context, Image) (Image, error)
	TrashImage(context.Context, string) (Image, error)
	RestoreImage(context.Context, string) (Image, error)
	PurgeImage(context.Context, string) error
	EmptyImageTrash(context.Context) error
	PurgeTrashedImages(context.Context, time.Time) error
	QueueImageProcessing(context.Context, string) (Image, error)
	ListImageAlbums(context.Context) ([]ImageAlbum, error)
	CreateImageAlbum(context.Context, ImageAlbum) (ImageAlbum, error)
	DeleteImageAlbum(context.Context, string) error
	ListImageLinks(context.Context, string) ([]ImageLink, error)
	CreateImageLink(context.Context, ImageLink) (ImageLink, error)
	RevokeImageLink(context.Context, string) (ImageLink, error)
	GetImageByLink(context.Context, string) (Image, error)
}

// WallpaperStore persists publication metadata for the public wallpaper catalog.
type WallpaperStore interface {
	GetWallpaper(context.Context, string) (Wallpaper, error)
	ListWallpapers(context.Context, WallpaperListQuery) (WallpaperPage, error)
	PublishWallpaper(context.Context, Wallpaper) (Wallpaper, error)
	UpdateWallpaper(context.Context, Wallpaper) (Wallpaper, error)
	UnpublishWallpaper(context.Context, string) error
}

// WorkerStore leases processing jobs and persists derived metadata.
type WorkerStore interface {
	HeartbeatWorker(context.Context, time.Time) error
	GetWorkerHealth(context.Context) (WorkerHealth, error)
	ClaimProcessingJob(context.Context, time.Time, time.Duration, []ProcessingJobKind) (*ProcessingJob, error)
	ExtendProcessingJobLease(context.Context, ProcessingJob, time.Time) error
	CompleteProcessingJob(context.Context, ProcessingJob) error
	FailProcessingJob(context.Context, ProcessingJob, string, bool, time.Time) error
	GetBookForProcessing(context.Context, string) (Book, ContentObject, error)
	CompleteBookProcessing(context.Context, ProcessingJob, string, []string, int, *Artifact) error
	GetTrackForProcessing(context.Context, string) (Track, ContentObject, error)
	CompleteTrackProcessing(context.Context, ProcessingJob, Track, *Artifact) error
	GetMusicDownload(context.Context, string) (MusicDownload, error)
	CompleteMusicDownload(context.Context, ProcessingJob, SealedContent) error
	GetImageForProcessing(context.Context, string) (Image, ContentObject, error)
	CompleteImageProcessing(context.Context, ProcessingJob, int, int, *Artifact) error
	GetWallpaperForProcessing(context.Context, string) (Wallpaper, ContentObject, error)
	CompleteWallpaperProcessing(context.Context, ProcessingJob, string, []Artifact) error
	GetUploadForFinalization(context.Context, string) (Upload, error)
	CompleteUploadFinalization(context.Context, ProcessingJob, SealedContent) error
	GetPlaylistImportForProcessing(context.Context, string) (PlaylistImport, error)
	CompletePlaylistImport(context.Context, ProcessingJob, Playlist, []PlayableTrack) error
}

// Clock supplies wall-clock time to application services.
type Clock interface {
	Now() time.Time
}

// SystemClock returns the current UTC wall-clock time.
type SystemClock struct{}

// Now returns the current UTC time.
func (SystemClock) Now() time.Time { return time.Now().UTC() }
