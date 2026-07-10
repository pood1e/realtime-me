package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *Store) ListImageAlbums(ctx context.Context) ([]domain.ImageAlbum, error) {
	rows, err := s.pool.Query(ctx, `SELECT album.uid, album.display_name,
		COUNT(image.uid) FILTER (WHERE image.delete_time IS NULL), album.create_time
		FROM image_albums album LEFT JOIN images image ON image.album_uid = album.uid
		GROUP BY album.uid ORDER BY album.display_name, album.uid`)
	if err != nil {
		return nil, fmt.Errorf("list image albums: %w", err)
	}
	defer rows.Close()
	var albums []domain.ImageAlbum
	for rows.Next() {
		var album domain.ImageAlbum
		if err := rows.Scan(&album.UID, &album.DisplayName, &album.ImageCount, &album.CreateTime); err != nil {
			return nil, fmt.Errorf("scan image album: %w", err)
		}
		albums = append(albums, album)
	}
	return albums, rows.Err()
}

// CreateImageAlbum creates a collection.
func (s *Store) CreateImageAlbum(ctx context.Context, album domain.ImageAlbum) (domain.ImageAlbum, error) {
	if err := s.pool.QueryRow(ctx, `INSERT INTO image_albums (uid, display_name) VALUES ($1, $2)
		RETURNING create_time`, album.UID, album.DisplayName).Scan(&album.CreateTime); err != nil {
		return domain.ImageAlbum{}, fmt.Errorf("create image album: %w", err)
	}
	return album, nil
}

// DeleteImageAlbum deletes a collection without deleting images.
func (s *Store) DeleteImageAlbum(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM image_albums WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("delete image album: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: image album", domain.ErrNotFound)
	}
	return nil
}

// ListImageLinks lists all link history for one image.
func (s *Store) ListImageLinks(ctx context.Context, imageUID string) ([]domain.ImageLink, error) {
	if _, err := s.GetImage(ctx, imageUID, true); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT uid, image_uid, create_time, revoke_time
		FROM image_links WHERE image_uid = $1 ORDER BY create_time DESC`, imageUID)
	if err != nil {
		return nil, fmt.Errorf("list image links: %w", err)
	}
	defer rows.Close()
	var links []domain.ImageLink
	for rows.Next() {
		link, err := scanImageLink(rows)
		if err != nil {
			return nil, fmt.Errorf("scan image link: %w", err)
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// CreateImageLink publishes one stable anonymous ID.
func (s *Store) CreateImageLink(ctx context.Context, link domain.ImageLink) (domain.ImageLink, error) {
	image, err := s.GetImage(ctx, link.ImageUID, false)
	if err != nil {
		return domain.ImageLink{}, err
	}
	if image.ProcessingStatus != domain.ProcessingStatusReady {
		return domain.ImageLink{}, fmt.Errorf("%w: image is not ready", domain.ErrConflict)
	}
	if err := s.pool.QueryRow(ctx, `INSERT INTO image_links (uid, image_uid) VALUES ($1, $2)
		RETURNING create_time`, link.UID, link.ImageUID).Scan(&link.CreateTime); err != nil {
		return domain.ImageLink{}, fmt.Errorf("create image link: %w", err)
	}
	return link, nil
}

// RevokeImageLink disables one link immediately.
func (s *Store) RevokeImageLink(ctx context.Context, uid string) (domain.ImageLink, error) {
	link, err := scanImageLink(s.pool.QueryRow(ctx, `UPDATE image_links SET revoke_time = now()
		WHERE uid = $1 AND revoke_time IS NULL RETURNING uid, image_uid, create_time, revoke_time`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ImageLink{}, fmt.Errorf("%w: active image link", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ImageLink{}, fmt.Errorf("revoke image link: %w", err)
	}
	return link, nil
}

// GetImageByLink resolves an active anonymous image link.
func (s *Store) GetImageByLink(ctx context.Context, uid string) (domain.Image, error) {
	image, err := scanImage(s.pool.QueryRow(ctx, `SELECT `+imageColumns+` FROM `+imageFrom+`
		JOIN image_links link ON link.image_uid = image.uid
		WHERE link.uid = $1 AND link.revoke_time IS NULL AND image.delete_time IS NULL
		AND image.processing_status = 'ready'`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Image{}, fmt.Errorf("%w: active image link", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Image{}, fmt.Errorf("resolve image link: %w", err)
	}
	return image, nil
}

// GetWallpaper returns one published wallpaper and available variants.
