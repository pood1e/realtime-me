package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

const (
	gitHubAPIURL = "https://api.github.com"

	// GitHub caps a page at 100, and stops us walking forever if it ever stops
	// shrinking the last page.
	gitHubPageSize  = 100
	gitHubMaxPages  = 10
	gitHubReadLimit = 10 * time.Second
)

// GitHubProjectsClient reads repositories. It holds a token of its own, separate
// from the one that writes the owner's GitHub status: this one only ever reads, and
// a fine-grained token with Metadata: read-only is enough for every call it makes.
type GitHubProjectsClient struct {
	token  string
	client *http.Client
}

func NewGitHubProjectsClient(token string) *GitHubProjectsClient {
	return &GitHubProjectsClient{
		token:  token,
		client: &http.Client{Timeout: gitHubReadLimit},
	}
}

// GitHubRepository is a repository as the listing endpoint describes it. One call
// carries every field a card needs except the languages and the commit activity.
type GitHubRepository struct {
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	StarCount   int32    `json:"stargazers_count"`
	HTMLURL     string   `json:"html_url"`
	Homepage    string   `json:"homepage"`
	PushedAt    string   `json:"pushed_at"`
	CreatedAt   string   `json:"created_at"`
	Archived    bool     `json:"archived"`
	Private     bool     `json:"private"`
}

// GitHubRepositoryDetail is what the listing does not carry, and what costs a call
// apiece to learn.
type GitHubRepositoryDetail struct {
	Languages      []*mev1.LanguageShare
	CommitActivity []int32
}

// Repositories lists every repository the token can see, keyed by lowercase name.
// The curated list names repositories, not owners, so the token's own account is
// what a name is resolved against.
func (github *GitHubProjectsClient) Repositories(ctx context.Context) (map[string]GitHubRepository, error) {
	if strings.TrimSpace(github.token) == "" {
		return nil, errors.New("GITHUB_PROJECTS_TOKEN is not set")
	}

	repositories := map[string]GitHubRepository{}
	for page := 1; page <= gitHubMaxPages; page++ {
		url := fmt.Sprintf(
			"%s/user/repos?affiliation=owner,organization_member&visibility=all&sort=pushed&per_page=%d&page=%d",
			gitHubAPIURL, gitHubPageSize, page,
		)
		var batch []GitHubRepository
		if err := github.get(ctx, url, &batch); err != nil {
			return nil, err
		}
		for _, repository := range batch {
			repositories[strings.ToLower(repository.Name)] = repository
		}
		if len(batch) < gitHubPageSize {
			break
		}
	}
	return repositories, nil
}

// Detail reads the two things the listing withholds. Neither is worth failing a
// refresh over: a repository whose language breakdown GitHub declines to compute
// still belongs on the page, just without its bar.
func (github *GitHubProjectsClient) Detail(ctx context.Context, repo GitHubRepository) (GitHubRepositoryDetail, error) {
	languages, err := github.languages(ctx, repo.FullName)
	if err != nil {
		return GitHubRepositoryDetail{}, err
	}
	return GitHubRepositoryDetail{
		Languages:      languages,
		CommitActivity: github.commitActivity(ctx, repo.FullName),
	}, nil
}

func (github *GitHubProjectsClient) languages(ctx context.Context, fullName string) ([]*mev1.LanguageShare, error) {
	var bytesByLanguage map[string]int64
	url := fmt.Sprintf("%s/repos/%s/languages", gitHubAPIURL, fullName)
	if err := github.get(ctx, url, &bytesByLanguage); err != nil {
		return nil, err
	}

	shares := make([]*mev1.LanguageShare, 0, len(bytesByLanguage))
	for name, size := range bytesByLanguage {
		shares = append(shares, &mev1.LanguageShare{Name: name, Bytes: size})
	}
	sort.SliceStable(shares, func(first int, second int) bool {
		return shares[first].GetBytes() > shares[second].GetBytes()
	})
	return shares, nil
}

// commitActivity reads the weekly commit counts behind the sparkline. GitHub
// computes these lazily and answers 202 with an empty body while it does, so a miss
// is normal and means "not yet", not "broken": draw no sparkline and try again on
// the next refresh.
func (github *GitHubProjectsClient) commitActivity(ctx context.Context, fullName string) []int32 {
	var participation struct {
		All []int32 `json:"all"`
	}
	url := fmt.Sprintf("%s/repos/%s/stats/participation", gitHubAPIURL, fullName)
	if err := github.get(ctx, url, &participation); err != nil {
		return nil
	}
	return participation.All
}

func (github *GitHubProjectsClient) get(ctx context.Context, url string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("Authorization", "Bearer "+github.token)

	response, err := github.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		// The status code is the whole story here, and the body may carry the token
		// back at us in an error message. Report the code and nothing else.
		return fmt.Errorf("github: HTTP %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
