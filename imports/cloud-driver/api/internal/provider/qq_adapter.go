package provider

import (
	"context"

	"example.com/cloud-drive/api/internal/domain"
)

// QQAdapter is the compile-time QQ Music plugin.
type QQAdapter struct{}

var (
	_ domain.MusicProviderAdapter  = QQAdapter{}
	_ domain.MusicLoginStarter     = QQAdapter{}
	_ domain.MusicQRLoginPoller    = QQAdapter{}
	_ domain.MusicCatalogSearcher  = QQAdapter{}
	_ domain.MusicPlaybackResolver = QQAdapter{}
	_ domain.MusicLyricsProvider   = QQAdapter{}
)

func (QQAdapter) Provider() domain.MusicProvider { return domain.MusicProviderQQ }

func (QQAdapter) Configured() bool { return true }

func (QQAdapter) BeginLogin(ctx context.Context) (domain.ProviderLoginChallenge, error) {
	return beginQQLogin(ctx)
}

func (QQAdapter) PollLogin(ctx context.Context, state []byte) (domain.ProviderLoginPoll, error) {
	return pollQQLogin(ctx, state)
}

func (QQAdapter) Search(ctx context.Context, credentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	return searchQQ(ctx, credentials, query, pageSize, pageToken)
}

func (QQAdapter) ResolvePlayback(ctx context.Context, credentials []byte, trackID string, quality domain.PlaybackQuality) (domain.PlaybackDescriptor, []byte, error) {
	return resolveQQPlayback(ctx, credentials, trackID, quality)
}

func (QQAdapter) Lyrics(ctx context.Context, credentials []byte, trackID string) (domain.Lyric, []byte, error) {
	return qqLyrics(ctx, credentials, trackID)
}
