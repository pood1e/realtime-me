package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	heartbeatInterval = 30 * time.Second
	pollInterval      = 2 * time.Second
	jobLease          = 10 * time.Minute
	commandTimeout    = 5 * time.Minute
)

// Files is the worker-owned filesystem boundary.
type Files interface {
	Open(context.Context, string) (*os.File, error)
	NewWorkDir(string) (string, error)
	RemoveWorkDir(string) error
	PublishArtifact(string, []byte, string, string, string) (string, error)
	PublishSource(context.Context, string, string) (domain.SealedContent, error)
	SealUpload(context.Context, string, string) (domain.SealedContent, error)
	FreeBytes() (int64, error)
}

// Worker performs bounded local metadata and derivative processing.
type Worker struct {
	store      domain.WorkerStore
	files      Files
	clock      domain.Clock
	logger     *slog.Logger
	processors map[domain.ProcessingJobKind]Processor
}

// Processor executes one typed job in an isolated working directory.
type Processor interface {
	Process(context.Context, domain.ProcessingJob, string) error
}

type processorFunc func(context.Context, domain.ProcessingJob, string) error

func (process processorFunc) Process(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	return process(ctx, job, workDir)
}

// MusicProcessors contains the independent provider-backed background jobs.
type MusicProcessors struct {
	Downloads       *MusicDownloader
	Artwork         *MusicArtworkImporter
	PlaylistImports *PlaylistImporter
}

type jobLane struct {
	name  string
	kinds []domain.ProcessingJobKind
}

var jobLanes = []jobLane{
	{
		name: "interactive",
		kinds: []domain.ProcessingJobKind{
			domain.ProcessingJobUploadFinalize,
			domain.ProcessingJobPlaylistImport,
			domain.ProcessingJobBook,
			domain.ProcessingJobTrack,
			domain.ProcessingJobImage,
			domain.ProcessingJobWallpaper,
		},
	},
	{
		name: "provider-download",
		kinds: []domain.ProcessingJobKind{
			domain.ProcessingJobMusicDownload,
			domain.ProcessingJobMusicArtwork,
		},
	},
}

// New constructs a single-concurrency content worker.
func New(store domain.WorkerStore, files Files, clock domain.Clock, logger *slog.Logger, music MusicProcessors) *Worker {
	worker := &Worker{
		store: store, files: files, clock: clock, logger: logger,
		processors: make(map[domain.ProcessingJobKind]Processor),
	}
	worker.register(domain.ProcessingJobBook, processorFunc(worker.processBook))
	worker.register(domain.ProcessingJobTrack, processorFunc(worker.processTrack))
	worker.register(domain.ProcessingJobImage, processorFunc(worker.processImage))
	worker.register(domain.ProcessingJobWallpaper, processorFunc(worker.processWallpaper))
	worker.register(domain.ProcessingJobUploadFinalize, processorFunc(worker.processUploadFinalize))
	if music.Downloads != nil {
		worker.register(domain.ProcessingJobMusicDownload, processorFunc(func(ctx context.Context, job domain.ProcessingJob, workDir string) error {
			return music.Downloads.Process(ctx, job, workDir)
		}))
	}
	if music.Artwork != nil {
		worker.register(domain.ProcessingJobMusicArtwork, processorFunc(func(ctx context.Context, job domain.ProcessingJob, workDir string) error {
			return music.Artwork.Process(ctx, job, workDir)
		}))
	}
	if music.PlaylistImports != nil {
		worker.register(domain.ProcessingJobPlaylistImport, processorFunc(music.PlaylistImports.Process))
	}
	return worker
}

// Run heartbeats, leases, and processes jobs until cancellation.
func (w *Worker) Run(ctx context.Context) error {
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()
	if err := w.store.HeartbeatWorker(ctx, w.clock.Now().UTC()); err != nil {
		return err
	}
	var lanes sync.WaitGroup
	for _, lane := range jobLanes {
		lanes.Add(1)
		go func() {
			defer lanes.Done()
			w.runLane(ctx, lane)
		}()
	}
	for {
		select {
		case <-ctx.Done():
			lanes.Wait()
			return ctx.Err()
		case now := <-heartbeat.C:
			w.recordHeartbeat(ctx, now)
		}
	}
}

func (w *Worker) runLane(ctx context.Context, lane jobLane) {
	for {
		processed, err := w.processNext(ctx, lane.kinds)
		if err != nil && !errors.Is(err, context.Canceled) {
			w.logger.Error("processing cycle failed", "lane", lane.name, "error", err)
		}
		if processed {
			continue
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (w *Worker) processNext(ctx context.Context, kinds []domain.ProcessingJobKind) (bool, error) {
	job, err := w.store.ClaimProcessingJob(ctx, w.clock.Now().UTC(), jobLease, kinds)
	if err != nil || job == nil {
		return false, err
	}
	started := time.Now()
	err = w.processWithLease(ctx, *job)
	if err == nil {
		if err := w.store.CompleteProcessingJob(ctx, *job); err != nil {
			return true, err
		}
		w.logger.Info("processing job completed", "job_id", job.UID, "kind", job.Kind, "duration", time.Since(started))
		return true, nil
	}
	if errors.Is(err, context.Canceled) && ctx.Err() != nil {
		return true, ctx.Err()
	}
	if errors.Is(err, domain.ErrNotFound) {
		// A resource can be purged or unpublished while its worker lease is active.
		// Completing is best-effort because the purge may also remove the job row.
		_ = w.store.CompleteProcessingJob(ctx, *job)
		w.logger.Info("processing job discarded", "job_id", job.UID, "kind", job.Kind)
		return true, nil
	}
	w.logger.Error("processing job failed", "job_id", job.UID, "kind", job.Kind, "attempt", job.Attempts, "error", err)
	decision := classifyFailure(err, job.Attempts, w.clock.Now().UTC())
	return true, w.store.FailProcessingJob(ctx, *job, decision.Code, decision.Retry, decision.RetryAt)
}

func (w *Worker) process(ctx context.Context, job domain.ProcessingJob) error {
	processor, found := w.processors[job.Kind]
	if !found {
		return &unsupportedProcessorError{kind: job.Kind}
	}
	workDir, err := w.files.NewWorkDir(job.UID)
	if err != nil {
		return err
	}
	defer w.files.RemoveWorkDir(workDir)
	return processor.Process(ctx, job, workDir)
}

func (w *Worker) processWithLease(ctx context.Context, job domain.ProcessingJob) error {
	jobContext, cancel := context.WithCancel(ctx)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- w.process(jobContext, job) }()
	renewal := time.NewTicker(jobLease / 3)
	defer renewal.Stop()
	for {
		select {
		case err := <-result:
			return err
		case <-ctx.Done():
			cancel()
			<-result
			return ctx.Err()
		case <-renewal.C:
			if err := w.store.ExtendProcessingJobLease(ctx, job, w.clock.Now().UTC().Add(jobLease)); err != nil {
				cancel()
				<-result
				return err
			}
		}
	}
}

func (w *Worker) register(kind domain.ProcessingJobKind, processor Processor) {
	if processor == nil {
		return
	}
	if _, duplicate := w.processors[kind]; duplicate {
		panic("duplicate processing job kind: " + string(kind))
	}
	w.processors[kind] = processor
}

func (w *Worker) recordHeartbeat(ctx context.Context, now time.Time) {
	if err := w.store.HeartbeatWorker(ctx, now.UTC()); err != nil {
		w.logger.Error("worker heartbeat failed", "error", err)
	}
}

func (w *Worker) materialize(ctx context.Context, content domain.ContentObject, workDir, extension string) (string, error) {
	file, err := w.files.Open(ctx, content.StorageKey)
	if err != nil {
		return "", err
	}
	defer file.Close()
	path := filepath.Join(workDir, "source"+extension)
	target, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("create processing source: %w", err)
	}
	_, copyErr := io.Copy(target, file)
	closeErr := target.Close()
	if copyErr != nil {
		return "", fmt.Errorf("copy processing source: %w", copyErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("close processing source: %w", closeErr)
	}
	return path, nil
}
