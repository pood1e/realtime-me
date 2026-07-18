package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func importNetEasePlaylist(ctx context.Context, rawCredentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	playlistID, err := parseNetEasePlaylistID(source)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	client, _, err := netEaseClient(rawCredentials)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, err
	}
	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		return domain.ProviderPlaylist{}, nil, mapProviderError(err)
	}
	tracks := make([]domain.PlayableTrack, 0, len(playlist.Tracks))
	for _, track := range playlist.Tracks {
		tracks = append(tracks, netEasePlayableTrack(track))
	}
	updated, err := json.Marshal(client.Credentials())
	if err != nil {
		return domain.ProviderPlaylist{}, nil, fmt.Errorf("%w: encode NetEase credentials", domain.ErrUnavailable)
	}
	id := strconv.FormatInt(playlist.ID, 10)
	return domain.ProviderPlaylist{
		ExternalID: id, DisplayName: playlist.Name, ArtworkURL: playlist.CoverURL,
		ProviderURL: "https://music.163.com/playlist?id=" + id, Tracks: tracks,
	}, updated, nil
}

func resolveNetEaseDownload(ctx context.Context, credentials []byte, trackID string) (domain.ProviderDownload, []byte, error) {
	playback, updated, err := resolveNetEasePlayback(ctx, credentials, trackID, domain.PlaybackQualityBest)
	if err != nil {
		return domain.ProviderDownload{}, nil, err
	}
	return domain.ProviderDownload{URL: playback.DirectURL, ContentType: playback.ContentType}, updated, nil
}

func parseNetEasePlaylistID(source string) (int64, error) {
	source = strings.TrimSpace(source)
	if id, err := strconv.ParseInt(source, 10, 64); err == nil && id > 0 {
		return id, nil
	}
	parsed, err := url.ParseRequestURI(source)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return 0, fmt.Errorf("%w: invalid NetEase playlist source", domain.ErrInvalidArgument)
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "music.163.com" && host != "y.music.163.com" {
		return 0, fmt.Errorf("%w: invalid NetEase playlist host", domain.ErrInvalidArgument)
	}
	if id, err := strconv.ParseInt(parsed.Query().Get("id"), 10, 64); err == nil && id > 0 {
		return id, nil
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for index, segment := range segments {
		if segment == "playlist" && index+1 < len(segments) {
			if id, err := strconv.ParseInt(segments[index+1], 10, 64); err == nil && id > 0 {
				return id, nil
			}
		}
	}
	return 0, fmt.Errorf("%w: NetEase playlist ID is missing", domain.ErrInvalidArgument)
}
