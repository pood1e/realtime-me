package domain

import "time"

// MusicProvider identifies an isolated catalog and playback source.
type MusicProvider string

const (
	MusicProviderLocal   MusicProvider = "local"
	MusicProviderQQ      MusicProvider = "qq_music"
	MusicProviderNetEase MusicProvider = "netease_cloud_music"
	MusicProviderSpotify MusicProvider = "spotify"
)

// ProviderConnectionStatus describes whether a provider account can be used.
type ProviderConnectionStatus string

const (
	ProviderDisconnected      ProviderConnectionStatus = "disconnected"
	ProviderConnected         ProviderConnectionStatus = "connected"
	ProviderReconnectRequired ProviderConnectionStatus = "reconnect_required"
	ProviderUnavailable       ProviderConnectionStatus = "unavailable"
	ProviderNotConfigured     ProviderConnectionStatus = "not_configured"
)

// ProviderAttemptStatus describes one bounded interactive login operation.
type ProviderAttemptStatus string

const (
	ProviderAttemptWaiting   ProviderAttemptStatus = "waiting"
	ProviderAttemptScanned   ProviderAttemptStatus = "scanned"
	ProviderAttemptConnected ProviderAttemptStatus = "connected"
	ProviderAttemptExpired   ProviderAttemptStatus = "expired"
	ProviderAttemptRefused   ProviderAttemptStatus = "refused"
	ProviderAttemptFailed    ProviderAttemptStatus = "failed"
)

// ProviderSearchStatus describes one independently loaded search group.
type ProviderSearchStatus string

const (
	ProviderSearchReady             ProviderSearchStatus = "ready"
	ProviderSearchNotConnected      ProviderSearchStatus = "not_connected"
	ProviderSearchUnavailable       ProviderSearchStatus = "unavailable"
	ProviderSearchReconnectRequired ProviderSearchStatus = "reconnect_required"
)

// PlaybackQuality selects a direct browser audio tier.
type PlaybackQuality string

const (
	PlaybackQualityBest     PlaybackQuality = "best_compatible"
	PlaybackQualityHigh     PlaybackQuality = "high"
	PlaybackQualityStandard PlaybackQuality = "standard"
)

// MusicProviderCapability identifies one optional plugin contract.
type MusicProviderCapability string

const (
	MusicProviderAccountConnection MusicProviderCapability = "account_connection"
	MusicProviderCatalogSearch     MusicProviderCapability = "catalog_search"
	MusicProviderPlayback          MusicProviderCapability = "playback"
	MusicProviderLyrics            MusicProviderCapability = "lyrics"
	MusicProviderBrowserToken      MusicProviderCapability = "browser_token"
)

// Track is one private audio catalog entry.
type Track struct {
	UID               string
	ContentUID        string
	Title             string
	Artists           []string
	Album             string
	AlbumArtist       string
	TrackNumber       int
	DiscNumber        int
	Year              int
	Duration          time.Duration
	OriginalFileName  string
	ContentType       string
	SizeBytes         int64
	ArtworkStorageKey string
	Favorite          bool
	ProcessingStatus  ProcessingStatus
	CreateTime        time.Time
	UpdateTime        time.Time
	DeleteTime        *time.Time
}

// TrackPage is one cursor page of tracks.
type TrackPage struct {
	Tracks        []Track
	NextPageToken string
}

// Album summarizes tracks sharing an album and album artist.
type Album struct {
	UID               string
	Title             string
	AlbumArtist       string
	Year              int
	TrackCount        int
	ArtworkStorageKey string
}

// Artist summarizes one display artist.
type Artist struct {
	UID         string
	DisplayName string
	TrackCount  int
}

// PlayableTrack is the normalized read model shared by search and history.
type PlayableTrack struct {
	Provider        MusicProvider
	TrackID         string
	Title           string
	Artists         []string
	Album           string
	Duration        time.Duration
	ArtworkURL      string
	ProviderURL     string
	Playable        bool
	LyricsAvailable bool
}

// ProviderConnection contains account state plus encrypted credentials for storage.
type ProviderConnection struct {
	Provider             MusicProvider
	Status               ProviderConnectionStatus
	AccountID            string
	DisplayName          string
	AvatarURL            string
	Membership           string
	MembershipExpireTime *time.Time
	EncryptedCredentials []byte
	Capabilities         []MusicProviderCapability
	CreateTime           time.Time
	UpdateTime           time.Time
}

// ProviderConnectionAttempt persists one interactive login flow.
type ProviderConnectionAttempt struct {
	UID              string
	Provider         MusicProvider
	Status           ProviderAttemptStatus
	QRImage          []byte
	QRContentType    string
	QRPayload        string
	AuthorizationURL string
	StateHash        []byte
	EncryptedState   []byte
	CreateTime       time.Time
	UpdateTime       time.Time
	ExpireTime       time.Time
	ConsumedTime     *time.Time
}

// ProviderSearchGroup contains one source's independent search result.
type ProviderSearchGroup struct {
	Provider      MusicProvider
	Status        ProviderSearchStatus
	Tracks        []PlayableTrack
	NextPageToken string
}

// PlaybackDescriptor selects direct audio or the official Spotify SDK.
type PlaybackDescriptor struct {
	Provider    MusicProvider
	DirectURL   string
	ContentType string
	Quality     PlaybackQuality
	SpotifyURI  string
	ExpireTime  time.Time
}

// Lyric contains provider-supplied text without cross-provider substitution.
type Lyric struct {
	PlainText      string
	SyncedText     string
	TranslatedText string
}

// ProviderPlaybackToken is a short-lived browser SDK credential.
type ProviderPlaybackToken struct {
	AccessToken string
	ExpireTime  time.Time
}

// PlaybackEntry records one meaningful playback.
type PlaybackEntry struct {
	UID      string
	Track    PlayableTrack
	PlayTime time.Time
}

// PlaybackPage is one cursor page of playback entries.
type PlaybackPage struct {
	Entries       []PlaybackEntry
	NextPageToken string
}
