package qqmusic

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/provider/failure"
)

// Credentials contains the reusable QQ Music login state returned by QQLogin.
// Treat every field as a secret even when it is not itself an access token.
type Credentials struct {
	OpenID            string `json:"openid,omitempty"`
	RefreshToken      string `json:"refresh_token,omitempty"`
	AccessToken       string `json:"access_token,omitempty"`
	ExpiresAt         int64  `json:"expires_at,omitempty"`
	MusicID           int64  `json:"music_id"`
	MusicKey          string `json:"music_key"`
	UnionID           string `json:"union_id,omitempty"`
	StringMusicID     string `json:"string_music_id,omitempty"`
	RefreshKey        string `json:"refresh_key,omitempty"`
	MusicKeyCreatedAt int64  `json:"music_key_created_at,omitempty"`
	KeyExpiresIn      int64  `json:"key_expires_in,omitempty"`
	EncryptedUIN      string `json:"encrypted_uin,omitempty"`
	LoginType         int    `json:"login_type,omitempty"`
}

// Valid reports whether the minimum credentials required by authenticated
// QQ Music requests are present.
func (credentials Credentials) Valid() bool {
	return credentials.MusicID > 0 && strings.TrimSpace(credentials.MusicKey) != ""
}

func (credentials Credentials) String() string {
	return "qqmusic.Credentials{redacted}"
}

func (credentials Credentials) GoString() string {
	return credentials.String()
}

func (credentials Credentials) LogValue() slog.Value {
	return slog.StringValue(credentials.String())
}

// LoginAttempt is the serializable state of a QQ QR-code login operation.
// QRCode is encoded as base64 by encoding/json.
type LoginAttempt struct {
	QRCode    []byte    `json:"qr_code"`
	MIMEType  string    `json:"mime_type"`
	QRSig     string    `json:"qrsig"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (attempt LoginAttempt) String() string {
	return "qqmusic.LoginAttempt{redacted}"
}

func (attempt LoginAttempt) GoString() string {
	return attempt.String()
}

func (attempt LoginAttempt) LogValue() slog.Value {
	return slog.StringValue(attempt.String())
}

type LoginStatus string

const (
	LoginStatusWaiting   LoginStatus = "waiting"
	LoginStatusScanned   LoginStatus = "scanned"
	LoginStatusSucceeded LoginStatus = "succeeded"
	LoginStatusExpired   LoginStatus = "expired"
	LoginStatusRejected  LoginStatus = "rejected"
)

type LoginResult struct {
	Status      LoginStatus  `json:"status"`
	Credentials *Credentials `json:"credentials,omitempty"`
}

type VIPInfo struct {
	VIP             bool   `json:"vip"`
	LuxuryVIP       bool   `json:"luxury_vip"`
	SuperVIP        bool   `json:"super_vip"`
	Annual          bool   `json:"annual"`
	Level           int    `json:"level"`
	ExpiresAt       int64  `json:"expires_at,omitempty"`
	LuxuryExpiresAt string `json:"luxury_expires_at,omitempty"`
	IconURL         string `json:"icon_url,omitempty"`
}

type SearchRequest struct {
	Query     string `json:"query"`
	PageSize  int    `json:"page_size,omitempty"`
	PageToken string `json:"page_token,omitempty"`
}

type SearchResponse struct {
	Tracks        []Track `json:"tracks"`
	Total         int64   `json:"total"`
	NextPageToken string  `json:"next_page_token,omitempty"`
}

type Track struct {
	ID          int64    `json:"id"`
	MID         string   `json:"mid"`
	MediaMID    string   `json:"media_mid,omitempty"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Artists     []Artist `json:"artists"`
	Album       Album    `json:"album"`
	Duration    int      `json:"duration_seconds"`
	PayToPlay   bool     `json:"pay_to_play"`
	VIPRequired bool     `json:"vip_required"`
	Available   bool     `json:"available"`
}

type Artist struct {
	ID   int64  `json:"id"`
	MID  string `json:"mid,omitempty"`
	Name string `json:"name"`
}

type Album struct {
	ID       int64  `json:"id"`
	MID      string `json:"mid,omitempty"`
	Title    string `json:"title,omitempty"`
	CoverURL string `json:"cover_url,omitempty"`
}

type AudioQuality string

const (
	AudioQualityAAC96  AudioQuality = "aac_96"
	AudioQualityAAC192 AudioQuality = "aac_192"
	AudioQualityMP3128 AudioQuality = "mp3_128"
	AudioQualityMP3320 AudioQuality = "mp3_320"
	AudioQualityFLAC   AudioQuality = "flac"
)

type PlaybackRequest struct {
	TrackMID string       `json:"track_mid"`
	MediaMID string       `json:"media_mid,omitempty"`
	Quality  AudioQuality `json:"quality,omitempty"`
}

type PlaybackURL struct {
	URL       string       `json:"url"`
	MIMEType  string       `json:"mime_type"`
	Quality   AudioQuality `json:"quality"`
	ExpiresAt time.Time    `json:"expires_at,omitempty"`
}

func (playback PlaybackURL) String() string {
	return "qqmusic.PlaybackURL{redacted}"
}

func (playback PlaybackURL) GoString() string {
	return playback.String()
}

func (playback PlaybackURL) LogValue() slog.Value {
	return slog.StringValue(playback.String())
}

type LyricsRequest struct {
	TrackMID string `json:"track_mid"`
}

type Lyrics struct {
	TrackMID    string `json:"track_mid"`
	Original    string `json:"original"`
	Translation string `json:"translation,omitempty"`
	Romanized   string `json:"romanized,omitempty"`
}

type ErrorKind string

const (
	ErrorKindInvalidInput ErrorKind = "invalid_input"
	ErrorKindNetwork      ErrorKind = "network"
	ErrorKindUpstream     ErrorKind = "upstream"
	ErrorKindUnauthorized ErrorKind = "unauthorized"
	ErrorKindForbidden    ErrorKind = "forbidden"
	ErrorKindUnavailable  ErrorKind = "unavailable"
	ErrorKindNotFound     ErrorKind = "not_found"
	ErrorKindRateLimited  ErrorKind = "rate_limited"
)

// Error deliberately contains no request URL, response body, cookie, or
// credential material.
type Error struct {
	Operation string
	Kind      ErrorKind
	Code      int
	Message   string
}

func (err *Error) Error() string {
	if err.Code != 0 {
		return fmt.Sprintf("qqmusic %s: %s (code %d)", err.Operation, err.Message, err.Code)
	}
	return fmt.Sprintf("qqmusic %s: %s", err.Operation, err.Message)
}

// FailureKind exposes only the provider-neutral category to adapter code.
func (err *Error) FailureKind() failure.Kind {
	switch err.Kind {
	case ErrorKindInvalidInput:
		return failure.Invalid
	case ErrorKindUnauthorized:
		return failure.Unauthorized
	case ErrorKindForbidden:
		return failure.Forbidden
	case ErrorKindNotFound:
		return failure.NotFound
	case ErrorKindRateLimited:
		return failure.RateLimited
	default:
		return failure.Unavailable
	}
}

func providerError(operation string, kind ErrorKind, message string, code int) error {
	return &Error{Operation: operation, Kind: kind, Code: code, Message: message}
}
