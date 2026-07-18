package gateway

import (
	"testing"
)

// The login is the only identity written down: the name, the avatar, and the GitHub
// link are read from it, so none of them can drift out of step with the others.
func TestProfileDerivesIdentityFromTheLogin(t *testing.T) {
	service := NewProfileService(ConfiguredProfile{
		GitHubLogin: "pood1e",
		Links:       []ConfiguredLink{{Label: "Email", URI: "mailto:me@example.com", Platform: "email"}},
	})

	profile, err := service.Profile()
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.GetDisplayName() != "pood1e" {
		t.Fatalf("the topbar name is the login, got %q", profile.GetDisplayName())
	}
	if profile.GetAvatarUrl() != "https://github.com/pood1e.png" {
		t.Fatalf("the avatar is read from the login, got %q", profile.GetAvatarUrl())
	}

	links := profile.GetLinks()
	if len(links) != 2 {
		t.Fatalf("want the derived GitHub link plus the configured one, got %+v", links)
	}
	if links[0].GetPlatform() != "github" || links[0].GetUri() != "https://github.com/pood1e" {
		t.Fatalf("the GitHub link is derived, never written down, got %+v", links[0])
	}
	if links[1].GetPlatform() != "email" {
		t.Fatalf("the links GitHub cannot supply survive the round trip, got %+v", links[1])
	}
}

// Every visible piece of the topbar is read from the login, so a profile without one
// has nothing to draw and nothing true to say.
func TestProfileWithoutALoginIsAFault(t *testing.T) {
	if _, err := NewProfileService(ConfiguredProfile{}).Profile(); err == nil {
		t.Fatal("a profile with no login must report the fault, not serve a nameless page")
	}
}
