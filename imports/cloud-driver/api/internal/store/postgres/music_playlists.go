package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

const playlistSummaryColumns = `playlist.uid, playlist.provider, playlist.external_id, playlist.display_name,
	playlist.artwork_url, playlist.provider_url, COUNT(item.uid),
	COUNT(item.uid) FILTER (WHERE item.playable),
	COUNT(item.uid) FILTER (WHERE item.download_status IN ('pending', 'running')),
	COUNT(item.uid) FILTER (WHERE item.download_status = 'completed' AND item.local_track_uid IS NOT NULL),
	COUNT(item.uid) FILTER (WHERE item.download_status = 'failed'), playlist.download_supported,
	playlist.create_time, playlist.update_time`

const playlistSummaryFrom = `music_playlists playlist
	LEFT JOIN music_playlist_tracks item ON item.playlist_uid = playlist.uid`

const playlistSummaryGroup = ` GROUP BY playlist.uid`

// ImportPlaylist replaces one provider playlist snapshot while retaining downloaded local tracks.
func (s *Store) ImportPlaylist(ctx context.Context, playlist domain.Playlist, tracks []domain.PlayableTrack) (domain.Playlist, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("begin playlist import: %w", err)
	}
	defer tx.Rollback(ctx)
	query := `INSERT INTO music_playlists (uid, provider, external_id, display_name, artwork_url,
		provider_url, download_supported, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (provider, external_id) DO UPDATE SET display_name = EXCLUDED.display_name,
		artwork_url = EXCLUDED.artwork_url, provider_url = EXCLUDED.provider_url,
		download_supported = EXCLUDED.download_supported, update_time = EXCLUDED.update_time
		RETURNING uid`
	if err := tx.QueryRow(ctx, query, playlist.UID, playlist.Provider, playlist.ExternalID, playlist.DisplayName,
		playlist.ArtworkURL, playlist.ProviderURL, playlist.DownloadSupported, playlist.CreateTime, playlist.UpdateTime).Scan(&playlist.UID); err != nil {
		return domain.Playlist{}, fmt.Errorf("save playlist: %w", err)
	}
	if running, err := playlistDownloadRunning(ctx, tx, playlist.UID); err != nil {
		return domain.Playlist{}, err
	} else if running {
		return domain.Playlist{}, fmt.Errorf("%w: playlist download is running", domain.ErrConflict)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM processing_jobs WHERE kind = 'music_download'
		AND resource_uid IN (SELECT uid FROM music_playlist_tracks WHERE playlist_uid = $1)`, playlist.UID); err != nil {
		return domain.Playlist{}, fmt.Errorf("clear playlist download jobs: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM music_playlist_tracks WHERE playlist_uid = $1", playlist.UID); err != nil {
		return domain.Playlist{}, fmt.Errorf("replace playlist tracks: %w", err)
	}
	for index, track := range tracks {
		localTrackUID, status, err := existingDownloadedTrack(ctx, tx, track.Provider, track.TrackID)
		if err != nil {
			return domain.Playlist{}, err
		}
		_, err = tx.Exec(ctx, `INSERT INTO music_playlist_tracks (uid, playlist_uid, position, provider,
			external_track_id, title, artists, album, duration_ms, artwork_url, provider_url, playable,
			lyrics_available, download_status, local_track_uid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
			uuid.NewString(), playlist.UID, index+1, track.Provider, track.TrackID, track.Title, track.Artists,
			track.Album, track.Duration.Milliseconds(), track.ArtworkURL, track.ProviderURL, track.Playable,
			track.LyricsAvailable, status, nullableText(localTrackUID))
		if err != nil {
			return domain.Playlist{}, fmt.Errorf("save playlist track: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Playlist{}, fmt.Errorf("commit playlist import: %w", err)
	}
	return s.GetPlaylist(ctx, playlist.UID)
}

// GetPlaylist returns one imported playlist with current download counts.
func (s *Store) GetPlaylist(ctx context.Context, uid string) (domain.Playlist, error) {
	playlist, err := scanPlaylist(s.pool.QueryRow(ctx, "SELECT "+playlistSummaryColumns+" FROM "+playlistSummaryFrom+
		" WHERE playlist.uid = $1"+playlistSummaryGroup, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Playlist{}, fmt.Errorf("%w: playlist", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("get playlist: %w", err)
	}
	return playlist, nil
}

// ListPlaylists lists newest imported playlists first.
func (s *Store) ListPlaylists(ctx context.Context, pageSize int, pageToken string) (domain.PlaylistPage, error) {
	pageSize = normalizePageSize(pageSize)
	query := "SELECT " + playlistSummaryColumns + " FROM " + playlistSummaryFrom
	arguments := []any{}
	if pageToken != "" {
		cursor, err := decodeCursor(pageToken)
		if err != nil {
			return domain.PlaylistPage{}, err
		}
		updateTime, err := time.Parse(time.RFC3339Nano, cursor.name)
		if err != nil {
			return domain.PlaylistPage{}, fmt.Errorf("%w: malformed playlist token", domain.ErrInvalidArgument)
		}
		arguments = append(arguments, updateTime, cursor.uid)
		query += " WHERE (playlist.update_time, playlist.uid) < ($1, $2)"
	}
	query += playlistSummaryGroup
	arguments = append(arguments, pageSize+1)
	query += fmt.Sprintf(" ORDER BY playlist.update_time DESC, playlist.uid DESC LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.PlaylistPage{}, fmt.Errorf("list playlists: %w", err)
	}
	defer rows.Close()
	page := domain.PlaylistPage{}
	for rows.Next() {
		playlist, err := scanPlaylist(rows)
		if err != nil {
			return domain.PlaylistPage{}, fmt.Errorf("scan playlist: %w", err)
		}
		page.Playlists = append(page.Playlists, playlist)
	}
	if len(page.Playlists) > pageSize {
		last := page.Playlists[pageSize-1]
		page.Playlists = page.Playlists[:pageSize]
		page.NextPageToken = encodeCursor(last.UpdateTime.Format(time.RFC3339Nano), last.UID)
	}
	return page, rows.Err()
}

// DeletePlaylist removes an imported snapshot without deleting local tracks.
func (s *Store) DeletePlaylist(ctx context.Context, uid string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin playlist deletion: %w", err)
	}
	defer tx.Rollback(ctx)
	if running, err := playlistDownloadRunning(ctx, tx, uid); err != nil {
		return err
	} else if running {
		return fmt.Errorf("%w: playlist download is running", domain.ErrConflict)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM processing_jobs WHERE kind = 'music_download'
		AND resource_uid IN (SELECT uid FROM music_playlist_tracks WHERE playlist_uid = $1)`, uid); err != nil {
		return fmt.Errorf("delete playlist jobs: %w", err)
	}
	command, err := tx.Exec(ctx, "DELETE FROM music_playlists WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("delete playlist: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: playlist", domain.ErrNotFound)
	}
	return tx.Commit(ctx)
}

func playlistDownloadRunning(ctx context.Context, tx pgx.Tx, playlistUID string) (bool, error) {
	var running bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM processing_jobs job JOIN music_playlist_tracks item ON item.uid = job.resource_uid
		WHERE job.kind = 'music_download' AND job.status = 'running' AND item.playlist_uid = $1
	)`, playlistUID).Scan(&running)
	if err != nil {
		return false, fmt.Errorf("check playlist downloads: %w", err)
	}
	return running, nil
}

func existingDownloadedTrack(ctx context.Context, tx pgx.Tx, provider domain.MusicProvider, trackID string) (string, domain.PlaylistTrackDownloadStatus, error) {
	var uid string
	err := tx.QueryRow(ctx, `SELECT uid FROM tracks WHERE source_provider = $1 AND source_track_id = $2
		AND delete_time IS NULL`, provider, trackID).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.PlaylistTrackDownloadNotStarted, nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find downloaded track: %w", err)
	}
	return uid, domain.PlaylistTrackDownloadCompleted, nil
}

func scanPlaylist(row rowScanner) (domain.Playlist, error) {
	var playlist domain.Playlist
	err := row.Scan(&playlist.UID, &playlist.Provider, &playlist.ExternalID, &playlist.DisplayName,
		&playlist.ArtworkURL, &playlist.ProviderURL, &playlist.TrackCount, &playlist.DownloadableTrackCount,
		&playlist.PendingTrackCount,
		&playlist.CompletedTrackCount, &playlist.FailedTrackCount, &playlist.DownloadSupported,
		&playlist.CreateTime, &playlist.UpdateTime)
	return playlist, err
}

func nullableText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
