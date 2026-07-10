package app

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

// ContentFiles is the filesystem boundary used by request-path services.
type ContentFiles interface {
	PrepareUpload(context.Context, string) error
	WriteChunk(context.Context, string, int64, []byte) error
	SealUpload(context.Context, string, string) (domain.SealedContent, error)
	InspectLegacyObject(context.Context, domain.ContentObject) (domain.SealedContent, error)
	RemoveUpload(context.Context, string) error
	Open(context.Context, string) (*os.File, error)
	Remove(context.Context, string) error
	FreeBytes() (int64, error)
}

// ContentService manages neutral uploads, migration, and content garbage collection.
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
	if err := s.uploads.DeleteUpload(ctx, uploadUID); err != nil {
		return err
	}
	return s.files.RemoveUpload(ctx, uploadUID)
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
	if upload.Status != domain.UploadStatusActive || !chunksComplete(upload) {
		unlock()
		return domain.Upload{}, domain.SealedContent{}, nil, fmt.Errorf("%w: upload is incomplete", domain.ErrConflict)
	}
	sealed, err := s.files.SealUpload(ctx, upload.UID, upload.FileName)
	if err != nil {
		unlock()
		return domain.Upload{}, domain.SealedContent{}, nil, err
	}
	if sealed.SizeBytes != upload.TotalSizeBytes {
		unlock()
		return domain.Upload{}, domain.SealedContent{}, nil, fmt.Errorf("%w: uploaded size changed", domain.ErrConflict)
	}
	return upload, sealed, unlock, nil
}

// FinishClaim removes the temporary hard link after metadata commit.
func (s *ContentService) FinishClaim(ctx context.Context, uploadUID string) {
	_ = s.files.RemoveUpload(ctx, uploadUID)
}

// MigrateLegacyContent converts all pre-suite blobs to content-addressed storage.
func (s *ContentService) MigrateLegacyContent(ctx context.Context) error {
	for {
		objects, err := s.contents.ListUnhashedContent(ctx, 100)
		if err != nil {
			return err
		}
		if len(objects) == 0 {
			break
		}
		for _, object := range objects {
			sealed, err := s.files.InspectLegacyObject(ctx, object)
			if err != nil {
				return fmt.Errorf("migrate content %s: %w", object.UID, err)
			}
			if _, err := s.contents.CommitContentMigration(ctx, object.UID, sealed); err != nil {
				return err
			}
			if object.StorageKey != sealed.StorageKey {
				if err := s.files.Remove(ctx, object.StorageKey); err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}
		}
	}
	return s.contents.FinalizeContentMigration(ctx)
}

// PurgeExpiredUploads removes temporary sessions past their expiry.
func (s *ContentService) PurgeExpiredUploads(ctx context.Context) error {
	uploads, err := s.uploads.ListExpiredUploads(ctx, s.clock.Now().UTC())
	if err != nil {
		return err
	}
	for _, upload := range uploads {
		_ = s.files.RemoveUpload(ctx, upload.UID)
		if err := s.uploads.DeleteUpload(ctx, upload.UID); err != nil && !errors.Is(err, domain.ErrNotFound) {
			return err
		}
	}
	return nil
}

// CollectGarbage deletes unreferenced content metadata and source bytes.
func (s *ContentService) CollectGarbage(ctx context.Context) error {
	for {
		objects, err := s.contents.ListUnreferencedContent(ctx, 100)
		if err != nil {
			return err
		}
		if len(objects) == 0 {
			return nil
		}
		for _, object := range objects {
			if err := s.contents.DeleteContent(ctx, object.UID); err != nil {
				if errors.Is(err, domain.ErrConflict) {
					continue
				}
				return err
			}
			if err := s.files.Remove(ctx, object.StorageKey); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
}

func chunksComplete(upload domain.Upload) bool {
	if upload.TotalSizeBytes == 0 {
		return len(upload.Chunks) == 0
	}
	var expected int64
	for _, chunk := range upload.Chunks {
		if chunk.StartOffset != expected {
			return false
		}
		expected = chunk.EndOffset
	}
	return expected == upload.TotalSizeBytes
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
