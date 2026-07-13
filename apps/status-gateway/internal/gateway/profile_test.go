package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

// A configured document that has gone missing must not read as one nobody filled
// in. This is the failure that served a healthy 200 over an empty page.
func TestLoadConfigMissingFileIsAnError(t *testing.T) {
	if _, err := LoadProfileConfig(filepath.Join(t.TempDir(), "profile.json")); err == nil {
		t.Fatal("a configured profile file that does not exist must be an error")
	}
	if _, err := LoadProjectsConfig(filepath.Join(t.TempDir(), "projects.json")); err == nil {
		t.Fatal("a configured projects file that does not exist must be an error")
	}
}

func TestLoadProfileConfigUnsetPathIsEmpty(t *testing.T) {
	config, err := LoadProfileConfig("")
	if err != nil {
		t.Fatalf("an unconfigured profile is not an error: %v", err)
	}
	if config.Profile.DisplayName != "" || len(config.Profile.Links) != 0 {
		t.Fatalf("an unconfigured profile is empty, got %+v", config)
	}
}

func TestProfileReportsLoadFailure(t *testing.T) {
	if _, err := NewProfileService(ProfileConfig{}, os.ErrNotExist).Profile(); err == nil {
		t.Fatal("a profile that failed to load must report the fault, not an empty identity")
	}
}

func TestProfileServesConfiguredLinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	document := `{"profile":{"display_name":"pood1e","links":[{"label":"Email","uri":"mailto:me@example.com","platform":"email"}]}}`
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	config, err := LoadProfileConfig(path)
	if err != nil {
		t.Fatalf("load profile: %v", err)
	}
	profile, err := NewProfileService(config, nil).Profile()
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.GetDisplayName() != "pood1e" {
		t.Fatalf("display name must survive the round trip, got %q", profile.GetDisplayName())
	}
	links := profile.GetLinks()
	if len(links) != 1 || links[0].GetPlatform() != "email" {
		t.Fatalf("the contact links the topbar draws must survive the round trip, got %+v", links)
	}
}
