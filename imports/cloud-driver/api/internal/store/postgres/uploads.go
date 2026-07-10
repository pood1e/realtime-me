package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *Store) CreateUpload(ctx context.Context, upload domain.Upload) (domain.Upload, error) {
	if _, err := s.pool.Exec(ctx, `INSERT INTO uploads
		(uid, file_name, content_type, total_size_bytes, received_bytes, chunk_size_bytes, status, create_time, expire_time)
		VALUES ($1, $2, $3, $4, 0, $5, $6, $7, $8)`, upload.UID, upload.FileName, upload.ContentType,
		upload.TotalSizeBytes, upload.ChunkSizeBytes, upload.Status, upload.CreateTime, upload.ExpireTime); err != nil {
		return domain.Upload{}, fmt.Errorf("create upload: %w", err)
	}
	return s.GetUpload(ctx, upload.UID)
}

// ReservedUploadBytes returns capacity held by active sessions.
func (s *Store) ReservedUploadBytes(ctx context.Context) (int64, error) {
	var reserved int64
	if err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_size_bytes - received_bytes), 0)
		FROM uploads WHERE status = 'active' AND expire_time > now()`).Scan(&reserved); err != nil {
		return 0, fmt.Errorf("sum reserved upload bytes: %w", err)
	}
	return reserved, nil
}

// GetUpload returns metadata and acknowledged ranges.
func (s *Store) GetUpload(ctx context.Context, uid string) (domain.Upload, error) {
	upload, err := scanUpload(s.pool.QueryRow(ctx, `SELECT uid, file_name, content_type, total_size_bytes,
		received_bytes, chunk_size_bytes, status, create_time, expire_time, COALESCE(claimed_resource_uid, '')
		FROM uploads WHERE uid = $1`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Upload{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Upload{}, fmt.Errorf("get upload: %w", err)
	}
	if upload.Status == domain.UploadStatusActive && !upload.ExpireTime.After(time.Now().UTC()) {
		if _, err := s.pool.Exec(ctx, "UPDATE uploads SET status = 'expired' WHERE uid = $1 AND status = 'active'", uid); err != nil {
			return domain.Upload{}, fmt.Errorf("expire upload: %w", err)
		}
		upload.Status = domain.UploadStatusExpired
	}
	rows, err := s.pool.Query(ctx, `SELECT start_offset, end_offset, checksum FROM upload_chunks
		WHERE upload_uid = $1 ORDER BY start_offset`, uid)
	if err != nil {
		return domain.Upload{}, fmt.Errorf("get upload chunks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var chunk domain.UploadChunk
		if err := rows.Scan(&chunk.StartOffset, &chunk.EndOffset, &chunk.Checksum); err != nil {
			return domain.Upload{}, fmt.Errorf("scan upload chunk: %w", err)
		}
		upload.Chunks = append(upload.Chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return domain.Upload{}, fmt.Errorf("iterate upload chunks: %w", err)
	}
	return upload, nil
}

// RecordUploadChunk atomically acknowledges one non-overlapping range.
func (s *Store) RecordUploadChunk(ctx context.Context, uploadUID string, chunk domain.UploadChunk) (domain.Upload, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Upload{}, fmt.Errorf("begin chunk transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	var status string
	var expireTime time.Time
	if err := tx.QueryRow(ctx, "SELECT status, expire_time FROM uploads WHERE uid = $1 FOR UPDATE", uploadUID).Scan(&status, &expireTime); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Upload{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
		}
		return domain.Upload{}, fmt.Errorf("lock upload: %w", err)
	}
	if domain.UploadStatus(status) != domain.UploadStatusActive || !expireTime.After(time.Now().UTC()) {
		return domain.Upload{}, fmt.Errorf("%w: upload is not active", domain.ErrConflict)
	}
	var existingEnd int64
	var existingChecksum []byte
	err = tx.QueryRow(ctx, `SELECT end_offset, checksum FROM upload_chunks
		WHERE upload_uid = $1 AND start_offset = $2`, uploadUID, chunk.StartOffset).Scan(&existingEnd, &existingChecksum)
	if err == nil {
		if existingEnd != chunk.EndOffset || !bytes.Equal(existingChecksum, chunk.Checksum) {
			return domain.Upload{}, fmt.Errorf("%w: chunk offset contains different content", domain.ErrConflict)
		}
		if err := tx.Commit(ctx); err != nil {
			return domain.Upload{}, fmt.Errorf("commit repeated chunk: %w", err)
		}
		return s.GetUpload(ctx, uploadUID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Upload{}, fmt.Errorf("read upload chunk: %w", err)
	}
	var overlaps bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM upload_chunks
		WHERE upload_uid = $1 AND start_offset < $3 AND end_offset > $2)`, uploadUID, chunk.StartOffset, chunk.EndOffset).Scan(&overlaps); err != nil {
		return domain.Upload{}, fmt.Errorf("check upload overlap: %w", err)
	}
	if overlaps {
		return domain.Upload{}, fmt.Errorf("%w: upload chunk overlaps", domain.ErrConflict)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO upload_chunks (upload_uid, start_offset, end_offset, checksum)
		VALUES ($1, $2, $3, $4)`, uploadUID, chunk.StartOffset, chunk.EndOffset, chunk.Checksum); err != nil {
		return domain.Upload{}, fmt.Errorf("insert upload chunk: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE uploads SET received_bytes = (
		SELECT COALESCE(SUM(end_offset - start_offset), 0) FROM upload_chunks WHERE upload_uid = $1)
		WHERE uid = $1`, uploadUID); err != nil {
		return domain.Upload{}, fmt.Errorf("update upload size: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Upload{}, fmt.Errorf("commit upload chunk: %w", err)
	}
	return s.GetUpload(ctx, uploadUID)
}

// DeleteUpload deletes unclaimed upload metadata.
func (s *Store) DeleteUpload(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM uploads WHERE uid = $1 AND status <> 'claimed'", uid)
	if err != nil {
		return fmt.Errorf("delete upload: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: unclaimed upload", domain.ErrNotFound)
	}
	return nil
}

// ListExpiredUploads returns sessions whose temporary files may be removed.
func (s *Store) ListExpiredUploads(ctx context.Context, cutoff time.Time) ([]domain.Upload, error) {
	rows, err := s.pool.Query(ctx, `SELECT uid, file_name, content_type, total_size_bytes, received_bytes,
		chunk_size_bytes, status, create_time, expire_time, COALESCE(claimed_resource_uid, '')
		FROM uploads WHERE status <> 'claimed' AND expire_time <= $1 ORDER BY expire_time`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("list expired uploads: %w", err)
	}
	defer rows.Close()
	var uploads []domain.Upload
	for rows.Next() {
		upload, err := scanUpload(rows)
		if err != nil {
			return nil, fmt.Errorf("scan expired upload: %w", err)
		}
		uploads = append(uploads, upload)
	}
	return uploads, rows.Err()
}

// ListUnhashedContent returns legacy source objects awaiting migration.
func lockCompleteUpload(ctx context.Context, tx pgx.Tx, uploadUID string) (domain.Upload, error) {
	upload, err := scanUpload(tx.QueryRow(ctx, `SELECT uid, file_name, content_type, total_size_bytes,
		received_bytes, chunk_size_bytes, status, create_time, expire_time, COALESCE(claimed_resource_uid, '')
		FROM uploads WHERE uid = $1 FOR UPDATE`, uploadUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Upload{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Upload{}, fmt.Errorf("lock upload: %w", err)
	}
	if upload.Status == domain.UploadStatusClaimed {
		return upload, nil
	}
	if upload.Status != domain.UploadStatusActive || !upload.ExpireTime.After(time.Now().UTC()) || upload.ReceivedBytes != upload.TotalSizeBytes {
		return domain.Upload{}, fmt.Errorf("%w: upload is incomplete", domain.ErrConflict)
	}
	var expected int64
	rows, err := tx.Query(ctx, `SELECT start_offset, end_offset FROM upload_chunks WHERE upload_uid = $1 ORDER BY start_offset`, uploadUID)
	if err != nil {
		return domain.Upload{}, fmt.Errorf("read upload ranges: %w", err)
	}
	for rows.Next() {
		var start, end int64
		if err := rows.Scan(&start, &end); err != nil {
			rows.Close()
			return domain.Upload{}, fmt.Errorf("scan upload range: %w", err)
		}
		if start != expected {
			rows.Close()
			return domain.Upload{}, fmt.Errorf("%w: upload has missing ranges", domain.ErrConflict)
		}
		expected = end
	}
	rows.Close()
	if expected != upload.TotalSizeBytes {
		return domain.Upload{}, fmt.Errorf("%w: upload has missing ranges", domain.ErrConflict)
	}
	return upload, nil
}

func upsertContent(ctx context.Context, tx pgx.Tx, sealed domain.SealedContent) (domain.ContentObject, error) {
	uid := uuid.NewString()
	content, err := scanContent(tx.QueryRow(ctx, `INSERT INTO content_objects
		(uid, sha256, size_bytes, content_type, storage_key) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sha256) DO UPDATE SET sha256 = EXCLUDED.sha256
		RETURNING uid, sha256, size_bytes, content_type, storage_key, create_time`, uid, sealed.SHA256,
		sealed.SizeBytes, sealed.ContentType, sealed.StorageKey))
	if err != nil {
		return domain.ContentObject{}, fmt.Errorf("create content object: %w", err)
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

type rowScanner interface{ Scan(...any) error }

func scanUpload(row rowScanner) (domain.Upload, error) {
	var upload domain.Upload
	var status string
	if err := row.Scan(&upload.UID, &upload.FileName, &upload.ContentType, &upload.TotalSizeBytes,
		&upload.ReceivedBytes, &upload.ChunkSizeBytes, &status, &upload.CreateTime, &upload.ExpireTime,
		&upload.ClaimedUID); err != nil {
		return domain.Upload{}, err
	}
	upload.Status = domain.UploadStatus(status)
	return upload, nil
}
