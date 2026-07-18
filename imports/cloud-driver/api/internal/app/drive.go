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
	"time"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	defaultShareLifetime = 7 * 24 * time.Hour
	maximumShareLifetime = 30 * 24 * time.Hour
)

// DriveService manages the generic drive as one peer application.
type DriveService struct {
	store          domain.DriveStore
	content        *ContentService
	files          ContentReader
	clock          domain.Clock
	shareAppOrigin string
}

// NewDriveService constructs the drive application.
func NewDriveService(store domain.DriveStore, content *ContentService, files ContentReader, clock domain.Clock, shareAppOrigin string) *DriveService {
	return &DriveService{store: store, content: content, files: files, clock: clock, shareAppOrigin: strings.TrimRight(shareAppOrigin, "/")}
}

func (s *DriveService) GetItem(ctx context.Context, uid string) (domain.Item, error) {
	return s.store.GetItem(ctx, uid, false)
}

func (s *DriveService) ListItems(ctx context.Context, parentUID *string, includeTrashed bool, pageSize int, pageToken string) (domain.Page, error) {
	return s.store.ListItems(ctx, domain.DriveListQuery{
		ParentUID: emptyToNil(parentUID), IncludeTrashed: includeTrashed,
		PageSize: pageSize, PageToken: pageToken,
	})
}

func (s *DriveService) ListTrashedItems(ctx context.Context, pageSize int, pageToken string) (domain.Page, error) {
	return s.store.ListTrashedItems(ctx, pageSize, pageToken)
}

func (s *DriveService) SearchItems(ctx context.Context, query string, pageSize int, pageToken string) (domain.Page, error) {
	if strings.TrimSpace(query) == "" {
		return domain.Page{}, fmt.Errorf("%w: query is required", domain.ErrInvalidArgument)
	}
	return s.store.SearchItems(ctx, query, pageSize, pageToken)
}

func (s *DriveService) CreateDirectory(ctx context.Context, parentUID *string, name string) (domain.Item, error) {
	if err := validateName(name); err != nil {
		return domain.Item{}, err
	}
	return s.store.CreateDirectory(ctx, emptyToNil(parentUID), name)
}

func (s *DriveService) RenameItem(ctx context.Context, uid, name string) (domain.Item, error) {
	if err := validateName(name); err != nil {
		return domain.Item{}, err
	}
	return s.store.RenameItem(ctx, uid, name)
}

func (s *DriveService) MoveItem(ctx context.Context, uid string, parentUID *string) (domain.Item, error) {
	return s.store.MoveItem(ctx, uid, emptyToNil(parentUID))
}

func (s *DriveService) TrashItem(ctx context.Context, uid string) (domain.Item, error) {
	return s.store.TrashItem(ctx, uid)
}

func (s *DriveService) RestoreItem(ctx context.Context, uid string) (domain.Item, error) {
	return s.store.RestoreItem(ctx, uid)
}

func (s *DriveService) PurgeItem(ctx context.Context, uid string) error {
	if err := s.store.PurgeTrashedItem(ctx, uid); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *DriveService) EmptyTrash(ctx context.Context) error {
	if err := s.store.EmptyTrash(ctx); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

// ImportFile claims a neutral upload as a drive file.
func (s *DriveService) ImportFile(ctx context.Context, uploadUID string, parentUID *string, name string) (domain.Item, error) {
	if err := validateName(name); err != nil {
		return domain.Item{}, err
	}
	upload, sealed, unlock, err := s.content.SealForClaim(ctx, uploadUID)
	if err != nil {
		return domain.Item{}, err
	}
	defer unlock()
	item, err := s.store.ImportDriveFile(ctx, uploadUID, emptyToNil(parentUID), name, sealed)
	if err != nil {
		return domain.Item{}, err
	}
	if upload.Status != domain.UploadStatusClaimed {
		_ = s.content.FinishClaim(ctx, uploadUID)
	}
	return item, nil
}

// OpenItem opens an authorized drive file.
func (s *DriveService) OpenItem(ctx context.Context, item domain.Item) (*os.File, error) {
	if item.Kind != domain.ItemKindFile {
		return nil, fmt.Errorf("%w: item is not a file", domain.ErrConflict)
	}
	file, err := s.files.Open(ctx, item.StorageKey)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: file content", domain.ErrNotFound)
	}
	return file, err
}

func (s *DriveService) CreateShare(ctx context.Context, targetUID string, requestedExpiry *time.Time) (domain.ShareLink, string, error) {
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

func (s *DriveService) ListShareLinks(ctx context.Context, targetUID string, pageSize int, pageToken string) (domain.SharePage, error) {
	return s.store.ListShareLinks(ctx, targetUID, pageSize, pageToken)
}

func (s *DriveService) RevokeShare(ctx context.Context, uid string) (domain.ShareLink, error) {
	return s.store.RevokeShare(ctx, uid)
}

func (s *DriveService) ResolveShare(ctx context.Context, token string) (domain.ShareLink, domain.Item, error) {
	if token == "" {
		return domain.ShareLink{}, domain.Item{}, fmt.Errorf("%w: share token is required", domain.ErrInvalidArgument)
	}
	hash := sha256.Sum256([]byte(token))
	share, err := s.store.GetShareByTokenHash(ctx, hash[:])
	if err != nil {
		return domain.ShareLink{}, domain.Item{}, err
	}
	item, err := s.store.GetItem(ctx, share.TargetUID, false)
	return share, item, err
}

func (s *DriveService) ListSharedItems(ctx context.Context, token string, parentUID *string, pageSize int, pageToken string) (domain.Page, error) {
	share, _, err := s.ResolveShare(ctx, token)
	if err != nil {
		return domain.Page{}, err
	}
	return s.store.ListSharedItems(ctx, share.UID, emptyToNil(parentUID), pageSize, pageToken)
}

func (s *DriveService) SharedItem(ctx context.Context, token, itemUID string) (domain.ShareLink, domain.Item, error) {
	share, _, err := s.ResolveShare(ctx, token)
	if err != nil {
		return domain.ShareLink{}, domain.Item{}, err
	}
	item, err := s.store.CanReadSharedItem(ctx, share.UID, itemUID)
	return share, item, err
}
