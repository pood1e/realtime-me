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

// ImportPlaylist resolves and stores one provider playlist snapshot.
func (s *MusicService) ImportPlaylist(ctx context.Context, provider domain.MusicProvider, source string) (domain.Playlist, error) {
	source = strings.TrimSpace(source)
	if provider == "" || provider == domain.MusicProviderLocal || source == "" || len(source) > maximumPlaylistSourceLength {
		return domain.Playlist{}, fmt.Errorf("%w: provider and playlist source are required", domain.ErrInvalidArgument)
	}
	connection, credentials, err := s.providerCredentials(ctx, provider)
	if err != nil {
		return domain.Playlist{}, err
	}
	adapter, err := s.providerAdapter(provider)
	if err != nil {
		return domain.Playlist{}, err
	}
	importer, supported := adapter.(domain.MusicPlaylistImporter)
	if !supported {
		return domain.Playlist{}, fmt.Errorf("%w: provider does not expose playlists", domain.ErrConflict)
	}
	providerPlaylist, updated, err := importer.ImportPlaylist(ctx, credentials, source)
	if err != nil {
		s.markProviderFailure(ctx, connection, err)
		return domain.Playlist{}, err
	}
	if err := validateProviderPlaylist(provider, providerPlaylist); err != nil {
		return domain.Playlist{}, err
	}
	if err := s.persistCredentialUpdate(ctx, connection, credentials, updated); err != nil {
		return domain.Playlist{}, err
	}
	now := s.clock.Now().UTC()
	_, downloadSupported := adapter.(domain.MusicTrackDownloader)
	return s.store.ImportPlaylist(ctx, domain.Playlist{
		UID: uuid.NewString(), Provider: provider, ExternalID: strings.TrimSpace(providerPlaylist.ExternalID),
		DisplayName: strings.TrimSpace(providerPlaylist.DisplayName), ArtworkURL: strings.TrimSpace(providerPlaylist.ArtworkURL),
		ProviderURL: strings.TrimSpace(providerPlaylist.ProviderURL), DownloadSupported: downloadSupported,
		CreateTime: now, UpdateTime: now,
	}, providerPlaylist.Tracks)
}

// GetPlaylist returns one imported playlist.
func (s *MusicService) GetPlaylist(ctx context.Context, uid string) (domain.Playlist, error) {
	return s.store.GetPlaylist(ctx, strings.TrimSpace(uid))
}

// ListPlaylists lists imported playlists.
func (s *MusicService) ListPlaylists(ctx context.Context, pageSize int, pageToken string) (domain.PlaylistPage, error) {
	return s.store.ListPlaylists(ctx, pageSize, pageToken)
}

// ListPlaylistTracks lists one playlist in provider order.
func (s *MusicService) ListPlaylistTracks(ctx context.Context, uid string, pageSize int, pageToken string) (domain.PlaylistTrackPage, error) {
	return s.store.ListPlaylistTracks(ctx, strings.TrimSpace(uid), pageSize, pageToken)
}

// DownloadPlaylist queues every locally missing direct-audio track.
func (s *MusicService) DownloadPlaylist(ctx context.Context, uid string) (domain.Playlist, error) {
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
func (s *MusicService) DeletePlaylist(ctx context.Context, uid string) error {
	return s.store.DeletePlaylist(ctx, strings.TrimSpace(uid))
}

func validateProviderPlaylist(provider domain.MusicProvider, playlist domain.ProviderPlaylist) error {
	if strings.TrimSpace(playlist.ExternalID) == "" || strings.TrimSpace(playlist.DisplayName) == "" ||
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
