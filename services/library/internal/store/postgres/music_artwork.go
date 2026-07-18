package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

// GetMusicArtwork returns the content hash and provider image for a visible local track.
func (s *Store) GetMusicArtwork(ctx context.Context, trackUID string) (domain.ContentObject, string, error) {
	track, err := s.GetTrack(ctx, trackUID, false)
	if err != nil {
		return domain.ContentObject{}, "", err
	}
	if track.ArtworkStorageKey != "" {
		return domain.ContentObject{}, "", fmt.Errorf("%w: track artwork already exists", domain.ErrNotFound)
	}
	content, err := s.GetContent(ctx, track.ContentUID)
	if err != nil {
		return domain.ContentObject{}, "", err
	}
	var artworkURL string
	err = s.pool.QueryRow(ctx, `SELECT item.artwork_url FROM music_playlist_tracks item
		WHERE item.local_track_uid = $1 AND item.artwork_url <> ''
		ORDER BY item.position, item.uid LIMIT 1`, trackUID).Scan(&artworkURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ContentObject{}, "", fmt.Errorf("%w: provider artwork", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ContentObject{}, "", fmt.Errorf("get provider artwork: %w", err)
	}
	return content, artworkURL, nil
}

// CompleteMusicArtwork links one local JPEG artifact to its audio content.
func (s *Store) CompleteMusicArtwork(ctx context.Context, job domain.ProcessingJob, artwork domain.Artifact) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin music artwork completion: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := lockProcessingJobLease(ctx, tx, job); err != nil {
		return err
	}
	if err := upsertArtifact(ctx, tx, artwork); err != nil {
		return err
	}
	command, err := tx.Exec(ctx, "UPDATE tracks SET update_time = now() WHERE content_uid = $1", artwork.ContentUID)
	if err != nil {
		return fmt.Errorf("touch music artwork track: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: music artwork track", domain.ErrNotFound)
	}
	return tx.Commit(ctx)
}

func queueProviderArtworkIfMissing(ctx context.Context, tx pgx.Tx, trackUID string) error {
	var available bool
	if err := tx.QueryRow(ctx, `SELECT
		EXISTS (SELECT 1 FROM music_playlist_tracks
			WHERE local_track_uid = $1 AND artwork_url <> '')
		AND NOT EXISTS (
			SELECT 1 FROM tracks track JOIN content_artifacts artifact
			ON artifact.content_uid = track.content_uid
			WHERE track.uid = $1 AND artifact.kind = 'track_artwork'
			AND artifact.variant = 'default'
		)`, trackUID).Scan(&available); err != nil {
		return fmt.Errorf("check provider artwork fallback: %w", err)
	}
	if !available {
		return nil
	}
	return enqueueJob(ctx, tx, domain.ProcessingJobMusicArtwork, trackUID)
}
