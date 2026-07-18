package provider

import (
	"context"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
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
	_ domain.MusicPlaylistImporter = NetEaseAdapter{}
	_ domain.MusicTrackDownloader  = NetEaseAdapter{}
)

func (NetEaseAdapter) Provider() domain.MusicProvider { return domain.MusicProviderNetEase }

func (NetEaseAdapter) DisplayName() string { return "网易云音乐" }

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

func (NetEaseAdapter) ImportPlaylist(ctx context.Context, credentials []byte, source string) (domain.ProviderPlaylist, []byte, error) {
	return importNetEasePlaylist(ctx, credentials, source)
}

func (NetEaseAdapter) ResolveDownload(ctx context.Context, credentials []byte, trackID string) (domain.ProviderDownload, []byte, error) {
	return resolveNetEaseDownload(ctx, credentials, trackID)
}
