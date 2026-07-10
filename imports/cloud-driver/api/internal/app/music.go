package app

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

// MusicService manages the private music application independently of the drive.
type MusicService struct {
	store    domain.MusicStore
	contents domain.ContentStore
	content  *ContentService
	files    ContentFiles
	clock    domain.Clock
}

// NewMusicService constructs the music application service.
func NewMusicService(store domain.MusicStore, contents domain.ContentStore, content *ContentService, files ContentFiles, clock domain.Clock) *MusicService {
	return &MusicService{store: store, contents: contents, content: content, files: files, clock: clock}
}

func (s *MusicService) Get(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.GetTrack(ctx, uid, false)
}

func (s *MusicService) List(ctx context.Context, query, album, artist string, favorites, trashed bool, pageSize int, pageToken string) (domain.TrackPage, error) {
	return s.store.ListTracks(ctx, strings.TrimSpace(query), strings.TrimSpace(album), strings.TrimSpace(artist), favorites, trashed, pageSize, pageToken)
}

func (s *MusicService) Import(ctx context.Context, uploadUID string) (domain.Track, error) {
	upload, sealed, unlock, err := s.content.SealForClaim(ctx, uploadUID)
	if err != nil {
		return domain.Track{}, err
	}
	defer unlock()
	track, err := s.store.ImportTrack(ctx, uploadUID, sealed)
	if err != nil {
		return domain.Track{}, err
	}
	if upload.Status != domain.UploadStatusClaimed {
		s.content.FinishClaim(ctx, uploadUID)
	}
	return track, nil
}

func (s *MusicService) SetFavorite(ctx context.Context, uid string, favorite bool) (domain.Track, error) {
	return s.store.SetTrackFavorite(ctx, uid, favorite)
}

func (s *MusicService) Trash(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.TrashTrack(ctx, uid)
}

func (s *MusicService) Restore(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.RestoreTrack(ctx, uid)
}

func (s *MusicService) Purge(ctx context.Context, uid string) error {
	if err := s.store.PurgeTrack(ctx, uid); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *MusicService) EmptyTrash(ctx context.Context) error {
	if err := s.store.EmptyTrackTrash(ctx); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *MusicService) RetryProcessing(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.QueueTrackProcessing(ctx, uid)
}

func (s *MusicService) ListAlbums(ctx context.Context, query string) ([]domain.Album, error) {
	return s.store.ListAlbums(ctx, strings.TrimSpace(query))
}

func (s *MusicService) ListArtists(ctx context.Context, query string) ([]domain.Artist, error) {
	return s.store.ListArtists(ctx, strings.TrimSpace(query))
}

func (s *MusicService) RecordPlayback(ctx context.Context, trackUID string) (domain.PlaybackEntry, error) {
	return s.store.RecordPlayback(ctx, domain.PlaybackEntry{UID: uuid.NewString(), Track: domain.Track{UID: trackUID}, PlayTime: s.clock.Now().UTC()})
}

func (s *MusicService) ListPlaybackHistory(ctx context.Context, pageSize int, pageToken string) (domain.PlaybackPage, error) {
	return s.store.ListPlaybackHistory(ctx, pageSize, pageToken)
}

func (s *MusicService) OpenContent(ctx context.Context, uid string) (*os.File, domain.Track, error) {
	track, err := s.store.GetTrack(ctx, uid, false)
	if err != nil {
		return nil, domain.Track{}, err
	}
	content, err := s.contents.GetContent(ctx, track.ContentUID)
	if err != nil {
		return nil, domain.Track{}, err
	}
	file, err := s.files.Open(ctx, content.StorageKey)
	return file, track, err
}

func (s *MusicService) OpenArtwork(ctx context.Context, uid string) (*os.File, domain.Track, error) {
	track, err := s.store.GetTrack(ctx, uid, false)
	if err != nil {
		return nil, domain.Track{}, err
	}
	if track.ArtworkStorageKey == "" {
		return nil, domain.Track{}, domainNotFound("track artwork")
	}
	file, err := s.files.Open(ctx, track.ArtworkStorageKey)
	return file, track, err
}
