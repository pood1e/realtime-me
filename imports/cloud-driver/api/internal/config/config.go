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

// Config contains all runtime settings for the API process.
type Config struct {
	DatabaseURL       string
	DataRoot          string
	ListenAddr        string
	PrivateAPIHost    string
	ShareAPIHost      string
	PrivateAppOrigin  string
	ShareAppOrigin    string
	PasswordHash      []byte
	SessionSecret     []byte
	ChunkSizeBytes    int64
	ReservedFreeBytes int64
	UploadTTL         time.Duration
}

// Load reads and validates API configuration from the environment.
func Load() (Config, error) {
	passwordHash, err := decodePasswordHash(os.Getenv("PASSWORD_HASH_BASE64"))
	if err != nil {
		return Config{}, err
	}
	sessionSecret, err := decodeSessionSecret(os.Getenv("SESSION_SECRET"))
	if err != nil {
		return Config{}, err
	}
	config := Config{
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DataRoot:          strings.TrimSpace(os.Getenv("DATA_ROOT")),
		ListenAddr:        valueOrDefault("LISTEN_ADDR", defaultListenAddr),
		PrivateAPIHost:    normalizeHost(os.Getenv("PRIVATE_API_HOST")),
		ShareAPIHost:      normalizeHost(os.Getenv("SHARE_API_HOST")),
		PrivateAppOrigin:  trimTrailingSlash(os.Getenv("PRIVATE_APP_ORIGIN")),
		ShareAppOrigin:    trimTrailingSlash(os.Getenv("SHARE_APP_ORIGIN")),
		PasswordHash:      passwordHash,
		SessionSecret:     sessionSecret,
		ChunkSizeBytes:    defaultChunkSizeBytes,
		ReservedFreeBytes: defaultReservedFreeByte,
		UploadTTL:         defaultUploadTTL,
	}

	if raw := strings.TrimSpace(os.Getenv("UPLOAD_CHUNK_SIZE_BYTES")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 || value > defaultChunkSizeBytes {
			return Config{}, errors.New("UPLOAD_CHUNK_SIZE_BYTES must be between 1 and 16777216")
		}
		config.ChunkSizeBytes = value
	}
	if raw := strings.TrimSpace(os.Getenv("RESERVED_FREE_BYTES")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			return Config{}, errors.New("RESERVED_FREE_BYTES must be a non-negative integer")
		}
		config.ReservedFreeBytes = value
	}
	if raw := strings.TrimSpace(os.Getenv("UPLOAD_TTL_HOURS")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 1 || value > 168 {
			return Config{}, errors.New("UPLOAD_TTL_HOURS must be between 1 and 168")
		}
		config.UploadTTL = time.Duration(value) * time.Hour
	}
	if config.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if config.DataRoot == "" {
		return Config{}, errors.New("DATA_ROOT is required")
	}
	if config.PrivateAPIHost == "" || config.ShareAPIHost == "" {
		return Config{}, errors.New("PRIVATE_API_HOST and SHARE_API_HOST are required")
	}
	if config.PrivateAPIHost == config.ShareAPIHost {
		return Config{}, errors.New("PRIVATE_API_HOST and SHARE_API_HOST must differ")
	}
	if err := validateOrigin("PRIVATE_APP_ORIGIN", config.PrivateAppOrigin); err != nil {
		return Config{}, err
	}
	if err := validateOrigin("SHARE_APP_ORIGIN", config.ShareAppOrigin); err != nil {
		return Config{}, err
	}
	return config, nil
}

func decodePasswordHash(value string) ([]byte, error) {
	encoded := strings.TrimSpace(value)
	if encoded == "" {
		return nil, errors.New("PASSWORD_HASH_BASE64 is required")
	}
	passwordHash, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(passwordHash) == 0 {
		return nil, errors.New("PASSWORD_HASH_BASE64 must be valid padded Base64")
	}
	return passwordHash, nil
}

func decodeSessionSecret(value string) ([]byte, error) {
	encoded := strings.TrimSpace(value)
	if len(encoded) < 64 {
		return nil, errors.New("SESSION_SECRET must contain at least 64 hexadecimal characters")
	}
	sessionSecret, err := hex.DecodeString(encoded)
	if err != nil || len(sessionSecret) < 32 {
		return nil, errors.New("SESSION_SECRET must contain at least 64 hexadecimal characters")
	}
	return sessionSecret, nil
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
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("%s must be an https origin", name)
	}
	return nil
}
