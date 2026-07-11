package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	maximumPlaylistSourceLength = 2048
	maximumImportedPlaylistSize = 5000
)

// QueuePlaylistImport validates access and creates one durable import operation.
func (s *MusicPlaylistService) QueuePlaylistImport(ctx context.Context, provider domain.MusicProvider, source string) (domain.PlaylistImport, error) {
	source = strings.TrimSpace(source)
	if provider == "" || provider == domain.MusicProviderLocal || source == "" || len(source) > maximumPlaylistSourceLength {
		return domain.PlaylistImport{}, fmt.Errorf("%w: provider and playlist source are required", domain.ErrInvalidArgument)
	}
	if _, _, err := s.providerCredentials(ctx, provider); err != nil {
		return domain.PlaylistImport{}, err
	}
	adapter, err := s.providerAdapter(provider)
	if err != nil {
		return domain.PlaylistImport{}, err
	}
	if _, supported := adapter.(domain.MusicPlaylistImporter); !supported {
		return domain.PlaylistImport{}, fmt.Errorf("%w: provider does not expose playlists", domain.ErrConflict)
	}
	now := s.clock.Now().UTC()
	return s.store.QueuePlaylistImport(ctx, domain.PlaylistImport{
		UID: uuid.NewString(), Provider: provider, Source: source,
		Status: domain.PlaylistImportPending, CreateTime: now, UpdateTime: now,
	})
}

// GetPlaylistImport returns one durable import operation.
func (s *MusicPlaylistService) GetPlaylistImport(ctx context.Context, uid string) (domain.PlaylistImport, error) {
	return s.store.GetPlaylistImport(ctx, strings.TrimSpace(uid))
}

// ResolvePlaylistImport loads and normalizes provider data for the local worker.
func (s *MusicPlaylistService) ResolvePlaylistImport(ctx context.Context, operation domain.PlaylistImport) (domain.Playlist, []domain.PlayableTrack, error) {
	connection, credentials, err := s.providerCredentials(ctx, operation.Provider)
	if err != nil {
		return domain.Playlist{}, nil, err
	}
	adapter, err := s.providerAdapter(operation.Provider)
	if err != nil {
		return domain.Playlist{}, nil, err
	}
	importer, supported := adapter.(domain.MusicPlaylistImporter)
	if !supported {
		return domain.Playlist{}, nil, fmt.Errorf("%w: provider does not expose playlists", domain.ErrConflict)
	}
	providerPlaylist, updated, err := importer.ImportPlaylist(ctx, credentials, operation.Source)
	if err != nil {
		s.markProviderFailure(ctx, connection, err)
		return domain.Playlist{}, nil, err
	}
	if err := validateProviderPlaylist(operation.Provider, providerPlaylist); err != nil {
		return domain.Playlist{}, nil, err
	}
	normalizedTracks := make([]domain.PlayableTrack, 0, len(providerPlaylist.Tracks))
	for _, track := range providerPlaylist.Tracks {
		normalized, err := s.tracks.Validate(ctx, track)
		if err != nil {
			return domain.Playlist{}, nil, fmt.Errorf("%w: provider playlist contains invalid metadata", domain.ErrUnavailable)
		}
		normalizedTracks = append(normalizedTracks, normalized)
	}
	if err := s.persistCredentialUpdate(ctx, connection, credentials, updated); err != nil {
		return domain.Playlist{}, nil, err
	}
	now := s.clock.Now().UTC()
	_, downloadSupported := adapter.(domain.MusicTrackDownloader)
	playlist := domain.Playlist{
		UID: uuid.NewString(), Provider: operation.Provider,
		ExternalID:  truncateText(strings.TrimSpace(providerPlaylist.ExternalID), 512),
		DisplayName: truncateText(strings.TrimSpace(providerPlaylist.DisplayName), 512),
		ArtworkURL:  validatedProviderURL(providerPlaylist.ArtworkURL),
		ProviderURL: validatedProviderURL(providerPlaylist.ProviderURL), DownloadSupported: downloadSupported,
		CreateTime: now, UpdateTime: now,
	}
	return playlist, normalizedTracks, nil
}

// GetPlaylist returns one imported playlist.
func (s *MusicPlaylistService) GetPlaylist(ctx context.Context, uid string) (domain.Playlist, error) {
	return s.store.GetPlaylist(ctx, strings.TrimSpace(uid))
}

// ListPlaylists lists imported playlists.
func (s *MusicPlaylistService) ListPlaylists(ctx context.Context, pageSize int, pageToken string) (domain.PlaylistPage, error) {
	return s.store.ListPlaylists(ctx, pageSize, pageToken)
}

// ListPlaylistTracks lists one playlist in provider order.
func (s *MusicPlaylistService) ListPlaylistTracks(ctx context.Context, uid string, pageSize int, pageToken string) (domain.PlaylistTrackPage, error) {
	return s.store.ListPlaylistTracks(ctx, strings.TrimSpace(uid), pageSize, pageToken)
}

// DownloadPlaylist queues every locally missing direct-audio track.
func (s *MusicPlaylistService) DownloadPlaylist(ctx context.Context, uid string) (domain.Playlist, error) {
	playlist, err := s.store.GetPlaylist(ctx, strings.TrimSpace(uid))
	if err != nil {
		return domain.Playlist{}, err
	}
	adapter, err := s.providerAdapter(playlist.Provider)
	if err != nil {
		return domain.Playlist{}, err
	}
	if _, supported := adapter.(domain.MusicTrackDownloader); !supported {
		return domain.Playlist{}, fmt.Errorf("%w: provider does not expose downloadable audio", domain.ErrConflict)
	}
	if _, _, err := s.providerCredentials(ctx, playlist.Provider); err != nil {
		return domain.Playlist{}, err
	}
	return s.store.QueuePlaylistDownload(ctx, playlist.UID)
}

// DeletePlaylist removes an imported snapshot without deleting downloaded tracks.
func (s *MusicPlaylistService) DeletePlaylist(ctx context.Context, uid string) error {
	return s.store.DeletePlaylist(ctx, strings.TrimSpace(uid))
}

func validateProviderPlaylist(provider domain.MusicProvider, playlist domain.ProviderPlaylist) error {
	if strings.TrimSpace(playlist.ExternalID) == "" || len(playlist.ExternalID) > 512 ||
		strings.TrimSpace(playlist.DisplayName) == "" || len(playlist.DisplayName) > 512 ||
		len(playlist.Tracks) > maximumImportedPlaylistSize {
		return fmt.Errorf("%w: provider playlist is incomplete or too large", domain.ErrUnavailable)
	}
	for _, track := range playlist.Tracks {
		if track.Provider != provider || strings.TrimSpace(track.TrackID) == "" || strings.TrimSpace(track.Title) == "" ||
			track.Duration < 0 {
			return fmt.Errorf("%w: provider playlist contains an invalid track", domain.ErrUnavailable)
		}
	}
	return nil
}
