package config

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultListenAddr       = ":8080"
	defaultChunkSizeBytes   = 16 << 20
	defaultReservedFreeByte = 20 << 30
	defaultUploadTTL        = 24 * time.Hour
)

// Config contains validated API runtime settings.
type Config struct {
	DatabaseURL       string
	DataRoot          string
	ListenAddr        string
	PrivateAPIHost    string
	PublicAPIHost     string
	PrivateAppOrigins map[string]struct{}
	PublicAppOrigins  map[string]struct{}
	ShareAppOrigin    string
	PasswordHash      []byte
	SessionSecret     []byte
	ChunkSizeBytes    int64
	ReservedFreeBytes int64
	UploadTTL         time.Duration
}

// WorkerConfig contains only settings required by the local processing worker.
type WorkerConfig struct {
	DatabaseURL string
	DataRoot    string
}

// Load reads and validates API configuration.
func Load() (Config, error) {
	passwordHash, err := decodePasswordHash(os.Getenv("PASSWORD_HASH_BASE64"))
	if err != nil {
		return Config{}, err
	}
	sessionSecret, err := decodeSessionSecret(os.Getenv("SESSION_SECRET"))
	if err != nil {
		return Config{}, err
	}
	privateOrigins, err := parseOrigins("PRIVATE_APP_ORIGINS", os.Getenv("PRIVATE_APP_ORIGINS"))
	if err != nil {
		return Config{}, err
	}
	publicOrigins, err := parseOrigins("PUBLIC_APP_ORIGINS", os.Getenv("PUBLIC_APP_ORIGINS"))
	if err != nil {
		return Config{}, err
	}
	config := Config{
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DataRoot:          strings.TrimSpace(os.Getenv("DATA_ROOT")),
		ListenAddr:        valueOrDefault("LISTEN_ADDR", defaultListenAddr),
		PrivateAPIHost:    normalizeHost(os.Getenv("PRIVATE_API_HOST")),
		PublicAPIHost:     normalizeHost(os.Getenv("PUBLIC_API_HOST")),
		PrivateAppOrigins: privateOrigins,
		PublicAppOrigins:  publicOrigins,
		ShareAppOrigin:    trimTrailingSlash(os.Getenv("SHARE_APP_ORIGIN")),
		PasswordHash:      passwordHash,
		SessionSecret:     sessionSecret,
		ChunkSizeBytes:    defaultChunkSizeBytes,
		ReservedFreeBytes: defaultReservedFreeByte,
		UploadTTL:         defaultUploadTTL,
	}
	if err := applyNumericOverrides(&config); err != nil {
		return Config{}, err
	}
	if config.DatabaseURL == "" || config.DataRoot == "" {
		return Config{}, errors.New("DATABASE_URL and DATA_ROOT are required")
	}
	if config.PrivateAPIHost == "" || config.PublicAPIHost == "" || config.PrivateAPIHost == config.PublicAPIHost {
		return Config{}, errors.New("PRIVATE_API_HOST and PUBLIC_API_HOST must be distinct non-empty hosts")
	}
	if err := validateOrigin("SHARE_APP_ORIGIN", config.ShareAppOrigin); err != nil {
		return Config{}, err
	}
	if _, found := config.PublicAppOrigins[config.ShareAppOrigin]; !found {
		return Config{}, errors.New("SHARE_APP_ORIGIN must be present in PUBLIC_APP_ORIGINS")
	}
	return config, nil
}

// LoadWorker reads worker-only configuration.
func LoadWorker() (WorkerConfig, error) {
	config := WorkerConfig{DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")), DataRoot: strings.TrimSpace(os.Getenv("DATA_ROOT"))}
	if config.DatabaseURL == "" || config.DataRoot == "" {
		return WorkerConfig{}, errors.New("DATABASE_URL and DATA_ROOT are required")
	}
	return config, nil
}

// PublicAPIOrigin returns the canonical externally visible API origin.
func (c Config) PublicAPIOrigin() string { return "https://" + c.PublicAPIHost }

// ReturnURL validates an authentication return URL against private application origins.
func (c Config) ReturnURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return "", errors.New("invalid return URL")
	}
	origin := parsed.Scheme + "://" + parsed.Host
	if _, allowed := c.PrivateAppOrigins[origin]; !allowed {
		return "", errors.New("return URL origin is not allowed")
	}
	return parsed.String(), nil
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

func parseOrigins(name, value string) (map[string]struct{}, error) {
	origins := make(map[string]struct{})
	for _, candidate := range strings.Split(value, ",") {
		candidate = trimTrailingSlash(candidate)
		if candidate == "" {
			continue
		}
		if err := validateOrigin(name, candidate); err != nil {
			return nil, err
		}
		origins[candidate] = struct{}{}
	}
	if len(origins) == 0 {
		return nil, fmt.Errorf("%s must contain at least one HTTPS origin", name)
	}
	return origins, nil
}

func decodePasswordHash(value string) ([]byte, error) {
	encoded := strings.TrimSpace(value)
	decoded, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if encoded == "" || err != nil || len(decoded) == 0 {
		return nil, errors.New("PASSWORD_HASH_BASE64 must be valid padded Base64")
	}
	return decoded, nil
}

func decodeSessionSecret(value string) ([]byte, error) {
	encoded := strings.TrimSpace(value)
	decoded, err := hex.DecodeString(encoded)
	if len(encoded) < 64 || err != nil || len(decoded) < 32 {
		return nil, errors.New("SESSION_SECRET must contain at least 64 hexadecimal characters")
	}
	return decoded, nil
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
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must contain only an HTTPS origin", name)
	}
	return nil
}
