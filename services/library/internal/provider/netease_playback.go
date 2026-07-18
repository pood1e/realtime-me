package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/netease"
)

func resolveNetEasePlayback(ctx context.Context, rawCredentials []byte, trackID string, quality domain.PlaybackQuality) (domain.PlaybackDescriptor, []byte, error) {
	client, _, err := netEaseClient(rawCredentials)
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, err
	}
	songID, err := parsePositiveID(trackID)
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, err
	}
	providerQuality := netease.AudioQualityBest
	fallback := true
	if quality == domain.PlaybackQualityHigh {
		providerQuality = netease.AudioQualityHigh
		fallback = false
	} else if quality == domain.PlaybackQualityStandard {
		providerQuality = netease.AudioQualityStandard
		fallback = false
	}
	resource, err := client.SongURL(ctx, netease.SongURLRequest{
		SongID: songID, Quality: providerQuality, FallbackTo320K: fallback,
	})
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, mapProviderError(err)
	}
	expiresAt := time.Now().Add(20 * time.Minute).UTC()
	if resource.ExpiresInSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(resource.ExpiresInSeconds) * time.Second).UTC()
	}
	updated, err := json.Marshal(client.Credentials())
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, fmt.Errorf("%w: encode NetEase credentials", domain.ErrUnavailable)
	}
	return domain.PlaybackDescriptor{
		Provider: domain.MusicProviderNetEase, DirectURL: resource.URL, ContentType: netEaseContentType(resource.Format),
		Quality: netEasePlaybackQuality(resource.Quality), ExpireTime: expiresAt,
	}, updated, nil
}

func netEaseLyrics(ctx context.Context, rawCredentials []byte, trackID string) (domain.Lyric, []byte, error) {
	client, _, err := netEaseClient(rawCredentials)
	if err != nil {
		return domain.Lyric{}, nil, err
	}
	songID, err := parsePositiveID(trackID)
	if err != nil {
		return domain.Lyric{}, nil, err
	}
	lyrics, err := client.Lyrics(ctx, songID)
	if err != nil {
		return domain.Lyric{}, nil, mapProviderError(err)
	}
	updated, err := json.Marshal(client.Credentials())
	if err != nil {
		return domain.Lyric{}, nil, fmt.Errorf("%w: encode NetEase credentials", domain.ErrUnavailable)
	}
	plainText := ""
	if lyrics.Instrumental {
		plainText = "纯音乐，无歌词"
	} else if lyrics.NotCollected {
		plainText = "暂未收录歌词"
	}
	return domain.Lyric{PlainText: plainText, SyncedText: lyrics.Original, TranslatedText: lyrics.Translation}, updated, nil
}

func netEaseClient(encoded []byte) (*netease.Client, netease.Credentials, error) {
	var credentials netease.Credentials
	if json.Unmarshal(encoded, &credentials) != nil || len(credentials.Cookies) == 0 {
		return nil, netease.Credentials{}, fmt.Errorf("%w: invalid NetEase credentials", domain.ErrProviderReconnectRequired)
	}
	client, err := netease.NewClient(netease.WithCredentials(credentials))
	if err != nil {
		return nil, netease.Credentials{}, mapProviderError(err)
	}
	return client, credentials, nil
}

func parsePositiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: invalid provider track ID", domain.ErrInvalidArgument)
	}
	return id, nil
}

func netEaseContentType(format string) string {
	switch strings.ToLower(format) {
	case "flac":
		return "audio/flac"
	case "m4a", "aac":
		return "audio/mp4"
	default:
		return "audio/mpeg"
	}
}

func netEasePlaybackQuality(quality string) domain.PlaybackQuality {
	if quality == "standard" || quality == "higher" {
		return domain.PlaybackQualityStandard
	}
	return domain.PlaybackQualityHigh
}
