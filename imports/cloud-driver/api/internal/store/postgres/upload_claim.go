package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

func lockCompleteUpload(ctx context.Context, tx pgx.Tx, uploadUID string) (domain.Upload, error) {
	upload, err := scanUpload(tx.QueryRow(ctx, "SELECT "+uploadColumns+" FROM uploads WHERE uid = $1 FOR UPDATE", uploadUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Upload{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Upload{}, fmt.Errorf("lock upload: %w", err)
	}
	if upload.Status == domain.UploadStatusClaimed {
		return upload, nil
	}
	if upload.Status != domain.UploadStatusSealed || upload.Sealed == nil {
		return domain.Upload{}, fmt.Errorf("%w: upload is not sealed", domain.ErrConflict)
	}
	return upload, nil
}

func contentForUpload(ctx context.Context, tx pgx.Tx, upload domain.Upload, sealed domain.SealedContent) (domain.ContentObject, error) {
	if upload.SealedContentUID == "" || upload.Sealed == nil ||
		!bytes.Equal(upload.Sealed.SHA256, sealed.SHA256) ||
		upload.Sealed.SizeBytes != sealed.SizeBytes ||
		upload.Sealed.ContentType != sealed.ContentType ||
		upload.Sealed.StorageKey != sealed.StorageKey {
		return domain.ContentObject{}, fmt.Errorf("%w: upload content metadata changed", domain.ErrConflict)
	}
	content, err := scanContent(tx.QueryRow(ctx, `SELECT uid, sha256, size_bytes, content_type, storage_key, create_time
		FROM content_objects WHERE uid = $1`, upload.SealedContentUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ContentObject{}, fmt.Errorf("%w: upload content", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ContentObject{}, fmt.Errorf("get upload content: %w", err)
	}
	return content, nil
}

func markUploadClaimed(ctx context.Context, tx pgx.Tx, uploadUID, resourceUID string) error {
	if _, err := tx.Exec(ctx, `UPDATE uploads SET status = 'claimed', claimed_resource_uid = $2, claim_time = now()
		WHERE uid = $1`, uploadUID, resourceUID); err != nil {
		return fmt.Errorf("claim upload: %w", err)
	}
	return nil
}
