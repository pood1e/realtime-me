package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	maximumProviderArtworkBytes int64 = 16 << 20
	providerArtworkTimeout            = time.Minute
)

// MusicArtworkStore resolves and persists local artwork for provider tracks.
type MusicArtworkStore interface {
	GetMusicArtwork(context.Context, string) (domain.ContentObject, string, error)
	CompleteMusicArtwork(context.Context, domain.ProcessingJob, domain.Artifact) error
}

// MusicArtworkImporter stores a bounded provider image as a local JPEG artifact.
type MusicArtworkImporter struct {
	store      MusicArtworkStore
	files      Files
	httpClient *http.Client
}

// NewMusicArtworkImporter constructs the provider artwork processor.
func NewMusicArtworkImporter(store MusicArtworkStore, files Files, httpClient *http.Client) *MusicArtworkImporter {
	return &MusicArtworkImporter{store: store, files: files, httpClient: httpClient}
}

// Process imports artwork for one already-downloaded provider track.
func (i *MusicArtworkImporter) Process(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	content, artworkURL, err := i.store.GetMusicArtwork(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	sourcePath, err := i.fetch(ctx, artworkURL, workDir)
	if err != nil {
		return err
	}
	artworkPath := filepath.Join(workDir, "artwork.jpg")
	if err := runCommand(
		ctx,
		"vipsthumbnail",
		sourcePath,
		"--size",
		"1200x1200",
		"--output",
		artworkPath+"[Q=88,strip]",
	); err != nil {
		return fmt.Errorf("convert provider artwork: %w", err)
	}
	width, err := vipsDimension(ctx, artworkPath, "width")
	if err != nil {
		return err
	}
	height, err := vipsDimension(ctx, artworkPath, "height")
	if err != nil {
		return err
	}
	storageKey, err := i.files.PublishArtifact(
		artworkPath,
		content.SHA256,
		"track_artwork",
		"default",
		"jpg",
	)
	if err != nil {
		return err
	}
	return i.store.CompleteMusicArtwork(ctx, job, domain.Artifact{
		UID: uuid.NewString(), ContentUID: content.UID, Kind: "track_artwork", Variant: "default",
		ContentType: "image/jpeg", StorageKey: storageKey, Width: width, Height: height,
	})
}

func (i *MusicArtworkImporter) fetch(ctx context.Context, rawURL, workDir string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
		return "", fmt.Errorf("%w: provider artwork URL is invalid", domain.ErrUnavailable)
	}
	requestContext, cancel := context.WithTimeout(ctx, providerArtworkTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create provider artwork request: %w", err)
	}
	request.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/gif")
	request.Header.Set("User-Agent", "cloud-drive-music/1")
	response, err := i.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("download provider artwork: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: provider artwork returned HTTP %d", domain.ErrUnavailable, response.StatusCode)
	}
	if response.ContentLength > maximumProviderArtworkBytes {
		return "", fmt.Errorf("%w: provider artwork exceeds size limit", domain.ErrResourceExhausted)
	}
	path := filepath.Join(workDir, "artwork.source")
	if err := writeArtwork(path, response.Body); err != nil {
		return "", err
	}
	return path, nil
}

func writeArtwork(path string, source io.Reader) error {
	target, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create provider artwork file: %w", err)
	}
	written, copyErr := io.Copy(target, io.LimitReader(source, maximumProviderArtworkBytes+1))
	closeErr := target.Close()
	if copyErr != nil {
		return fmt.Errorf("write provider artwork: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close provider artwork: %w", closeErr)
	}
	if written == 0 {
		return fmt.Errorf("%w: provider returned empty artwork", domain.ErrUnavailable)
	}
	if written > maximumProviderArtworkBytes {
		return fmt.Errorf("%w: provider artwork exceeds size limit", domain.ErrResourceExhausted)
	}
	return nil
}
