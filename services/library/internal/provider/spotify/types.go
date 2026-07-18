package spotify

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/provider/failure"
)

const (
	ScopeStreaming               = "streaming"
	ScopeUserReadPrivate         = "user-read-private"
	ScopeUserReadPlaybackState   = "user-read-playback-state"
	ScopeUserModifyPlaybackState = "user-modify-playback-state"
	ScopePlaylistReadPrivate     = "playlist-read-private"
)

// Config contains the Spotify application credentials registered for this
// client. ClientSecret is optional when the client uses PKCE as a public client.
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"-"`
	RedirectURI  string `json:"redirect_uri"`
}

func (c Config) String() string {
	return "spotify.Config{redacted}"
}

func (c Config) GoString() string {
	return c.String()
}

func (c Config) LogValue() slog.Value {
	return slog.StringValue(c.String())
}

// Credentials are the reusable result of an OAuth code exchange or refresh.
// Callers should persist this value in secret storage.
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	Scopes       []string  `json:"scopes,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (c Credentials) String() string {
	return "spotify.Credentials{redacted}"
}

func (c Credentials) GoString() string {
	return c.String()
}

func (c Credentials) LogValue() slog.Value {
	return slog.StringValue(c.String())
}

func (c Credentials) HasScope(scope string) bool {
	for _, granted := range c.Scopes {
		if granted == scope {
			return true
		}
	}
	return false
}

// LoginAttempt contains the state needed to complete one PKCE authorization.
// CodeVerifier must be treated as a short-lived secret.
type LoginAttempt struct {
	State        string    `json:"state"`
	CodeVerifier string    `json:"code_verifier"`
	Scopes       []string  `json:"scopes,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (a LoginAttempt) String() string {
	return "spotify.LoginAttempt{redacted}"
}

func (a LoginAttempt) GoString() string {
	return a.String()
}

func (a LoginAttempt) LogValue() slog.Value {
	return slog.StringValue(a.String())
}

func (a LoginAttempt) validState(returnedState string, now time.Time) bool {
	if a.State == "" || returnedState == "" || a.ExpiresAt.IsZero() || !now.Before(a.ExpiresAt) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a.State), []byte(returnedState)) == 1
}

// PlaybackToken is safe to return to a Web Playback SDK client. It never
// contains the reusable refresh token.
type PlaybackToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (t PlaybackToken) String() string {
	return "spotify.PlaybackToken{redacted}"
}

func (t PlaybackToken) GoString() string {
	return t.String()
}

func (t PlaybackToken) LogValue() slog.Value {
	return slog.StringValue(t.String())
}

type User struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name,omitempty"`
}

type TrackSearchRequest struct {
	Query  string
	Market string
	Limit  int
	Offset int
}

type TrackSearchPage struct {
	Items      []Track `json:"items"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	Total      int     `json:"total"`
	NextOffset *int    `json:"next_offset,omitempty"`
}

type Track struct {
	ID                   string   `json:"id"`
	URI                  string   `json:"uri"`
	SpotifyURL           string   `json:"spotify_url"`
	Name                 string   `json:"name"`
	Artists              []Artist `json:"artists"`
	AlbumID              string   `json:"album_id"`
	AlbumName            string   `json:"album_name"`
	ArtworkURL           string   `json:"artwork_url,omitempty"`
	ISRC                 string   `json:"isrc,omitempty"`
	DurationMilliseconds int      `json:"duration_ms"`
	Explicit             bool     `json:"explicit"`
}

type Artist struct {
	ID   string `json:"id"`
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// APIError describes a non-successful response without retaining request
// credentials or raw response bodies.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "spotify API error"
	}

	details := httpStatusLabel(e.StatusCode)
	if e.Code != "" {
		details += ": " + e.Code
	}
	if e.Message != "" {
		details += ": " + e.Message
	}
	if e.RetryAfter > 0 {
		details += fmt.Sprintf(" (retry after %s)", e.RetryAfter)
	}
	return "spotify API error: " + details
}

func (e *APIError) Temporary() bool {
	return e != nil && (e.StatusCode == 429 || e.StatusCode >= 500)
}

// FailureKind exposes only the provider-neutral category to adapter code.
func (e *APIError) FailureKind() failure.Kind {
	if e == nil {
		return failure.Unavailable
	}
	switch {
	case e.StatusCode == 401 || e.Code == "invalid_grant":
		return failure.Unauthorized
	case e.StatusCode == 403:
		return failure.Forbidden
	case e.StatusCode == 404:
		return failure.NotFound
	case e.StatusCode == 429:
		return failure.RateLimited
	default:
		return failure.Unavailable
	}
}

func httpStatusLabel(statusCode int) string {
	if statusCode == 0 {
		return "request failed"
	}
	return fmt.Sprintf("HTTP %d", statusCode)
}

func normalizedScopes(scopes []string) []string {
	seen := make(map[string]struct{}, len(scopes))
	normalized := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	return normalized
}
