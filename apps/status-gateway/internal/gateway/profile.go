package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// ProfileConfig is the static profile served on the profile page. Projects are
// collected once (see scripts/operator/collect-projects.py) and stored here as data; the
// gateway never fetches or refreshes them at runtime.
type ProfileConfig struct {
	Profile  ConfiguredProfile   `json:"profile"`
	Projects []ConfiguredProject `json:"projects"`
}

// ConfiguredProfile holds the personal introduction and contact links.
type ConfiguredProfile struct {
	DisplayName string           `json:"display_name"`
	Headline    string           `json:"headline"`
	Bio         string           `json:"bio"`
	AvatarURL   string           `json:"avatar_url"`
	Location    string           `json:"location"`
	GitHubLogin string           `json:"github_login"`
	Links       []ConfiguredLink `json:"links"`
}

// ConfiguredLink is a single configurable contact or social link.
type ConfiguredLink struct {
	Label    string `json:"label"`
	URI      string `json:"uri"`
	Platform string `json:"platform"`
}

// ConfiguredProject is one collected repository, in display order.
type ConfiguredProject struct {
	DisplayName     string               `json:"display_name"`
	Description     string               `json:"description"`
	Summary         string               `json:"summary"`
	Visibility      string               `json:"visibility"`
	PrimaryLanguage string               `json:"primary_language"`
	Topics          []string             `json:"topics"`
	StarCount       int32                `json:"star_count"`
	RepositoryURL   string               `json:"repository_url"`
	HomepageURL     string               `json:"homepage_url"`
	LastPushTime    string               `json:"last_push_time"`
	CreateTime      string               `json:"create_time"`
	Archived        bool                 `json:"archived"`
	Languages       []ConfiguredLanguage `json:"languages"`
	CommitActivity  []int32              `json:"commit_activity"`
}

// ConfiguredLanguage is one language's source-byte count for a collected repository.
type ConfiguredLanguage struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
}

// LoadProfileConfig reads the profile configuration file. A missing file or empty
// path yields an empty configuration so the profile page degrades gracefully.
func LoadProfileConfig(path string) (ProfileConfig, error) {
	if strings.TrimSpace(path) == "" {
		return ProfileConfig{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ProfileConfig{}, nil
		}
		return ProfileConfig{}, err
	}
	var config ProfileConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ProfileConfig{}, err
	}
	return config, nil
}

// ProfileService assembles the public profile document from static configuration.
type ProfileService struct {
	config ProfileConfig
}

func NewProfileService(config ProfileConfig) *ProfileService {
	return &ProfileService{config: config}
}

// Page builds the public profile document, withholding private repository
// locations from the public surface.
func (service *ProfileService) Page(now time.Time) *mev1.ProfilePage {
	projects := make([]*mev1.Project, 0, len(service.config.Projects))
	for _, project := range service.config.Projects {
		projects = append(projects, publicProject(project))
	}
	return &mev1.ProfilePage{
		Profile:    profileMessage(service.config.Profile),
		Projects:   projects,
		UpdateTime: timestamppb.New(now.UTC()),
	}
}

func profileMessage(profile ConfiguredProfile) *mev1.Profile {
	links := make([]*mev1.ProfileLink, 0, len(profile.Links))
	for _, link := range profile.Links {
		links = append(links, &mev1.ProfileLink{
			Label:    link.Label,
			Uri:      link.URI,
			Platform: link.Platform,
		})
	}
	return &mev1.Profile{
		DisplayName: profile.DisplayName,
		Headline:    profile.Headline,
		Bio:         profile.Bio,
		AvatarUrl:   profile.AvatarURL,
		Location:    profile.Location,
		GithubLogin: profile.GitHubLogin,
		Links:       links,
	}
}

func publicProject(project ConfiguredProject) *mev1.Project {
	return &mev1.Project{
		Uid:             projectUID(project),
		DisplayName:     project.DisplayName,
		Description:     project.Description,
		Summary:         project.Summary,
		Visibility:      projectVisibility(project.Visibility),
		PrimaryLanguage: project.PrimaryLanguage,
		Topics:          project.Topics,
		StarCount:       project.StarCount,
		RepositoryUrl:   publicRepositoryURL(project),
		HomepageUrl:     project.HomepageURL,
		LastPushTime:    parseTimestamp(project.LastPushTime),
		CreateTime:      parseTimestamp(project.CreateTime),
		Archived:        project.Archived,
		Languages:       projectLanguages(project.Languages),
		CommitActivity:  project.CommitActivity,
	}
}

func projectLanguages(languages []ConfiguredLanguage) []*mev1.LanguageShare {
	shares := make([]*mev1.LanguageShare, 0, len(languages))
	for _, language := range languages {
		shares = append(shares, &mev1.LanguageShare{Name: language.Name, Bytes: language.Bytes})
	}
	return shares
}

func projectVisibility(value string) mev1.ProjectVisibility {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "private":
		return mev1.ProjectVisibility_PROJECT_VISIBILITY_PRIVATE
	case "public":
		return mev1.ProjectVisibility_PROJECT_VISIBILITY_PUBLIC
	default:
		return mev1.ProjectVisibility_PROJECT_VISIBILITY_UNSPECIFIED
	}
}

// publicRepositoryURL withholds the GitHub link for private repositories so the
// public surface never exposes a private repository location.
func publicRepositoryURL(project ConfiguredProject) string {
	if strings.EqualFold(strings.TrimSpace(project.Visibility), "private") {
		return ""
	}
	return project.RepositoryURL
}

// projectUID derives a stable, opaque identifier so callers never construct or
// depend on the underlying repository identity.
func projectUID(project ConfiguredProject) string {
	seed := firstString(project.RepositoryURL, project.DisplayName)
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
