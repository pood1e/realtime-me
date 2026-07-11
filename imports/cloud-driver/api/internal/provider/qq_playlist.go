package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/qqmusic"
)

func importQQPlaylist(ctx context.Context, rawCredentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	playlistID, err := parseQQPlaylistID(source)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	credentials, err := decodeQQCredentials(rawCredentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	client, err := qqmusic.NewClient()
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

func resolveQQDownload(ctx context.Context, credentials []byte, trackID string) (domain.ProviderDownload, []byte, error) {
	playback, updated, err := resolveQQPlayback(ctx, credentials, trackID, domain.PlaybackQualityBest)
	if err != nil {
		return domain.ProviderDownload{}, nil, err
	}
	return domain.ProviderDownload{URL: playback.DirectURL, ContentType: playback.ContentType}, updated, nil
}

func parseQQPlaylistID(source string) (int64, error) {
	source = strings.TrimSpace(source)
	if id, err := strconv.ParseInt(source, 10, 64); err == nil && id > 0 {
		return id, nil
	}
	parsed, err := url.ParseRequestURI(source)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return 0, fmt.Errorf("%w: invalid QQ Music playlist source", domain.ErrInvalidArgument)
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "y.qq.com" && host != "i.y.qq.com" && host != "c.y.qq.com" {
		return 0, fmt.Errorf("%w: invalid QQ Music playlist host", domain.ErrInvalidArgument)
	}
	for _, key := range []string{"id", "disstid"} {
		if id, err := strconv.ParseInt(parsed.Query().Get(key), 10, 64); err == nil && id > 0 {
			return id, nil
		}
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for index, segment := range segments {
		if (segment == "playlist" || segment == "playsquare") && index+1 < len(segments) {
			if id, err := strconv.ParseInt(segments[index+1], 10, 64); err == nil && id > 0 {
				return id, nil
			}
		}
	}
	return 0, fmt.Errorf("%w: QQ Music playlist ID is missing", domain.ErrInvalidArgument)
}
