package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const maximumTrackDownloadBytes int64 = 2 << 30

// MusicDownloadStore persists provider downloads and their local track links.
type MusicDownloadStore interface {
	GetMusicDownload(context.Context, string) (domain.MusicDownload, error)
	CompleteMusicDownload(context.Context, domain.ProcessingJob, domain.SealedContent) error
}

// CredentialProtector decrypts and rotates provider credentials.
type CredentialProtector interface {
	Seal(string, []byte) ([]byte, error)
	Open(string, []byte) ([]byte, error)
}

// MusicDownloader persists direct provider audio into the local content store.
type MusicDownloader struct {
	store             MusicDownloadStore
	providerStore     domain.MusicProviderStore
	providers         domain.MusicProviderRegistry
	credentials       CredentialProtector
	files             Files
	httpClient        *http.Client
	clock             domain.Clock
	reservedFreeBytes int64
}

// NewMusicDownloader constructs the background provider download service.
func NewMusicDownloader(
	store MusicDownloadStore,
	providerStore domain.MusicProviderStore,
	providers domain.MusicProviderRegistry,
	credentials CredentialProtector,
	files Files,
	httpClient *http.Client,
	clock domain.Clock,
	reservedFreeBytes int64,
) *MusicDownloader {
	return &MusicDownloader{
		store: store, providerStore: providerStore, providers: providers, credentials: credentials,
		files: files, httpClient: httpClient, clock: clock, reservedFreeBytes: reservedFreeBytes,
	}
}

// Process resolves and persists one playlist track.
func (d *MusicDownloader) Process(ctx context.Context, job domain.ProcessingJob, workDir string) error {
	download, err := d.store.GetMusicDownload(ctx, job.ResourceUID)
	if err != nil {
		return err
	}
	if download.PlaylistTrack.LocalTrackUID != "" {
		return nil
	}
	if download.Connection.Status == domain.ProviderReconnectRequired {
		return domain.ErrProviderReconnectRequired
	}
	adapter, found := d.providers.Get(download.PlaylistTrack.Track.Provider)
	if !found || !adapter.Configured() {
		return fmt.Errorf("%w: provider download plugin is unavailable", domain.ErrConflict)
	}
	downloader, supported := adapter.(domain.MusicTrackDownloader)
	if !supported {
		return fmt.Errorf("%w: provider does not expose downloadable audio", domain.ErrConflict)
	}
	purpose := domain.MusicProviderCredentialPurpose(download.Connection.Provider)
	rawCredentials, err := d.credentials.Open(purpose, download.Connection.EncryptedCredentials)
	if err != nil {
		return err
	}
	resource, updatedCredentials, err := downloader.ResolveDownload(ctx, rawCredentials, download.PlaylistTrack.Track.TrackID)
	if err != nil {
		d.markProviderFailure(ctx, download.Connection, err)
		return err
	}
	if err := d.persistCredentials(ctx, download.Connection, rawCredentials, updatedCredentials); err != nil {
		return err
	}
	sourcePath, fileName, err := d.fetch(ctx, resource, download.PlaylistTrack, workDir)
	if err != nil {
		return err
	}
	sealed, err := d.files.PublishSource(ctx, sourcePath, fileName)
	if err != nil {
		return err
	}
	return d.store.CompleteMusicDownload(ctx, job, sealed)
}

func (d *MusicDownloader) fetch(ctx context.Context, resource domain.ProviderDownload, item domain.PlaylistTrack, workDir string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(resource.URL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
		return "", "", fmt.Errorf("%w: provider download URL is invalid", domain.ErrUnavailable)
	}
	extension := extensionForContent(resource.ContentType)
	if extension == "" {
		return "", "", fmt.Errorf("%w: provider download format is unsupported", domain.ErrUnavailable)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", "", fmt.Errorf("create music download request: %w", err)
	}
	request.Header.Set("Accept", "audio/*,application/octet-stream")
	request.Header.Set("User-Agent", "cloud-drive-music/1")
	response, err := d.httpClient.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("download provider audio: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", "", fmt.Errorf("%w: provider audio returned HTTP %d", domain.ErrUnavailable, response.StatusCode)
	}
	limit, err := d.downloadLimit(response.ContentLength)
	if err != nil {
		return "", "", err
	}
	fileName := item.UID + extension
	sourcePath := filepath.Join(workDir, fileName)
	target, err := os.OpenFile(sourcePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", "", fmt.Errorf("create music download file: %w", err)
	}
	written, copyErr := io.Copy(target, io.LimitReader(response.Body, limit+1))
	closeErr := target.Close()
	if copyErr != nil {
		return "", "", fmt.Errorf("write music download: %w", copyErr)
	}
	if closeErr != nil {
		return "", "", fmt.Errorf("close music download: %w", closeErr)
	}
	if written == 0 {
		return "", "", fmt.Errorf("%w: provider returned an empty audio file", domain.ErrUnavailable)
	}
	if written > limit {
		return "", "", fmt.Errorf("%w: provider audio exceeds download limit", domain.ErrResourceExhausted)
	}
	return sourcePath, fileName, nil
}

func (d *MusicDownloader) downloadLimit(contentLength int64) (int64, error) {
	freeBytes, err := d.files.FreeBytes()
	if err != nil {
		return 0, err
	}
	available := freeBytes - d.reservedFreeBytes
	if available <= 0 {
		return 0, fmt.Errorf("%w: reserved free space reached", domain.ErrResourceExhausted)
	}
	limit := min(available, maximumTrackDownloadBytes)
	if contentLength > limit {
		return 0, fmt.Errorf("%w: provider audio exceeds available storage", domain.ErrResourceExhausted)
	}
	return limit, nil
}

func (d *MusicDownloader) persistCredentials(ctx context.Context, connection domain.ProviderConnection, previous, updated []byte) error {
	if len(updated) == 0 || bytes.Equal(previous, updated) {
		return nil
	}
	encrypted, err := d.credentials.Seal(domain.MusicProviderCredentialPurpose(connection.Provider), updated)
	if err != nil {
		return err
	}
	connection.EncryptedCredentials = encrypted
	connection.Status = domain.ProviderConnected
	connection.UpdateTime = d.clock.Now().UTC()
	_, err = d.providerStore.UpsertProviderConnection(ctx, connection)
	return err
}

func (d *MusicDownloader) markProviderFailure(ctx context.Context, connection domain.ProviderConnection, providerErr error) {
	if !errors.Is(providerErr, domain.ErrProviderReconnectRequired) {
		return
	}
	connection.Status = domain.ProviderReconnectRequired
	connection.UpdateTime = d.clock.Now().UTC()
	_, _ = d.providerStore.UpsertProviderConnection(ctx, connection)
}
