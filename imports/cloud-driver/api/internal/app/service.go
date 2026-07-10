package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	defaultShareLifetime = 7 * 24 * time.Hour
	maximumShareLifetime = 30 * 24 * time.Hour
	trashRetention       = 30 * 24 * time.Hour
)

// BlobStore is the file-content boundary owned by the application service.
type BlobStore interface {
	PrepareUpload(context.Context, string) error
	WriteChunk(context.Context, string, int64, []byte) error
	FinalizeUpload(context.Context, string, string) error
	RemoveUpload(context.Context, string) error
	OpenBlob(context.Context, string) (*os.File, error)
	RemoveBlob(context.Context, string) error
	FreeBytes() (int64, error)
}

// Service coordinates metadata, content storage, and lifecycle policies.
type Service struct {
	store             domain.Store
	blobs             BlobStore
	clock             domain.Clock
	chunkSizeBytes    int64
	reservedFreeBytes int64
	uploadTTL         time.Duration
	shareAppOrigin    string
	locks             *keyedLocks
	capacityMu        sync.Mutex
}

// NewService constructs a drive application service.
func NewService(store domain.Store, blobs BlobStore, clock domain.Clock, chunkSizeBytes, reservedFreeBytes int64, uploadTTL time.Duration, shareAppOrigin string) *Service {
	return &Service{
		store:             store,
		blobs:             blobs,
		clock:             clock,
		chunkSizeBytes:    chunkSizeBytes,
		reservedFreeBytes: reservedFreeBytes,
		uploadTTL:         uploadTTL,
		shareAppOrigin:    strings.TrimRight(shareAppOrigin, "/"),
		locks:             newKeyedLocks(),
	}
}

// Ping checks metadata storage readiness.
func (s *Service) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}

// GetItem returns one visible drive item.
func (s *Service) GetItem(ctx context.Context, itemUID string) (domain.Item, error) {
	return s.store.GetItem(ctx, itemUID, false)
}

// ListItems returns visible direct children.
func (s *Service) ListItems(ctx context.Context, parentUID *string, includeTrashed bool, pageSize int, pageToken string) (domain.Page, error) {
	return s.store.ListItems(ctx, emptyToNil(parentUID), includeTrashed, pageSize, pageToken)
}

// ListTrashedItems returns trash roots.
func (s *Service) ListTrashedItems(ctx context.Context, pageSize int, pageToken string) (domain.Page, error) {
	return s.store.ListTrashedItems(ctx, pageSize, pageToken)
}

// SearchItems returns visible basename matches.
func (s *Service) SearchItems(ctx context.Context, query string, pageSize int, pageToken string) (domain.Page, error) {
	if strings.TrimSpace(query) == "" {
		return domain.Page{}, fmt.Errorf("%w: query is required", domain.ErrInvalidArgument)
	}
	return s.store.SearchItems(ctx, query, pageSize, pageToken)
}

// CreateDirectory creates a directory after validating its basename.
func (s *Service) CreateDirectory(ctx context.Context, parentUID *string, name string) (domain.Item, error) {
	if err := validateName(name); err != nil {
		return domain.Item{}, err
	}
	return s.store.CreateDirectory(ctx, emptyToNil(parentUID), name)
}

// RenameItem validates and updates an item's basename.
func (s *Service) RenameItem(ctx context.Context, itemUID, name string) (domain.Item, error) {
	if err := validateName(name); err != nil {
		return domain.Item{}, err
	}
	return s.store.RenameItem(ctx, itemUID, name)
}

// MoveItem moves an item to a directory or the root.
func (s *Service) MoveItem(ctx context.Context, itemUID string, parentUID *string) (domain.Item, error) {
	return s.store.MoveItem(ctx, itemUID, emptyToNil(parentUID))
}

// TrashItem moves an item hierarchy to trash.
func (s *Service) TrashItem(ctx context.Context, itemUID string) (domain.Item, error) {
	return s.store.TrashItem(ctx, itemUID)
}

// RestoreItem restores an item hierarchy from trash.
func (s *Service) RestoreItem(ctx context.Context, itemUID string) (domain.Item, error) {
	return s.store.RestoreItem(ctx, itemUID)
}

// StartUpload reserves capacity and creates a resumable session.
func (s *Service) StartUpload(ctx context.Context, parentUID *string, fileName, contentType string, totalSizeBytes int64) (domain.Upload, error) {
	if err := validateName(fileName); err != nil {
		return domain.Upload{}, err
	}
	if totalSizeBytes < 0 {
		return domain.Upload{}, fmt.Errorf("%w: total size must not be negative", domain.ErrInvalidArgument)
	}
	s.capacityMu.Lock()
	defer s.capacityMu.Unlock()

	freeBytes, err := s.blobs.FreeBytes()
	if err != nil {
		return domain.Upload{}, fmt.Errorf("read free storage: %w", err)
	}
	reservedBytes, err := s.store.ReservedUploadBytes(ctx)
	if err != nil {
		return domain.Upload{}, err
	}
	available := freeBytes - s.reservedFreeBytes - reservedBytes
	if available < 0 || totalSizeBytes > available {
		return domain.Upload{}, fmt.Errorf("%w: insufficient free storage", domain.ErrResourceExhausted)
	}
	now := s.clock.Now().UTC()
	upload := domain.Upload{
		UID:            uuid.NewString(),
		ItemUID:        uuid.NewString(),
		ParentUID:      emptyToNil(parentUID),
		FileName:       fileName,
		ContentType:    contentType,
		TotalSizeBytes: totalSizeBytes,
		ChunkSizeBytes: s.chunkSizeBytes,
		Status:         domain.UploadStatusActive,
		CreateTime:     now,
		ExpireTime:     now.Add(s.uploadTTL),
	}
	if err := s.blobs.PrepareUpload(ctx, upload.UID); err != nil {
		return domain.Upload{}, err
	}
	created, err := s.store.CreateUpload(ctx, upload)
	if err != nil {
		_ = s.blobs.RemoveUpload(ctx, upload.UID)
		return domain.Upload{}, err
	}
	return created, nil
}

// GetUpload returns the current upload session state.
func (s *Service) GetUpload(ctx context.Context, uploadUID string) (domain.Upload, error) {
	return s.store.GetUpload(ctx, uploadUID)
}

// WriteUploadChunk serializes a file write and checksum acknowledgement per upload.
func (s *Service) WriteUploadChunk(ctx context.Context, uploadUID string, startOffset, totalSizeBytes int64, data []byte) (domain.Upload, error) {
	unlock := s.locks.Lock(uploadUID)
	defer unlock()

	upload, err := s.store.GetUpload(ctx, uploadUID)
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
			return domain.Upload{}, fmt.Errorf("%w: chunk offset already has different content", domain.ErrConflict)
		}
		if existing.StartOffset < endOffset && existing.EndOffset > startOffset {
			return domain.Upload{}, fmt.Errorf("%w: chunk overlaps an acknowledged range", domain.ErrConflict)
		}
	}
	if err := s.blobs.WriteChunk(ctx, uploadUID, startOffset, data); err != nil {
		return domain.Upload{}, err
	}
	return s.store.RecordUploadChunk(ctx, uploadUID, domain.UploadChunk{
		StartOffset: startOffset,
		EndOffset:   endOffset,
		Checksum:    checksum[:],
	})
}

// CompleteUpload publishes a fully acknowledged upload file.
func (s *Service) CompleteUpload(ctx context.Context, uploadUID string) (domain.Item, error) {
	unlock := s.locks.Lock(uploadUID)
	defer unlock()

	upload, err := s.store.GetUpload(ctx, uploadUID)
	if err != nil {
		return domain.Item{}, err
	}
	if upload.Status == domain.UploadStatusCompleted {
		return s.store.CompleteUpload(ctx, uploadUID)
	}
	if upload.Status != domain.UploadStatusActive || !chunksComplete(upload) {
		return domain.Item{}, fmt.Errorf("%w: upload is incomplete", domain.ErrConflict)
	}
	if err := s.blobs.FinalizeUpload(ctx, upload.UID, upload.ItemUID); err != nil {
		return domain.Item{}, err
	}
	return s.store.CompleteUpload(ctx, uploadUID)
}

// CreateShare creates an expiring read-only bearer link.
func (s *Service) CreateShare(ctx context.Context, targetUID string, requestedExpiry *time.Time) (domain.ShareLink, string, error) {
	now := s.clock.Now().UTC()
	expiry := now.Add(defaultShareLifetime)
	if requestedExpiry != nil {
		expiry = requestedExpiry.UTC()
	}
	if !expiry.After(now) || expiry.After(now.Add(maximumShareLifetime)) {
		return domain.ShareLink{}, "", fmt.Errorf("%w: share expiry must be within 30 days", domain.ErrInvalidArgument)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return domain.ShareLink{}, "", fmt.Errorf("generate share token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	share := domain.ShareLink{UID: uuid.NewString(), TargetUID: targetUID, CreateTime: now, ExpireTime: expiry}
	created, err := s.store.CreateShare(ctx, share, tokenHash[:])
	if err != nil {
		return domain.ShareLink{}, "", err
	}
	return created, s.shareAppOrigin + "/s/" + token, nil
}

// ListShareLinks lists owner-visible metadata for a target item.
func (s *Service) ListShareLinks(ctx context.Context, targetUID string, pageSize int, pageToken string) (domain.SharePage, error) {
	return s.store.ListShareLinks(ctx, targetUID, pageSize, pageToken)
}

// RevokeShare disables a share link.
func (s *Service) RevokeShare(ctx context.Context, shareUID string) (domain.ShareLink, error) {
	return s.store.RevokeShare(ctx, shareUID)
}

// ResolveShare resolves an active public token.
func (s *Service) ResolveShare(ctx context.Context, token string) (domain.ShareLink, domain.Item, error) {
	if token == "" {
		return domain.ShareLink{}, domain.Item{}, fmt.Errorf("%w: share token is required", domain.ErrInvalidArgument)
	}
	hash := sha256.Sum256([]byte(token))
	share, err := s.store.GetShareByTokenHash(ctx, hash[:])
	if err != nil {
		return domain.ShareLink{}, domain.Item{}, err
	}
	item, err := s.store.GetItem(ctx, share.TargetUID, false)
	if err != nil {
		return domain.ShareLink{}, domain.Item{}, err
	}
	return share, item, nil
}

// ListSharedItems resolves a token and lists one visible shared directory.
func (s *Service) ListSharedItems(ctx context.Context, token string, parentUID *string, pageSize int, pageToken string) (domain.Page, error) {
	share, _, err := s.ResolveShare(ctx, token)
	if err != nil {
		return domain.Page{}, err
	}
	return s.store.ListSharedItems(ctx, share.UID, emptyToNil(parentUID), pageSize, pageToken)
}

// SharedItem resolves a token and verifies an item is inside its scope.
func (s *Service) SharedItem(ctx context.Context, token, itemUID string) (domain.ShareLink, domain.Item, error) {
	share, _, err := s.ResolveShare(ctx, token)
	if err != nil {
		return domain.ShareLink{}, domain.Item{}, err
	}
	item, err := s.store.CanReadSharedItem(ctx, share.UID, itemUID)
	if err != nil {
		return domain.ShareLink{}, domain.Item{}, err
	}
	return share, item, nil
}

// OpenFile opens a file's blob after metadata authorization has completed.
func (s *Service) OpenFile(ctx context.Context, item domain.Item) (*os.File, error) {
	if item.Kind != domain.ItemKindFile {
		return nil, fmt.Errorf("%w: item is not a file", domain.ErrConflict)
	}
	file, err := s.blobs.OpenBlob(ctx, item.StorageKey)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: file content", domain.ErrNotFound)
	}
	return file, err
}

// PurgeExpired removes expired temporary uploads and trash past retention.
func (s *Service) PurgeExpired(ctx context.Context) error {
	now := s.clock.Now().UTC()
	uploads, err := s.store.ListExpiredUploads(ctx, now)
	if err != nil {
		return err
	}
	for _, upload := range uploads {
		if err := s.blobs.RemoveUpload(ctx, upload.UID); err != nil {
			return err
		}
		if err := s.store.DeleteUpload(ctx, upload.UID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return err
		}
		if err := s.blobs.RemoveBlob(ctx, upload.ItemUID); err != nil {
			return err
		}
	}
	cutoff := now.Add(-trashRetention)
	items, err := s.store.PurgeTrashedItems(ctx, cutoff)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Kind == domain.ItemKindFile {
			if err := s.blobs.RemoveBlob(ctx, item.StorageKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateName(name string) error {
	if name == "" || name == "." || name == ".." || len(name) > 255 || !utf8.ValidString(name) {
		return fmt.Errorf("%w: invalid item name", domain.ErrInvalidArgument)
	}
	for _, runeValue := range name {
		if runeValue == '/' || runeValue == '\\' || runeValue == 0 || unicode.IsControl(runeValue) {
			return fmt.Errorf("%w: invalid item name", domain.ErrInvalidArgument)
		}
	}
	return nil
}

func emptyToNil(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	copied := *value
	return &copied
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

type keyedLocks struct {
	mu    sync.Mutex
	locks map[string]*keyedLock
}

type keyedLock struct {
	mu   sync.Mutex
	refs int
}

func newKeyedLocks() *keyedLocks {
	return &keyedLocks{locks: make(map[string]*keyedLock)}
}

func (locks *keyedLocks) Lock(key string) func() {
	locks.mu.Lock()
	lock := locks.locks[key]
	if lock == nil {
		lock = &keyedLock{}
		locks.locks[key] = lock
	}
	lock.refs++
	locks.mu.Unlock()
	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		locks.mu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(locks.locks, key)
		}
		locks.mu.Unlock()
	}
}
