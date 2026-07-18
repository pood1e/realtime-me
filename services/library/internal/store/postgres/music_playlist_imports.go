package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const playlistImportColumns = `uid, provider_id, source, status,
	COALESCE(playlist_uid, ''), failure_code, create_time, update_time`

// QueuePlaylistImport creates one durable provider operation.
func (s *Store) QueuePlaylistImport(ctx context.Context, operation domain.PlaylistImport) (domain.PlaylistImport, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.PlaylistImport{}, fmt.Errorf("begin playlist import queue: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO music_playlist_imports
		(uid, provider_id, source, status, create_time, update_time)
		VALUES ($1, $2, $3, 'pending', $4, $4)`, operation.UID, operation.Provider,
		operation.Source, operation.CreateTime); err != nil {
		return domain.PlaylistImport{}, fmt.Errorf("create playlist import: %w", err)
	}
	if err := enqueueJob(ctx, tx, domain.ProcessingJobPlaylistImport, operation.UID); err != nil {
		return domain.PlaylistImport{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PlaylistImport{}, fmt.Errorf("commit playlist import queue: %w", err)
	}
	return s.GetPlaylistImport(ctx, operation.UID)
}

// GetPlaylistImport returns current operation state.
func (s *Store) GetPlaylistImport(ctx context.Context, uid string) (domain.PlaylistImport, error) {
	operation, err := scanPlaylistImport(s.pool.QueryRow(ctx,
		"SELECT "+playlistImportColumns+" FROM music_playlist_imports WHERE uid = $1", uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PlaylistImport{}, fmt.Errorf("%w: playlist import", domain.ErrNotFound)
	}
	if err != nil {
		return domain.PlaylistImport{}, fmt.Errorf("get playlist import: %w", err)
	}
	return operation, nil
}

// GetPlaylistImportForProcessing returns one leased operation.
func (s *Store) GetPlaylistImportForProcessing(ctx context.Context, uid string) (domain.PlaylistImport, error) {
	operation, err := s.GetPlaylistImport(ctx, uid)
	if err != nil {
		return domain.PlaylistImport{}, err
	}
	if operation.Status != domain.PlaylistImportRunning {
		return domain.PlaylistImport{}, fmt.Errorf("%w: playlist import is not running", domain.ErrNotFound)
	}
	return operation, nil
}

func scanPlaylistImport(row rowScanner) (domain.PlaylistImport, error) {
	var operation domain.PlaylistImport
	err := row.Scan(&operation.UID, &operation.Provider, &operation.Source, &operation.Status,
		&operation.PlaylistUID, &operation.FailureCode, &operation.CreateTime, &operation.UpdateTime)
	return operation, err
}
