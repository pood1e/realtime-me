package gateway

import (
	"errors"
	"strings"

	mev1 "github.com/pood1e/realtime-me/services/status/internal/genproto/realtime/me/v1"
)

// ConfiguredProfile holds the owner's login and the ways to reach them that GitHub
// knows nothing about. Their name, avatar, and GitHub link are not among them: all
// three are functions of the login, and writing it into four fields is four chances
// to change three of them.
type ConfiguredProfile struct {
	GitHubLogin string
	Links       []ConfiguredLink
}

// ConfiguredLink is a single public contact link.
type ConfiguredLink struct {
	Label    string
	URI      string
	Platform string
}

// ProfileService serves the site owner's identity.
type ProfileService struct {
	profile ConfiguredProfile
}

func NewProfileService(profile ConfiguredProfile) *ProfileService {
	return &ProfileService{profile: profile}
}

// Profile builds the owner's public identity, deriving from the login everything the
// login already decides. GitHub is not asked for any of it: github.com/<login>.png
// always resolves to the current avatar, so a call would fetch a string the login
// already spells — and would give the topbar a way to go nameless when GitHub is
// down, in exchange for nothing.
//
// A profile with no login is a fault, not an anonymous page: every visible piece of
// the topbar is read from it, so without it there is nothing to draw and nothing true
// to say.
func (service *ProfileService) Profile() (*mev1.Profile, error) {
	login := strings.TrimSpace(service.profile.GitHubLogin)
	if login == "" {
		return nil, errors.New("profile.github_login is not configured")
	}

	links := make([]*mev1.ProfileLink, 0, len(service.profile.Links)+1)
	links = append(links, &mev1.ProfileLink{
		Label:    "GitHub",
		Uri:      gitHubProfileURL(login),
		Platform: "github",
	})
	for _, link := range service.profile.Links {
		links = append(links, &mev1.ProfileLink{
			Label:    link.Label,
			Uri:      link.URI,
			Platform: link.Platform,
		})
	}

	return &mev1.Profile{
		DisplayName: login,
		AvatarUrl:   gitHubAvatarURL(login),
		GithubLogin: login,
		Links:       links,
	}, nil
}

func gitHubProfileURL(login string) string {
	return "https://github.com/" + login
}

// gitHubAvatarURL always resolves to the owner's current avatar, so the page follows
// a change on GitHub without anyone editing a file or the gateway asking a question.
func gitHubAvatarURL(login string) string {
	return gitHubProfileURL(login) + ".png"
}
