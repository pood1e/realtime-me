package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/qqmusic"
)

func resolveQQPlayback(ctx context.Context, rawCredentials []byte, trackID string, quality domain.PlaybackQuality) (domain.PlaybackDescriptor, []byte, error) {
	credentials, err := decodeQQCredentials(rawCredentials)
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, err
	}
	reference, err := decodeQQTrackReference(trackID)
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, err
	}
	client, err := qqmusic.NewClient()
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, mapProviderError(err)
	}
	credentials, err = refreshQQCredentials(ctx, client, credentials)
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, err
	}
	qualities := qqPlaybackQualities(quality)
	var playback qqmusic.PlaybackURL
	var playbackErr error
	for _, candidate := range qualities {
		playback, playbackErr = client.GetPlaybackURL(ctx, credentials, qqmusic.PlaybackRequest{
			TrackMID: reference.MID, MediaMID: reference.MediaMID, Quality: candidate,
		})
		if playbackErr == nil {
			break
		}
		if !qqPlaybackCanFallback(playbackErr) {
			break
		}
	}
	if playbackErr != nil {
		return domain.PlaybackDescriptor{}, nil, mapProviderError(playbackErr)
	}
	updated, err := json.Marshal(credentials)
	if err != nil {
		return domain.PlaybackDescriptor{}, nil, fmt.Errorf("%w: encode QQ Music credentials", domain.ErrUnavailable)
	}
	return domain.PlaybackDescriptor{
		Provider: domain.MusicProviderQQ, DirectURL: playback.URL, ContentType: playback.MIMEType,
		Quality: qqPlaybackQuality(playback.Quality), ExpireTime: playback.ExpiresAt,
	}, updated, nil
}

func qqPlaybackCanFallback(err error) bool {
	var providerError *qqmusic.Error
	return errors.As(err, &providerError) &&
		(providerError.Kind == qqmusic.ErrorKindUpstream || providerError.Kind == qqmusic.ErrorKindUnavailable)
}

func qqLyrics(ctx context.Context, rawCredentials []byte, trackID string) (domain.Lyric, []byte, error) {
	_, err := decodeQQCredentials(rawCredentials)
	if err != nil {
		return domain.Lyric{}, nil, err
	}
	reference, err := decodeQQTrackReference(trackID)
	if err != nil {
		return domain.Lyric{}, nil, err
	}
	client, err := qqmusic.NewClient()
	if err != nil {
		return domain.Lyric{}, nil, mapProviderError(err)
	}
	lyrics, err := client.GetLyrics(ctx, qqmusic.LyricsRequest{TrackMID: reference.MID})
	if err != nil {
		return domain.Lyric{}, nil, mapProviderError(err)
	}
	return domain.Lyric{SyncedText: lyrics.Original, TranslatedText: lyrics.Translation}, rawCredentials, nil
}

func decodeQQCredentials(encoded []byte) (qqmusic.Credentials, error) {
	var credentials qqmusic.Credentials
	if json.Unmarshal(encoded, &credentials) != nil || !credentials.Valid() {
		return qqmusic.Credentials{}, fmt.Errorf("%w: invalid QQ Music credentials", domain.ErrProviderReconnectRequired)
	}
	return credentials, nil
}

func refreshQQCredentials(ctx context.Context, client *qqmusic.Client, credentials qqmusic.Credentials) (qqmusic.Credentials, error) {
	if credentials.KeyExpiresIn <= 0 || credentials.MusicKeyCreatedAt <= 0 {
		return credentials, nil
	}
	expiresAt := time.Unix(credentials.MusicKeyCreatedAt+credentials.KeyExpiresIn, 0)
	if time.Now().Add(5 * time.Minute).Before(expiresAt) {
		return credentials, nil
	}
	refreshed, err := client.RefreshCredentials(ctx, credentials)
	if err != nil {
		return qqmusic.Credentials{}, mapProviderError(err)
	}
	return refreshed, nil
}

func qqPlaybackQualities(quality domain.PlaybackQuality) []qqmusic.AudioQuality {
	switch quality {
	case domain.PlaybackQualityStandard:
		return []qqmusic.AudioQuality{qqmusic.AudioQualityMP3128, qqmusic.AudioQualityAAC96}
	case domain.PlaybackQualityHigh:
		return []qqmusic.AudioQuality{qqmusic.AudioQualityMP3320, qqmusic.AudioQualityAAC192, qqmusic.AudioQualityMP3128}
	default:
		return []qqmusic.AudioQuality{qqmusic.AudioQualityFLAC, qqmusic.AudioQualityMP3320, qqmusic.AudioQualityAAC192}
	}
}

func qqPlaybackQuality(quality qqmusic.AudioQuality) domain.PlaybackQuality {
	if quality == qqmusic.AudioQualityMP3128 || quality == qqmusic.AudioQualityAAC96 {
		return domain.PlaybackQualityStandard
	}
	return domain.PlaybackQualityHigh
}
