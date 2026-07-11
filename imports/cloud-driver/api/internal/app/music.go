package app

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

// MusicSuite groups independently addressable music application services.
type MusicSuite struct {
	Library   *MusicLibraryService
	Providers *MusicProviderService
	Playlists *MusicPlaylistService
}

// MusicLibraryService owns the local catalog and playback history.
type MusicLibraryService struct {
	store    domain.MusicLibraryStore
	contents domain.ContentStore
	content  *ContentService
	files    ContentReader
	clock    domain.Clock
	tracks   *musicTrackValidator
}

// MusicProviderService owns external account, search, and playback operations.
type MusicProviderService struct {
	*musicProviderSession
	store  domain.MusicLibraryStore
	tracks *musicTrackValidator
}

// MusicPlaylistService owns imported playlist snapshots and downloads.
type MusicPlaylistService struct {
	*musicProviderSession
	store  domain.MusicPlaylistStore
	tracks *musicTrackValidator
}

type musicProviderSession struct {
	providerStore domain.MusicProviderStore
	providers     domain.MusicProviderRegistry
	credentials   CredentialProtector
	clock         domain.Clock
}

// CredentialProtector encrypts provider state before persistence.
type CredentialProtector interface {
	Seal(string, []byte) ([]byte, error)
	Open(string, []byte) ([]byte, error)
}

// MusicProviderDependencies contains external provider infrastructure.
type MusicProviderDependencies struct {
	Store       domain.MusicProviderStore
	Registry    domain.MusicProviderRegistry
	Credentials CredentialProtector
}

// NewMusicSuite constructs the music application around one shared provider session.
func NewMusicSuite(store domain.MusicStore, contents domain.ContentStore, content *ContentService, files ContentReader, clock domain.Clock, providers MusicProviderDependencies) *MusicSuite {
	session := newMusicProviderSession(clock, providers)
	tracks := &musicTrackValidator{store: store, providers: providers.Registry}
	return &MusicSuite{
		Library: &MusicLibraryService{
			store: store, contents: contents, content: content, files: files, clock: clock, tracks: tracks,
		},
		Providers: &MusicProviderService{musicProviderSession: session, store: store, tracks: tracks},
		Playlists: &MusicPlaylistService{musicProviderSession: session, store: store, tracks: tracks},
	}
}

// NewMusicPlaylistService constructs the worker-safe playlist import service.
func NewMusicPlaylistService(store domain.MusicStore, clock domain.Clock, providers MusicProviderDependencies) *MusicPlaylistService {
	return &MusicPlaylistService{
		musicProviderSession: newMusicProviderSession(clock, providers), store: store,
		tracks: &musicTrackValidator{store: store, providers: providers.Registry},
	}
}

func newMusicProviderSession(clock domain.Clock, providers MusicProviderDependencies) *musicProviderSession {
	return &musicProviderSession{
		providerStore: providers.Store, providers: providers.Registry,
		credentials: providers.Credentials, clock: clock,
	}
}

func (s *MusicLibraryService) Get(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.GetTrack(ctx, uid, false)
}

func (s *MusicLibraryService) List(ctx context.Context, query, album, artist string, favorites, trashed bool, pageSize int, pageToken string) (domain.TrackPage, error) {
	return s.store.ListTracks(ctx, domain.TrackListQuery{
		Query: strings.TrimSpace(query), Album: strings.TrimSpace(album), Artist: strings.TrimSpace(artist),
		Favorites: favorites, Trashed: trashed, PageSize: pageSize, PageToken: pageToken,
	})
}

func (s *MusicLibraryService) Import(ctx context.Context, uploadUID string) (domain.Track, error) {
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
		_ = s.content.FinishClaim(ctx, uploadUID)
	}
	return track, nil
}

func (s *MusicLibraryService) SetFavorite(ctx context.Context, uid string, favorite bool) (domain.Track, error) {
	return s.store.SetTrackFavorite(ctx, uid, favorite)
}

func (s *MusicLibraryService) Trash(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.TrashTrack(ctx, uid)
}

func (s *MusicLibraryService) Restore(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.RestoreTrack(ctx, uid)
}

func (s *MusicLibraryService) Purge(ctx context.Context, uid string) error {
	if err := s.store.PurgeTrack(ctx, uid); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *MusicLibraryService) EmptyTrash(ctx context.Context) error {
	if err := s.store.EmptyTrackTrash(ctx); err != nil {
		return err
	}
	return s.content.CollectGarbage(ctx)
}

func (s *MusicLibraryService) RetryProcessing(ctx context.Context, uid string) (domain.Track, error) {
	return s.store.QueueTrackProcessing(ctx, uid)
}

func (s *MusicLibraryService) RecordPlayback(ctx context.Context, track domain.PlayableTrack) (domain.PlaybackEntry, error) {
	validated, err := s.tracks.Validate(ctx, track)
	if err != nil {
		return domain.PlaybackEntry{}, err
	}
	return s.store.RecordPlayback(ctx, domain.PlaybackEntry{UID: uuid.NewString(), Track: validated, PlayTime: s.clock.Now().UTC()})
}

func (s *MusicLibraryService) ListPlaybackHistory(ctx context.Context, pageSize int, pageToken string) (domain.PlaybackPage, error) {
	return s.store.ListPlaybackHistory(ctx, pageSize, pageToken)
}

func (s *MusicLibraryService) OpenContent(ctx context.Context, uid string) (*os.File, domain.Track, error) {
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

func (s *MusicLibraryService) OpenArtwork(ctx context.Context, uid string) (*os.File, domain.Track, error) {
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
