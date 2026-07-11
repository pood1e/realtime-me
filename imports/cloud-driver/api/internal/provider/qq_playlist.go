package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/qqmusic"
)

func importQQPlaylist(ctx context.Context, client *qqmusic.Client, rawCredentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	credentials, err := decodeQQCredentials(rawCredentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	playlistID, err := client.ResolvePlaylistID(ctx, source)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, mapProviderError(err)
	}
	credentials, err = refreshQQCredentials(ctx, client, credentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	playlist, err := client.GetPlaylist(ctx, credentials, playlistID)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, mapProviderError(err)
	}
	tracks := make([]domain.PlayableTrack, 0, len(playlist.Tracks))
	for _, track := range playlist.Tracks {
		playable, err := qqPlayableTrack(track)
		if err != nil {
			return domain.ProviderPlaylist{}, nil, err
		}
		tracks = append(tracks, playable)
	}
	updated, err := json.Marshal(credentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, fmt.Errorf("%w: encode QQ Music credentials", domain.ErrUnavailable)
	}
	id := strconv.FormatInt(playlist.ID, 10)
	return domain.ProviderPlaylist{
		ExternalID: id, DisplayName: playlist.Name, ArtworkURL: playlist.CoverURL,
		ProviderURL: "https://y.qq.com/n/ryqq/playlist/" + id, Tracks: tracks,
	}, updated, nil
}

func resolveQQDownload(ctx context.Context, client *qqmusic.Client, credentials []byte, trackID string) (domain.ProviderDownload, []byte, error) {
	playback, updated, err := resolveQQPlayback(ctx, client, credentials, trackID, domain.PlaybackQualityBest)
	if err != nil {
		return domain.ProviderDownload{}, nil, err
	}
	return domain.ProviderDownload{URL: playback.DirectURL, ContentType: playback.ContentType}, updated, nil
}
