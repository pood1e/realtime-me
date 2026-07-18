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

const (
	trackColumns = `track.uid, track.content_uid, track.title, track.artists, track.album, track.album_artist,
		track.track_number, track.disc_number, track.year, track.duration_ms, track.original_file_name,
		content.content_type, content.size_bytes, COALESCE(art.storage_key, ''), track.favorite,
		track.processing_status, track.create_time, track.update_time, track.delete_time`
	trackFrom = `tracks track JOIN content_objects content ON content.uid = track.content_uid
		LEFT JOIN content_artifacts art ON art.content_uid = track.content_uid
		AND art.kind = 'track_artwork' AND art.variant = 'default'`
)

// GetTrack returns one audio catalog entry.
func (s *Store) GetTrack(ctx context.Context, uid string, includeTrashed bool) (domain.Track, error) {
	query := "SELECT " + trackColumns + " FROM " + trackFrom + " WHERE track.uid = $1"
	if !includeTrashed {
		query += " AND track.delete_time IS NULL"
	}
	track, err := scanTrack(s.pool.QueryRow(ctx, query, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Track{}, fmt.Errorf("%w: track", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Track{}, fmt.Errorf("get track: %w", err)
	}
	return track, nil
}

// GetTrackBySource returns one visible local copy of an external provider track.
func (s *Store) GetTrackBySource(ctx context.Context, provider domain.MusicProvider, trackID string) (domain.Track, error) {
	track, err := scanTrack(s.pool.QueryRow(ctx, "SELECT "+trackColumns+" FROM "+trackFrom+
		` JOIN music_track_sources source ON source.track_uid = track.uid
		WHERE source.provider_id = $1 AND source.external_track_id = $2 AND track.delete_time IS NULL`,
		provider, trackID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Track{}, fmt.Errorf("%w: source track", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Track{}, fmt.Errorf("get source track: %w", err)
	}
	return track, nil
}

// ListTracks lists audio entries with catalog filters.
func (s *Store) ListTracks(ctx context.Context, filter domain.TrackListQuery) (domain.TrackPage, error) {
	cursor, err := decodeCursor(filter.PageToken)
	if err != nil {
		return domain.TrackPage{}, err
	}
	pageSize := normalizePageSize(filter.PageSize)
	query := "SELECT " + trackColumns + " FROM " + trackFrom
	arguments := []any{}
	conditions := []string{"track.delete_time IS " + map[bool]string{true: "NOT NULL", false: "NULL"}[filter.Trashed]}
	addCondition := func(value, expression string) {
		if value == "" {
			return
		}
		arguments = append(arguments, value)
		conditions = append(conditions, fmt.Sprintf(expression, len(arguments)))
	}
	addCondition(filter.Query, "(track.title ILIKE '%%' || $%[1]d || '%%' OR track.album ILIKE '%%' || $%[1]d || '%%' OR array_to_string(track.artists, ' ') ILIKE '%%' || $%[1]d || '%%')")
	addCondition(filter.Album, "track.album = $%d")
	addCondition(filter.Artist, "$%d = ANY(track.artists)")
	if filter.Favorites {
		conditions = append(conditions, "track.favorite")
	}
	if cursor != nil {
		arguments = append(arguments, cursor.name, cursor.uid)
		conditions = append(conditions, fmt.Sprintf("(track.title, track.uid) > ($%d, $%d)", len(arguments)-1, len(arguments)))
	}
	arguments = append(arguments, pageSize+1)
	query += " WHERE " + strings.Join(conditions, " AND ") + fmt.Sprintf(" ORDER BY track.title, track.uid LIMIT $%d", len(arguments))
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.TrackPage{}, fmt.Errorf("list tracks: %w", err)
	}
	defer rows.Close()
	page := domain.TrackPage{}
	for rows.Next() {
		track, err := scanTrack(rows)
		if err != nil {
			return domain.TrackPage{}, fmt.Errorf("scan track: %w", err)
		}
		page.Tracks = append(page.Tracks, track)
	}
	if len(page.Tracks) > pageSize {
		last := page.Tracks[pageSize-1]
		page.Tracks = page.Tracks[:pageSize]
		page.NextPageToken = encodeCursor(last.Title, last.UID)
	}
	return page, rows.Err()
}

// ImportTrack claims a supported audio upload.
func (s *Store) ImportTrack(ctx context.Context, uploadUID string, sealed domain.SealedContent) (domain.Track, error) {
	if !supportedAudioType(sealed.ContentType) {
		return domain.Track{}, fmt.Errorf("%w: unsupported audio type", domain.ErrInvalidArgument)
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Track{}, fmt.Errorf("begin track import: %w", err)
	}
	defer tx.Rollback(ctx)
	upload, err := lockCompleteUpload(ctx, tx, uploadUID)
	if err != nil {
		return domain.Track{}, err
	}
	if upload.Status == domain.UploadStatusClaimed {
		if err := tx.Commit(ctx); err != nil {
			return domain.Track{}, fmt.Errorf("commit repeated track import: %w", err)
		}
		return s.GetTrack(ctx, upload.ClaimedUID, false)
	}
	content, err := contentForUpload(ctx, tx, upload, sealed)
	if err != nil {
		return domain.Track{}, err
	}
	var existingUID string
	err = tx.QueryRow(ctx, "SELECT uid FROM tracks WHERE content_uid = $1", content.UID).Scan(&existingUID)
	if err == nil {
		if err := markUploadClaimed(ctx, tx, uploadUID, existingUID); err != nil {
			return domain.Track{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return domain.Track{}, fmt.Errorf("commit deduplicated track import: %w", err)
		}
		return s.GetTrack(ctx, existingUID, true)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Track{}, fmt.Errorf("find imported track: %w", err)
	}
	trackUID := uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO tracks
		(uid, content_uid, title, original_file_name, processing_status)
		VALUES ($1, $2, $3, $4, 'pending')`, trackUID, content.UID, displayName(upload.FileName), upload.FileName); err != nil {
		return domain.Track{}, fmt.Errorf("create track: %w", err)
	}
	if err := enqueueJob(ctx, tx, domain.ProcessingJobTrack, trackUID); err != nil {
		return domain.Track{}, err
	}
	if err := markUploadClaimed(ctx, tx, uploadUID, trackUID); err != nil {
		return domain.Track{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Track{}, fmt.Errorf("commit track import: %w", err)
	}
	return s.GetTrack(ctx, trackUID, false)
}

// SetTrackFavorite changes the owner favorite state.
func (s *Store) SetTrackFavorite(ctx context.Context, uid string, favorite bool) (domain.Track, error) {
	command, err := s.pool.Exec(ctx, `UPDATE tracks SET favorite = $2, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid, favorite)
	if err != nil {
		return domain.Track{}, fmt.Errorf("set track favorite: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Track{}, fmt.Errorf("%w: track", domain.ErrNotFound)
	}
	return s.GetTrack(ctx, uid, false)
}

// TrashTrack moves a track to trash.
func (s *Store) TrashTrack(ctx context.Context, uid string) (domain.Track, error) {
	command, err := s.pool.Exec(ctx, `UPDATE tracks SET delete_time = now(), update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid)
	if err != nil {
		return domain.Track{}, fmt.Errorf("trash track: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Track{}, fmt.Errorf("%w: track", domain.ErrNotFound)
	}
	return s.GetTrack(ctx, uid, true)
}

// RestoreTrack restores a trashed track.
func (s *Store) RestoreTrack(ctx context.Context, uid string) (domain.Track, error) {
	command, err := s.pool.Exec(ctx, `UPDATE tracks SET delete_time = NULL, update_time = now()
		WHERE uid = $1 AND delete_time IS NOT NULL`, uid)
	if err != nil {
		return domain.Track{}, fmt.Errorf("restore track: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Track{}, fmt.Errorf("%w: trashed track", domain.ErrNotFound)
	}
	return s.GetTrack(ctx, uid, false)
}

// PurgeTrack permanently deletes one trashed track.
func (s *Store) PurgeTrack(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM tracks WHERE uid = $1 AND delete_time IS NOT NULL", uid)
	if err != nil {
		return fmt.Errorf("purge track: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: trashed track", domain.ErrNotFound)
	}
	return nil
}

// EmptyTrackTrash permanently removes every trashed track.
func (s *Store) EmptyTrackTrash(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM tracks WHERE delete_time IS NOT NULL")
	return wrapDatabaseError("empty track trash", err)
}

// PurgeTrashedTracks removes tracks past retention.
func (s *Store) PurgeTrashedTracks(ctx context.Context, cutoff time.Time) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM tracks WHERE delete_time <= $1", cutoff)
	return wrapDatabaseError("purge expired tracks", err)
}

// QueueTrackProcessing retries tag extraction.
func (s *Store) QueueTrackProcessing(ctx context.Context, uid string) (domain.Track, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Track{}, fmt.Errorf("begin track processing retry: %w", err)
	}
	defer tx.Rollback(ctx)
	command, err := tx.Exec(ctx, `UPDATE tracks SET processing_status = 'pending', update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid)
	if err != nil {
		return domain.Track{}, fmt.Errorf("mark track pending: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Track{}, fmt.Errorf("%w: track", domain.ErrNotFound)
	}
	if err := enqueueJob(ctx, tx, domain.ProcessingJobTrack, uid); err != nil {
		return domain.Track{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Track{}, fmt.Errorf("commit track retry: %w", err)
	}
	return s.GetTrack(ctx, uid, false)
}

// ListAlbums returns visible album summaries.
