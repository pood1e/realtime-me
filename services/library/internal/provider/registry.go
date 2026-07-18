package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider/failure"
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
	qqAdapter, err := newQQAdapter()
	if err != nil {
		return nil, err
	}
	spotifyAdapter, err := newSpotifyAdapter(config)
	if err != nil {
		return nil, err
	}
	registry := &Registry{byProvider: make(map[domain.MusicProvider]domain.MusicProviderAdapter)}
	for _, adapter := range []domain.MusicProviderAdapter{
		qqAdapter,
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
	if adapter == nil || !domain.ValidMusicProviderID(adapter.Provider()) ||
		adapter.Provider() == domain.MusicProviderLocal || strings.TrimSpace(adapter.DisplayName()) == "" {
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
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	kind, classified := failure.Classify(err)
	if !classified {
		return fmt.Errorf("%w: provider request failed", domain.ErrUnavailable)
	}
	var domainError error
	switch kind {
	case failure.Invalid:
		domainError = domain.ErrInvalidArgument
	case failure.Unauthorized:
		domainError = domain.ErrProviderReconnectRequired
	case failure.Forbidden:
		domainError = domain.ErrForbidden
	case failure.NotFound:
		domainError = domain.ErrNotFound
	case failure.RateLimited:
		domainError = domain.ErrRateLimited
	default:
		domainError = domain.ErrUnavailable
	}
	return fmt.Errorf("%w: %v", domainError, err)
}
