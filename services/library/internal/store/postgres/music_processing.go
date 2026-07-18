package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func (s *Store) GetTrackForProcessing(ctx context.Context, uid string) (domain.Track, domain.ContentObject, error) {
	track, err := s.GetTrack(ctx, uid, true)
	if err != nil {
		return domain.Track{}, domain.ContentObject{}, err
	}
	content, err := s.GetContent(ctx, track.ContentUID)
	return track, content, err
}

// CompleteTrackProcessing persists extracted tags and optional artwork.
func (s *Store) CompleteTrackProcessing(ctx context.Context, job domain.ProcessingJob, track domain.Track, artwork *domain.Artifact) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin track processing completion: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := lockProcessingJobLease(ctx, tx, job); err != nil {
		return err
	}
	if artwork != nil {
		if err := upsertArtifact(ctx, tx, *artwork); err != nil {
			return err
		}
	} else if err := queueProviderArtworkIfMissing(ctx, tx, job.ResourceUID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE tracks SET title = CASE WHEN $2 <> '' THEN $2 ELSE title END,
		artists = CASE WHEN cardinality($3::text[]) > 0 THEN $3 ELSE artists END,
		album = $4, album_artist = $5, track_number = $6, disc_number = $7, year = $8,
		duration_ms = $9, processing_status = 'ready', update_time = now() WHERE uid = $1`,
		job.ResourceUID, track.Title, track.Artists, track.Album, track.AlbumArtist, track.TrackNumber,
		track.DiscNumber, track.Year, track.Duration.Milliseconds())
	if err != nil {
		return fmt.Errorf("complete track processing: %w", err)
	}
	return tx.Commit(ctx)
}

func scanTrack(row rowScanner) (domain.Track, error) {
	var track domain.Track
	var durationMS int64
	var status string
	var deleteTime pgtype.Timestamptz
	if err := row.Scan(&track.UID, &track.ContentUID, &track.Title, &track.Artists, &track.Album,
		&track.AlbumArtist, &track.TrackNumber, &track.DiscNumber, &track.Year, &durationMS,
		&track.OriginalFileName, &track.ContentType, &track.SizeBytes, &track.ArtworkStorageKey,
		&track.Favorite, &status, &track.CreateTime, &track.UpdateTime, &deleteTime); err != nil {
		return domain.Track{}, err
	}
	track.Duration = time.Duration(durationMS) * time.Millisecond
	track.ProcessingStatus = domain.ProcessingStatus(status)
	if deleteTime.Valid {
		value := deleteTime.Time.UTC()
		track.DeleteTime = &value
	}
	return track, nil
}

func supportedAudioType(contentType string) bool {
	switch contentType {
	case "audio/mpeg", "audio/mp4", "audio/aac", "audio/flac", "audio/ogg", "audio/opus", "audio/wav", "audio/x-wav":
		return true
	default:
		return false
	}
}
