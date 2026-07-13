package gateway

import (
	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// ProfileConfig is the site owner's identity. It ships with the stack rather than
// living in the data volume: it is small, public, hand-written, and nothing can
// regenerate it, so it is the one document that must never quietly go missing.
type ProfileConfig struct {
	Profile ConfiguredProfile `json:"profile"`
}

// ConfiguredProfile holds the owner's identity and contact links.
type ConfiguredProfile struct {
	DisplayName string           `json:"display_name"`
	AvatarURL   string           `json:"avatar_url"`
	GitHubLogin string           `json:"github_login"`
	Links       []ConfiguredLink `json:"links"`
}

// ConfiguredLink is a single public contact link.
type ConfiguredLink struct {
	Label    string `json:"label"`
	URI      string `json:"uri"`
	Platform string `json:"platform"`
}

// LoadProfileConfig reads the profile configuration file.
func LoadProfileConfig(path string) (ProfileConfig, error) {
	return loadJSONConfig[ProfileConfig](path)
}

// ProfileService serves the site owner's identity.
type ProfileService struct {
	config  ProfileConfig
	loadErr error
}

// NewProfileService holds either the loaded configuration or the error that
// prevented it from loading, so a profile that was configured but unreadable is
// reported as a fault rather than served as a nameless page.
func NewProfileService(config ProfileConfig, loadErr error) *ProfileService {
	return &ProfileService{config: config, loadErr: loadErr}
}

// Profile builds the owner's public identity.
func (service *ProfileService) Profile() (*mev1.Profile, error) {
	if service.loadErr != nil {
		return nil, service.loadErr
	}
	profile := service.config.Profile
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
		AvatarUrl:   profile.AvatarURL,
		GithubLogin: profile.GitHubLogin,
		Links:       links,
	}, nil
}
