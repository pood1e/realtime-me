package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	cookieName             = "__Host-cloud-drive-session"
	sessionDuration        = 24 * time.Hour
	sessionNonceBytes      = 32
	maximumPasswordBytes   = 72
	bcryptConcurrency      = 2
	clientRateWindow       = time.Minute
	clientRateLimit        = 5
	maximumClientEntries   = 1024
	unknownClientKey       = "unknown"
	minimumBcryptCost      = bcrypt.DefaultCost
	minimumSigningKeyBytes = 32
)

// ErrUnauthenticated is the single externally visible authentication failure.
var ErrUnauthenticated = errors.New("authentication required")

// Clock supplies wall-clock time to the session manager.
type Clock interface {
	Now() time.Time
}

// Session is the verified lifetime of a browser session.
type Session struct {
	ExpireTime time.Time
}

// Manager verifies the owner password and issues signed session cookies.
type Manager struct {
	passwordHash []byte
	signingKey   []byte
	clock        Clock
	random       io.Reader
	bcryptSlots  chan struct{}
	clients      clientRateLimiter
}

// NewManager validates authentication material and constructs a session manager.
func NewManager(passwordHash, signingKey []byte, clock Clock, random io.Reader) (*Manager, error) {
	cost, err := bcrypt.Cost(passwordHash)
	if err != nil || cost < minimumBcryptCost {
		return nil, fmt.Errorf("password hash must be bcrypt with cost %d or greater", minimumBcryptCost)
	}
	if len(signingKey) < minimumSigningKeyBytes {
		return nil, fmt.Errorf("session signing key must contain at least %d bytes", minimumSigningKeyBytes)
	}
	if clock == nil || random == nil {
		return nil, errors.New("session clock and random source are required")
	}
	return &Manager{
		passwordHash: append([]byte(nil), passwordHash...),
		signingKey:   append([]byte(nil), signingKey...),
		clock:        clock,
		random:       random,
		bcryptSlots:  make(chan struct{}, bcryptConcurrency),
		clients: clientRateLimiter{
			entries: make(map[string]clientRateEntry, maximumClientEntries),
		},
	}, nil
}

// Login verifies one password attempt and returns a new session cookie.
func (manager *Manager) Login(password, clientAddress string) (*http.Cookie, error) {
	if !manager.clients.allow(clientKey(clientAddress), manager.clock.Now().UTC()) {
		return nil, ErrUnauthenticated
	}
	if !manager.passwordMatches(password) {
		return nil, ErrUnauthenticated
	}

	now := manager.clock.Now().UTC()
	nonce := make([]byte, sessionNonceBytes)
	if _, err := io.ReadFull(manager.random, nonce); err != nil {
		return nil, fmt.Errorf("generate session nonce: %w", err)
	}
	expireTime := now.Add(sessionDuration)
	payload := strings.Join([]string{
		"v1",
		strconv.FormatInt(expireTime.Unix(), 10),
		base64.RawURLEncoding.EncodeToString(nonce),
	}, ".")
	token := payload + "." + base64.RawURLEncoding.EncodeToString(manager.signature(payload))
	return sessionCookie(token, expireTime), nil
}

func (manager *Manager) passwordMatches(password string) bool {
	select {
	case manager.bcryptSlots <- struct{}{}:
	default:
		return false
	}
	defer func() { <-manager.bcryptSlots }()

	passwordBytes := []byte(password)
	validLength := len(passwordBytes) <= maximumPasswordBytes
	candidate := passwordBytes
	if !validLength {
		candidate = passwordBytes[:maximumPasswordBytes]
	}
	passwordValid := bcrypt.CompareHashAndPassword(manager.passwordHash, candidate) == nil && validLength
	clear(passwordBytes)
	return passwordValid
}

// Authenticate verifies the signed session cookie on an HTTP request.
func (manager *Manager) Authenticate(request *http.Request) (Session, error) {
	token := ""
	if cookie, err := request.Cookie(cookieName); err == nil {
		token = cookie.Value
	}
	return manager.validateToken(token)
}

// LogoutCookie clears the host-only session cookie.
func (manager *Manager) LogoutCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (manager *Manager) validateToken(token string) (Session, error) {
	parts := strings.Split(token, ".")
	payload := token
	signatureText := ""
	validStructure := len(parts) == 4
	if validStructure {
		payload = strings.Join(parts[:3], ".")
		signatureText = parts[3]
	}

	expectedSignature := manager.signature(payload)
	providedSignature := make([]byte, sha256.Size)
	decodedSignature, signatureErr := base64.RawURLEncoding.DecodeString(signatureText)
	validSignatureEncoding := signatureErr == nil && len(decodedSignature) == sha256.Size
	copy(providedSignature, decodedSignature)
	validSignature := hmac.Equal(expectedSignature, providedSignature)

	if !validStructure || !validSignatureEncoding || !validSignature || parts[0] != "v1" {
		return Session{}, ErrUnauthenticated
	}
	expireUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return Session{}, ErrUnauthenticated
	}
	nonce, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(nonce) != sessionNonceBytes {
		return Session{}, ErrUnauthenticated
	}
	expireTime := time.Unix(expireUnix, 0).UTC()
	now := manager.clock.Now().UTC()
	if !now.Before(expireTime) || expireTime.After(now.Add(sessionDuration)) {
		return Session{}, ErrUnauthenticated
	}
	return Session{ExpireTime: expireTime}, nil
}

func (manager *Manager) signature(payload string) []byte {
	mac := hmac.New(sha256.New, manager.signingKey)
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}

func sessionCookie(token string, expireTime time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		Expires:  expireTime,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

type sessionContextKey struct{}

// ContextWithSession attaches one authenticated session to a request context.
func ContextWithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, session)
}

// SessionFromContext returns the authenticated session attached by the HTTP boundary.
func SessionFromContext(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(Session)
	return session, ok
}

func clientKey(value string) string {
	address := net.ParseIP(strings.TrimSpace(value))
	if address == nil {
		return unknownClientKey
	}
	return address.String()
}

type clientRateEntry struct {
	attemptTimes [clientRateLimit]time.Time
	attemptCount int
	lastUsed     time.Time
}

func (entry clientRateEntry) expired(now time.Time) bool {
	return now.Before(entry.lastUsed) || now.Sub(entry.lastUsed) >= clientRateWindow
}

func (entry *clientRateEntry) allow(now time.Time) bool {
	if entry.attemptCount > 0 && now.Before(entry.attemptTimes[0]) {
		entry.attemptCount = 0
	}
	retained := 0
	for index := 0; index < entry.attemptCount; index++ {
		attemptTime := entry.attemptTimes[index]
		if now.Sub(attemptTime) < clientRateWindow {
			entry.attemptTimes[retained] = attemptTime
			retained++
		}
	}
	entry.attemptCount = retained
	entry.lastUsed = now
	if entry.attemptCount >= clientRateLimit {
		return false
	}
	entry.attemptTimes[entry.attemptCount] = now
	entry.attemptCount++
	return true
}

type clientRateLimiter struct {
	mu      sync.Mutex
	entries map[string]clientRateEntry
}

func (limiter *clientRateLimiter) allow(key string, now time.Time) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	entry, found := limiter.entries[key]
	if !found {
		if len(limiter.entries) >= maximumClientEntries {
			limiter.evict(now)
		}
		entry = clientRateEntry{}
	}
	allowed := entry.allow(now)
	limiter.entries[key] = entry
	return allowed
}

func (limiter *clientRateLimiter) evict(now time.Time) {
	expiredKey, oldestKey := "", ""
	var expiredUse, oldestUse time.Time
	for key, entry := range limiter.entries {
		if oldestKey == "" || entry.lastUsed.Before(oldestUse) {
			oldestKey = key
			oldestUse = entry.lastUsed
		}
		if entry.expired(now) && (expiredKey == "" || entry.lastUsed.Before(expiredUse)) {
			expiredKey = key
			expiredUse = entry.lastUsed
		}
	}
	if expiredKey != "" {
		delete(limiter.entries, expiredKey)
		return
	}
	delete(limiter.entries, oldestKey)
}
