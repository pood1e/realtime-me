package domain

import "time"

// ItemKind is the persisted kind of a drive item.
type ItemKind string

const (
	// ItemKindFile identifies a file backed by a blob.
	ItemKindFile ItemKind = "file"
	// ItemKindDirectory identifies a directory in the hierarchy.
	ItemKindDirectory ItemKind = "directory"
)

// Item is the domain representation of a drive file or directory.
type Item struct {
	UID         string
	ParentUID   *string
	Name        string
	Kind        ItemKind
	SizeBytes   int64
	ContentType string
	StorageKey  string
	CreateTime  time.Time
	UpdateTime  time.Time
	DeleteTime  *time.Time
}

// UploadStatus is the lifecycle state of an upload.
type UploadStatus string

const (
	// UploadStatusActive accepts new chunks.
	UploadStatusActive UploadStatus = "active"
	// UploadStatusCompleted has published a file.
	UploadStatusCompleted UploadStatus = "completed"
	// UploadStatusExpired no longer accepts chunks.
	UploadStatusExpired UploadStatus = "expired"
)

// Upload describes a resumable file transfer.
type Upload struct {
	UID            string
	ItemUID        string
	ParentUID      *string
	FileName       string
	ContentType    string
	TotalSizeBytes int64
	ReceivedBytes  int64
	ChunkSizeBytes int64
	Status         UploadStatus
	CreateTime     time.Time
	ExpireTime     time.Time
	Chunks         []UploadChunk
}

// UploadChunk is an acknowledged byte range with an exclusive end.
type UploadChunk struct {
	StartOffset int64
	EndOffset   int64
	Checksum    []byte
}

// ShareLink is a read-only link. Its bearer token is never stored in this type.
type ShareLink struct {
	UID        string
	TargetUID  string
	CreateTime time.Time
	ExpireTime time.Time
	RevokeTime *time.Time
}

// Page is a cursor page of items.
type Page struct {
	Items         []Item
	NextPageToken string
}

// SharePage is a cursor page of owner-visible share links.
type SharePage struct {
	ShareLinks    []ShareLink
	NextPageToken string
}
