package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const maximumProcessingAttempts = 5

// HeartbeatWorker records the worker's liveness.
func (s *Store) HeartbeatWorker(ctx context.Context, now time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO worker_state (singleton, heartbeat_time) VALUES (TRUE, $1)
		ON CONFLICT (singleton) DO UPDATE SET heartbeat_time = EXCLUDED.heartbeat_time`, now)
	return wrapDatabaseError("record worker heartbeat", err)
}

// GetWorkerHealth returns the latest heartbeat and queued job count.
func (s *Store) GetWorkerHealth(ctx context.Context) (domain.WorkerHealth, error) {
	var health domain.WorkerHealth
	var heartbeat *time.Time
	if err := s.pool.QueryRow(ctx, `SELECT
		(SELECT heartbeat_time FROM worker_state WHERE singleton = TRUE),
		(SELECT COUNT(*) FROM processing_jobs WHERE status IN ('pending', 'running'))`).Scan(&heartbeat, &health.PendingJobs); err != nil {
		return domain.WorkerHealth{}, fmt.Errorf("get worker health: %w", err)
	}
	health.HeartbeatTime = heartbeat
	return health, nil
}

// ClaimProcessingJob leases one available or abandoned job.
func (s *Store) ClaimProcessingJob(ctx context.Context, now time.Time, leaseDuration time.Duration, kinds []domain.ProcessingJobKind) (*domain.ProcessingJob, error) {
	if len(kinds) == 0 {
		return nil, fmt.Errorf("%w: processing job kinds are required", domain.ErrInvalidArgument)
	}
	kindValues := make([]string, len(kinds))
	for index, kind := range kinds {
		kindValues[index] = string(kind)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin job claim: %w", err)
	}
	defer tx.Rollback(ctx)
	var job domain.ProcessingJob
	err = tx.QueryRow(ctx, `SELECT uid, kind, resource_uid, attempts FROM processing_jobs
		WHERE kind = ANY($2) AND ((status = 'pending' AND available_time <= $1)
		OR (status = 'running' AND lease_until <= $1))
		ORDER BY CASE kind
			WHEN 'upload_finalize' THEN 0
			WHEN 'playlist_import' THEN 1
			ELSE 2
		END, available_time, create_time FOR UPDATE SKIP LOCKED LIMIT 1`, now, kindValues).Scan(
		&job.UID, &job.Kind, &job.ResourceUID, &job.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select processing job: %w", err)
	}
	job.Attempts++
	job.LeaseToken = uuid.NewString()
	if _, err := tx.Exec(ctx, `UPDATE processing_jobs SET status = 'running', attempts = $2,
		lease_until = $3, lease_token = $4, update_time = $1 WHERE uid = $5`,
		now, job.Attempts, now.Add(leaseDuration), job.LeaseToken, job.UID); err != nil {
		return nil, fmt.Errorf("lease processing job: %w", err)
	}
	if job.Kind == domain.ProcessingJobMusicDownload {
		if _, err := tx.Exec(ctx, `UPDATE music_playlist_tracks SET download_status = 'running'
			WHERE uid = $1 AND local_track_uid IS NULL`, job.ResourceUID); err != nil {
			return nil, fmt.Errorf("mark music download running: %w", err)
		}
	}
	if job.Kind == domain.ProcessingJobPlaylistImport {
		if _, err := tx.Exec(ctx, `UPDATE music_playlist_imports SET status = 'running', update_time = $2
			WHERE uid = $1 AND status IN ('pending', 'running')`, job.ResourceUID, now); err != nil {
			return nil, fmt.Errorf("mark playlist import running: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit job claim: %w", err)
	}
	return &job, nil
}

// ExtendProcessingJobLease renews only the caller's fenced lease.
func (s *Store) ExtendProcessingJobLease(ctx context.Context, job domain.ProcessingJob, leaseUntil time.Time) error {
	command, err := s.pool.Exec(ctx, `UPDATE processing_jobs SET lease_until = $3, update_time = now()
		WHERE uid = $1 AND lease_token = $2 AND status = 'running'`, job.UID, job.LeaseToken, leaseUntil)
	if err != nil {
		return fmt.Errorf("renew processing job lease: %w", err)
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("%w: processing job lease", domain.ErrConflict)
	}
	return nil
}

func lockProcessingJobLease(ctx context.Context, tx pgx.Tx, job domain.ProcessingJob) error {
	var uid string
	err := tx.QueryRow(ctx, `SELECT uid FROM processing_jobs
		WHERE uid = $1 AND kind = $2 AND resource_uid = $3 AND lease_token = $4
		AND status = 'running' AND lease_until > now() FOR UPDATE`,
		job.UID, job.Kind, job.ResourceUID, job.LeaseToken).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: processing job lease", domain.ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("lock processing job lease: %w", err)
	}
	return nil
}

// CompleteProcessingJob removes a successfully persisted fenced job.
func (s *Store) CompleteProcessingJob(ctx context.Context, job domain.ProcessingJob) error {
	command, err := s.pool.Exec(ctx, `DELETE FROM processing_jobs
		WHERE uid = $1 AND lease_token = $2 AND status = 'running'`, job.UID, job.LeaseToken)
	if err != nil {
		return fmt.Errorf("complete processing job: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: processing job lease", domain.ErrConflict)
	}
	return nil
}

// FailProcessingJob retries a bounded number of times before marking the resource failed.
func (s *Store) FailProcessingJob(ctx context.Context, job domain.ProcessingJob, errorCode string, retry bool, retryTime time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin processing failure: %w", err)
	}
	defer tx.Rollback(ctx)
	status := "pending"
	if !retry || job.Attempts >= maximumProcessingAttempts {
		status = "failed"
	}
	command, err := tx.Exec(ctx, `UPDATE processing_jobs SET status = $3, available_time = $4,
		lease_until = NULL, lease_token = NULL, error_code = $5, update_time = now()
		WHERE uid = $1 AND lease_token = $2 AND status = 'running'`,
		job.UID, job.LeaseToken, status, retryTime, errorCode)
	if err != nil {
		return fmt.Errorf("fail processing job: %w", err)
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("%w: processing job lease", domain.ErrConflict)
	}
	if status == "failed" {
		table, ok := processingTable(job.Kind)
		if !ok && !allowsJobOnlyFailure(job.Kind) {
			return fmt.Errorf("unknown processing job kind %q", job.Kind)
		}
		if ok {
			if _, err := tx.Exec(ctx, "UPDATE "+table+" SET processing_status = 'failed', update_time = now() WHERE uid = $1", job.ResourceUID); err != nil {
				return fmt.Errorf("mark resource processing failed: %w", err)
			}
		}
	}
	if job.Kind == domain.ProcessingJobMusicDownload {
		if _, err := tx.Exec(ctx, `UPDATE music_playlist_tracks SET download_status = $2
			WHERE uid = $1`, job.ResourceUID, status); err != nil {
			return fmt.Errorf("mark music download %s: %w", status, err)
		}
	}
	if job.Kind == domain.ProcessingJobUploadFinalize && status == "failed" {
		if _, err := tx.Exec(ctx, `UPDATE uploads SET status = 'failed', failure_code = $2
			WHERE uid = $1 AND status = 'finalizing'`, job.ResourceUID, errorCode); err != nil {
			return fmt.Errorf("mark upload finalization failed: %w", err)
		}
	}
	if job.Kind == domain.ProcessingJobPlaylistImport {
		if _, err := tx.Exec(ctx, `UPDATE music_playlist_imports SET status = $2,
			failure_code = $3, update_time = now() WHERE uid = $1 AND status = 'running'`,
			job.ResourceUID, status, errorCode); err != nil {
			return fmt.Errorf("mark playlist import %s: %w", status, err)
		}
	}
	return tx.Commit(ctx)
}

func enqueueJob(ctx context.Context, tx pgx.Tx, kind domain.ProcessingJobKind, resourceUID string) error {
	_, err := tx.Exec(ctx, `INSERT INTO processing_jobs (uid, kind, resource_uid, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (kind, resource_uid) DO UPDATE SET status = 'pending', attempts = 0,
		available_time = now(), lease_until = NULL, lease_token = NULL, error_code = '', update_time = now()
		WHERE processing_jobs.status <> 'running'`,
		uuid.NewString(), string(kind), resourceUID)
	if err != nil {
		return fmt.Errorf("queue %s processing: %w", kind, err)
	}
	return nil
}

func (s *Store) queueProcessing(ctx context.Context, table string, kind domain.ProcessingJobKind, uid string, get func() (domain.Book, error)) (domain.Book, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Book{}, fmt.Errorf("begin processing retry: %w", err)
	}
	defer tx.Rollback(ctx)
	command, err := tx.Exec(ctx, "UPDATE "+table+" SET processing_status = 'pending', update_time = now() WHERE uid = $1 AND delete_time IS NULL", uid)
	if err != nil {
		return domain.Book{}, fmt.Errorf("mark processing pending: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Book{}, fmt.Errorf("%w: resource", domain.ErrNotFound)
	}
	if err := enqueueJob(ctx, tx, kind, uid); err != nil {
		return domain.Book{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Book{}, fmt.Errorf("commit processing retry: %w", err)
	}
	return get()
}

func upsertArtifact(ctx context.Context, tx pgx.Tx, artifact domain.Artifact) error {
	_, err := tx.Exec(ctx, `INSERT INTO content_artifacts
		(uid, content_uid, kind, variant, content_type, storage_key, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (content_uid, kind, variant) DO UPDATE SET content_type = EXCLUDED.content_type,
		storage_key = EXCLUDED.storage_key, width = EXCLUDED.width, height = EXCLUDED.height,
		create_time = now()`, artifact.UID, artifact.ContentUID, artifact.Kind, artifact.Variant,
		artifact.ContentType, artifact.StorageKey, artifact.Width, artifact.Height)
	if err != nil {
		return fmt.Errorf("save content artifact: %w", err)
	}
	return nil
}

func processingTable(kind domain.ProcessingJobKind) (string, bool) {
	switch kind {
	case domain.ProcessingJobBook:
		return "books", true
	case domain.ProcessingJobTrack:
		return "tracks", true
	case domain.ProcessingJobImage:
		return "images", true
	default:
		return "", false
	}
}

func allowsJobOnlyFailure(kind domain.ProcessingJobKind) bool {
	switch kind {
	case domain.ProcessingJobWallpaper,
		domain.ProcessingJobMusicDownload,
		domain.ProcessingJobMusicArtwork,
		domain.ProcessingJobUploadFinalize,
		domain.ProcessingJobPlaylistImport:
		return true
	default:
		return false
	}
}

func wrapDatabaseError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
