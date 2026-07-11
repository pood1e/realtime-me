package provider

import (
	"context"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/qqmusic"
)

// QQAdapter is the compile-time QQ Music plugin.
type QQAdapter struct {
	client *qqmusic.Client
}

var (
	_ domain.MusicProviderAdapter  = (*QQAdapter)(nil)
	_ domain.MusicLoginStarter     = (*QQAdapter)(nil)
	_ domain.MusicQRLoginPoller    = (*QQAdapter)(nil)
	_ domain.MusicCatalogSearcher  = (*QQAdapter)(nil)
	_ domain.MusicPlaybackResolver = (*QQAdapter)(nil)
	_ domain.MusicLyricsProvider   = (*QQAdapter)(nil)
	_ domain.MusicPlaylistImporter = (*QQAdapter)(nil)
	_ domain.MusicTrackDownloader  = (*QQAdapter)(nil)
)

func newQQAdapter() (*QQAdapter, error) {
	client, err := qqmusic.NewClient()
	if err != nil {
		return nil, err
	}
	return &QQAdapter{client: client}, nil
}

func (*QQAdapter) Provider() domain.MusicProvider { return domain.MusicProviderQQ }

func (*QQAdapter) DisplayName() string { return "QQ 音乐" }

func (*QQAdapter) Configured() bool { return true }

func (*QQAdapter) BeginLogin(ctx context.Context) (domain.ProviderLoginChallenge, error) {
	return beginQQLogin(ctx)
}

func (*QQAdapter) PollLogin(ctx context.Context, state []byte) (domain.ProviderLoginPoll, error) {
	return pollQQLogin(ctx, state)
}

func (a *QQAdapter) Search(ctx context.Context, credentials []byte, query string, pageSize int, pageToken string) ([]domain.PlayableTrack, string, []byte, error) {
	return searchQQ(ctx, a.client, credentials, query, pageSize, pageToken)
}

func (a *QQAdapter) ResolvePlayback(ctx context.Context, credentials []byte, trackID string, quality domain.PlaybackQuality) (domain.PlaybackDescriptor, []byte, error) {
	return resolveQQPlayback(ctx, a.client, credentials, trackID, quality)
}

func (a *QQAdapter) Lyrics(ctx context.Context, credentials []byte, trackID string) (domain.Lyric, []byte, error) {
	return qqLyrics(ctx, a.client, credentials, trackID)
}

func (a *QQAdapter) ImportPlaylist(ctx context.Context, credentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	return importQQPlaylist(ctx, a.client, credentials, source)
}

func (a *QQAdapter) ResolveDownload(ctx context.Context, credentials []byte, trackID string) (domain.ProviderDownload, []byte, error) {
	return resolveQQDownload(ctx, a.client, credentials, trackID)
}
