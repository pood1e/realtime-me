package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

const playlistTrackColumns = `item.uid, item.playlist_uid, item.position, item.provider_id, item.external_track_id,
	item.title, item.artists, item.album, item.duration_ms, item.artwork_url, item.provider_url,
	item.playable, item.lyrics_available, item.download_status, COALESCE(item.local_track_uid, '')`

// ListPlaylistTracks lists tracks in provider order.
func (s *Store) ListPlaylistTracks(ctx context.Context, playlistUID string, pageSize int, pageToken string) (domain.PlaylistTrackPage, error) {
	if _, err := s.GetPlaylist(ctx, playlistUID); err != nil {
		return domain.PlaylistTrackPage{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := "SELECT " + playlistTrackColumns + " FROM music_playlist_tracks item WHERE item.playlist_uid = $1"
	arguments := []any{playlistUID}
	if pageToken != "" {
		cursor, err := decodeCursor(pageToken)
		if err != nil {
			return domain.PlaylistTrackPage{}, err
		}
		position, err := strconv.Atoi(cursor.name)
		if err != nil || position < 1 {
			return domain.PlaylistTrackPage{}, fmt.Errorf("%w: malformed playlist track token", domain.ErrInvalidArgument)
		}
		arguments = append(arguments, position, cursor.uid)
		query += " AND (item.position, item.uid) > ($2, $3)"
	}
	arguments = append(arguments, pageSize+1)
	query += fmt.Sprintf(" ORDER BY item.position, item.uid LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.PlaylistTrackPage{}, fmt.Errorf("list playlist tracks: %w", err)
	}
	defer rows.Close()
	page := domain.PlaylistTrackPage{}
	for rows.Next() {
		track, err := scanPlaylistTrack(rows)
		if err != nil {
			return domain.PlaylistTrackPage{}, fmt.Errorf("scan playlist track: %w", err)
		}
		page.Tracks = append(page.Tracks, track)
	}
	if len(page.Tracks) > pageSize {
		last := page.Tracks[pageSize-1]
		page.Tracks = page.Tracks[:pageSize]
		page.NextPageToken = encodeCursor(strconv.Itoa(last.Position), last.UID)
	}
	return page, rows.Err()
}

// QueuePlaylistDownload enqueues every missing or failed direct-audio track.
func (s *Store) QueuePlaylistDownload(ctx context.Context, playlistUID string) (domain.Playlist, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("begin playlist download: %w", err)
	}
	defer tx.Rollback(ctx)
	var supported bool
	if err := tx.QueryRow(ctx, "SELECT download_supported FROM music_playlists WHERE uid = $1 FOR UPDATE", playlistUID).Scan(&supported); errors.Is(err, pgx.ErrNoRows) {
		return domain.Playlist{}, fmt.Errorf("%w: playlist", domain.ErrNotFound)
	} else if err != nil {
		return domain.Playlist{}, fmt.Errorf("lock playlist: %w", err)
	}
	if !supported {
		return domain.Playlist{}, fmt.Errorf("%w: playlist source does not expose downloadable audio", domain.ErrConflict)
	}
	itemUIDs, err := lockPlaylistDownloads(ctx, tx, playlistUID)
	if err != nil {
		return domain.Playlist{}, err
	}
	for _, uid := range itemUIDs {
		if _, err := tx.Exec(ctx, "UPDATE music_playlist_tracks SET download_status = 'pending' WHERE uid = $1", uid); err != nil {
			return domain.Playlist{}, fmt.Errorf("mark playlist download pending: %w", err)
		}
		if err := enqueueJob(ctx, tx, domain.ProcessingJobMusicDownload, uid); err != nil {
			return domain.Playlist{}, err
		}
	}
	if _, err := tx.Exec(ctx, "UPDATE music_playlists SET update_time = now() WHERE uid = $1", playlistUID); err != nil {
		return domain.Playlist{}, fmt.Errorf("touch playlist: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Playlist{}, fmt.Errorf("commit playlist download: %w", err)
	}
	return s.GetPlaylist(ctx, playlistUID)
}

func lockPlaylistDownloads(ctx context.Context, tx pgx.Tx, playlistUID string) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT uid FROM music_playlist_tracks
		WHERE playlist_uid = $1 AND playable AND (download_status <> 'completed' OR local_track_uid IS NULL)
		ORDER BY position FOR UPDATE`, playlistUID)
	if err != nil {
		return nil, fmt.Errorf("select playlist downloads: %w", err)
	}
	defer rows.Close()
	var itemUIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan playlist download: %w", err)
		}
		itemUIDs = append(itemUIDs, uid)
	}
	return itemUIDs, rows.Err()
}

func scanPlaylistTrack(row rowScanner) (domain.PlaylistTrack, error) {
	var item domain.PlaylistTrack
	var durationMS int64
	err := row.Scan(&item.UID, &item.PlaylistUID, &item.Position, &item.Track.Provider, &item.Track.TrackID,
		&item.Track.Title, &item.Track.Artists, &item.Track.Album, &durationMS, &item.Track.ArtworkURL,
		&item.Track.ProviderURL, &item.Track.Playable, &item.Track.LyricsAvailable, &item.DownloadStatus,
		&item.LocalTrackUID)
	item.Track.Duration = time.Duration(durationMS) * time.Millisecond
	return item, err
}
