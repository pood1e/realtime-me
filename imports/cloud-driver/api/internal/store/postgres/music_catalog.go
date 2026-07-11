package postgres

import (
	"context"
	"fmt"
	"time"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *Store) ListAlbums(ctx context.Context, query string) ([]domain.Album, error) {
	rows, err := s.pool.Query(ctx, `SELECT md5(album || E'\000' || album_artist), album, album_artist,
		MAX(year), COUNT(*), COALESCE(MAX(art.storage_key), '')
		FROM tracks track LEFT JOIN content_artifacts art ON art.content_uid = track.content_uid
		AND art.kind = 'track_artwork' AND art.variant = 'default'
		WHERE track.delete_time IS NULL AND album <> '' AND ($1 = '' OR album ILIKE '%' || $1 || '%')
		GROUP BY album, album_artist ORDER BY album, album_artist`, query)
	if err != nil {
		return nil, fmt.Errorf("list albums: %w", err)
	}
	defer rows.Close()
	var albums []domain.Album
	for rows.Next() {
		var album domain.Album
		if err := rows.Scan(&album.UID, &album.Title, &album.AlbumArtist, &album.Year, &album.TrackCount, &album.ArtworkStorageKey); err != nil {
			return nil, fmt.Errorf("scan album: %w", err)
		}
		albums = append(albums, album)
	}
	return albums, rows.Err()
}

// ListArtists returns visible display-artist summaries.
func (s *Store) ListArtists(ctx context.Context, query string) ([]domain.Artist, error) {
	rows, err := s.pool.Query(ctx, `SELECT md5(artist), artist, COUNT(*) FROM tracks track,
		LATERAL unnest(track.artists) artist WHERE track.delete_time IS NULL
		AND ($1 = '' OR artist ILIKE '%' || $1 || '%') GROUP BY artist ORDER BY artist`, query)
	if err != nil {
		return nil, fmt.Errorf("list artists: %w", err)
	}
	defer rows.Close()
	var artists []domain.Artist
	for rows.Next() {
		var artist domain.Artist
		if err := rows.Scan(&artist.UID, &artist.DisplayName, &artist.TrackCount); err != nil {
			return nil, fmt.Errorf("scan artist: %w", err)
		}
		artists = append(artists, artist)
	}
	return artists, rows.Err()
}

// RecordPlayback persists one meaningful playback event.
func (s *Store) RecordPlayback(ctx context.Context, entry domain.PlaybackEntry) (domain.PlaybackEntry, error) {
	var localTrackUID any
	if entry.Track.Provider == domain.MusicProviderLocal {
		localTrackUID = entry.Track.TrackID
	}
	err := s.pool.QueryRow(ctx, `INSERT INTO playback_history (uid, track_uid, play_time, provider,
		external_track_id, title, artists, album, duration_ms, artwork_url, provider_url, playable, lyrics_available)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING play_time`,
		entry.UID, localTrackUID, entry.PlayTime, entry.Track.Provider, entry.Track.TrackID, entry.Track.Title,
		entry.Track.Artists, entry.Track.Album, entry.Track.Duration.Milliseconds(), entry.Track.ArtworkURL,
		entry.Track.ProviderURL, entry.Track.Playable, entry.Track.LyricsAvailable).Scan(&entry.PlayTime)
	if err != nil {
		return domain.PlaybackEntry{}, fmt.Errorf("record playback: %w", err)
	}
	return entry, nil
}

// ListPlaybackHistory lists newest events first.
func (s *Store) ListPlaybackHistory(ctx context.Context, pageSize int, pageToken string) (domain.PlaybackPage, error) {
	pageSize = normalizePageSize(pageSize)
	query := `SELECT uid, play_time, provider, external_track_id, title, artists, album, duration_ms,
		artwork_url, provider_url, playable, lyrics_available FROM playback_history history`
	arguments := []any{}
	if pageToken != "" {
		cursor, err := decodeCursor(pageToken)
		if err != nil {
			return domain.PlaybackPage{}, err
		}
		playTime, err := time.Parse(time.RFC3339Nano, cursor.name)
		if err != nil {
			return domain.PlaybackPage{}, fmt.Errorf("%w: malformed history token", domain.ErrInvalidArgument)
		}
		arguments = append(arguments, playTime, cursor.uid)
		query += " WHERE (history.play_time, history.uid) < ($1, $2)"
	}
	arguments = append(arguments, pageSize+1)
	query += fmt.Sprintf(" ORDER BY history.play_time DESC, history.uid DESC LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.PlaybackPage{}, fmt.Errorf("list playback history: %w", err)
	}
	defer rows.Close()
	page := domain.PlaybackPage{}
	for rows.Next() {
		var entry domain.PlaybackEntry
		var durationMS int64
		if err := rows.Scan(&entry.UID, &entry.PlayTime, &entry.Track.Provider, &entry.Track.TrackID,
			&entry.Track.Title, &entry.Track.Artists, &entry.Track.Album, &durationMS, &entry.Track.ArtworkURL,
			&entry.Track.ProviderURL, &entry.Track.Playable, &entry.Track.LyricsAvailable); err != nil {
			return domain.PlaybackPage{}, fmt.Errorf("scan playback history: %w", err)
		}
		entry.Track.Duration = time.Duration(durationMS) * time.Millisecond
		page.Entries = append(page.Entries, entry)
	}
	if len(page.Entries) > pageSize {
		last := page.Entries[pageSize-1]
		page.Entries = page.Entries[:pageSize]
		page.NextPageToken = encodeCursor(last.PlayTime.Format(time.RFC3339Nano), last.UID)
	}
	return page, rows.Err()
}

// GetTrackForProcessing returns source metadata for the worker.
