package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/spotify"
)

func resolveSpotifyPlayback(trackID string, credentials []byte) (domain.PlaybackDescriptor, []byte, error) {
	trackID = strings.TrimSpace(trackID)
	if len(trackID) < 1 || len(trackID) > 64 || strings.ContainsAny(trackID, ":/ ") {
		return domain.PlaybackDescriptor{}, nil, fmt.Errorf("%w: invalid Spotify track ID", domain.ErrInvalidArgument)
	}
	return domain.PlaybackDescriptor{
		Provider: domain.MusicProviderSpotify, SDKID: "spotify_web_playback", ResourceURI: "spotify:track:" + trackID,
	}, credentials, nil
}

func (a *SpotifyAdapter) spotifyPlaybackToken(ctx context.Context, rawCredentials []byte) (domain.ProviderPlaybackToken, []byte, error) {
	credentials, err := decodeSpotifyCredentials(rawCredentials)
	if err != nil {
		return domain.ProviderPlaybackToken{}, nil, err
	}
	token, updatedCredentials, err := a.client.WebPlaybackToken(ctx, credentials)
	if err != nil {
		return domain.ProviderPlaybackToken{}, nil, mapProviderError(err)
	}
	updated, err := json.Marshal(updatedCredentials)
	if err != nil {
		return domain.ProviderPlaybackToken{}, nil, fmt.Errorf("%w: encode Spotify credentials", domain.ErrUnavailable)
	}
	return domain.ProviderPlaybackToken{AccessToken: token.AccessToken, ExpireTime: token.ExpiresAt}, updated, nil
}

func (a *SpotifyAdapter) spotifyCredentials(ctx context.Context, encoded []byte) (spotify.Credentials, error) {
	credentials, err := decodeSpotifyCredentials(encoded)
	if err != nil {
		return spotify.Credentials{}, err
	}
	if time.Now().Add(2 * time.Minute).Before(credentials.ExpiresAt) {
		return credentials, nil
	}
	credentials, err = a.client.Refresh(ctx, credentials)
	if err != nil {
		return spotify.Credentials{}, mapProviderError(err)
	}
	return credentials, nil
}

func decodeSpotifyCredentials(encoded []byte) (spotify.Credentials, error) {
	var credentials spotify.Credentials
	if json.Unmarshal(encoded, &credentials) != nil || credentials.RefreshToken == "" {
		return spotify.Credentials{}, fmt.Errorf("%w: invalid Spotify credentials", domain.ErrProviderReconnectRequired)
	}
	return credentials, nil
}
