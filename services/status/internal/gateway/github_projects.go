package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	sitev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/site/v1"
)

const (
	gitHubAPIURL    = "https://api.github.com"
	gitHubReadLimit = 10 * time.Second
)

// GitHubProjectsClient reads repositories, and only ever reads them. It holds tokens
// of its own, separate from the one that writes the owner's GitHub status.
//
// It holds more than one because a fine-grained token reaches the repositories of a
// single user or organization and no further, while the curated list spans several
// owners. One token per owner, each needing no more than Metadata: read-only — which
// is the whole point: a classic token has no read-only grade for private
// repositories, only `repo`, which is read *and write* over every repository the
// owner has, on a server that needs to write to none of them.
type GitHubProjectsClient struct {
	tokens []string
	client *http.Client

	// Which token turned out to be the one that can see a given owner. Resolved on
	// first use and remembered, so a refresh does not re-probe every owner from the
	// top of the list for every repository it holds.
	mutex        sync.Mutex
	tokenByOwner map[string]string
}

func NewGitHubProjectsClient(tokens []string) *GitHubProjectsClient {
	return &GitHubProjectsClient{
		tokens:       tokens,
		client:       &http.Client{Timeout: gitHubReadLimit},
		tokenByOwner: map[string]string{},
	}
}

// GitHubRepository is a repository as GitHub describes it. It carries every field a
// card needs except the languages and the commit activity, which cost a call apiece.
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

// GitHubRepositoryDetail is the decoration: what the repository itself does not
// carry, and what a card can do without.
type GitHubRepositoryDetail struct {
	Languages      []*sitev1.LanguageShare
	CommitActivity []int32
}

// Project reads one curated repository by its full name, "owner/name". The curated
// list is the only thing that decides which repositories are read: there is no
// listing call, and so no way for a repository nobody curated to reach the page.
func (github *GitHubProjectsClient) Project(ctx context.Context, fullName string) (GitHubRepository, GitHubRepositoryDetail, error) {
	if len(github.tokens) == 0 {
		return GitHubRepository{}, GitHubRepositoryDetail{}, errors.New("GITHUB_PROJECTS_TOKENS is not set")
	}

	token, repo, err := github.resolve(ctx, fullName)
	if err != nil {
		return GitHubRepository{}, GitHubRepositoryDetail{}, err
	}

	// Neither of these is worth losing a card over. A repository whose languages
	// GitHub declines to serve still belongs on the page, just without its bar.
	return repo, GitHubRepositoryDetail{
		Languages:      github.languages(ctx, fullName, token),
		CommitActivity: github.commitActivity(ctx, fullName, token),
	}, nil
}

// resolve finds the token that can see this repository. GitHub answers 404, not 403,
// for a repository a token cannot see, so "not visible" and "not there" look alike
// from here — which is why every token is tried before the repository is given up on.
func (github *GitHubProjectsClient) resolve(ctx context.Context, fullName string) (string, GitHubRepository, error) {
	owner, _, ok := strings.Cut(fullName, "/")
	if !ok {
		return "", GitHubRepository{}, fmt.Errorf("github: %q is not an owner/name", fullName)
	}
	owner = strings.ToLower(owner)

	for _, token := range github.orderedFor(owner) {
		repo, err := github.repository(ctx, fullName, token)
		if err != nil {
			continue
		}
		github.remember(owner, token)
		return token, repo, nil
	}
	return "", GitHubRepository{}, fmt.Errorf("github: no token can see %s", fullName)
}

// orderedFor puts the token already known to work for this owner first, and keeps
// the rest as fallbacks in case a token was rotated or its grant withdrawn.
func (github *GitHubProjectsClient) orderedFor(owner string) []string {
	github.mutex.Lock()
	known, ok := github.tokenByOwner[owner]
	github.mutex.Unlock()
	if !ok {
		return github.tokens
	}

	ordered := make([]string, 0, len(github.tokens))
	ordered = append(ordered, known)
	for _, token := range github.tokens {
		if token != known {
			ordered = append(ordered, token)
		}
	}
	return ordered
}

func (github *GitHubProjectsClient) remember(owner string, token string) {
	github.mutex.Lock()
	github.tokenByOwner[owner] = token
	github.mutex.Unlock()
}

func (github *GitHubProjectsClient) repository(ctx context.Context, fullName string, token string) (GitHubRepository, error) {
	var repo GitHubRepository
	err := github.get(ctx, fmt.Sprintf("%s/repos/%s", gitHubAPIURL, fullName), token, &repo)
	return repo, err
}

func (github *GitHubProjectsClient) languages(ctx context.Context, fullName string, token string) []*sitev1.LanguageShare {
	var bytesByLanguage map[string]int64
	url := fmt.Sprintf("%s/repos/%s/languages", gitHubAPIURL, fullName)
	if err := github.get(ctx, url, token, &bytesByLanguage); err != nil {
		slog.Warn("failed to read repository languages", "repo", fullName, "error", err)
		return nil
	}

	shares := make([]*sitev1.LanguageShare, 0, len(bytesByLanguage))
	for name, size := range bytesByLanguage {
		shares = append(shares, &sitev1.LanguageShare{Name: name, Bytes: size})
	}
	sort.SliceStable(shares, func(first int, second int) bool {
		return shares[first].GetBytes() > shares[second].GetBytes()
	})
	return shares
}

// commitActivity reads the weekly commit counts behind the sparkline. GitHub computes
// these lazily and answers 202 with an empty body while it does, so a miss is normal
// and means "not yet", not "broken": draw no sparkline, and ask again tomorrow.
func (github *GitHubProjectsClient) commitActivity(ctx context.Context, fullName string, token string) []int32 {
	var participation struct {
		All []int32 `json:"all"`
	}
	url := fmt.Sprintf("%s/repos/%s/stats/participation", gitHubAPIURL, fullName)
	if err := github.get(ctx, url, token, &participation); err != nil {
		return nil
	}
	return participation.All
}

func (github *GitHubProjectsClient) get(ctx context.Context, url string, token string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := github.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		// The status code is the whole story here, and the body may quote the request
		// back at us. Report the code, and never what was sent to get it.
		return fmt.Errorf("github: HTTP %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
