package worker

import (
	"context"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

// PlaylistImportResolver loads normalized provider playlist data.
type PlaylistImportResolver interface {
	ResolvePlaylistImport(context.Context, domain.PlaylistImport) (domain.Playlist, []domain.PlayableTrack, error)
}

// PlaylistImporter persists one durable provider playlist operation.
type PlaylistImporter struct {
	store    domain.WorkerStore
	resolver PlaylistImportResolver
}

// NewPlaylistImporter constructs the provider-backed import processor.
func NewPlaylistImporter(store domain.WorkerStore, resolver PlaylistImportResolver) *PlaylistImporter {
	return &PlaylistImporter{store: store, resolver: resolver}
}

// Process resolves and commits one playlist snapshot under its fenced lease.
func (p *PlaylistImporter) Process(ctx context.Context, job domain.ProcessingJob, _ string) error {
	operation, err := p.store.GetPlaylistImportForProcessing(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	playlist, tracks, err := p.resolver.ResolvePlaylistImport(ctx, operation)
	if err != nil {
		return err
	}
	return p.store.CompletePlaylistImport(ctx, job, playlist, tracks)
}
