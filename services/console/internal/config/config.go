// Package config loads the Console BFF runtime wiring.
package config

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pood1e/realtime-me/libs/go/authn"
	"github.com/pood1e/realtime-me/libs/go/serviceauth"
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
	InternalAPIKey   serviceauth.Key
	StatusUpstream   *url.URL
	LibraryUpstream  *url.URL
	ManagerUpstream  *url.URL
}

// Load reads and validates Console configuration from the environment.
func Load() (Config, error) {
	managementPrefixes, err := parseManagementPrefixes(os.Getenv("MANAGEMENT_CIDRS"))
	if err != nil {
		return Config{}, err
	}
	publicOrigin, err := parsePublicOrigin("CONSOLE_PUBLIC_ORIGIN", os.Getenv("CONSOLE_PUBLIC_ORIGIN"))
	if err != nil {
		return Config{}, err
	}
	statusUpstream, err := parseUpstream("STATUS_UPSTREAM", os.Getenv("STATUS_UPSTREAM"), managementPrefixes)
	if err != nil {
		return Config{}, err
	}
	libraryUpstream, err := parseUpstream("LIBRARY_UPSTREAM", os.Getenv("LIBRARY_UPSTREAM"), managementPrefixes)
	if err != nil {
		return Config{}, err
	}
	managerUpstream, err := parseUpstream("MANAGER_UPSTREAM", os.Getenv("MANAGER_UPSTREAM"), managementPrefixes)
	if err != nil {
		return Config{}, err
	}
	internalAPIKey, err := serviceauth.LoadFile(strings.TrimSpace(os.Getenv("INTERNAL_API_KEY_FILE")))
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
		InternalAPIKey:   internalAPIKey,
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
		listenAddress, err := netip.ParseAddrPort(config.ListenAddress)
		if err != nil || !listenAddress.Addr().IsLoopback() {
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

func parseHTTPOrigin(name, value string) (*url.URL, error) {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return nil, fmt.Errorf("%s must be an HTTP origin", name)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("%s must use HTTP or HTTPS", name)
	}
	parsed.Path = ""
	return parsed, nil
}

func parsePublicOrigin(name, value string) (*url.URL, error) {
	parsed, err := parseHTTPOrigin(name, value)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "https" && !loopbackHost(parsed.Hostname()) {
		return nil, fmt.Errorf("%s must use HTTPS outside loopback development", name)
	}
	return parsed, nil
}

func parseUpstream(name, value string, managementPrefixes []netip.Prefix) (*url.URL, error) {
	parsed, err := parseHTTPOrigin(name, value)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "https" || loopbackHost(parsed.Hostname()) {
		return parsed, nil
	}
	address, err := netip.ParseAddr(parsed.Hostname())
	if err != nil {
		return nil, fmt.Errorf("%s HTTP host must be a literal management address", name)
	}
	for _, prefix := range managementPrefixes {
		if prefix.Contains(address) {
			return parsed, nil
		}
	}
	return nil, fmt.Errorf("%s HTTP host is outside MANAGEMENT_CIDRS", name)
}

func parseManagementPrefixes(value string) ([]netip.Prefix, error) {
	values := strings.FieldsFunc(value, func(character rune) bool {
		return character == ',' || character == ' ' || character == '\t' || character == '\n'
	})
	if len(values) == 0 {
		return nil, errors.New("MANAGEMENT_CIDRS is required")
	}
	privateNetworks := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("fc00::/7"),
	}
	prefixes := make([]netip.Prefix, 0, len(values))
	seen := make(map[netip.Prefix]struct{}, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("MANAGEMENT_CIDRS contains invalid prefix %q", value)
		}
		prefix = prefix.Masked()
		private := false
		for _, network := range privateNetworks {
			if prefix.Addr().BitLen() == network.Addr().BitLen() && prefix.Bits() >= network.Bits() && network.Contains(prefix.Addr()) {
				private = true
				break
			}
		}
		if !private {
			return nil, fmt.Errorf("MANAGEMENT_CIDRS prefix %q must be inside a private network", value)
		}
		if _, duplicate := seen[prefix]; duplicate {
			continue
		}
		seen[prefix] = struct{}{}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address, err := netip.ParseAddr(host)
	return err == nil && address.IsLoopback()
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
