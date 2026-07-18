package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

// WallpaperService manages private publication and the public read-only catalog.
type WallpaperService struct {
	store            domain.WallpaperStore
	contents         domain.ContentStore
	files            ContentReader
	publicSiteOrigin string
}

// NewWallpaperService constructs wallpaper publication and public access.
func NewWallpaperService(store domain.WallpaperStore, contents domain.ContentStore, files ContentReader, publicSiteOrigin string) *WallpaperService {
	return &WallpaperService{store: store, contents: contents, files: files, publicSiteOrigin: strings.TrimRight(publicSiteOrigin, "/")}
}

func (s *WallpaperService) Get(ctx context.Context, uid string) (domain.Wallpaper, error) {
	return s.store.GetWallpaper(ctx, uid)
}

func (s *WallpaperService) List(ctx context.Context, query, tag, orientation string, pageSize int, pageToken string) (domain.WallpaperPage, error) {
	return s.store.ListWallpapers(ctx, domain.WallpaperListQuery{
		Query: strings.TrimSpace(query), Tag: strings.TrimSpace(tag), Orientation: orientation,
		PageSize: pageSize, PageToken: pageToken,
	})
}

func (s *WallpaperService) Publish(ctx context.Context, imageUID, title string, tags []string) (domain.Wallpaper, error) {
	if err := validateDisplayName(title); err != nil {
		return domain.Wallpaper{}, err
	}
	return s.store.PublishWallpaper(ctx, domain.Wallpaper{
		UID: uuid.NewString(), ImageUID: imageUID, Title: strings.TrimSpace(title), Tags: normalizedStrings(tags, 20, 64),
	})
}

func (s *WallpaperService) Update(ctx context.Context, uid, title string, tags []string) (domain.Wallpaper, error) {
	if err := validateDisplayName(title); err != nil {
		return domain.Wallpaper{}, err
	}
	return s.store.UpdateWallpaper(ctx, domain.Wallpaper{UID: uid, Title: strings.TrimSpace(title), Tags: normalizedStrings(tags, 20, 64)})
}

func (s *WallpaperService) Unpublish(ctx context.Context, uid string) error {
	return s.store.UnpublishWallpaper(ctx, uid)
}

func (s *WallpaperService) OriginalURL(uid string) string {
	return s.publicSiteOrigin + "/v1/wallpapers/" + uid + "/original"
}

func (s *WallpaperService) VariantURL(uid string, width int) string {
	return s.publicSiteOrigin + "/v1/wallpapers/" + uid + "/" + strconv.Itoa(width)
}

func (s *WallpaperService) OpenOriginal(ctx context.Context, uid string) (*os.File, domain.Wallpaper, error) {
	wallpaper, err := s.store.GetWallpaper(ctx, uid)
	if err != nil {
		return nil, domain.Wallpaper{}, err
	}
	file, err := s.files.Open(ctx, wallpaper.StorageKey)
	return file, wallpaper, err
}

func (s *WallpaperService) OpenVariant(ctx context.Context, uid string, width int) (*os.File, domain.Wallpaper, domain.Artifact, error) {
	wallpaper, err := s.store.GetWallpaper(ctx, uid)
	if err != nil {
		return nil, domain.Wallpaper{}, domain.Artifact{}, err
	}
	for _, variant := range wallpaper.Variants {
		if variant.Width == width {
			file, err := s.files.Open(ctx, variant.StorageKey)
			return file, wallpaper, variant, err
		}
	}
	return nil, domain.Wallpaper{}, domain.Artifact{}, fmt.Errorf("%w: wallpaper variant", domain.ErrNotFound)
}
