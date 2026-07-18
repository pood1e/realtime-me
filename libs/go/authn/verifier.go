// Package authn verifies OIDC JWTs at service boundaries.
package authn

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	authv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/auth/v1"
)

const discoveryTimeout = 10 * time.Second

// ErrUnauthenticated means that no valid identity was presented.
var ErrUnauthenticated = errors.New("authentication required")

// ErrPermissionDenied means that an authenticated identity lacks a capability.
var ErrPermissionDenied = errors.New("permission denied")

// ErrUnavailable means that the configured identity provider cannot establish trust.
var ErrUnavailable = errors.New("identity service unavailable")

// Config identifies one OIDC trust boundary and token audience.
type Config struct {
	Issuer   string
	Audience string
}

// Validate rejects an incomplete trust boundary.
func (config Config) Validate() error {
	issuer, err := url.Parse(strings.TrimSpace(config.Issuer))
	if err != nil || issuer.Host == "" || issuer.User != nil || issuer.RawQuery != "" || issuer.Fragment != "" {
		return errors.New("OIDC issuer must be an HTTP origin or issuer URL without credentials, query, or fragment")
	}
	if issuer.Scheme != "https" && !(issuer.Scheme == "http" && loopbackHost(issuer.Hostname())) {
		return errors.New("OIDC issuer must use HTTPS outside loopback development")
	}
	audience := strings.TrimSpace(config.Audience)
	if audience == "" || strings.ContainsAny(audience, " \t\r\n") {
		return errors.New("OIDC audience must be one non-empty token")
	}
	return nil
}

func loopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || net.ParseIP(host).IsLoopback()
}

// Principal is the verified identity projected from one access token.
type Principal struct {
	Subject     string
	DisplayName string
	ExpireTime  time.Time
	Permissions []authv1.Permission
}

// Has reports whether the principal carries a permission.
func (principal Principal) Has(required authv1.Permission) bool {
	for _, permission := range principal.Permissions {
		if permission == required {
			return true
		}
	}
	return false
}

// Verifier validates signed OIDC JWTs and their permission claim.
type Verifier struct {
	config Config
	client *http.Client
	mu     sync.Mutex
	tokens *oidc.IDTokenVerifier
}

// NewVerifier validates configuration. Discovery is lazy so public and workload paths do not
// acquire an availability dependency on the human identity provider.
func NewVerifier(config Config) (*Verifier, error) {
	config.Issuer = strings.TrimRight(strings.TrimSpace(config.Issuer), "/")
	config.Audience = strings.TrimSpace(config.Audience)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Verifier{config: config, client: &http.Client{
		Timeout: discoveryTimeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          16,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: time.Second,
			TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}}, nil
}

// Authenticate verifies a bearer token and enforces one required permission.
func (verifier *Verifier) Authenticate(ctx context.Context, authorization string, required authv1.Permission) (Principal, error) {
	rawToken, ok := bearerToken(authorization)
	if !ok {
		return Principal{}, ErrUnauthenticated
	}
	tokens, err := verifier.tokenVerifier(ctx)
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	token, err := tokens.Verify(ctx, rawToken)
	if err != nil {
		if strings.Contains(err.Error(), "failed to verify signature: fetching keys ") {
			return Principal{}, ErrUnavailable
		}
		return Principal{}, ErrUnauthenticated
	}
	var claims struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := token.Claims(&claims); err != nil || strings.TrimSpace(token.Subject) == "" {
		return Principal{}, ErrUnauthenticated
	}
	principal := Principal{
		Subject:     strings.TrimSpace(token.Subject),
		DisplayName: strings.TrimSpace(claims.Name),
		ExpireTime:  token.Expiry.UTC(),
		Permissions: ParsePermissionNames(claims.Permissions),
	}
	if required != authv1.Permission_PERMISSION_UNSPECIFIED && !principal.Has(required) {
		return Principal{}, ErrPermissionDenied
	}
	return principal, nil
}

func (verifier *Verifier) tokenVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	verifier.mu.Lock()
	defer verifier.mu.Unlock()
	if verifier.tokens != nil {
		return verifier.tokens, nil
	}
	discoveryContext, cancel := context.WithTimeout(oidc.ClientContext(ctx, verifier.client), discoveryTimeout)
	defer cancel()
	provider, err := oidc.NewProvider(discoveryContext, strings.TrimRight(verifier.config.Issuer, "/"))
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	verifier.tokens = provider.Verifier(&oidc.Config{ClientID: verifier.config.Audience})
	return verifier.tokens, nil
}

func bearerToken(authorization string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return "", false
	}
	token := strings.TrimPrefix(authorization, prefix)
	return token, token != "" && !strings.ContainsAny(token, " \t\r\n")
}

// ParsePermissionNames converts the canonical JWT claim names into generated values.
func ParsePermissionNames(names []string) []authv1.Permission {
	result := make([]authv1.Permission, 0, len(names))
	seen := make(map[authv1.Permission]struct{}, len(names))
	for _, name := range names {
		value, found := authv1.Permission_value[name]
		permission := authv1.Permission(value)
		if !found || permission == authv1.Permission_PERMISSION_UNSPECIFIED {
			continue
		}
		if _, duplicate := seen[permission]; duplicate {
			continue
		}
		seen[permission] = struct{}{}
		result = append(result, permission)
	}
	return result
}
