package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

// ContentFiles is the filesystem boundary used by request-path services.
type ContentFiles interface {
	PrepareUpload(context.Context, string) error
	WriteChunk(context.Context, string, int64, []byte) error
	RemoveUpload(context.Context, string) error
	Open(context.Context, string) (*os.File, error)
	Remove(context.Context, string) error
	WalkStoredFiles(context.Context, func(domain.StoredFile) error) error
	RemoveStoredFileIfOlder(context.Context, string, time.Time) error
	FreeBytes() (int64, error)
}

// ContentMigrationFiles exposes only offline legacy migration operations.
type ContentMigrationFiles interface {
	InspectLegacyObject(context.Context, domain.ContentObject) (domain.SealedContent, error)
	Remove(context.Context, string) error
}

// ContentReader opens an authorized storage key for application delivery.
type ContentReader interface {
	Open(context.Context, string) (*os.File, error)
}

// CapacityFiles reports capacity without exposing content mutation methods.
type CapacityFiles interface {
	FreeBytes() (int64, error)
}

// ContentService manages neutral uploads and their shared content lifecycle.
type ContentService struct {
	uploads           domain.UploadStore
	contents          domain.ContentStore
	files             ContentFiles
	clock             domain.Clock
	chunkSizeBytes    int64
	reservedFreeBytes int64
	uploadTTL         time.Duration
	locks             *keyedLocks
	capacityMu        sync.Mutex
}

// NewContentService constructs the shared content core.
func NewContentService(uploads domain.UploadStore, contents domain.ContentStore, files ContentFiles, clock domain.Clock, chunkSizeBytes, reservedFreeBytes int64, uploadTTL time.Duration) *ContentService {
	return &ContentService{
		uploads: uploads, contents: contents, files: files, clock: clock,
		chunkSizeBytes: chunkSizeBytes, reservedFreeBytes: reservedFreeBytes, uploadTTL: uploadTTL,
		locks: newKeyedLocks(),
	}
}

// StartUpload reserves capacity and creates a neutral session.
func (s *ContentService) StartUpload(ctx context.Context, fileName, contentType string, totalSizeBytes int64) (domain.Upload, error) {
	if err := validateName(fileName); err != nil {
		return domain.Upload{}, err
	}
	if totalSizeBytes < 0 {
		return domain.Upload{}, fmt.Errorf("%w: total size must not be negative", domain.ErrInvalidArgument)
	}
	s.capacityMu.Lock()
	defer s.capacityMu.Unlock()
	freeBytes, err := s.files.FreeBytes()
	if err != nil {
		return domain.Upload{}, fmt.Errorf("read free storage: %w", err)
	}
	reservedBytes, err := s.uploads.ReservedUploadBytes(ctx)
	if err != nil {
		return domain.Upload{}, err
	}
	if available := freeBytes - s.reservedFreeBytes - reservedBytes; available < 0 || totalSizeBytes > available {
		return domain.Upload{}, fmt.Errorf("%w: insufficient free storage", domain.ErrResourceExhausted)
	}
	now := s.clock.Now().UTC()
	upload := domain.Upload{
		UID: uuid.NewString(), FileName: fileName, ContentType: contentType,
		TotalSizeBytes: totalSizeBytes, ChunkSizeBytes: s.chunkSizeBytes,
		Status: domain.UploadStatusActive, CreateTime: now, ExpireTime: now.Add(s.uploadTTL),
	}
	if err := s.files.PrepareUpload(ctx, upload.UID); err != nil {
		return domain.Upload{}, err
	}
	created, err := s.uploads.CreateUpload(ctx, upload)
	if err != nil {
		_ = s.files.RemoveUpload(ctx, upload.UID)
		return domain.Upload{}, err
	}
	return created, nil
}

// GetUpload returns current upload state.
func (s *ContentService) GetUpload(ctx context.Context, uid string) (domain.Upload, error) {
	return s.uploads.GetUpload(ctx, uid)
}

// FinalizeUpload queues local hashing after every byte range is acknowledged.
func (s *ContentService) FinalizeUpload(ctx context.Context, uid string) (domain.Upload, error) {
	unlock := s.locks.Lock(uid)
	defer unlock()
	return s.uploads.BeginUploadFinalization(ctx, uid)
}

// WriteUploadChunk writes and acknowledges one bounded range.
func (s *ContentService) WriteUploadChunk(ctx context.Context, uploadUID string, startOffset, totalSizeBytes int64, data []byte) (domain.Upload, error) {
	unlock := s.locks.Lock(uploadUID)
	defer unlock()
	upload, err := s.uploads.GetUpload(ctx, uploadUID)
	if err != nil {
		return domain.Upload{}, err
	}
	if upload.Status != domain.UploadStatusActive {
		return domain.Upload{}, fmt.Errorf("%w: upload is not active", domain.ErrConflict)
	}
	if totalSizeBytes != upload.TotalSizeBytes || startOffset < 0 || len(data) == 0 {
		return domain.Upload{}, fmt.Errorf("%w: invalid upload chunk range", domain.ErrInvalidArgument)
	}
	if int64(len(data)) > upload.ChunkSizeBytes || startOffset > upload.TotalSizeBytes-int64(len(data)) {
		return domain.Upload{}, fmt.Errorf("%w: chunk exceeds upload bounds", domain.ErrInvalidArgument)
	}
	endOffset := startOffset + int64(len(data))
	checksum := sha256.Sum256(data)
	for _, existing := range upload.Chunks {
		if existing.StartOffset == startOffset {
			if existing.EndOffset == endOffset && bytesEqual(existing.Checksum, checksum[:]) {
				return upload, nil
			}
			return domain.Upload{}, fmt.Errorf("%w: chunk offset contains different content", domain.ErrConflict)
		}
		if existing.StartOffset < endOffset && existing.EndOffset > startOffset {
			return domain.Upload{}, fmt.Errorf("%w: chunk overlaps an acknowledged range", domain.ErrConflict)
		}
	}
	if err := s.files.WriteChunk(ctx, uploadUID, startOffset, data); err != nil {
		return domain.Upload{}, err
	}
	return s.uploads.RecordUploadChunk(ctx, uploadUID, domain.UploadChunk{StartOffset: startOffset, EndOffset: endOffset, Checksum: checksum[:]})
}

// AbandonUpload removes an unclaimed upload.
func (s *ContentService) AbandonUpload(ctx context.Context, uploadUID string) error {
	unlock := s.locks.Lock(uploadUID)
	defer unlock()
	upload, err := s.uploads.GetUpload(ctx, uploadUID)
	if err != nil {
		return err
	}
	if upload.Status == domain.UploadStatusClaimed {
		return fmt.Errorf("%w: upload is already claimed", domain.ErrConflict)
	}
	if upload.Status == domain.UploadStatusFinalizing {
		return fmt.Errorf("%w: upload finalization is running", domain.ErrConflict)
	}
	// Remove bytes first so a transient database failure leaves durable metadata
	// that the retention pass can retry instead of an unowned temporary file.
	if err := s.files.RemoveUpload(ctx, uploadUID); err != nil {
		return err
	}
	if err := s.uploads.DeleteUpload(ctx, uploadUID); err != nil {
		return err
	}
	return nil
}

// SealForClaim validates and publishes one upload without claiming it.
func (s *ContentService) SealForClaim(ctx context.Context, uploadUID string) (domain.Upload, domain.SealedContent, func(), error) {
	unlock := s.locks.Lock(uploadUID)
	upload, err := s.uploads.GetUpload(ctx, uploadUID)
	if err != nil {
		unlock()
		return domain.Upload{}, domain.SealedContent{}, nil, err
	}
	if upload.Status == domain.UploadStatusClaimed {
		return upload, domain.SealedContent{}, unlock, nil
	}
	if upload.Status != domain.UploadStatusSealed || upload.Sealed == nil {
		unlock()
		return domain.Upload{}, domain.SealedContent{}, nil, fmt.Errorf("%w: upload is not sealed", domain.ErrConflict)
	}
	return upload, *upload.Sealed, unlock, nil
}

// FinishClaim removes the temporary hard link after metadata commit.
func (s *ContentService) FinishClaim(ctx context.Context, uploadUID string) error {
	return s.files.RemoveUpload(ctx, uploadUID)
}

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var difference byte
	for index := range left {
		difference |= left[index] ^ right[index]
	}
	return difference == 0
}
