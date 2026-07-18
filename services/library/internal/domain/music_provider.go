package domain

import (
	"context"
	"time"
)

// ProviderAccount is the normalized result of a successful provider login.
type ProviderAccount struct {
	AccountID            string
	DisplayName          string
	AvatarURL            string
	Membership           string
	MembershipExpireTime *time.Time
	Credentials          []byte
}

// ProviderLoginChallenge starts either a QR or redirect login operation.
type ProviderLoginChallenge struct {
	QRImage          []byte
	QRContentType    string
	QRPayload        string
	AuthorizationURL string
	OAuthState       string
	State            []byte
	ExpireTime       time.Time
}

// ProviderLoginPoll is one provider-side QR status transition.
type ProviderLoginPoll struct {
	Status  ProviderAttemptStatus
	State   []byte
	Account *ProviderAccount
}

// MusicProviderAdapter is the common identity shared by every compile-time plugin.
type MusicProviderAdapter interface {
	Provider() MusicProvider
	DisplayName() string
	Configured() bool
}

// MusicProviderRegistry resolves registered plugins without platform switches.
type MusicProviderRegistry interface {
	Get(MusicProvider) (MusicProviderAdapter, bool)
	List() []MusicProviderAdapter
}

// MusicLoginStarter begins either a QR or redirect login operation.
type MusicLoginStarter interface {
	BeginLogin(context.Context) (ProviderLoginChallenge, error)
}

// MusicQRLoginPoller is implemented only by QR-based account plugins.
type MusicQRLoginPoller interface {
	PollLogin(context.Context, []byte) (ProviderLoginPoll, error)
}

// MusicRedirectLoginCompleter is implemented only by OAuth redirect plugins.
type MusicRedirectLoginCompleter interface {
	CompleteRedirectLogin(context.Context, string, []byte) (ProviderAccount, error)
}

// MusicCatalogSearcher provides normalized, source-local search pagination.
type MusicCatalogSearcher interface {
	Search(context.Context, []byte, string, int, string) ([]PlayableTrack, string, []byte, error)
}

// MusicPlaybackResolver provides direct audio or an official SDK descriptor.
type MusicPlaybackResolver interface {
	ResolvePlayback(context.Context, []byte, string, PlaybackQuality) (PlaybackDescriptor, []byte, error)
}

// MusicLyricsProvider is an optional provider-supplied lyrics capability.
type MusicLyricsProvider interface {
	Lyrics(context.Context, []byte, string) (Lyric, []byte, error)
}

// MusicPlaybackTokenProvider is an optional browser SDK token capability.
type MusicPlaybackTokenProvider interface {
	PlaybackToken(context.Context, []byte) (ProviderPlaybackToken, []byte, error)
}

// MusicPlaylistImporter resolves a provider playlist URL or identifier.
type MusicPlaylistImporter interface {
	ImportPlaylist(context.Context, []byte, string) (ProviderPlaylist, []byte, error)
}

// MusicTrackDownloader resolves provider audio that can be persisted locally.
type MusicTrackDownloader interface {
	ResolveDownload(context.Context, []byte, string) (ProviderDownload, []byte, error)
}
