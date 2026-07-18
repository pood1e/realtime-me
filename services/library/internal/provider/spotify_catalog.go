package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/spotify"
)

func (a *SpotifyAdapter) searchSpotify(ctx context.Context, rawCredentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	credentials, err := a.spotifyCredentials(ctx, rawCredentials)
	if err != nil {
		return nil, "", nil, err
	}
	offset, err := parseOffsetToken(pageToken)
	if err != nil {
		return nil, "", nil, err
	}
	page, err := a.client.SearchTracks(ctx, credentials, spotify.TrackSearchRequest{Query: query, Limit: pageSize, Offset: offset})
	if err != nil {
		return nil, "", nil, mapProviderError(err)
	}
	tracks := make([]domain.PlayableTrack, 0, len(page.Items))
	for _, track := range page.Items {
		tracks = append(tracks, spotifyPlayableTrack(track))
	}
	nextPageToken := ""
	if page.NextOffset != nil {
		nextPageToken = strconv.Itoa(*page.NextOffset)
	}
	updated, err := json.Marshal(credentials)
	if err != nil {
		return nil, "", nil, fmt.Errorf("%w: encode Spotify credentials", domain.ErrUnavailable)
	}
	return tracks, nextPageToken, updated, nil
}

func spotifyPlayableTrack(track spotify.Track) domain.PlayableTrack {
	artists := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		artists = append(artists, artist.Name)
	}
	return domain.PlayableTrack{
		Provider: domain.MusicProviderSpotify, TrackID: track.ID, Title: track.Name, Artists: artists,
		Album: track.AlbumName, Duration: time.Duration(track.DurationMilliseconds) * time.Millisecond,
		ArtworkURL: track.ArtworkURL, ProviderURL: track.SpotifyURL, Playable: true,
	}
}
