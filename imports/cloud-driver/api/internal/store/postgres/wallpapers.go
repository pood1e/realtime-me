package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	wallpaperColumns = `wallpaper.uid, wallpaper.image_uid, wallpaper.title, wallpaper.tags,
		wallpaper.dominant_color, image.width, image.height, content.content_type,
		content.storage_key, wallpaper.publish_time, wallpaper.update_time`
	wallpaperFrom = `wallpapers wallpaper JOIN images image ON image.uid = wallpaper.image_uid
		JOIN content_objects content ON content.uid = image.content_uid`
)

func (s *Store) GetWallpaper(ctx context.Context, uid string) (domain.Wallpaper, error) {
	wallpaper, err := scanWallpaper(s.pool.QueryRow(ctx, "SELECT "+wallpaperColumns+" FROM "+wallpaperFrom+
		" WHERE wallpaper.uid = $1 AND image.delete_time IS NULL", uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Wallpaper{}, fmt.Errorf("%w: wallpaper", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Wallpaper{}, fmt.Errorf("get wallpaper: %w", err)
	}
	variants, err := s.listWallpaperArtifacts(ctx, wallpaper.ImageUID)
	if err != nil {
		return domain.Wallpaper{}, err
	}
	wallpaper.Variants = variants
	return wallpaper, nil
}

// ListWallpapers lists the published catalog.
func (s *Store) ListWallpapers(ctx context.Context, filter domain.WallpaperListQuery) (domain.WallpaperPage, error) {
	pageSize := normalizePageSize(filter.PageSize)
	query := "SELECT " + wallpaperColumns + " FROM " + wallpaperFrom + " WHERE image.delete_time IS NULL"
	arguments := []any{}
	if filter.Query != "" {
		arguments = append(arguments, filter.Query)
		query += fmt.Sprintf(" AND wallpaper.title ILIKE '%%' || $%d || '%%'", len(arguments))
	}
	if filter.Tag != "" {
		arguments = append(arguments, filter.Tag)
		query += fmt.Sprintf(" AND $%d = ANY(wallpaper.tags)", len(arguments))
	}
	if filter.Orientation != "" {
		condition := map[string]string{
			"landscape": "image.width > image.height",
			"portrait":  "image.height > image.width",
			"square":    "image.height = image.width",
		}[filter.Orientation]
		if condition == "" {
			return domain.WallpaperPage{}, fmt.Errorf("%w: invalid orientation", domain.ErrInvalidArgument)
		}
		query += " AND " + condition
	}
	if filter.PageToken != "" {
		cursor, err := decodeCursor(filter.PageToken)
		if err != nil {
			return domain.WallpaperPage{}, err
		}
		publishTime, err := time.Parse(time.RFC3339Nano, cursor.name)
		if err != nil {
			return domain.WallpaperPage{}, fmt.Errorf("%w: malformed wallpaper token", domain.ErrInvalidArgument)
		}
		arguments = append(arguments, publishTime, cursor.uid)
		query += fmt.Sprintf(" AND (wallpaper.publish_time, wallpaper.uid) < ($%d, $%d)", len(arguments)-1, len(arguments))
	}
	arguments = append(arguments, pageSize+1)
	query += fmt.Sprintf(" ORDER BY wallpaper.publish_time DESC, wallpaper.uid DESC LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.WallpaperPage{}, fmt.Errorf("list wallpapers: %w", err)
	}
	page := domain.WallpaperPage{}
	for rows.Next() {
		wallpaper, err := scanWallpaper(rows)
		if err != nil {
			rows.Close()
			return domain.WallpaperPage{}, fmt.Errorf("scan wallpaper: %w", err)
		}
		page.Wallpapers = append(page.Wallpapers, wallpaper)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.WallpaperPage{}, fmt.Errorf("iterate wallpapers: %w", err)
	}
	rows.Close()
	if len(page.Wallpapers) > pageSize {
		last := page.Wallpapers[pageSize-1]
		page.Wallpapers = page.Wallpapers[:pageSize]
		page.NextPageToken = encodeCursor(last.PublishTime.Format(time.RFC3339Nano), last.UID)
	}
	imageUIDs := make([]string, len(page.Wallpapers))
	for index, wallpaper := range page.Wallpapers {
		imageUIDs[index] = wallpaper.ImageUID
	}
	variants, err := s.listWallpaperArtifactsByImage(ctx, imageUIDs)
	if err != nil {
		return domain.WallpaperPage{}, err
	}
	for index := range page.Wallpapers {
		page.Wallpapers[index].Variants = variants[page.Wallpapers[index].ImageUID]
	}
	return page, nil
}

// PublishWallpaper manually publishes a ready image.
func (s *Store) PublishWallpaper(ctx context.Context, wallpaper domain.Wallpaper) (domain.Wallpaper, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Wallpaper{}, fmt.Errorf("begin wallpaper publication: %w", err)
	}
	defer tx.Rollback(ctx)
	var ready bool
	if err := tx.QueryRow(ctx, `SELECT processing_status = 'ready' FROM images
		WHERE uid = $1 AND delete_time IS NULL`, wallpaper.ImageUID).Scan(&ready); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Wallpaper{}, fmt.Errorf("%w: image", domain.ErrNotFound)
		}
		return domain.Wallpaper{}, fmt.Errorf("get wallpaper image: %w", err)
	}
	if !ready {
		return domain.Wallpaper{}, fmt.Errorf("%w: image is not ready", domain.ErrConflict)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO wallpapers (uid, image_uid, title, tags)
		VALUES ($1, $2, $3, $4)`, wallpaper.UID, wallpaper.ImageUID, wallpaper.Title, wallpaper.Tags); err != nil {
		return domain.Wallpaper{}, fmt.Errorf("publish wallpaper: %w", err)
	}
	if err := enqueueJob(ctx, tx, domain.ProcessingJobWallpaper, wallpaper.UID); err != nil {
		return domain.Wallpaper{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Wallpaper{}, fmt.Errorf("commit wallpaper publication: %w", err)
	}
	return s.GetWallpaper(ctx, wallpaper.UID)
}

// UpdateWallpaper changes public metadata without modifying its image.
func (s *Store) UpdateWallpaper(ctx context.Context, wallpaper domain.Wallpaper) (domain.Wallpaper, error) {
	command, err := s.pool.Exec(ctx, `UPDATE wallpapers SET title = $2, tags = $3, update_time = now()
		WHERE uid = $1`, wallpaper.UID, wallpaper.Title, wallpaper.Tags)
	if err != nil {
		return domain.Wallpaper{}, fmt.Errorf("update wallpaper: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Wallpaper{}, fmt.Errorf("%w: wallpaper", domain.ErrNotFound)
	}
	return s.GetWallpaper(ctx, wallpaper.UID)
}

// UnpublishWallpaper removes a wallpaper without deleting its private image.
func (s *Store) UnpublishWallpaper(ctx context.Context, uid string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin wallpaper removal: %w", err)
	}
	defer tx.Rollback(ctx)
	command, err := tx.Exec(ctx, "DELETE FROM wallpapers WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("unpublish wallpaper: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: wallpaper", domain.ErrNotFound)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM processing_jobs
		WHERE kind = 'wallpaper' AND resource_uid = $1`, uid); err != nil {
		return fmt.Errorf("delete wallpaper jobs: %w", err)
	}
	return tx.Commit(ctx)
}

// GetImageForProcessing returns source metadata for the worker.
