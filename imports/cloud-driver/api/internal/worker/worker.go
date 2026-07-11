package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
	PublishSource(string, string) (domain.SealedContent, error)
	FreeBytes() (int64, error)
}

// Worker performs bounded local metadata and derivative processing.
type Worker struct {
	store     domain.WorkerStore
	files     Files
	clock     domain.Clock
	logger    *slog.Logger
	downloads *MusicDownloader
}

// New constructs a single-concurrency content worker.
func New(store domain.WorkerStore, files Files, clock domain.Clock, logger *slog.Logger, downloads *MusicDownloader) *Worker {
	return &Worker{store: store, files: files, clock: clock, logger: logger, downloads: downloads}
}

// Run heartbeats, leases, and processes jobs until cancellation.
func (w *Worker) Run(ctx context.Context) error {
	heartbeat := time.NewTicker(heartbeatInterval)
	poll := time.NewTicker(pollInterval)
	defer heartbeat.Stop()
	defer poll.Stop()
	if err := w.store.HeartbeatWorker(ctx, w.clock.Now().UTC()); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-heartbeat.C:
			if err := w.store.HeartbeatWorker(ctx, now.UTC()); err != nil {
				w.logger.Error("worker heartbeat failed", "error", err)
			}
		case <-poll.C:
			if err := w.processNext(ctx); err != nil {
				w.logger.Error("processing cycle failed", "error", err)
			}
		}
	}
}

func (w *Worker) processNext(ctx context.Context) error {
	job, err := w.store.ClaimProcessingJob(ctx, w.clock.Now().UTC(), jobLease)
	if err != nil || job == nil {
		return err
	}
	started := time.Now()
	err = w.process(ctx, *job)
	if err == nil {
		if err := w.store.CompleteProcessingJob(ctx, *job); err != nil {
			return err
		}
		w.logger.Info("processing job completed", "job_id", job.UID, "kind", job.Kind, "duration", time.Since(started))
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) && job.Kind != "music_download" {
		// A resource can be purged or unpublished while its worker lease is active.
		// Completing is best-effort because the purge may also remove the job row.
		_ = w.store.CompleteProcessingJob(ctx, *job)
		w.logger.Info("processing job discarded", "job_id", job.UID, "kind", job.Kind)
		return nil
	}
	w.logger.Error("processing job failed", "job_id", job.UID, "kind", job.Kind, "attempt", job.Attempts, "error", err)
	retryTime := w.clock.Now().UTC().Add(time.Duration(job.Attempts*job.Attempts) * 30 * time.Second)
	return w.store.FailProcessingJob(ctx, *job, errorCode(err), retryTime)
}

func (w *Worker) process(ctx context.Context, job domain.ProcessingJob) error {
	workDir, err := w.files.NewWorkDir(job.UID)
	if err != nil {
		return err
	}
	defer w.files.RemoveWorkDir(workDir)
	switch job.Kind {
	case "book":
		return w.processBook(ctx, job, workDir)
	case "track":
		return w.processTrack(ctx, job, workDir)
	case "image":
		return w.processImage(ctx, job, workDir)
	case "wallpaper":
		return w.processWallpaper(ctx, job, workDir)
	case "music_download":
		if w.downloads == nil {
			return errors.New("music downloader is unavailable")
		}
		return w.downloads.Process(ctx, job.ResourceUID, workDir)
	default:
		return fmt.Errorf("unsupported job kind %q", job.Kind)
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
