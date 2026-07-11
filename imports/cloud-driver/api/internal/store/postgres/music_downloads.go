package postgres

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

// GetMusicDownload returns a leased playlist item and its encrypted provider account.
func (s *Store) GetMusicDownload(ctx context.Context, itemUID string) (domain.MusicDownload, error) {
	item, err := scanPlaylistTrack(s.pool.QueryRow(ctx, "SELECT "+playlistTrackColumns+
		" FROM music_playlist_tracks item WHERE item.uid = $1", itemUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MusicDownload{}, fmt.Errorf("%w: music download", domain.ErrNotFound)
	}
	if err != nil {
		return domain.MusicDownload{}, fmt.Errorf("get music download: %w", err)
	}
	connection, err := s.GetProviderConnection(ctx, item.Track.Provider)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.MusicDownload{}, fmt.Errorf("%w: provider account is disconnected", domain.ErrProviderReconnectRequired)
	}
	if err != nil {
		return domain.MusicDownload{}, err
	}
	return domain.MusicDownload{PlaylistTrack: item, Connection: connection}, nil
}

// CompleteMusicDownload imports downloaded audio and links matching playlist entries.
func (s *Store) CompleteMusicDownload(ctx context.Context, job domain.ProcessingJob, sealed domain.SealedContent) error {
	if !supportedAudioType(sealed.ContentType) {
		return fmt.Errorf("%w: downloaded source is not supported audio", domain.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin music download completion: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := lockProcessingJobLease(ctx, tx, job); err != nil {
		return err
	}
	locked, err := scanPlaylistTrack(tx.QueryRow(ctx, "SELECT "+playlistTrackColumns+
		" FROM music_playlist_tracks item WHERE item.uid = $1 FOR UPDATE", job.ResourceUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: playlist track", domain.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("lock playlist track: %w", err)
	}
	content, err := upsertContent(ctx, tx, sealed)
	if err != nil {
		return err
	}
	trackUID, err := findDownloadedTrack(ctx, tx, locked.Track.Provider, locked.Track.TrackID, content.UID)
	if err != nil {
		return err
	}
	createdTrack := trackUID == ""
	if createdTrack {
		trackUID = uuid.NewString()
		fileName := downloadedTrackFileName(locked.Track, content.ContentType)
		_, err = tx.Exec(ctx, `INSERT INTO tracks (uid, content_uid, title, artists, album, duration_ms,
			original_file_name, processing_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending')`, trackUID, content.UID,
			locked.Track.Title, locked.Track.Artists, locked.Track.Album, locked.Track.Duration.Milliseconds(),
			fileName)
		if err != nil {
			return fmt.Errorf("create downloaded track: %w", err)
		}
		if err := enqueueJob(ctx, tx, domain.ProcessingJobTrack, trackUID); err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(ctx, "UPDATE tracks SET delete_time = NULL, update_time = now() WHERE uid = $1", trackUID)
		if err != nil {
			return fmt.Errorf("restore downloaded track: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO music_track_sources (provider_id, external_track_id, track_uid)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider_id, external_track_id) DO UPDATE SET track_uid = EXCLUDED.track_uid`,
		locked.Track.Provider, locked.Track.TrackID, trackUID); err != nil {
		return fmt.Errorf("link downloaded track source: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE music_playlist_tracks SET download_status = 'completed',
		local_track_uid = $3 WHERE provider_id = $1 AND external_track_id = $2`, locked.Track.Provider,
		locked.Track.TrackID, trackUID); err != nil {
		return fmt.Errorf("link downloaded playlist tracks: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE music_playlists SET update_time = now() WHERE uid IN
		(SELECT playlist_uid FROM music_playlist_tracks WHERE provider_id = $1 AND external_track_id = $2)`,
		locked.Track.Provider, locked.Track.TrackID); err != nil {
		return fmt.Errorf("touch downloaded playlists: %w", err)
	}
	if !createdTrack {
		if err := queueProviderArtworkIfMissing(ctx, tx, trackUID); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit music download: %w", err)
	}
	return nil
}

func findDownloadedTrack(ctx context.Context, tx pgx.Tx, provider domain.MusicProvider, trackID, contentUID string) (string, error) {
	var uid string
	err := tx.QueryRow(ctx, `SELECT track_uid FROM music_track_sources WHERE provider_id = $1 AND external_track_id = $2`,
		provider, trackID).Scan(&uid)
	if err == nil {
		return uid, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("find source track: %w", err)
	}
	err = tx.QueryRow(ctx, "SELECT uid FROM tracks WHERE content_uid = $1", contentUID).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find content track: %w", err)
	}
	return uid, nil
}

func downloadedTrackFileName(track domain.PlayableTrack, contentType string) string {
	artist := ""
	if len(track.Artists) > 0 {
		artist = track.Artists[0]
	}
	base := strings.TrimSpace(strings.Join([]string{artist, track.Title}, " - "))
	base = strings.Trim(base, " -")
	base = strings.Map(func(character rune) rune {
		if strings.ContainsRune(`/\\\x00`, character) || character < ' ' {
			return '_'
		}
		return character
	}, base)
	if base == "" {
		base = "downloaded-track"
	}
	runes := []rune(base)
	if len(runes) > 160 {
		base = string(runes[:160])
	}
	extension := map[string]string{
		"audio/mpeg": ".mp3", "audio/mp4": ".m4a", "audio/aac": ".aac", "audio/flac": ".flac",
		"audio/ogg": ".ogg", "audio/opus": ".opus", "audio/wav": ".wav", "audio/x-wav": ".wav",
	}[contentType]
	return strings.TrimSuffix(base, filepath.Ext(base)) + extension
}
