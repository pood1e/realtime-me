package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

// ImageService manages private images and stable anonymous raw links.
type ImageService struct {
	store           domain.ImageStore
	contents        domain.ContentStore
	content         *ContentService
	files           ContentFiles
	publicAPIOrigin string
}

// NewImageService constructs the private image application.
func NewImageService(store domain.ImageStore, contents domain.ContentStore, content *ContentService, files ContentFiles, publicAPIOrigin string) *ImageService {
	return &ImageService{store: store, contents: contents, content: content, files: files, publicAPIOrigin: strings.TrimRight(publicAPIOrigin, "/")}
}

func (s *ImageService) Get(ctx context.Context, uid string) (domain.Image, error) {
	return s.store.GetImage(ctx, uid, false)
}

func (s *ImageService) List(ctx context.Context, query string, albumUID *string, trashed bool, pageSize int, pageToken string) (domain.ImagePage, error) {
	return s.store.ListImages(ctx, strings.TrimSpace(query), emptyToNil(albumUID), trashed, pageSize, pageToken)
}

func (s *ImageService) Import(ctx context.Context, uploadUID string, albumUID *string) (domain.Image, error) {
	upload, sealed, unlock, err := s.content.SealForClaim(ctx, uploadUID)
	if err != nil {
		return domain.Image{}, err
	}
	defer unlock()
	image, err := s.store.ImportImage(ctx, uploadUID, emptyToNil(albumUID), sealed)
	if err != nil {
		return domain.Image{}, err
	}
	if upload.Status != domain.UploadStatusClaimed {
		s.content.FinishClaim(ctx, uploadUID)
	}
	return image, nil
}

func (s *ImageService) Update(ctx context.Context, image domain.Image) (domain.Image, error) {
	if err := validateDisplayName(image.DisplayName); err != nil {
		return domain.Image{}, err
	}
	image.DisplayName = strings.TrimSpace(image.DisplayName)
	image.AlbumUID = emptyToNil(image.AlbumUID)
	return s.store.UpdateImage(ctx, image)
}

func (s *ImageService) Trash(ctx context.Context, uid string) (domain.Image, error) {
	return s.store.TrashImage(ctx, uid)
}

func (s *ImageService) Restore(ctx context.Context, uid string) (domain.Image, error) {
	return s.store.RestoreImage(ctx, uid)
}

func (s *ImageService) Purge(ctx context.Context, uid string) error {
	if err := s.store.PurgeImage(ctx, uid); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *ImageService) EmptyTrash(ctx context.Context) error {
	if err := s.store.EmptyImageTrash(ctx); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *ImageService) RetryProcessing(ctx context.Context, uid string) (domain.Image, error) {
	return s.store.QueueImageProcessing(ctx, uid)
}

func (s *ImageService) ListAlbums(ctx context.Context) ([]domain.ImageAlbum, error) {
	return s.store.ListImageAlbums(ctx)
}

func (s *ImageService) CreateAlbum(ctx context.Context, displayName string) (domain.ImageAlbum, error) {
	if err := validateDisplayName(displayName); err != nil {
		return domain.ImageAlbum{}, err
	}
	return s.store.CreateImageAlbum(ctx, domain.ImageAlbum{UID: uuid.NewString(), DisplayName: strings.TrimSpace(displayName)})
}

func (s *ImageService) DeleteAlbum(ctx context.Context, uid string) error {
	return s.store.DeleteImageAlbum(ctx, uid)
}

func (s *ImageService) ListLinks(ctx context.Context, imageUID string) ([]domain.ImageLink, error) {
	return s.store.ListImageLinks(ctx, imageUID)
}

func (s *ImageService) CreateLink(ctx context.Context, imageUID string) (domain.ImageLink, error) {
	random := make([]byte, 18)
	if _, err := rand.Read(random); err != nil {
		return domain.ImageLink{}, fmt.Errorf("generate image link: %w", err)
	}
	return s.store.CreateImageLink(ctx, domain.ImageLink{UID: base64.RawURLEncoding.EncodeToString(random), ImageUID: imageUID})
}

func (s *ImageService) RevokeLink(ctx context.Context, uid string) (domain.ImageLink, error) {
	return s.store.RevokeImageLink(ctx, uid)
}

func (s *ImageService) PublicLinkURL(uid string) string {
	return s.publicAPIOrigin + "/i/" + uid
}

func (s *ImageService) OpenOriginal(ctx context.Context, uid string) (*os.File, domain.Image, error) {
	image, err := s.store.GetImage(ctx, uid, false)
	if err != nil {
		return nil, domain.Image{}, err
	}
	content, err := s.contents.GetContent(ctx, image.ContentUID)
	if err != nil {
		return nil, domain.Image{}, err
	}
	file, err := s.files.Open(ctx, content.StorageKey)
	return file, image, err
}

func (s *ImageService) OpenPreview(ctx context.Context, uid string) (*os.File, domain.Image, error) {
	image, err := s.store.GetImage(ctx, uid, false)
	if err != nil {
		return nil, domain.Image{}, err
	}
	if image.PreviewStorageKey == "" {
		return nil, domain.Image{}, domainNotFound("image preview")
	}
	file, err := s.files.Open(ctx, image.PreviewStorageKey)
	return file, image, err
}

// OpenPublicLink resolves a link and selects a safe embeddable representation.
func (s *ImageService) OpenPublicLink(ctx context.Context, uid string) (*os.File, domain.Image, string, error) {
	image, err := s.store.GetImageByLink(ctx, uid)
	if err != nil {
		return nil, domain.Image{}, "", err
	}
	storageKey := ""
	contentType := image.ContentType
	if image.ContentType == "image/svg+xml" {
		storageKey = image.PreviewStorageKey
		contentType = "image/webp"
	} else {
		content, err := s.contents.GetContent(ctx, image.ContentUID)
		if err != nil {
			return nil, domain.Image{}, "", err
		}
		storageKey = content.StorageKey
	}
	if storageKey == "" {
		return nil, domain.Image{}, "", domainNotFound("public image representation")
	}
	file, err := s.files.Open(ctx, storageKey)
	return file, image, contentType, err
}
