package domain

import (
	"context"
	"time"
)

// Store is the durable metadata boundary for the drive application.
type Store interface {
	Ping(context.Context) error
	GetItem(context.Context, string, bool) (Item, error)
	ListItems(context.Context, *string, bool, int, string) (Page, error)
	ListTrashedItems(context.Context, int, string) (Page, error)
	SearchItems(context.Context, string, int, string) (Page, error)
	CreateDirectory(context.Context, *string, string) (Item, error)
	RenameItem(context.Context, string, string) (Item, error)
	MoveItem(context.Context, string, *string) (Item, error)
	TrashItem(context.Context, string) (Item, error)
	RestoreItem(context.Context, string) (Item, error)
	PurgeTrashedItem(context.Context, string) ([]Item, error)
	EmptyTrash(context.Context) ([]Item, error)
	CreateUpload(context.Context, Upload) (Upload, error)
	ReservedUploadBytes(context.Context) (int64, error)
	GetUpload(context.Context, string) (Upload, error)
	RecordUploadChunk(context.Context, string, UploadChunk) (Upload, error)
	CompleteUpload(context.Context, string) (Item, error)
	CreateShare(context.Context, ShareLink, []byte) (ShareLink, error)
	ListShareLinks(context.Context, string, int, string) (SharePage, error)
	GetShareByTokenHash(context.Context, []byte) (ShareLink, error)
	RevokeShare(context.Context, string) (ShareLink, error)
	ListSharedItems(context.Context, string, *string, int, string) (Page, error)
	CanReadSharedItem(context.Context, string, string) (Item, error)
	ListExpiredUploads(context.Context, time.Time) ([]Upload, error)
	DeleteUpload(context.Context, string) error
	PurgeTrashedItems(context.Context, time.Time) ([]Item, error)
}

// Clock supplies wall-clock time to application services.
type Clock interface {
	Now() time.Time
}

// SystemClock returns the current UTC wall-clock time.
type SystemClock struct{}

// Now returns the current UTC time.
func (SystemClock) Now() time.Time { return time.Now().UTC() }
