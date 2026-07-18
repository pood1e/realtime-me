// Package session owns bounded, server-side OIDC sessions.
package session

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"golang.org/x/oauth2"

	authv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/auth/v1"
)

const (
	sessionLifetime = 24 * time.Hour
	pendingLifetime = 10 * time.Minute
	maximumSessions = 64
	maximumPending  = 32
)

// ErrUnauthenticated means that the Console session is absent or expired.
var ErrUnauthenticated = errors.New("authentication required")

// PendingLogin is one short-lived authorization-code exchange.
type PendingLogin struct {
	State      string
	Nonce      string
	Verifier   string
	ReturnPath string
	ExpireTime time.Time
}

// Identity contains the claims needed by the Console shell.
type Identity struct {
	Subject     string
	DisplayName string
	Permissions []authv1.Permission
}

// Resolved is an authenticated session and its current access token.
type Resolved struct {
	ID          string
	Identity    Identity
	AccessToken string
	ExpireTime  time.Time
}

type storedSession struct {
	identity   Identity
	tokens     oauth2.TokenSource
	expireTime time.Time
}

// Store keeps a single-instance Console's sessions in memory.
type Store struct {
	mu       sync.Mutex
	sessions map[string]*storedSession
	pending  map[string]PendingLogin
	now      func() time.Time
}

// NewStore constructs an empty bounded session store.
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*storedSession),
		pending:  make(map[string]PendingLogin),
		now:      time.Now,
	}
}

// Begin creates one PKCE and nonce-bound login transaction.
func (store *Store) Begin(returnPath string) (PendingLogin, error) {
	state, err := randomValue()
	if err != nil {
		return PendingLogin{}, err
	}
	nonce, err := randomValue()
	if err != nil {
		return PendingLogin{}, err
	}
	pending := PendingLogin{
		State:      state,
		Nonce:      nonce,
		Verifier:   oauth2.GenerateVerifier(),
		ReturnPath: returnPath,
		ExpireTime: store.now().UTC().Add(pendingLifetime),
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.sweepLocked()
	if len(store.pending) >= maximumPending {
		store.evictEarliestPendingLocked()
	}
	store.pending[state] = pending
	return pending, nil
}

// Consume accepts a state value exactly once.
func (store *Store) Consume(state string) (PendingLogin, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.sweepLocked()
	pending, found := store.pending[state]
	delete(store.pending, state)
	if !found {
		return PendingLogin{}, ErrUnauthenticated
	}
	return pending, nil
}

// NonceMatches compares an ID-token nonce without timing-dependent early exits.
func NonceMatches(expected, provided string) bool {
	if len(expected) != len(provided) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

// Create stores one authenticated OIDC token source and returns its opaque ID.
func (store *Store) Create(identity Identity, tokens oauth2.TokenSource) (string, error) {
	id, err := randomValue()
	if err != nil {
		return "", err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.sweepLocked()
	if len(store.sessions) >= maximumSessions {
		store.evictEarliestSessionLocked()
	}
	store.sessions[id] = &storedSession{
		identity:   identity,
		tokens:     tokens,
		expireTime: store.now().UTC().Add(sessionLifetime),
	}
	return id, nil
}

// Resolve refreshes a token when necessary and projects the current session.
func (store *Store) Resolve(ctx context.Context, id string) (Resolved, error) {
	store.mu.Lock()
	store.sweepLocked()
	current, found := store.sessions[id]
	store.mu.Unlock()
	if !found {
		return Resolved{}, ErrUnauthenticated
	}
	token, err := current.tokens.Token()
	if err != nil || token.AccessToken == "" || token.Expiry.IsZero() || !token.Valid() {
		store.Delete(id)
		return Resolved{}, ErrUnauthenticated
	}
	select {
	case <-ctx.Done():
		return Resolved{}, ctx.Err()
	default:
	}
	return Resolved{
		ID:          id,
		Identity:    current.identity,
		AccessToken: token.AccessToken,
		ExpireTime:  token.Expiry.UTC(),
	}, nil
}

// Delete removes one active session.
func (store *Store) Delete(id string) {
	store.mu.Lock()
	delete(store.sessions, id)
	store.mu.Unlock()
}

func (store *Store) sweepLocked() {
	now := store.now().UTC()
	for state, pending := range store.pending {
		if !now.Before(pending.ExpireTime) {
			delete(store.pending, state)
		}
	}
	for id, current := range store.sessions {
		if !now.Before(current.expireTime) {
			delete(store.sessions, id)
		}
	}
}

func (store *Store) evictEarliestPendingLocked() {
	var selected string
	var earliest time.Time
	for state, pending := range store.pending {
		if selected == "" || pending.ExpireTime.Before(earliest) {
			selected, earliest = state, pending.ExpireTime
		}
	}
	delete(store.pending, selected)
}

func (store *Store) evictEarliestSessionLocked() {
	var selected string
	var earliest time.Time
	for id, current := range store.sessions {
		if selected == "" || current.expireTime.Before(earliest) {
			selected, earliest = id, current.expireTime
		}
	}
	delete(store.sessions, selected)
}

func randomValue() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
