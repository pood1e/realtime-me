package provider

import (
	"errors"
	"fmt"

	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider/netease"
	"example.com/cloud-drive/api/internal/provider/qqmusic"
	"example.com/cloud-drive/api/internal/provider/spotify"
)

// RegistryConfig contains provider application credentials, never user credentials.
type RegistryConfig struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
}

// Registry contains ordered compile-time plugins keyed by their stable provider ID.
type Registry struct {
	ordered    []domain.MusicProviderAdapter
	byProvider map[domain.MusicProvider]domain.MusicProviderAdapter
}

// NewRegistry constructs and validates the complete plugin registry.
func NewRegistry(config RegistryConfig) (*Registry, error) {
	spotifyAdapter, err := newSpotifyAdapter(config)
	if err != nil {
		return nil, err
	}
	registry := &Registry{byProvider: make(map[domain.MusicProvider]domain.MusicProviderAdapter)}
	for _, adapter := range []domain.MusicProviderAdapter{
		QQAdapter{},
		NetEaseAdapter{},
		spotifyAdapter,
	} {
		if err := registry.register(adapter); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// Get resolves one registered plugin.
func (r *Registry) Get(provider domain.MusicProvider) (domain.MusicProviderAdapter, bool) {
	adapter, found := r.byProvider[provider]
	return adapter, found
}

// List returns providers in stable UI order without exposing mutable registry state.
func (r *Registry) List() []domain.MusicProviderAdapter {
	return append([]domain.MusicProviderAdapter(nil), r.ordered...)
}

func (r *Registry) register(adapter domain.MusicProviderAdapter) error {
	if adapter == nil || adapter.Provider() == "" || adapter.Provider() == domain.MusicProviderLocal {
		return errors.New("music provider plugin has an invalid identity")
	}
	if _, duplicate := r.byProvider[adapter.Provider()]; duplicate {
		return fmt.Errorf("duplicate music provider plugin: %s", adapter.Provider())
	}
	r.byProvider[adapter.Provider()] = adapter
	r.ordered = append(r.ordered, adapter)
	return nil
}

func mapProviderError(err error) error {
	if err == nil {
		return nil
	}
	var qqError *qqmusic.Error
	if errors.As(err, &qqError) {
		switch qqError.Kind {
		case qqmusic.ErrorKindInvalidInput:
			return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
		case qqmusic.ErrorKindUnauthorized:
			return fmt.Errorf("%w: %v", domain.ErrProviderReconnectRequired, err)
		case qqmusic.ErrorKindUnavailable:
			return fmt.Errorf("%w: %v", domain.ErrNotFound, err)
		default:
			return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
		}
	}
	var netEaseError *netease.ProviderError
	if errors.As(err, &netEaseError) {
		switch netEaseError.Kind {
		case netease.ErrorKindInvalid:
			return fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err)
		case netease.ErrorKindUnauthorized:
			return fmt.Errorf("%w: %v", domain.ErrProviderReconnectRequired, err)
		case netease.ErrorKindUnavailable:
			return fmt.Errorf("%w: %v", domain.ErrNotFound, err)
		default:
			return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
		}
	}
	if errors.Is(err, spotify.ErrInvalidCredentials) || errors.Is(err, spotify.ErrInvalidLoginAttempt) {
		return fmt.Errorf("%w: Spotify credentials are invalid", domain.ErrProviderReconnectRequired)
	}
	var spotifyError *spotify.APIError
	if errors.As(err, &spotifyError) {
		if spotifyError.StatusCode == 401 || spotifyError.Code == "invalid_grant" {
			return fmt.Errorf("%w: Spotify authorization expired", domain.ErrProviderReconnectRequired)
		}
		return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
	}
	return fmt.Errorf("%w: provider request failed", domain.ErrUnavailable)
}
