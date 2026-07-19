// Package serviceauth authenticates trusted management-plane requests with a
// deployment-scoped key. It is an additional service boundary, not a human
// identity mechanism.
package serviceauth

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	// Header carries the internal management-plane credential.
	Header  = "X-Realtime-Internal-Key"
	keySize = 32
)

// Key is a validated 256-bit management-plane credential.
type Key struct {
	value [keySize]byte
	valid bool
}

// LoadFile reads a 64-character hexadecimal key. A single trailing newline is
// accepted so `openssl rand -hex 32 > file` produces a valid credential.
func LoadFile(path string) (Key, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Key{}, errors.New("internal API key file is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Key{}, fmt.Errorf("read internal API key: %w", err)
	}
	if len(data) == keySize*2+1 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	if len(data) != keySize*2 {
		return Key{}, errors.New("internal API key must contain exactly 32 hexadecimal bytes")
	}
	var key Key
	if _, err := hex.Decode(key.value[:], data); err != nil {
		return Key{}, errors.New("internal API key must contain exactly 32 hexadecimal bytes")
	}
	key.valid = true
	return key, nil
}

// HeaderValue returns the canonical value injected by a trusted proxy.
func (key Key) HeaderValue() string {
	if !key.valid {
		return ""
	}
	return hex.EncodeToString(key.value[:])
}

// Matches compares a presented key in constant time after strict decoding.
func (key Key) Matches(presented string) bool {
	if !key.valid || len(presented) != keySize*2 {
		return false
	}
	var decoded [keySize]byte
	if _, err := hex.Decode(decoded[:], []byte(presented)); err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(key.value[:], decoded[:]) == 1
}
