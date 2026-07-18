package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

func upsertContent(ctx context.Context, tx pgx.Tx, sealed domain.SealedContent) (domain.ContentObject, error) {
	content, err := scanContent(tx.QueryRow(ctx, `INSERT INTO content_objects
		(uid, sha256, size_bytes, content_type, storage_key) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sha256) DO UPDATE SET sha256 = EXCLUDED.sha256
		RETURNING uid, sha256, size_bytes, content_type, storage_key, create_time`, uuid.NewString(), sealed.SHA256,
		sealed.SizeBytes, sealed.ContentType, sealed.StorageKey))
	if err != nil {
		return domain.ContentObject{}, fmt.Errorf("create content object: %w", err)
	}
	return content, nil
}
