// Package provider contains application-neutral support for external music providers.
package provider

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const secretEnvelopeVersion byte = 1

// SecretBox encrypts provider credentials with authenticated, purpose-bound envelopes.
type SecretBox struct {
	aead cipher.AEAD
	rand io.Reader
}

// NewSecretBox constructs an AES-256-GCM credential protector.
func NewSecretBox(key []byte) (*SecretBox, error) {
	if len(key) != 32 {
		return nil, errors.New("provider credential key must contain exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create provider credential cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create provider credential protector: %w", err)
	}
	return &SecretBox{aead: aead, rand: rand.Reader}, nil
}

// Seal encrypts plaintext and binds it to a stable storage purpose.
func (b *SecretBox) Seal(purpose string, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(b.rand, nonce); err != nil {
		return nil, fmt.Errorf("generate provider credential nonce: %w", err)
	}
	envelope := make([]byte, 1+len(nonce), 1+len(nonce)+len(plaintext)+b.aead.Overhead())
	envelope[0] = secretEnvelopeVersion
	copy(envelope[1:], nonce)
	return b.aead.Seal(envelope, nonce, plaintext, []byte(purpose)), nil
}

// Open authenticates and decrypts one purpose-bound envelope.
func (b *SecretBox) Open(purpose string, envelope []byte) ([]byte, error) {
	nonceSize := b.aead.NonceSize()
	if len(envelope) < 1+nonceSize+b.aead.Overhead() || envelope[0] != secretEnvelopeVersion {
		return nil, errors.New("invalid provider credential envelope")
	}
	plaintext, err := b.aead.Open(nil, envelope[1:1+nonceSize], envelope[1+nonceSize:], []byte(purpose))
	if err != nil {
		return nil, errors.New("decrypt provider credential envelope")
	}
	return plaintext, nil
}
