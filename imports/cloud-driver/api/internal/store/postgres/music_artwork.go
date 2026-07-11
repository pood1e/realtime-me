package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
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
func (s *Store) CompleteMusicArtwork(ctx context.Context, artwork domain.Artifact) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin music artwork completion: %w", err)
	}
	defer tx.Rollback(ctx)
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
