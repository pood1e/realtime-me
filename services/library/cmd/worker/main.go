package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pood1e/realtime-me/services/library/internal/app"
	"github.com/pood1e/realtime-me/services/library/internal/config"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider"
	"github.com/pood1e/realtime-me/services/library/internal/storage"
	"github.com/pood1e/realtime-me/services/library/internal/store/postgres"
	contentworker "github.com/pood1e/realtime-me/services/library/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(logger); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.LoadWorker()
	if err != nil {
		return err
	}
	startupContext, cancelStartup := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelStartup()
	store, err := postgres.Open(startupContext, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	files, err := storage.NewFilesystem(cfg.DataRoot)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	credentialBox, err := provider.NewSecretBox(cfg.ProviderCredentialKey)
	if err != nil {
		return err
	}
	providerRegistry, err := provider.NewRegistry(provider.RegistryConfig{
		SpotifyClientID: cfg.SpotifyClientID, SpotifyClientSecret: cfg.SpotifyClientSecret,
		SpotifyRedirectURI: cfg.SpotifyRedirectURI,
	})
	if err != nil {
		return err
	}
	downloadHTTPClient, err := newDownloadHTTPClient()
	if err != nil {
		return err
	}
	clock := domain.SystemClock{}
	playlistService := app.NewMusicPlaylistService(store, clock, app.MusicProviderDependencies{
		Store: store, Registry: providerRegistry, Credentials: credentialBox,
	})
	music := contentworker.MusicProcessors{
		Downloads: contentworker.NewMusicDownloader(
			store, store, providerRegistry, credentialBox, files, downloadHTTPClient, clock, cfg.ReservedFreeBytes,
		),
		Artwork:         contentworker.NewMusicArtworkImporter(store, files, downloadHTTPClient),
		PlaylistImports: contentworker.NewPlaylistImporter(store, playlistService),
	}
	logger.Info("content worker started")
	return contentworker.New(store, files, clock, logger, music).Run(ctx)
}

func newDownloadHTTPClient() (*http.Client, error) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("standard HTTP transport is unavailable")
	}
	transport := defaultTransport.Clone()
	transport.DialContext = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.ResponseHeaderTimeout = 20 * time.Second
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Minute,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many music download redirects")
			}
			return nil
		},
	}, nil
}
