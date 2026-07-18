package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const providerSearchPageSize = 10

func (s *MusicProviderService) SearchMusic(ctx context.Context, query string, cursors map[domain.MusicProvider]string) ([]domain.ProviderSearchGroup, error) {
	query = strings.TrimSpace(query)
	if query == "" || len([]rune(query)) > 256 {
		return nil, fmt.Errorf("%w: music search query must contain 1 to 256 characters", domain.ErrInvalidArgument)
	}
	for provider := range cursors {
		if provider == domain.MusicProviderLocal {
			continue
		}
		if _, registered := s.providers.Get(provider); !registered {
			return nil, fmt.Errorf("%w: unknown music provider", domain.ErrInvalidArgument)
		}
	}
	providers := s.orderedSearchProviders(cursors)
	groups := make([]domain.ProviderSearchGroup, len(providers))
	var wait sync.WaitGroup
	for index, provider := range providers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			groups[index] = s.searchProvider(ctx, provider, query, cursors[provider])
		}()
	}
	wait.Wait()
	return groups, nil
}

// ResolvePlayback creates a fresh source-specific descriptor.
func (s *MusicProviderService) searchProvider(ctx context.Context, provider domain.MusicProvider, query, pageToken string) domain.ProviderSearchGroup {
	group := domain.ProviderSearchGroup{Provider: provider}
	if provider == domain.MusicProviderLocal {
		page, err := s.store.ListTracks(ctx, domain.TrackListQuery{
			Query: query, PageSize: providerSearchPageSize, PageToken: pageToken,
		})
		if err != nil {
			group.Status = domain.ProviderSearchUnavailable
			return group
		}
		group.Status = domain.ProviderSearchReady
		group.NextPageToken = page.NextPageToken
		for _, track := range page.Tracks {
			group.Tracks = append(group.Tracks, playableTrackFromLocal(track))
		}
		return group
	}
	adapter, adapterErr := s.providerAdapter(provider)
	searcher, searchable := adapter.(domain.MusicCatalogSearcher)
	if adapterErr != nil || !adapter.Configured() || !searchable {
		group.Status = domain.ProviderSearchNotConnected
		return group
	}
	connection, credentials, err := s.providerCredentials(ctx, provider)
	if errors.Is(err, domain.ErrNotFound) {
		group.Status = domain.ProviderSearchNotConnected
		return group
	}
	if err != nil {
		group.Status = domain.ProviderSearchUnavailable
		return group
	}
	tracks, nextPageToken, updated, err := searcher.Search(ctx, credentials, query, providerSearchPageSize, pageToken)
	if err != nil {
		if errors.Is(err, domain.ErrProviderReconnectRequired) {
			group.Status = domain.ProviderSearchReconnectRequired
		} else {
			group.Status = domain.ProviderSearchUnavailable
		}
		s.markProviderFailure(ctx, connection, err)
		return group
	}
	if err := s.persistCredentialUpdate(ctx, connection, credentials, updated); err != nil {
		group.Status = domain.ProviderSearchUnavailable
		return group
	}
	for index, track := range tracks {
		if track.Provider != provider {
			group.Status = domain.ProviderSearchUnavailable
			return group
		}
		normalized, err := s.tracks.Validate(ctx, track)
		if err != nil {
			group.Status = domain.ProviderSearchUnavailable
			return group
		}
		tracks[index] = normalized
	}
	group.Status = domain.ProviderSearchReady
	group.Tracks = tracks
	group.NextPageToken = nextPageToken
	return group
}

func (s *MusicProviderService) orderedSearchProviders(cursors map[domain.MusicProvider]string) []domain.MusicProvider {
	order := []domain.MusicProvider{domain.MusicProviderLocal}
	for _, adapter := range s.providers.List() {
		order = append(order, adapter.Provider())
	}
	if len(cursors) == 0 {
		return order
	}
	providers := make([]domain.MusicProvider, 0, len(cursors))
	for _, provider := range order {
		if _, found := cursors[provider]; found {
			providers = append(providers, provider)
		}
	}
	return providers
}
