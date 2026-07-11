package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/qqmusic"
)

type qqTrackReference struct {
	MID      string `json:"mid"`
	MediaMID string `json:"media_mid,omitempty"`
}

func searchQQ(ctx context.Context, client *qqmusic.Client, rawCredentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	credentials, err := decodeQQCredentials(rawCredentials)
	if err != nil {
		return nil, "", nil, err
	}
	credentials, err = refreshQQCredentials(ctx, client, credentials)
	if err != nil {
		return nil, "", nil, err
	}
	page, err := client.Search(ctx, qqmusic.SearchRequest{Query: query, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		return nil, "", nil, mapProviderError(err)
	}
	tracks := make([]domain.PlayableTrack, 0, len(page.Tracks))
	for _, track := range page.Tracks {
		playable, err := qqPlayableTrack(track)
		if err != nil {
			return nil, "", nil, err
		}
		tracks = append(tracks, playable)
	}
	updated, err := json.Marshal(credentials)
	if err != nil {
		return nil, "", nil, fmt.Errorf("%w: encode QQ Music credentials", domain.ErrUnavailable)
	}
	return tracks, page.NextPageToken, updated, nil
}

func qqPlayableTrack(track qqmusic.Track) (domain.PlayableTrack, error) {
	artists := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		artists = append(artists, artist.Name)
	}
	reference, err := encodeQQTrackReference(qqTrackReference{MID: track.MID, MediaMID: track.MediaMID})
	if err != nil {
		return domain.PlayableTrack{}, err
	}
	return domain.PlayableTrack{
		Provider: domain.MusicProviderQQ, TrackID: reference, Title: track.Title, Artists: artists,
		Album: track.Album.Title, Duration: time.Duration(track.Duration) * time.Second,
		ArtworkURL: track.Album.CoverURL, ProviderURL: "https://y.qq.com/n/ryqq/songDetail/" + track.MID,
		Playable: track.Available, LyricsAvailable: true,
	}, nil
}

func encodeQQTrackReference(reference qqTrackReference) (string, error) {
	if reference.MID == "" {
		return "", fmt.Errorf("%w: invalid QQ Music track", domain.ErrUnavailable)
	}
	encoded, err := json.Marshal(reference)
	if err != nil {
		return "", fmt.Errorf("%w: encode QQ Music track", domain.ErrUnavailable)
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decodeQQTrackReference(value string) (qqTrackReference, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return qqTrackReference{}, fmt.Errorf("%w: invalid QQ Music track", domain.ErrInvalidArgument)
	}
	var reference qqTrackReference
	if json.Unmarshal(decoded, &reference) != nil || reference.MID == "" {
		return qqTrackReference{}, fmt.Errorf("%w: invalid QQ Music track", domain.ErrInvalidArgument)
	}
	return reference, nil
}
