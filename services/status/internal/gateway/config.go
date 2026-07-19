package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pood1e/realtime-me/libs/go/authn"
	"go.yaml.in/yaml/v3"
)

// Config is everything the gateway runs on. It is assembled from two places, and the
// line between them is what a person chose versus what the container is.
//
// Settings — the tokens, the GitHub credentials, the owner's profile — come from one
// YAML file, because a person writes them and should have one place to look. Paths,
// ports, and the address of Prometheus come from the environment, because Compose
// decides those and nobody should have to keep them in step by hand.
type Config struct {
	// The container: from the environment.
	Port              string
	StateFile         string
	IdentityStateFile string
	PrometheusURL     string
	ProjectsFile      string
	OIDCIssuer        string
	OIDCAudience      string

	// The settings: from the YAML.
	IngestTokens                   map[string]struct{}
	DiscoveryTokens                map[string]struct{}
	GitHubToken                    string
	GitHubStatusMinIntervalSeconds int
	GitHubStatusTTLMinutes         int
	GitHubProjectsTokens           []string
	GitHubProjectsRefreshHours     int
	PublicAgentPlaceholder         bool
	Profile                        ConfiguredProfile
	ScrapeTargets                  ScrapeTargetPolicy
}

// Settings is the hand-written half, exactly as it appears in the YAML.
type Settings struct {
	Tokens                 TokenSettings   `yaml:"tokens"`
	GitHub                 GitHubSettings  `yaml:"github"`
	Probe                  ProbeSettings   `yaml:"probe"`
	Profile                ProfileSettings `yaml:"profile"`
	PublicAgentPlaceholder bool            `yaml:"public_agent_placeholder"`
}

// TokenSettings are workload credentials for ingest and Prometheus discovery.
type TokenSettings struct {
	Ingest    string `yaml:"ingest"`
	Discovery string `yaml:"discovery"`
}

// ProbeSettings bounds the network endpoint Prometheus may learn from an
// ingest caller.
type ProbeSettings struct {
	AllowedCIDRs []string `yaml:"allowed_cidrs"`
	Port         int      `yaml:"port"`
}

// GitHubSettings holds credentials that must not be confused for one another. The
// status token writes and cannot read repositories; the projects tokens read and
// cannot write anything at all.
type GitHubSettings struct {
	StatusToken              string   `yaml:"status_token"`
	StatusMinIntervalSeconds int      `yaml:"status_min_interval_seconds"`
	StatusTTLMinutes         int      `yaml:"status_ttl_minutes"`
	ProjectsTokens           []string `yaml:"projects_tokens"`
	ProjectsRefreshHours     int      `yaml:"projects_refresh_hours"`
}

// ProfileSettings is the owner's login and the ways to reach them that GitHub cannot
// supply. Their name, avatar, and GitHub link are read from the login, never written.
type ProfileSettings struct {
	GitHubLogin string         `yaml:"github_login"`
	Links       []LinkSettings `yaml:"links"`
}

type LinkSettings struct {
	Platform string `yaml:"platform"`
	Label    string `yaml:"label"`
	URI      string `yaml:"uri"`
}

// LoadConfig reads the settings file named by STATUS_GATEWAY_CONFIG and the container's
// own environment. An unreadable settings file is fatal by construction: it holds the
// tokens, and Validate refuses to run without them.
func LoadConfig() (Config, error) {
	settings, err := loadSettings(os.Getenv("STATUS_GATEWAY_CONFIG"))
	if err != nil {
		return Config{}, err
	}
	scrapeTargets, err := NewScrapeTargetPolicy(settings.Probe.AllowedCIDRs, settings.Probe.Port)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:              env("PORT", "8080"),
		StateFile:         env("STATUS_STATE_FILE", "/data/status-state.json"),
		IdentityStateFile: env("STATUS_IDENTITY_STATE_FILE", "/data/identity-state.json"),
		PrometheusURL:     strings.TrimRight(env("PROMETHEUS_URL", "http://prometheus:9090"), "/"),
		ProjectsFile:      strings.TrimSpace(os.Getenv("PROJECTS_CONFIG_FILE")),
		OIDCIssuer:        strings.TrimRight(strings.TrimSpace(os.Getenv("OIDC_ISSUER")), "/"),
		OIDCAudience:      strings.TrimSpace(os.Getenv("STATUS_AUTH_AUDIENCE")),

		IngestTokens:                   bearerToken(settings.Tokens.Ingest),
		DiscoveryTokens:                bearerToken(settings.Tokens.Discovery),
		GitHubToken:                    secret(settings.GitHub.StatusToken),
		GitHubStatusMinIntervalSeconds: atLeastOne(settings.GitHub.StatusMinIntervalSeconds, 10),
		GitHubStatusTTLMinutes:         atLeastOne(settings.GitHub.StatusTTLMinutes, 20),
		GitHubProjectsTokens:           secrets(settings.GitHub.ProjectsTokens),
		GitHubProjectsRefreshHours:     atLeastOne(settings.GitHub.ProjectsRefreshHours, 24),
		PublicAgentPlaceholder:         settings.PublicAgentPlaceholder,
		Profile:                        configuredProfile(settings.Profile),
		ScrapeTargets:                  scrapeTargets,
	}, nil
}

func loadSettings(path string) (Settings, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Settings{}, errors.New("STATUS_GATEWAY_CONFIG is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, fmt.Errorf("read %s: %w", path, err)
	}

	var settings Settings
	// KnownFields turns a typo into a startup error rather than a setting that
	// silently does nothing — which is the whole failure this file exists to end.
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&settings); err != nil {
		return Settings{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return settings, nil
}

func configuredProfile(profile ProfileSettings) ConfiguredProfile {
	links := make([]ConfiguredLink, 0, len(profile.Links))
	for _, link := range profile.Links {
		links = append(links, ConfiguredLink{
			Label:    link.Label,
			URI:      link.URI,
			Platform: link.Platform,
		})
	}
	return ConfiguredProfile{
		GitHubLogin: strings.TrimSpace(profile.GitHubLogin),
		Links:       links,
	}
}

// Validate reports the first incomplete authentication boundary.
func (config Config) Validate() error {
	if err := config.ScrapeTargets.validate(); err != nil {
		return err
	}
	if len(config.IngestTokens) == 0 {
		return errors.New("tokens.ingest is required")
	}
	if len(config.DiscoveryTokens) == 0 {
		return errors.New("tokens.discovery is required")
	}
	for token := range config.DiscoveryTokens {
		if _, shared := config.IngestTokens[token]; shared {
			return errors.New("tokens.discovery must differ from tokens.ingest")
		}
	}
	return config.AccessConfig().Validate()
}

// AccessConfig returns the human OIDC trust boundary.
func (config Config) AccessConfig() authn.Config {
	return authn.Config{Issuer: config.OIDCIssuer, Audience: config.OIDCAudience}
}

// AuthorizedDiscovery reports whether Prometheus may fetch scrape discovery.
func (config Config) AuthorizedDiscovery(header string) bool {
	return authorizedWith(config.DiscoveryTokens, header)
}

func authorizedWith(tokens map[string]struct{}, header string) bool {
	if len(tokens) == 0 || !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	_, ok := tokens[strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))]
	return ok
}

func env(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

// secret drops the placeholder the example file ships with, so an unedited copy reads
// as unset rather than as a credential that will 401 on every call.
func secret(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "replace-with-") {
		return ""
	}
	return value
}

func secrets(values []string) []string {
	kept := make([]string, 0, len(values))
	for _, value := range values {
		if value := secret(value); value != "" {
			kept = append(kept, value)
		}
	}
	return kept
}

// bearerToken wraps the one configured token in the set the interceptor matches
// against, so a blank setting yields an empty set that admits nobody.
func bearerToken(value string) map[string]struct{} {
	tokens := map[string]struct{}{}
	if value := secret(value); value != "" {
		tokens[value] = struct{}{}
	}
	return tokens
}

func atLeastOne(value int, fallback int) int {
	if value < 1 {
		return fallback
	}
	return value
}

// loadJSONConfig reads a configured JSON document. An empty path means the document
// was not configured, and yields the zero value.
//
// A configured file that cannot be read is an error, not an empty document. The two
// are indistinguishable once served — an empty page reads as one nobody has filled
// in — so swallowing the missing file is what lets a lost config sit unnoticed
// behind a healthy 200.
func loadJSONConfig[T any](path string) (T, error) {
	var config T
	if strings.TrimSpace(path) == "" {
		return config, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		var zero T
		return zero, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		var zero T
		return zero, err
	}
	return config, nil
}
