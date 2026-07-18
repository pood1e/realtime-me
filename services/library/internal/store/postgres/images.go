package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const (
	imageColumns = `image.uid, image.content_uid, image.album_uid, image.display_name, image.original_file_name,
		content.content_type, content.size_bytes, image.width, image.height, COALESCE(preview.storage_key, ''),
		image.processing_status, image.create_time, image.update_time, image.delete_time`
	imageFrom = `images image JOIN content_objects content ON content.uid = image.content_uid
		LEFT JOIN content_artifacts preview ON preview.content_uid = image.content_uid
		AND preview.kind = 'image_preview' AND preview.variant = 'default'`
)

// GetImage returns one private image.
func (s *Store) GetImage(ctx context.Context, uid string, includeTrashed bool) (domain.Image, error) {
	query := "SELECT " + imageColumns + " FROM " + imageFrom + " WHERE image.uid = $1"
	if !includeTrashed {
		query += " AND image.delete_time IS NULL"
	}
	image, err := scanImage(s.pool.QueryRow(ctx, query, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Image{}, fmt.Errorf("%w: image", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Image{}, fmt.Errorf("get image: %w", err)
	}
	return image, nil
}

// ListImages lists private image assets.
func (s *Store) ListImages(ctx context.Context, filter domain.ImageListQuery) (domain.ImagePage, error) {
	cursor, err := decodeCursor(filter.PageToken)
	if err != nil {
		return domain.ImagePage{}, err
	}
	pageSize := normalizePageSize(filter.PageSize)
	query := "SELECT " + imageColumns + " FROM " + imageFrom
	arguments := []any{}
	conditions := []string{"image.delete_time IS " + map[bool]string{true: "NOT NULL", false: "NULL"}[filter.Trashed]}
	if filter.Query != "" {
		arguments = append(arguments, filter.Query)
		conditions = append(conditions, fmt.Sprintf("image.display_name ILIKE '%%' || $%d || '%%'", len(arguments)))
	}
	if filter.AlbumUID != nil {
		arguments = append(arguments, *filter.AlbumUID)
		conditions = append(conditions, fmt.Sprintf("image.album_uid = $%d", len(arguments)))
	}
	if cursor != nil {
		arguments = append(arguments, cursor.name, cursor.uid)
		conditions = append(conditions, fmt.Sprintf("(image.display_name, image.uid) > ($%d, $%d)", len(arguments)-1, len(arguments)))
	}
	arguments = append(arguments, pageSize+1)
	query += " WHERE " + strings.Join(conditions, " AND ") + fmt.Sprintf(" ORDER BY image.display_name, image.uid LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.ImagePage{}, fmt.Errorf("list images: %w", err)
	}
	defer rows.Close()
	page := domain.ImagePage{}
	for rows.Next() {
		image, err := scanImage(rows)
		if err != nil {
			return domain.ImagePage{}, fmt.Errorf("scan image: %w", err)
		}
		page.Images = append(page.Images, image)
	}
	if len(page.Images) > pageSize {
		last := page.Images[pageSize-1]
		page.Images = page.Images[:pageSize]
		page.NextPageToken = encodeCursor(last.DisplayName, last.UID)
	}
	return page, rows.Err()
}

// ImportImage claims a supported image upload.
func (s *Store) ImportImage(ctx context.Context, uploadUID string, albumUID *string, sealed domain.SealedContent) (domain.Image, error) {
	if !supportedImageType(sealed.ContentType) {
		return domain.Image{}, fmt.Errorf("%w: unsupported image type", domain.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Image{}, fmt.Errorf("begin image import: %w", err)
	}
	defer tx.Rollback(ctx)
	upload, err := lockCompleteUpload(ctx, tx, uploadUID)
	if err != nil {
		return domain.Image{}, err
	}
	if upload.Status == domain.UploadStatusClaimed {
		if err := tx.Commit(ctx); err != nil {
			return domain.Image{}, fmt.Errorf("commit repeated image import: %w", err)
		}
		return s.GetImage(ctx, upload.ClaimedUID, false)
	}
	if err := validateImageAlbum(ctx, tx, albumUID); err != nil {
		return domain.Image{}, err
	}
	content, err := contentForUpload(ctx, tx, upload, sealed)
	if err != nil {
		return domain.Image{}, err
	}
	var existingUID string
	err = tx.QueryRow(ctx, "SELECT uid FROM images WHERE content_uid = $1", content.UID).Scan(&existingUID)
	if err == nil {
		if albumUID != nil {
			if _, err := tx.Exec(ctx, "UPDATE images SET album_uid = $2, update_time = now() WHERE uid = $1", existingUID, *albumUID); err != nil {
				return domain.Image{}, fmt.Errorf("move deduplicated image: %w", err)
			}
		}
		if err := markUploadClaimed(ctx, tx, uploadUID, existingUID); err != nil {
			return domain.Image{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return domain.Image{}, fmt.Errorf("commit deduplicated image import: %w", err)
		}
		return s.GetImage(ctx, existingUID, true)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Image{}, fmt.Errorf("find imported image: %w", err)
	}
	imageUID := uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO images
		(uid, content_uid, album_uid, display_name, original_file_name, processing_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')`, imageUID, content.UID, nullableString(albumUID),
		displayName(upload.FileName), upload.FileName); err != nil {
		return domain.Image{}, fmt.Errorf("create image: %w", err)
	}
	if err := enqueueJob(ctx, tx, domain.ProcessingJobImage, imageUID); err != nil {
		return domain.Image{}, err
	}
	if err := markUploadClaimed(ctx, tx, uploadUID, imageUID); err != nil {
		return domain.Image{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Image{}, fmt.Errorf("commit image import: %w", err)
	}
	return s.GetImage(ctx, imageUID, false)
}

// UpdateImage changes display metadata and collection placement.
func (s *Store) UpdateImage(ctx context.Context, image domain.Image) (domain.Image, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Image{}, fmt.Errorf("begin image update: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := validateImageAlbum(ctx, tx, image.AlbumUID); err != nil {
		return domain.Image{}, err
	}
	command, err := tx.Exec(ctx, `UPDATE images SET display_name = $2, album_uid = $3, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, image.UID, image.DisplayName, nullableString(image.AlbumUID))
	if err != nil {
		return domain.Image{}, fmt.Errorf("update image: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Image{}, fmt.Errorf("%w: image", domain.ErrNotFound)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Image{}, fmt.Errorf("commit image update: %w", err)
	}
	return s.GetImage(ctx, image.UID, false)
}

// TrashImage revokes public links and unpublishes the wallpaper before soft deletion.
func (s *Store) TrashImage(ctx context.Context, uid string) (domain.Image, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Image{}, fmt.Errorf("begin image trash: %w", err)
	}
	defer tx.Rollback(ctx)
	command, err := tx.Exec(ctx, `UPDATE images SET delete_time = now(), update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid)
	if err != nil {
		return domain.Image{}, fmt.Errorf("trash image: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Image{}, fmt.Errorf("%w: image", domain.ErrNotFound)
	}
	if _, err := tx.Exec(ctx, "UPDATE image_links SET revoke_time = now() WHERE image_uid = $1 AND revoke_time IS NULL", uid); err != nil {
		return domain.Image{}, fmt.Errorf("revoke image links: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM wallpapers WHERE image_uid = $1", uid); err != nil {
		return domain.Image{}, fmt.Errorf("unpublish trashed image: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Image{}, fmt.Errorf("commit image trash: %w", err)
	}
	return s.GetImage(ctx, uid, true)
}

// RestoreImage restores a trashed image without restoring public links.
func (s *Store) RestoreImage(ctx context.Context, uid string) (domain.Image, error) {
	command, err := s.pool.Exec(ctx, `UPDATE images SET delete_time = NULL, update_time = now()
		WHERE uid = $1 AND delete_time IS NOT NULL`, uid)
	if err != nil {
		return domain.Image{}, fmt.Errorf("restore image: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Image{}, fmt.Errorf("%w: trashed image", domain.ErrNotFound)
	}
	return s.GetImage(ctx, uid, false)
}

// PurgeImage permanently removes one trashed image.
func (s *Store) PurgeImage(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM images WHERE uid = $1 AND delete_time IS NOT NULL", uid)
	if err != nil {
		return fmt.Errorf("purge image: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: trashed image", domain.ErrNotFound)
	}
	return nil
}

// EmptyImageTrash permanently removes all trashed images.
func (s *Store) EmptyImageTrash(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM images WHERE delete_time IS NOT NULL")
	return wrapDatabaseError("empty image trash", err)
}

// PurgeTrashedImages removes images past retention.
func (s *Store) PurgeTrashedImages(ctx context.Context, cutoff time.Time) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM images WHERE delete_time <= $1", cutoff)
	return wrapDatabaseError("purge expired images", err)
}

// QueueImageProcessing retries validation and preview generation.
func (s *Store) QueueImageProcessing(ctx context.Context, uid string) (domain.Image, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Image{}, fmt.Errorf("begin image processing retry: %w", err)
	}
	defer tx.Rollback(ctx)
	command, err := tx.Exec(ctx, `UPDATE images SET processing_status = 'pending', update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid)
	if err != nil {
		return domain.Image{}, fmt.Errorf("mark image pending: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Image{}, fmt.Errorf("%w: image", domain.ErrNotFound)
	}
	if err := enqueueJob(ctx, tx, domain.ProcessingJobImage, uid); err != nil {
		return domain.Image{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Image{}, fmt.Errorf("commit image retry: %w", err)
	}
	return s.GetImage(ctx, uid, false)
}

// ListImageAlbums lists collections and visible counts.
