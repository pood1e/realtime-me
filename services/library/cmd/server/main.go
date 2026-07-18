package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pood1e/realtime-me/libs/go/authn"
	"github.com/pood1e/realtime-me/services/library/internal/app"
	"github.com/pood1e/realtime-me/services/library/internal/config"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
	"github.com/pood1e/realtime-me/services/library/internal/provider"
	"github.com/pood1e/realtime-me/services/library/internal/storage"
	"github.com/pood1e/realtime-me/services/library/internal/store/postgres"
	"github.com/pood1e/realtime-me/services/library/internal/transport"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	startupContext, cancelStartup := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelStartup()
	access, err := authn.NewVerifier(cfg.AccessConfig())
	if err != nil {
		return err
	}
	clock := domain.SystemClock{}
	store, err := postgres.Open(startupContext, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	files, err := storage.NewFilesystem(cfg.DataRoot)
	if err != nil {
		return err
	}
	credentialBox, err := provider.NewSecretBox(cfg.ProviderCredentialKey)
	if err != nil {
		return err
	}
	providerRegistry, err := provider.NewRegistry(provider.RegistryConfig{
		SpotifyClientID: cfg.SpotifyClientID, SpotifyClientSecret: cfg.SpotifyClientSecret,
		SpotifyRedirectURI: cfg.PrivateAPIOrigin() + "/v1/music/providers/spotify/callback",
	})
	if err != nil {
		return err
	}
	content := app.NewContentService(store, store, files, clock, cfg.ChunkSizeBytes, cfg.ReservedFreeBytes, cfg.UploadTTL)
	suite := &app.Suite{
		Content: content,
		Drive:   app.NewDriveService(store, content, files, clock, cfg.PublicSiteOrigin),
		Books:   app.NewBookService(store, store, content, files),
		Music: app.NewMusicSuite(store, store, content, files, clock, app.MusicProviderDependencies{
			Store: store, Registry: providerRegistry, Credentials: credentialBox,
		}),
		Images:     app.NewImageService(store, store, content, files, cfg.PublicSiteOrigin),
		Wallpapers: app.NewWallpaperService(store, store, files, cfg.PublicSiteOrigin),
		System:     app.NewSystemService(store, store, files, clock),
		Retention:  app.NewRetentionService(store, content, clock),
	}
	purgeContext, cancelPurge := context.WithTimeout(context.Background(), 10*time.Minute)
	if err := suite.Retention.PurgeExpired(purgeContext); err != nil {
		logger.Error("startup retention cleanup failed", "error", err)
	}
	cancelPurge()
	go runPurger(logger, suite.Retention)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           transport.NewHTTPHandler(cfg, suite, access, logger),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server listening", "address", cfg.ListenAddr)
		serverErrors <- server.ListenAndServe()
	}()
	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case signalValue := <-shutdownSignals:
		logger.Info("shutdown signal received", "signal", signalValue.String())
		shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelShutdown()
		return server.Shutdown(shutdownContext)
	}
}

func runPurger(logger *slog.Logger, retention *app.RetentionService) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		if err := retention.PurgeExpired(ctx); err != nil {
			logger.Error("retention cleanup failed", "error", err)
		}
		cancel()
	}
}
