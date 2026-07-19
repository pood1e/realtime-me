package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pood1e/realtime-me/libs/go/authn"
	"github.com/pood1e/realtime-me/libs/go/serviceauth"
)

const (
	defaultListenAddr       = ":8080"
	defaultChunkSizeBytes   = 16 << 20
	defaultReservedFreeByte = 20 << 30
	defaultUploadTTL        = 24 * time.Hour
)

// Config contains validated API runtime settings.
type Config struct {
	DatabaseURL           string
	DataRoot              string
	ListenAddr            string
	PrivateAPIHost        string
	PublicAPIHost         string
	PublicSiteOrigin      string
	ConsoleOrigin         string
	OIDCIssuer            string
	OIDCAudience          string
	InternalAPIKey        serviceauth.Key
	ProviderCredentialKey []byte
	SpotifyClientID       string
	SpotifyClientSecret   string
	ChunkSizeBytes        int64
	ReservedFreeBytes     int64
	UploadTTL             time.Duration
}

// WorkerConfig contains only settings required by the local processing worker.
type WorkerConfig struct {
	DatabaseURL           string
	DataRoot              string
	ProviderCredentialKey []byte
	ReservedFreeBytes     int64
	SpotifyClientID       string
	SpotifyClientSecret   string
	SpotifyRedirectURI    string
}

// MigrationConfig contains only settings required by schema and content migrations.
type MigrationConfig struct {
	DatabaseURL string
	DataRoot    string
}

// Load reads and validates API configuration.
func Load() (Config, error) {
	config := Config{
		DatabaseURL:         strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DataRoot:            strings.TrimSpace(os.Getenv("DATA_ROOT")),
		ListenAddr:          valueOrDefault("LISTEN_ADDR", defaultListenAddr),
		PrivateAPIHost:      normalizeHost(os.Getenv("PRIVATE_API_HOST")),
		PublicAPIHost:       normalizeHost(os.Getenv("PUBLIC_API_HOST")),
		PublicSiteOrigin:    trimTrailingSlash(os.Getenv("PUBLIC_SITE_ORIGIN")),
		ConsoleOrigin:       trimTrailingSlash(os.Getenv("CONSOLE_ORIGIN")),
		OIDCIssuer:          trimTrailingSlash(os.Getenv("OIDC_ISSUER")),
		OIDCAudience:        strings.TrimSpace(os.Getenv("LIBRARY_AUTH_AUDIENCE")),
		SpotifyClientID:     strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID")),
		SpotifyClientSecret: strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_SECRET")),
		ChunkSizeBytes:      defaultChunkSizeBytes,
		ReservedFreeBytes:   defaultReservedFreeByte,
		UploadTTL:           defaultUploadTTL,
	}
	providerCredentialKey, err := decodeProviderCredentialKey(os.Getenv("MUSIC_PROVIDER_CREDENTIAL_KEY"))
	if err != nil {
		return Config{}, err
	}
	config.ProviderCredentialKey = providerCredentialKey
	internalAPIKey, err := serviceauth.LoadFile(os.Getenv("INTERNAL_API_KEY_FILE"))
	if err != nil {
		return Config{}, err
	}
	config.InternalAPIKey = internalAPIKey
	if err := applyNumericOverrides(&config); err != nil {
		return Config{}, err
	}
	if config.DatabaseURL == "" || config.DataRoot == "" {
		return Config{}, errors.New("DATABASE_URL and DATA_ROOT are required")
	}
	if config.PrivateAPIHost == "" || config.PublicAPIHost == "" || config.PrivateAPIHost == config.PublicAPIHost {
		return Config{}, errors.New("PRIVATE_API_HOST and PUBLIC_API_HOST must be distinct non-empty hosts")
	}
	if err := validateHost("PRIVATE_API_HOST", config.PrivateAPIHost); err != nil {
		return Config{}, err
	}
	if err := validateHost("PUBLIC_API_HOST", config.PublicAPIHost); err != nil {
		return Config{}, err
	}
	if err := validateOrigin("PUBLIC_SITE_ORIGIN", config.PublicSiteOrigin); err != nil {
		return Config{}, err
	}
	if err := validateOrigin("CONSOLE_ORIGIN", config.ConsoleOrigin); err != nil {
		return Config{}, err
	}
	if err := config.AccessConfig().Validate(); err != nil {
		return Config{}, err
	}
	if err := validateSpotifyConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

// LoadWorker reads worker-only configuration.
func LoadWorker() (WorkerConfig, error) {
	config := WorkerConfig{
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")), DataRoot: strings.TrimSpace(os.Getenv("DATA_ROOT")),
		ReservedFreeBytes: defaultReservedFreeByte, SpotifyClientID: strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID")),
		SpotifyClientSecret: strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_SECRET")),
	}
	if config.DatabaseURL == "" || config.DataRoot == "" {
		return WorkerConfig{}, errors.New("DATABASE_URL and DATA_ROOT are required")
	}
	providerCredentialKey, err := decodeProviderCredentialKey(os.Getenv("MUSIC_PROVIDER_CREDENTIAL_KEY"))
	if err != nil {
		return WorkerConfig{}, err
	}
	config.ProviderCredentialKey = providerCredentialKey
	consoleOrigin := trimTrailingSlash(os.Getenv("CONSOLE_ORIGIN"))
	if err := validateOrigin("CONSOLE_ORIGIN", consoleOrigin); err != nil {
		return WorkerConfig{}, err
	}
	spotifyConfigured := config.SpotifyClientID != "" || config.SpotifyClientSecret != ""
	if spotifyConfigured {
		if config.SpotifyClientID == "" || config.SpotifyClientSecret == "" {
			return WorkerConfig{}, errors.New("Spotify worker configuration requires client ID and client secret")
		}
		config.SpotifyRedirectURI = spotifyRedirectURI(consoleOrigin)
	}
	if raw := strings.TrimSpace(os.Getenv("RESERVED_FREE_BYTES")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return WorkerConfig{}, errors.New("RESERVED_FREE_BYTES must be a non-negative integer")
		}
		config.ReservedFreeBytes = value
	}
	return config, nil
}

// LoadMigration reads migration-only configuration without requiring application secrets.
func LoadMigration() (MigrationConfig, error) {
	config := MigrationConfig{
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DataRoot:    strings.TrimSpace(os.Getenv("DATA_ROOT")),
	}
	if config.DatabaseURL == "" || config.DataRoot == "" {
		return MigrationConfig{}, errors.New("DATABASE_URL and DATA_ROOT are required")
	}
	return config, nil
}

// SpotifyRedirectURI returns the Console-owned provider callback URL.
func (c Config) SpotifyRedirectURI() string { return spotifyRedirectURI(c.ConsoleOrigin) }

// AccessConfig returns the human OIDC trust boundary.
func (c Config) AccessConfig() authn.Config {
	return authn.Config{Issuer: c.OIDCIssuer, Audience: c.OIDCAudience}
}

func applyNumericOverrides(config *Config) error {
	if raw := strings.TrimSpace(os.Getenv("UPLOAD_CHUNK_SIZE_BYTES")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 || value > defaultChunkSizeBytes {
			return errors.New("UPLOAD_CHUNK_SIZE_BYTES must be between 1 and 16777216")
		}
		config.ChunkSizeBytes = value
	}
	if raw := strings.TrimSpace(os.Getenv("RESERVED_FREE_BYTES")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return errors.New("RESERVED_FREE_BYTES must be a non-negative integer")
		}
		config.ReservedFreeBytes = value
	}
	if raw := strings.TrimSpace(os.Getenv("UPLOAD_TTL_HOURS")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 1 || value > 168 {
			return errors.New("UPLOAD_TTL_HOURS must be between 1 and 168")
		}
		config.UploadTTL = time.Duration(value) * time.Hour
	}
	return nil
}

func decodeProviderCredentialKey(value string) ([]byte, error) {
	encoded := strings.TrimSpace(value)
	decoded, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(decoded) != 32 {
		return nil, errors.New("MUSIC_PROVIDER_CREDENTIAL_KEY must be padded Base64 containing exactly 32 bytes")
	}
	return decoded, nil
}

func validateSpotifyConfig(config Config) error {
	configured := config.SpotifyClientID != "" || config.SpotifyClientSecret != ""
	if configured && (config.SpotifyClientID == "" || config.SpotifyClientSecret == "") {
		return errors.New("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be configured together")
	}
	if !configured {
		return nil
	}
	return nil
}

func spotifyRedirectURI(consoleOrigin string) string {
	return consoleOrigin + "/api/library/v1/music/providers/spotify/callback"
}

func valueOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func normalizeHost(value string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
}

func trimTrailingSlash(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func validateOrigin(name, value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must contain only an HTTPS origin", name)
	}
	return nil
}

func validateHost(name, value string) error {
	parsed, err := url.Parse("https://" + value)
	if err != nil || parsed.Host != value || parsed.Hostname() == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must contain only one HTTP host", name)
	}
	return nil
}
