package postgres

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func encodeCursor(name, uid string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(name + "\x00" + uid))
}

func decodeCursor(token string) (*cursor, error) {
	if token == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	parts := strings.Split(string(decoded), "\x00")
	if err != nil || len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("%w: malformed page token", domain.ErrInvalidArgument)
	}
	return &cursor{name: parts[0], uid: parts[1]}, nil
}

func encodeSimpleCursor(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeSimpleCursor(token string) (string, error) {
	if token == "" {
		return "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) == 0 {
		return "", fmt.Errorf("%w: malformed page token", domain.ErrInvalidArgument)
	}
	return string(decoded), nil
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 100
	}
	if pageSize > 200 {
		return 200
	}
	return pageSize
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func copyString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
