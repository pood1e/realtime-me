package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

// GetUploadForFinalization returns one worker-owned upload operation.
func (s *Store) GetUploadForFinalization(ctx context.Context, uid string) (domain.Upload, error) {
	upload, err := s.GetUpload(ctx, uid)
	if err != nil {
		return domain.Upload{}, err
	}
	if upload.Status != domain.UploadStatusFinalizing {
		return domain.Upload{}, fmt.Errorf("%w: upload is not finalizing", domain.ErrNotFound)
	}
	return upload, nil
}

// CompleteUploadFinalization records a sealed content address under a fenced lease.
func (s *Store) CompleteUploadFinalization(ctx context.Context, job domain.ProcessingJob, sealed domain.SealedContent) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin upload finalization completion: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := lockProcessingJobLease(ctx, tx, job); err != nil {
		return err
	}
	var totalSize int64
	err = tx.QueryRow(ctx, `SELECT total_size_bytes FROM uploads
		WHERE uid = $1 AND status = 'finalizing' FOR UPDATE`, job.ResourceUID).Scan(&totalSize)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: finalizing upload", domain.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("lock finalizing upload: %w", err)
	}
	if len(sealed.SHA256) != 32 || sealed.SizeBytes != totalSize || sealed.ContentType == "" || sealed.StorageKey == "" {
		return fmt.Errorf("%w: finalized upload metadata is invalid", domain.ErrConflict)
	}
	content, err := upsertContent(ctx, tx, sealed)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE uploads SET status = 'sealed', sealed_sha256 = $2,
		sealed_size_bytes = $3, sealed_content_type = $4, sealed_storage_key = $5,
		sealed_content_uid = $6, failure_code = '', finalize_time = now() WHERE uid = $1`, job.ResourceUID,
		sealed.SHA256, sealed.SizeBytes, sealed.ContentType, sealed.StorageKey, content.UID); err != nil {
		return fmt.Errorf("complete upload finalization: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM upload_chunks WHERE upload_uid = $1", job.ResourceUID); err != nil {
		return fmt.Errorf("clear finalized upload ranges: %w", err)
	}
	return tx.Commit(ctx)
}
