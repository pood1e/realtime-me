package postgres

import (
	"context"
	"fmt"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *Store) GetImageForProcessing(ctx context.Context, uid string) (domain.Image, domain.ContentObject, error) {
	image, err := s.GetImage(ctx, uid, true)
	if err != nil {
		return domain.Image{}, domain.ContentObject{}, err
	}
	content, err := s.GetContent(ctx, image.ContentUID)
	return image, content, err
}

// CompleteImageProcessing stores dimensions and a safe preview.
func (s *Store) CompleteImageProcessing(ctx context.Context, uid string, width, height int, preview *domain.Artifact) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin image processing completion: %w", err)
	}
	defer tx.Rollback(ctx)
	if preview != nil {
		if err := upsertArtifact(ctx, tx, *preview); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE images SET width = $2, height = $3,
		processing_status = 'ready', update_time = now() WHERE uid = $1`, uid, width, height); err != nil {
		return fmt.Errorf("complete image processing: %w", err)
	}
	return tx.Commit(ctx)
}

// GetWallpaperForProcessing returns the source image for variant generation.
func (s *Store) GetWallpaperForProcessing(ctx context.Context, uid string) (domain.Wallpaper, domain.ContentObject, error) {
	wallpaper, err := s.GetWallpaper(ctx, uid)
	if err != nil {
		return domain.Wallpaper{}, domain.ContentObject{}, err
	}
	var contentUID string
	if err := s.pool.QueryRow(ctx, "SELECT content_uid FROM images WHERE uid = $1", wallpaper.ImageUID).Scan(&contentUID); err != nil {
		return domain.Wallpaper{}, domain.ContentObject{}, fmt.Errorf("get wallpaper content id: %w", err)
	}
	content, err := s.GetContent(ctx, contentUID)
	return wallpaper, content, err
}

// CompleteWallpaperProcessing stores dominant color and responsive variants.
func (s *Store) CompleteWallpaperProcessing(ctx context.Context, uid, dominantColor string, variants []domain.Artifact) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin wallpaper processing completion: %w", err)
	}
	defer tx.Rollback(ctx)
	for _, variant := range variants {
		if err := upsertArtifact(ctx, tx, variant); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE wallpapers SET dominant_color = $2, update_time = now()
		WHERE uid = $1`, uid, dominantColor); err != nil {
		return fmt.Errorf("complete wallpaper processing: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Store) listWallpaperArtifacts(ctx context.Context, imageUID string) ([]domain.Artifact, error) {
	rows, err := s.pool.Query(ctx, `SELECT artifact.uid, artifact.content_uid, artifact.kind, artifact.variant,
		artifact.content_type, artifact.storage_key, artifact.width, artifact.height, artifact.create_time
		FROM content_artifacts artifact JOIN images image ON image.content_uid = artifact.content_uid
		WHERE image.uid = $1 AND artifact.kind = 'wallpaper' ORDER BY artifact.width`, imageUID)
	if err != nil {
		return nil, fmt.Errorf("list wallpaper variants: %w", err)
	}
	defer rows.Close()
	var artifacts []domain.Artifact
	for rows.Next() {
		var artifact domain.Artifact
		if err := rows.Scan(&artifact.UID, &artifact.ContentUID, &artifact.Kind, &artifact.Variant,
			&artifact.ContentType, &artifact.StorageKey, &artifact.Width, &artifact.Height, &artifact.CreateTime); err != nil {
			return nil, fmt.Errorf("scan wallpaper variant: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, rows.Err()
}
