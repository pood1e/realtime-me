package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
)

func (a *SpotifyAdapter) importSpotifyPlaylist(ctx context.Context, rawCredentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	playlistID, err := parseSpotifyPlaylistID(source)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	credentials, err := a.spotifyCredentials(ctx, rawCredentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	playlist, err := a.client.GetPlaylist(ctx, credentials, playlistID)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, mapProviderError(err)
	}
	tracks := make([]domain.PlayableTrack, 0, len(playlist.Tracks))
	for _, track := range playlist.Tracks {
		tracks = append(tracks, spotifyPlayableTrack(track))
	}
	updated, err := json.Marshal(credentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, fmt.Errorf("%w: encode Spotify credentials", domain.ErrUnavailable)
	}
	return domain.ProviderPlaylist{
		ExternalID: playlist.ID, DisplayName: playlist.Name, ArtworkURL: playlist.ArtworkURL,
		ProviderURL: playlist.SpotifyURL, Tracks: tracks,
	}, updated, nil
}

func parseSpotifyPlaylistID(source string) (string, error) {
	source = strings.TrimSpace(source)
	if strings.HasPrefix(source, "spotify:playlist:") {
		source = strings.TrimPrefix(source, "spotify:playlist:")
	}
	if validSpotifySourceID(source) {
		return source, nil
	}
	parsed, err := url.ParseRequestURI(source)
	if err != nil || parsed.Scheme != "https" || strings.ToLower(parsed.Hostname()) != "open.spotify.com" {
		return "", fmt.Errorf("%w: invalid Spotify playlist source", domain.ErrInvalidArgument)
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) == 2 && segments[0] == "playlist" && validSpotifySourceID(segments[1]) {
		return segments[1], nil
	}
	return "", fmt.Errorf("%w: Spotify playlist ID is missing", domain.ErrInvalidArgument)
}

func validSpotifySourceID(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') {
			continue
		}
		return false
	}
	return true
}
