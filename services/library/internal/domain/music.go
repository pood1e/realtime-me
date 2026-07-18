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

// ValidMusicProviderID validates the open, DNS-label-like plugin identifier.
func ValidMusicProviderID(provider MusicProvider) bool {
	value := string(provider)
	if len(value) < 1 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '_' || character == '.' || character == '-' {
			continue
		}
		return false
	}
	return true
}

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
	MusicProviderPlaylistImport    MusicProviderCapability = "playlist_import"
	MusicProviderLocalDownload     MusicProviderCapability = "local_download"
)

// ProviderDescriptor describes one registered plugin without exposing credentials.
type ProviderDescriptor struct {
	ID           MusicProvider
	DisplayName  string
	Capabilities []MusicProviderCapability
	Configured   bool
}

// PlaylistTrackDownloadStatus describes local persistence progress.
type PlaylistTrackDownloadStatus string

const (
	PlaylistTrackDownloadNotStarted PlaylistTrackDownloadStatus = "not_started"
	PlaylistTrackDownloadPending    PlaylistTrackDownloadStatus = "pending"
	PlaylistTrackDownloadRunning    PlaylistTrackDownloadStatus = "running"
	PlaylistTrackDownloadCompleted  PlaylistTrackDownloadStatus = "completed"
	PlaylistTrackDownloadFailed     PlaylistTrackDownloadStatus = "failed"
)

// PlaylistImportStatus describes one durable provider import operation.
type PlaylistImportStatus string

const (
	PlaylistImportPending   PlaylistImportStatus = "pending"
	PlaylistImportRunning   PlaylistImportStatus = "running"
	PlaylistImportCompleted PlaylistImportStatus = "completed"
	PlaylistImportFailed    PlaylistImportStatus = "failed"
)

// PlaylistImport is one queued provider playlist resolution.
type PlaylistImport struct {
	UID         string
	Provider    MusicProvider
	Source      string
	Status      PlaylistImportStatus
	PlaylistUID string
	FailureCode string
	CreateTime  time.Time
	UpdateTime  time.Time
}

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

// TrackListQuery contains the complete cursor filter for the local music catalog.
type TrackListQuery struct {
	Query     string
	Album     string
	Artist    string
	Favorites bool
	Trashed   bool
	PageSize  int
	PageToken string
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

// PlaybackDescriptor selects direct audio or a provider browser SDK.
type PlaybackDescriptor struct {
	Provider    MusicProvider
	DirectURL   string
	ContentType string
	Quality     PlaybackQuality
	SDKID       string
	ResourceURI string
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

// ProviderPlaylist is the normalized snapshot returned by a provider plugin.
type ProviderPlaylist struct {
	ExternalID  string
	DisplayName string
	ArtworkURL  string
	ProviderURL string
	Tracks      []PlayableTrack
}

// Playlist is one locally imported provider playlist.
type Playlist struct {
	UID                    string
	Provider               MusicProvider
	ExternalID             string
	DisplayName            string
	ArtworkURL             string
	ProviderURL            string
	TrackCount             int
	DownloadableTrackCount int
	PendingTrackCount      int
	CompletedTrackCount    int
	FailedTrackCount       int
	DownloadSupported      bool
	CreateTime             time.Time
	UpdateTime             time.Time
}

// PlaylistPage is one cursor page of imported playlists.
type PlaylistPage struct {
	Playlists     []Playlist
	NextPageToken string
}

// PlaylistTrack is one ordered provider track and its local persistence state.
type PlaylistTrack struct {
	UID            string
	PlaylistUID    string
	Position       int
	Track          PlayableTrack
	DownloadStatus PlaylistTrackDownloadStatus
	LocalTrackUID  string
}

// PlaylistTrackPage is one cursor page of playlist tracks.
type PlaylistTrackPage struct {
	Tracks        []PlaylistTrack
	NextPageToken string
}

// ProviderDownload is a short-lived direct audio resource.
type ProviderDownload struct {
	URL         string
	ContentType string
}

// MusicDownload contains everything required by the background downloader.
type MusicDownload struct {
	PlaylistTrack PlaylistTrack
	Connection    ProviderConnection
}

// MusicProviderCredentialPurpose separates provider credentials from other encrypted values.
func MusicProviderCredentialPurpose(provider MusicProvider) string {
	return "music-provider-connection:" + string(provider)
}
