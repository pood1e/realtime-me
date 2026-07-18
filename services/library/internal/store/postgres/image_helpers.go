package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func scanImage(row rowScanner) (domain.Image, error) {
	var image domain.Image
	var albumUID pgtype.Text
	var status string
	var deleteTime pgtype.Timestamptz
	if err := row.Scan(&image.UID, &image.ContentUID, &albumUID, &image.DisplayName, &image.OriginalFileName,
		&image.ContentType, &image.SizeBytes, &image.Width, &image.Height, &image.PreviewStorageKey,
		&status, &image.CreateTime, &image.UpdateTime, &deleteTime); err != nil {
		return domain.Image{}, err
	}
	if albumUID.Valid {
		image.AlbumUID = copyString(&albumUID.String)
	}
	image.ProcessingStatus = domain.ProcessingStatus(status)
	if deleteTime.Valid {
		value := deleteTime.Time.UTC()
		image.DeleteTime = &value
	}
	return image, nil
}

func scanImageLink(row rowScanner) (domain.ImageLink, error) {
	var link domain.ImageLink
	var revokeTime pgtype.Timestamptz
	if err := row.Scan(&link.UID, &link.ImageUID, &link.CreateTime, &revokeTime); err != nil {
		return domain.ImageLink{}, err
	}
	if revokeTime.Valid {
		value := revokeTime.Time.UTC()
		link.RevokeTime = &value
	}
	return link, nil
}

func scanWallpaper(row rowScanner) (domain.Wallpaper, error) {
	var wallpaper domain.Wallpaper
	if err := row.Scan(&wallpaper.UID, &wallpaper.ImageUID, &wallpaper.Title, &wallpaper.Tags,
		&wallpaper.DominantColor, &wallpaper.Width, &wallpaper.Height, &wallpaper.ContentType,
		&wallpaper.StorageKey, &wallpaper.PublishTime, &wallpaper.UpdateTime); err != nil {
		return domain.Wallpaper{}, err
	}
	return wallpaper, nil
}

func supportedImageType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml":
		return true
	default:
		return false
	}
}

func validateImageAlbum(ctx context.Context, tx pgx.Tx, albumUID *string) error {
	if albumUID == nil {
		return nil
	}
	var found bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM image_albums WHERE uid = $1)", *albumUID).Scan(&found); err != nil {
		return fmt.Errorf("validate image album: %w", err)
	}
	if !found {
		return fmt.Errorf("%w: image album", domain.ErrNotFound)
	}
	return nil
}
