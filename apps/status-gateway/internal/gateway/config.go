package gateway

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                           string
	StateFile                      string
	IngestTokens                   map[string]struct{}
	PrometheusURL                  string
	PublicAgentPlaceholder         bool
	AgentFreshSeconds              int
	GitHubToken                    string
	GitHubStatusMinIntervalSeconds int
	GitHubStatusTTLMinutes         int
}

func LoadConfig() Config {
	return Config{
		Port:                           env("PORT", "8080"),
		StateFile:                      env("STATUS_STATE_FILE", "/data/status-state.json"),
		IngestTokens:                   parseTokens(os.Getenv("STATUS_INGEST_TOKEN")),
		PrometheusURL:                  strings.TrimRight(env("PROMETHEUS_URL", "http://prometheus:9090"), "/"),
		PublicAgentPlaceholder:         os.Getenv("PUBLIC_AGENT_PLACEHOLDER") != "false",
		AgentFreshSeconds:              positiveInt("STATUS_AGENT_FRESH_SECONDS", 120),
		GitHubToken:                    secretEnv("GITHUB_TOKEN"),
		GitHubStatusMinIntervalSeconds: positiveInt("GITHUB_STATUS_MIN_INTERVAL_SECONDS", 10),
		GitHubStatusTTLMinutes:         positiveInt("GITHUB_STATUS_TTL_MINUTES", 20),
	}
}

func (config Config) Authorized(header string) bool {
	if len(config.IngestTokens) == 0 || !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	_, ok := config.IngestTokens[strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))]
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
