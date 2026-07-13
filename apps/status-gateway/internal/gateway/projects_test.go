package gateway

import (
	"os"
	"testing"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

func TestProjectsReportUnreadableCuration(t *testing.T) {
	service := NewProjectsService(ProjectsConfig{}, os.ErrNotExist, NewGitHubProjectsClient(nil), 30)
	if _, err := service.List(); err == nil {
		t.Fatal("a curated list that failed to load must report the fault, not an empty page")
	}
}

// Until GitHub has answered once there is nothing to show, and saying so is not the
// same as saying the owner has built nothing.
func TestProjectsReportNotYetFetched(t *testing.T) {
	service := NewProjectsService(
		ProjectsConfig{Projects: []CuratedProject{{Repo: "pood1e/realtime-me"}}},
		nil,
		NewGitHubProjectsClient(nil),
		30,
	)
	if _, err := service.List(); err == nil {
		t.Fatal("projects not yet read from GitHub must report a fault, not an empty list")
	}
}

// The public surface must never say where a private repository lives, however the
// project reached the page.
func TestPublicProjectWithholdsPrivateRepositoryURL(t *testing.T) {
	private := publicProject(
		GitHubRepository{Name: "secret", HTMLURL: "https://github.com/pood1e/secret", Private: true},
		GitHubRepositoryDetail{},
		"",
	)
	if private.GetVisibility() != mev1.ProjectVisibility_PROJECT_VISIBILITY_PRIVATE {
		t.Fatalf("want private visibility, got %v", private.GetVisibility())
	}
	if private.GetRepositoryUrl() != "" {
		t.Fatalf("a private repository's location must never reach the page, got %q", private.GetRepositoryUrl())
	}

	public := publicProject(
		GitHubRepository{Name: "open", HTMLURL: "https://github.com/pood1e/open"},
		GitHubRepositoryDetail{},
		"",
	)
	if public.GetRepositoryUrl() == "" {
		t.Fatal("a public repository keeps its link")
	}
	if private.GetUid() == "" || private.GetUid() == public.GetUid() {
		t.Fatal("every project carries its own opaque uid")
	}
}

// The summary is the one field GitHub cannot give back, so the curated list is what
// must supply it; the repository's own description stands in when it is empty.
func TestPublicProjectCarriesTheCuratedSummary(t *testing.T) {
	repo := GitHubRepository{Name: "realtime-me", Description: "from github"}

	withSummary := publicProject(repo, GitHubRepositoryDetail{}, "written by hand")
	if withSummary.GetSummary() != "written by hand" {
		t.Fatalf("the curated summary must reach the card, got %q", withSummary.GetSummary())
	}
	if withSummary.GetDescription() != "from github" {
		t.Fatalf("GitHub's own description rides along as the fallback, got %q", withSummary.GetDescription())
	}

	if withoutSummary := publicProject(repo, GitHubRepositoryDetail{}, ""); withoutSummary.GetSummary() != "" {
		t.Fatalf("an uncurated summary stays empty so the description can stand in, got %q", withoutSummary.GetSummary())
	}
}
