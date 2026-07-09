package gateway

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                           string
	StateFile                      string
	IdentityStateFile              string
	StaticDir                      string
	IngestTokens                   map[string]struct{}
	QueryTokens                    map[string]struct{}
	PrometheusURL                  string
	PublicAgentPlaceholder         bool
	AgentFreshSeconds              int
	GitHubToken                    string
	GitHubStatusMinIntervalSeconds int
	GitHubStatusTTLMinutes         int
	ProfileConfigFile              string
}

// LoadConfig reads the gateway configuration from the environment. The read and
// write tokens are deliberately separate secrets: a query token grants only the
// internal status document, metric queries, and scrape discovery, so handing it
// to a dashboard or to Prometheus never confers the ability to enroll devices or
// register scrape targets.
func LoadConfig() Config {
	ingestTokens := parseTokens(os.Getenv("STATUS_INGEST_TOKEN"))
	queryTokens := parseTokens(os.Getenv("STATUS_QUERY_TOKEN"))
	return Config{
		Port:                           env("PORT", "8080"),
		StateFile:                      env("STATUS_STATE_FILE", "/data/status-state.json"),
		IdentityStateFile:              env("STATUS_IDENTITY_STATE_FILE", "/data/identity-state.json"),
		StaticDir:                      strings.TrimRight(strings.TrimSpace(os.Getenv("STATUS_WEB_DIR")), "/"),
		IngestTokens:                   ingestTokens,
		QueryTokens:                    queryTokens,
		PrometheusURL:                  strings.TrimRight(env("PROMETHEUS_URL", "http://prometheus:9090"), "/"),
		PublicAgentPlaceholder:         os.Getenv("PUBLIC_AGENT_PLACEHOLDER") == "true",
		AgentFreshSeconds:              positiveInt("STATUS_AGENT_FRESH_SECONDS", 120),
		GitHubToken:                    secretEnv("GITHUB_TOKEN"),
		GitHubStatusMinIntervalSeconds: positiveInt("GITHUB_STATUS_MIN_INTERVAL_SECONDS", 10),
		GitHubStatusTTLMinutes:         positiveInt("GITHUB_STATUS_TTL_MINUTES", 20),
		ProfileConfigFile:              strings.TrimSpace(os.Getenv("PROFILE_CONFIG_FILE")),
	}
}

// Validate reports the first configuration error that would leave the gateway
// unable to authenticate any caller. An empty token set rejects everyone, which
// reads as a mysterious outage rather than a deliberate lockdown, so both sets
// must be configured explicitly.
func (config Config) Validate() error {
	if len(config.IngestTokens) == 0 {
		return errors.New("STATUS_INGEST_TOKEN is required")
	}
	if len(config.QueryTokens) == 0 {
		return errors.New("STATUS_QUERY_TOKEN is required")
	}
	return nil
}

// AuthorizedQuery reports whether the bearer token may read internal status,
// query metrics, and fetch scrape discovery.
func (config Config) AuthorizedQuery(header string) bool {
	return authorizedWith(config.QueryTokens, header)
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

func secretEnv(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" || strings.HasPrefix(value, "replace-with-") {
		return ""
	}
	return value
}

func parseTokens(value string) map[string]struct{} {
	tokens := map[string]struct{}{}
	for _, token := range strings.Split(value, ",") {
		token = strings.TrimSpace(token)
		if token != "" {
			tokens[token] = struct{}{}
		}
	}
	return tokens
}

func positiveInt(name string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
