// Package config loads the Console BFF runtime wiring.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pood1e/realtime-me/libs/go/authn"
)

// Config contains the Console's OIDC, static asset, and upstream wiring.
type Config struct {
	ListenAddress    string
	PublicOrigin     *url.URL
	WebDirectory     string
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCScopes       []string
	StatusUpstream   *url.URL
	LibraryUpstream  *url.URL
	ManagerUpstream  *url.URL
}

// Load reads and validates Console configuration from the environment.
func Load() (Config, error) {
	publicOrigin, err := parseOrigin("CONSOLE_PUBLIC_ORIGIN", os.Getenv("CONSOLE_PUBLIC_ORIGIN"))
	if err != nil {
		return Config{}, err
	}
	statusUpstream, err := parseOrigin("STATUS_UPSTREAM", os.Getenv("STATUS_UPSTREAM"))
	if err != nil {
		return Config{}, err
	}
	libraryUpstream, err := parseOrigin("LIBRARY_UPSTREAM", os.Getenv("LIBRARY_UPSTREAM"))
	if err != nil {
		return Config{}, err
	}
	managerUpstream, err := parseOrigin("MANAGER_UPSTREAM", os.Getenv("MANAGER_UPSTREAM"))
	if err != nil {
		return Config{}, err
	}
	config := Config{
		ListenAddress:    valueOrDefault("CONSOLE_LISTEN_ADDRESS", "127.0.0.1:8090"),
		PublicOrigin:     publicOrigin,
		WebDirectory:     strings.TrimRight(strings.TrimSpace(os.Getenv("CONSOLE_WEB_DIR")), "/"),
		OIDCIssuer:       strings.TrimRight(strings.TrimSpace(os.Getenv("OIDC_ISSUER")), "/"),
		OIDCClientID:     strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID")),
		OIDCClientSecret: strings.TrimSpace(os.Getenv("OIDC_CLIENT_SECRET")),
		OIDCScopes:       scopes(os.Getenv("OIDC_SCOPES")),
		StatusUpstream:   statusUpstream,
		LibraryUpstream:  libraryUpstream,
		ManagerUpstream:  managerUpstream,
	}
	if config.OIDCClientSecret == "" {
		return Config{}, errors.New("OIDC_CLIENT_SECRET is required")
	}
	if err := (authn.Config{Issuer: config.OIDCIssuer, Audience: config.OIDCClientID}).Validate(); err != nil {
		return Config{}, err
	}
	if config.SecureCookies() {
		if config.WebDirectory == "" || !filepath.IsAbs(config.WebDirectory) {
			return Config{}, errors.New("CONSOLE_WEB_DIR must be an absolute path in production")
		}
		host, _, err := net.SplitHostPort(config.ListenAddress)
		if err != nil || !loopbackHost(host) {
			return Config{}, errors.New("CONSOLE_LISTEN_ADDRESS must use a loopback host in production")
		}
	}
	return config, nil
}

// CallbackURL returns the registered OIDC redirect URL.
func (config Config) CallbackURL() string {
	return config.PublicOrigin.ResolveReference(&url.URL{Path: "/auth/callback"}).String()
}

// SecureCookies reports whether cookies must use their production host-only name.
func (config Config) SecureCookies() bool {
	return config.PublicOrigin.Scheme == "https"
}

func parseOrigin(name, value string) (*url.URL, error) {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return nil, fmt.Errorf("%s must be an HTTP origin", name)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("%s must use HTTP or HTTPS", name)
	}
	if parsed.Scheme != "https" && !loopbackHost(parsed.Hostname()) {
		return nil, fmt.Errorf("%s must use HTTPS outside loopback development", name)
	}
	parsed.Path = ""
	return parsed, nil
}

func loopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || net.ParseIP(host).IsLoopback()
}

func scopes(value string) []string {
	result := []string{"openid", "profile"}
	seen := map[string]struct{}{"openid": {}, "profile": {}}
	for _, scope := range strings.Fields(value) {
		if _, duplicate := seen[scope]; duplicate {
			continue
		}
		seen[scope] = struct{}{}
		result = append(result, scope)
	}
	return result
}

func valueOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
