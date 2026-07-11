package provider

import (
	"context"

	"example.com/cloud-drive/api/internal/domain"
)

// NetEaseAdapter is the compile-time NetEase Cloud Music plugin.
type NetEaseAdapter struct{}

var (
	_ domain.MusicProviderAdapter  = NetEaseAdapter{}
	_ domain.MusicLoginStarter     = NetEaseAdapter{}
	_ domain.MusicQRLoginPoller    = NetEaseAdapter{}
	_ domain.MusicCatalogSearcher  = NetEaseAdapter{}
	_ domain.MusicPlaybackResolver = NetEaseAdapter{}
	_ domain.MusicLyricsProvider   = NetEaseAdapter{}
)

func (NetEaseAdapter) Provider() domain.MusicProvider { return domain.MusicProviderNetEase }

func (NetEaseAdapter) Configured() bool { return true }

func (NetEaseAdapter) BeginLogin(ctx context.Context) (domain.ProviderLoginChallenge, error) {
	return beginNetEaseLogin(ctx)
}

func (NetEaseAdapter) PollLogin(ctx context.Context, state []byte) (domain.ProviderLoginPoll, error) {
	return pollNetEaseLogin(ctx, state)
}

func (NetEaseAdapter) Search(ctx context.Context, credentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	return searchNetEase(ctx, credentials, query, pageSize, pageToken)
}

func (NetEaseAdapter) ResolvePlayback(ctx context.Context, credentials []byte, trackID string, quality domain.PlaybackQuality) (domain.PlaybackDescriptor, []byte, error) {
	return resolveNetEasePlayback(ctx, credentials, trackID, quality)
}

func (NetEaseAdapter) Lyrics(ctx context.Context, credentials []byte, trackID string) (domain.Lyric, []byte, error) {
	return netEaseLyrics(ctx, credentials, trackID)
}
