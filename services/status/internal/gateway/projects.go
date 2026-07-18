package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/services/status/internal/genproto/realtime/me/v1"
)

// projectFetchConcurrency bounds the fan-out of a refresh. Two calls per project
// run sequentially would leave the page unavailable for the fifteen seconds after
// every restart; GitHub's secondary limit is a hundred concurrent requests, so a
// handful is both quick and polite.
const projectFetchConcurrency = 6

// ProjectsConfig is the curated set of repositories the page may show. It names
// them and nothing else: every other field a card carries — the languages, the
// stars, the archived flag, the commit sparkline — is read from GitHub at runtime,
// because a snapshot of those ages the moment it is written.
//
// Curation is explicit on purpose. Publishing whatever the token can see would put
// every private repository the owner creates from now on onto a public page.
type ProjectsConfig struct {
	Projects []CuratedProject `json:"projects"`
}

// CuratedProject is one repository the owner has chosen to show, with the one
// thing GitHub cannot give back.
type CuratedProject struct {
	// Repo is the repository's full name on GitHub, "owner/name". The owner is
	// spelled out because the token reaches organizations as well as the account,
	// and a bare name would silently resolve to whichever organization happened to
	// paginate last.
	Repo string `json:"repo"`

	// Summary is the owner's own blurb, shown in place of GitHub's description.
	// It is optional; the description stands in when it is empty.
	Summary string `json:"summary"`
}

// LoadProjectsConfig reads the curated project list.
func LoadProjectsConfig(path string) (ProjectsConfig, error) {
	return loadJSONConfig[ProjectsConfig](path)
}

// ProjectsService serves the curated projects, refreshed from GitHub once a day and
// served from memory. A repository's languages and stars do not move quickly enough
// to be worth asking about more often than that.
//
// A page cannot fetch on demand. One refresh costs a call for the repository list
// plus two per project, and against a 5,000-request hourly budget a per-visitor
// fetch would be spent inside seventy page loads — and would make every visitor
// wait on seventy-odd round trips to GitHub.
type ProjectsService struct {
	curated   []CuratedProject
	configErr error
	github    *GitHubProjectsClient
	interval  time.Duration

	mutex    sync.RWMutex
	projects []*mev1.Project
	fetched  bool
}

func NewProjectsService(config ProjectsConfig, configErr error, github *GitHubProjectsClient, refreshHours int) *ProjectsService {
	return &ProjectsService{
		curated:   config.Projects,
		configErr: configErr,
		github:    github,
		interval:  time.Duration(refreshHours) * time.Hour,
	}
}

// Run refreshes the projects until the context is cancelled. The first refresh is
// immediate, so the page is only unavailable for the seconds a restart takes.
func (service *ProjectsService) Run(ctx context.Context) {
	if service.configErr != nil || len(service.curated) == 0 {
		return
	}
	for {
		service.refresh(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(service.interval):
		}
	}
}

// List returns the projects last read from GitHub. It reports a fault rather than
// an empty list: a page with nothing on it reads as a life with nothing built in
// it, which is how a broken projects feed goes unnoticed.
func (service *ProjectsService) List() ([]*mev1.Project, error) {
	if service.configErr != nil {
		return nil, service.configErr
	}
	service.mutex.RLock()
	defer service.mutex.RUnlock()
	if !service.fetched {
		return nil, errors.New("projects have not been read from GitHub yet")
	}
	return service.projects, nil
}

// refresh reads every curated repository from GitHub and swaps in the result. One
// repository that cannot be read costs its own card and no other: a page is not
// worth emptying over a repository that was renamed, deleted, or blocked.
func (service *ProjectsService) refresh(ctx context.Context) {
	built := make([]*mev1.Project, len(service.curated))
	semaphore := make(chan struct{}, projectFetchConcurrency)
	var group sync.WaitGroup

	for index, curated := range service.curated {
		group.Add(1)
		go func(index int, fullName string, summary string) {
			defer group.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			repo, detail, err := service.github.Project(ctx, strings.TrimSpace(fullName))
			if err != nil {
				// A curated repository nobody can see is a typo, a rename, or one
				// GitHub has blocked. Name it: dropping it in silence would leave the
				// owner staring at a page one project short, for no stated reason.
				slog.Warn("failed to read curated repository", "repo", fullName, "error", err)
				return
			}
			built[index] = publicProject(repo, detail, summary)
		}(index, curated.Repo, curated.Summary)
	}
	group.Wait()

	projects := make([]*mev1.Project, 0, len(built))
	for _, project := range built {
		if project != nil {
			projects = append(projects, project)
		}
	}
	if len(projects) == 0 {
		// Every repository failed, which is GitHub being unreachable or every token
		// being dead — not the owner having deleted their life's work. Keep the last
		// good snapshot rather than blanking a page that was right this morning.
		slog.Error("no curated repository could be read; keeping the previous projects")
		return
	}

	// The timeline labels a card with the month of its last push, and only when that
	// month changes, so the order it is handed is the order it draws.
	sort.SliceStable(projects, func(first int, second int) bool {
		return projects[first].GetLastPushTime().AsTime().After(projects[second].GetLastPushTime().AsTime())
	})

	service.mutex.Lock()
	service.projects = projects
	service.fetched = true
	service.mutex.Unlock()
	slog.Info("refreshed GitHub projects", "count", len(projects))
}

func publicProject(repo GitHubRepository, detail GitHubRepositoryDetail, summary string) *mev1.Project {
	return &mev1.Project{
		Uid:             projectUID(repo),
		DisplayName:     repo.Name,
		Description:     repo.Description,
		Summary:         summary,
		Visibility:      projectVisibility(repo.Private),
		PrimaryLanguage: repo.Language,
		Topics:          repo.Topics,
		StarCount:       repo.StarCount,
		RepositoryUrl:   publicRepositoryURL(repo),
		HomepageUrl:     repo.Homepage,
		LastPushTime:    parseTimestamp(repo.PushedAt),
		CreateTime:      parseTimestamp(repo.CreatedAt),
		Archived:        repo.Archived,
		Languages:       detail.Languages,
		CommitActivity:  detail.CommitActivity,
	}
}

func projectVisibility(private bool) mev1.ProjectVisibility {
	if private {
		return mev1.ProjectVisibility_PROJECT_VISIBILITY_PRIVATE
	}
	return mev1.ProjectVisibility_PROJECT_VISIBILITY_PUBLIC
}

// publicRepositoryURL withholds the GitHub link for private repositories so the
// public surface never exposes a private repository location.
func publicRepositoryURL(repo GitHubRepository) string {
	if repo.Private {
		return ""
	}
	return repo.HTMLURL
}

// projectUID derives a stable, opaque identifier so callers never construct or
// depend on the underlying repository identity.
func projectUID(repo GitHubRepository) string {
	seed := firstString(repo.HTMLURL, repo.Name)
	sum := sha256.Sum256([]byte("realtime.me.project:" + seed))
	return hex.EncodeToString(sum[:8])
}

func parseTimestamp(value string) *timestamppb.Timestamp {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return timestamppb.New(parsed.UTC())
}
