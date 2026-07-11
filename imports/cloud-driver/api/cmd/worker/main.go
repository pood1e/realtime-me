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

	"example.com/cloud-drive/api/internal/config"
	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/provider"
	"example.com/cloud-drive/api/internal/storage"
	"example.com/cloud-drive/api/internal/store/postgres"
	contentworker "example.com/cloud-drive/api/internal/worker"
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
	providerRegistry, err := provider.NewRegistry(provider.RegistryConfig{})
	if err != nil {
		return err
	}
	downloadHTTPClient, err := newDownloadHTTPClient()
	if err != nil {
		return err
	}
	clock := domain.SystemClock{}
	downloads := contentworker.NewMusicDownloader(
		store, store, providerRegistry, credentialBox, files, downloadHTTPClient, clock, cfg.ReservedFreeBytes,
	)
	logger.Info("content worker started")
	return contentworker.New(store, files, clock, logger, downloads).Run(ctx)
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
