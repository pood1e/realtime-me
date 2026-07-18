package netease

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/provider/failure"
)

// CredentialCookie is one persisted NetEase session cookie.
type CredentialCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (CredentialCookie) String() string {
	return "netease.CredentialCookie{redacted}"
}

func (cookie CredentialCookie) GoString() string {
	return cookie.String()
}

// Credentials can be serialized and restored into a Client.
type Credentials struct {
	Cookies []CredentialCookie `json:"cookies"`
}

func (Credentials) String() string {
	return "netease.Credentials{redacted}"
}

func (credentials Credentials) GoString() string {
	return credentials.String()
}

func (credentials Credentials) LogValue() slog.Value {
	return slog.StringValue(credentials.String())
}

// LoginAttempt contains everything required to resume QR-code polling.
type LoginAttempt struct {
	Key         string      `json:"key"`
	LoginURL    string      `json:"loginUrl"`
	Credentials Credentials `json:"credentials"`
	ExpiresAt   time.Time   `json:"expiresAt"`
}

func (LoginAttempt) String() string {
	return "netease.LoginAttempt{redacted}"
}

func (attempt LoginAttempt) GoString() string {
	return attempt.String()
}

func (attempt LoginAttempt) LogValue() slog.Value {
	return slog.StringValue(attempt.String())
}

// LoginState describes the current QR-code authorization state.
type LoginState string

const (
	LoginStateWaiting    LoginState = "waiting"
	LoginStateScanned    LoginState = "scanned"
	LoginStateAuthorized LoginState = "authorized"
	LoginStateExpired    LoginState = "expired"
)

// LoginStatus is returned while polling a LoginAttempt.
type LoginStatus struct {
	State       LoginState   `json:"state"`
	Nickname    string       `json:"nickname,omitempty"`
	AvatarURL   string       `json:"avatarUrl,omitempty"`
	Credentials *Credentials `json:"credentials,omitempty"`
}

// Account contains the authenticated account and its public profile.
type Account struct {
	ID       int64   `json:"id"`
	Username string  `json:"username,omitempty"`
	Status   int     `json:"status"`
	VIPType  int     `json:"vipType"`
	Profile  Profile `json:"profile"`
}

// Profile contains the account fields needed by the music application.
type Profile struct {
	UserID    int64  `json:"userId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	VIPType   int    `json:"vipType"`
}

// VIPInfo describes the authenticated user's music memberships.
type VIPInfo struct {
	UserID         int64          `json:"userId"`
	RedVIPLevel    int            `json:"redVipLevel"`
	RedAnnualCount int            `json:"redAnnualCount"`
	ServerTime     time.Time      `json:"serverTime,omitempty"`
	Associator     *VIPMembership `json:"associator,omitempty"`
	MusicPackage   *VIPMembership `json:"musicPackage,omitempty"`
}

// VIPMembership is one membership product returned by NetEase.
type VIPMembership struct {
	Code      int       `json:"code"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

// SearchRequest requests one offset page of songs.
type SearchRequest struct {
	Keywords string `json:"keywords"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

// SearchPage is one offset page of matching songs.
type SearchPage struct {
	Songs      []SearchSong `json:"songs"`
	Total      int          `json:"total"`
	Offset     int          `json:"offset"`
	Limit      int          `json:"limit"`
	HasMore    bool         `json:"hasMore"`
	NextOffset int          `json:"nextOffset,omitempty"`
}

// SearchSong is the normalized catalog representation of a NetEase song.
type SearchSong struct {
	ID                   int64          `json:"id"`
	Name                 string         `json:"name"`
	Artists              []SearchArtist `json:"artists"`
	Album                SearchAlbum    `json:"album"`
	DurationMilliseconds int64          `json:"durationMilliseconds"`
	Fee                  int            `json:"fee"`
}

// SearchArtist identifies one song artist.
type SearchArtist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// SearchAlbum identifies a song album.
type SearchAlbum struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	CoverURL string `json:"coverUrl,omitempty"`
}

// AudioQuality identifies one browser-compatible NetEase playback tier.
type AudioQuality string

const (
	AudioQualityBest     AudioQuality = "hires"
	AudioQualityHigh     AudioQuality = "exhigh"
	AudioQualityStandard AudioQuality = "standard"
)

// SongURLRequest requests one browser-compatible quality and optional fallback.
type SongURLRequest struct {
	SongID         int64        `json:"songId"`
	Quality        AudioQuality `json:"quality,omitempty"`
	FallbackTo320K bool         `json:"fallbackTo320K"`
}

// SongURL is a short-lived authenticated playback resource.
type SongURL struct {
	SongID           int64      `json:"songId"`
	URL              string     `json:"url"`
	Quality          string     `json:"quality"`
	Format           string     `json:"format"`
	Bitrate          int        `json:"bitrate"`
	SizeBytes        int64      `json:"sizeBytes"`
	MD5              string     `json:"md5,omitempty"`
	ExpiresInSeconds int        `json:"expiresInSeconds"`
	Fee              int        `json:"fee"`
	FreeTrial        *FreeTrial `json:"freeTrial,omitempty"`
}

// FreeTrial is the playable interval of a trial-only URL.
type FreeTrial struct {
	StartMilliseconds int64 `json:"startMilliseconds"`
	EndMilliseconds   int64 `json:"endMilliseconds"`
}

// Lyrics contains the available lyric variants for one song.
type Lyrics struct {
	SongID       int64  `json:"songId"`
	Original     string `json:"original,omitempty"`
	Translation  string `json:"translation,omitempty"`
	Romanization string `json:"romanization,omitempty"`
	Instrumental bool   `json:"instrumental"`
	NotCollected bool   `json:"notCollected"`
}

// ErrorKind classifies a provider failure without exposing response data.
type ErrorKind string

const (
	ErrorKindInvalid      ErrorKind = "invalid"
	ErrorKindTransport    ErrorKind = "transport"
	ErrorKindTimeout      ErrorKind = "timeout"
	ErrorKindHTTP         ErrorKind = "http"
	ErrorKindMalformed    ErrorKind = "malformed"
	ErrorKindUnauthorized ErrorKind = "unauthorized"
	ErrorKindUpstream     ErrorKind = "upstream"
	ErrorKindUnavailable  ErrorKind = "unavailable"
	ErrorKindNotFound     ErrorKind = "not_found"
	ErrorKindRateLimited  ErrorKind = "rate_limited"
)

// ProviderError is deliberately limited to non-sensitive diagnostics.
type ProviderError struct {
	Operation    string
	Kind         ErrorKind
	HTTPStatus   int
	UpstreamCode int
}

func (e *ProviderError) Error() string {
	switch e.Kind {
	case ErrorKindInvalid:
		return "netease " + e.Operation + ": invalid request"
	case ErrorKindTransport:
		return "netease " + e.Operation + ": network request failed"
	case ErrorKindTimeout:
		return "netease " + e.Operation + ": request timed out"
	case ErrorKindHTTP:
		return "netease " + e.Operation + ": unexpected HTTP status " + strconv.Itoa(e.HTTPStatus)
	case ErrorKindMalformed:
		return "netease " + e.Operation + ": malformed upstream response"
	case ErrorKindUnauthorized:
		return "netease " + e.Operation + ": authentication required"
	case ErrorKindUnavailable:
		return "netease " + e.Operation + ": song is unavailable"
	case ErrorKindNotFound:
		return "netease " + e.Operation + ": resource not found"
	case ErrorKindRateLimited:
		return "netease " + e.Operation + ": request rate limited"
	default:
		return "netease " + e.Operation + ": upstream request rejected with code " + strconv.Itoa(e.UpstreamCode)
	}
}

// FailureKind exposes only the provider-neutral category to adapter code.
func (e *ProviderError) FailureKind() failure.Kind {
	switch e.Kind {
	case ErrorKindInvalid:
		return failure.Invalid
	case ErrorKindUnauthorized:
		return failure.Unauthorized
	case ErrorKindNotFound:
		return failure.NotFound
	case ErrorKindRateLimited:
		return failure.RateLimited
	default:
		return failure.Unavailable
	}
}
