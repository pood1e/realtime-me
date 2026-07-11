package provider

import (
	"context"
	"fmt"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/spotify"
)

// SpotifyAdapter is the compile-time official Spotify plugin.
type SpotifyAdapter struct {
	client *spotify.Client
}

var (
	_ domain.MusicProviderAdapter        = (*SpotifyAdapter)(nil)
	_ domain.MusicLoginStarter           = (*SpotifyAdapter)(nil)
	_ domain.MusicRedirectLoginCompleter = (*SpotifyAdapter)(nil)
	_ domain.MusicCatalogSearcher        = (*SpotifyAdapter)(nil)
	_ domain.MusicPlaybackResolver       = (*SpotifyAdapter)(nil)
	_ domain.MusicPlaybackTokenProvider  = (*SpotifyAdapter)(nil)
	_ domain.MusicPlaylistImporter       = (*SpotifyAdapter)(nil)
)

func newSpotifyAdapter(config RegistryConfig) (*SpotifyAdapter, error) {
	adapter := &SpotifyAdapter{}
	if config.SpotifyClientID == "" && config.SpotifyClientSecret == "" {
		return adapter, nil
	}
	client, err := spotify.New(spotify.Config{
		ClientID: config.SpotifyClientID, ClientSecret: config.SpotifyClientSecret, RedirectURI: config.SpotifyRedirectURI,
	})
	if err != nil {
		return nil, err
	}
	adapter.client = client
	return adapter, nil
}

func (*SpotifyAdapter) Provider() domain.MusicProvider { return domain.MusicProviderSpotify }

func (*SpotifyAdapter) DisplayName() string { return "Spotify" }

func (a *SpotifyAdapter) Configured() bool { return a.client != nil }

func (a *SpotifyAdapter) BeginLogin(_ context.Context) (domain.ProviderLoginChallenge, error) {
	return a.beginSpotifyLogin()
}

func (a *SpotifyAdapter) CompleteRedirectLogin(ctx context.Context, code string, state []byte) (domain.ProviderAccount, error) {
	if !a.Configured() {
		return domain.ProviderAccount{}, fmt.Errorf("%w: Spotify is not configured", domain.ErrConflict)
	}
	return a.completeSpotifyLogin(ctx, code, state)
}

func (a *SpotifyAdapter) Search(ctx context.Context, credentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	return a.searchSpotify(ctx, credentials, query, pageSize, pageToken)
}

func (a *SpotifyAdapter) ResolvePlayback(_ context.Context, credentials []byte, trackID string, _ domain.PlaybackQuality) (domain.PlaybackDescriptor, []byte, error) {
	return resolveSpotifyPlayback(trackID, credentials)
}

func (a *SpotifyAdapter) PlaybackToken(ctx context.Context, credentials []byte) (domain.ProviderPlaybackToken, []byte, error) {
	return a.spotifyPlaybackToken(ctx, credentials)
}

func (a *SpotifyAdapter) ImportPlaylist(ctx context.Context, credentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	return a.importSpotifyPlaylist(ctx, credentials, source)
}
