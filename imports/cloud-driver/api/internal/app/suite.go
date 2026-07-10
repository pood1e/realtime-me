package app

import (
	"context"
	"time"

	"example.com/cloud-drive/api/internal/domain"
)

const trashRetention = 30 * 24 * time.Hour

// Suite contains independently addressable peer application services.
type Suite struct {
	Content    *ContentService
	Drive      *DriveService
	Books      *BookService
	Music      *MusicService
	Images     *ImageService
	Wallpapers *WallpaperService
	System     *SystemService
	Retention  *RetentionService
}

// SystemService reports request-path and worker health.
type SystemService struct {
	drive domain.DriveStore
	jobs  domain.WorkerStore
	files ContentFiles
	clock domain.Clock
}

// NewSystemService constructs health reporting.
func NewSystemService(drive domain.DriveStore, jobs domain.WorkerStore, files ContentFiles, clock domain.Clock) *SystemService {
	return &SystemService{drive: drive, jobs: jobs, files: files, clock: clock}
}

// Check returns database, worker, queue, and storage state.
func (s *SystemService) Check(ctx context.Context) (bool, bool, domain.WorkerHealth, int64, error) {
	if err := s.drive.Ping(ctx); err != nil {
		return false, false, domain.WorkerHealth{}, 0, err
	}
	freeBytes, err := s.files.FreeBytes()
	if err != nil {
		return false, false, domain.WorkerHealth{}, 0, err
	}
	health, err := s.jobs.GetWorkerHealth(ctx)
	if err != nil {
		return false, false, domain.WorkerHealth{}, 0, err
	}
	workerHealthy := health.HeartbeatTime != nil && s.clock.Now().UTC().Sub(*health.HeartbeatTime) <= 2*time.Minute
	return true, workerHealthy, health, freeBytes, nil
}

// RetentionStore contains application trash retention operations.
type RetentionStore interface {
	PurgeTrashedItems(context.Context, time.Time) error
	PurgeTrashedBooks(context.Context, time.Time) error
	PurgeTrashedTracks(context.Context, time.Time) error
	PurgeTrashedImages(context.Context, time.Time) error
}

// RetentionService applies upload, trash, and content collection policies.
type RetentionService struct {
	store   RetentionStore
	content *ContentService
	clock   domain.Clock
}

// NewRetentionService constructs periodic cleanup.
func NewRetentionService(store RetentionStore, content *ContentService, clock domain.Clock) *RetentionService {
	return &RetentionService{store: store, content: content, clock: clock}
}

// PurgeExpired applies all bounded retention policies.
func (s *RetentionService) PurgeExpired(ctx context.Context) error {
	if err := s.content.PurgeExpiredUploads(ctx); err != nil {
		return err
	}
	cutoff := s.clock.Now().UTC().Add(-trashRetention)
	for _, purge := range []func(context.Context, time.Time) error{
		s.store.PurgeTrashedItems,
		s.store.PurgeTrashedBooks,
		s.store.PurgeTrashedTracks,
		s.store.PurgeTrashedImages,
	} {
		if err := purge(ctx, cutoff); err != nil {
			return err
		}
	}
	return s.content.CollectGarbage(ctx)
}
