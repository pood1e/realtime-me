package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRequiresBothTokens(t *testing.T) {
	targets := testScrapeTargetPolicy(t)
	cases := []struct {
		name      string
		ingest    map[string]struct{}
		discovery map[string]struct{}
		wantErr   bool
	}{
		{name: "both set", ingest: tokenSet("write"), discovery: tokenSet("read"), wantErr: false},
		{name: "no discovery token", ingest: tokenSet("write"), discovery: nil, wantErr: true},
		{name: "no ingest token", ingest: nil, discovery: tokenSet("read"), wantErr: true},
		{name: "neither", ingest: nil, discovery: nil, wantErr: true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			config := Config{
				IngestTokens:    testCase.ingest,
				DiscoveryTokens: testCase.discovery,
				OIDCIssuer:      "https://identity.example.com",
				OIDCAudience:    "status",
				ScrapeTargets:   targets,
			}
			if err := config.Validate(); (err != nil) != testCase.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

// The discovery token must never authorize a write.
func TestReadTokenIsNotAWriteToken(t *testing.T) {
	config := Config{IngestTokens: tokenSet("write"), DiscoveryTokens: tokenSet("read")}

	if config.AuthorizedDiscovery("Bearer read") != true {
		t.Error("discovery token must authorize reads")
	}
	if config.AuthorizedDiscovery("Bearer write") != false {
		t.Error("ingest token must not authorize reads")
	}
	if authorizedWith(config.IngestTokens, "Bearer read") != false {
		t.Error("discovery token must not authorize writes")
	}
	if authorizedWith(config.IngestTokens, "Bearer write") != true {
		t.Error("ingest token must authorize writes")
	}
}

func TestAuthorizedWith(t *testing.T) {
	tokens := tokenSet("alpha", "beta")
	cases := []struct {
		header string
		want   bool
	}{
		{header: "Bearer alpha", want: true},
		{header: "Bearer beta", want: true},
		{header: "Bearer  alpha ", want: true}, // surrounding space is trimmed
		{header: "Bearer gamma", want: false},
		{header: "alpha", want: false},        // missing scheme
		{header: "bearer alpha", want: false}, // scheme is case-sensitive
		{header: "Basic alpha", want: false},
		{header: "", want: false},
		{header: "Bearer ", want: false},
	}
	for _, testCase := range cases {
		if got := authorizedWith(tokens, testCase.header); got != testCase.want {
			t.Errorf("authorizedWith(%q) = %v, want %v", testCase.header, got, testCase.want)
		}
	}
}

// An empty token set rejects every caller rather than admitting them.
func TestAuthorizedWithEmptySetRejects(t *testing.T) {
	if authorizedWith(map[string]struct{}{}, "Bearer anything") {
		t.Fatal("an empty token set must reject every caller")
	}
}

func loadSettingsFile(t *testing.T, document string) (Config, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gateway.yaml")
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}
	t.Setenv("STATUS_GATEWAY_CONFIG", path)
	t.Setenv("OIDC_ISSUER", "https://identity.example.com")
	t.Setenv("STATUS_AUTH_AUDIENCE", "status")
	return LoadConfig()
}

// One file holds every setting a person writes: the two bearer tokens, the two kinds
// of GitHub credential, and the owner's profile.
func TestSettingsCarryTheTokensAndTheProfile(t *testing.T) {
	config, err := loadSettingsFile(t, `
tokens:
  ingest: write-me
  discovery: read-me
github:
  status_token: ghp_status
  projects_tokens:
    - github_pat_one
    - github_pat_two
profile:
  github_login: pood1e
  links:
    - platform: telegram
      label: Telegram
      uri: https://t.me/a-handle
probe:
  allowed_cidrs:
    - 10.40.0.0/16
  port: 18082
`)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if _, ok := config.IngestTokens["write-me"]; !ok {
		t.Fatal("the ingest token must reach the interceptor")
	}
	if config.GitHubToken != "ghp_status" {
		t.Fatalf("the status token is the one that writes, got %q", config.GitHubToken)
	}
	if len(config.GitHubProjectsTokens) != 2 {
		t.Fatalf("one read-only token per owner, got %d", len(config.GitHubProjectsTokens))
	}
	if config.Profile.GitHubLogin != "pood1e" || len(config.Profile.Links) != 1 {
		t.Fatalf("the profile rides in the settings now, got %+v", config.Profile)
	}
	// An omitted interval must not become zero: a zero refresh would spin on GitHub.
	if config.GitHubProjectsRefreshHours != 24 || config.GitHubStatusTTLMinutes != 20 {
		t.Fatalf("an omitted interval falls back to its default, got %+v", config)
	}
}

// Prometheus holds the discovery token. It must never authorize device enrollment.
func TestOneTokenCannotBeBothHalvesOfTheAPI(t *testing.T) {
	config, err := loadSettingsFile(t, "tokens:\n  ingest: same\n  discovery: same\nprobe:\n  allowed_cidrs: [10.40.0.0/16]\n  port: 18082\n")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := config.Validate(); err == nil {
		t.Fatal("the read and write tokens must be different secrets")
	}
}

func TestProbePolicyIsRequired(t *testing.T) {
	if _, err := loadSettingsFile(t, "tokens:\n  ingest: write\n  discovery: read\n"); err == nil {
		t.Fatal("missing probe policy must fail configuration loading")
	}
}

func TestProbePolicyRejectsInvalidNetwork(t *testing.T) {
	if _, err := loadSettingsFile(t, "tokens:\n  ingest: write\n  discovery: read\nprobe:\n  allowed_cidrs: [not-a-network]\n  port: 18082\n"); err == nil {
		t.Fatal("invalid probe CIDR must fail configuration loading")
	}
}

// A misspelled setting that silently does nothing is the very failure this file was
// introduced to end. Say so at startup instead of serving without it.
func TestAMisspelledSettingIsAStartupError(t *testing.T) {
	if _, err := loadSettingsFile(t, "tokens:\n  ingest: a\n  querry: b\n"); err == nil {
		t.Fatal("an unknown field must fail the load, not be ignored")
	}
}

func TestAMissingSettingsFileIsAStartupError(t *testing.T) {
	t.Setenv("STATUS_GATEWAY_CONFIG", filepath.Join(t.TempDir(), "absent.yaml"))
	if _, err := LoadConfig(); err == nil {
		t.Fatal("a settings file that does not exist must fail the load")
	}
}

func tokenSet(values ...string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	tokens := make(map[string]struct{}, len(values))
	for _, value := range values {
		tokens[value] = struct{}{}
	}
	return tokens
}
