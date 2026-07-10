package main

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/auth"
	"example.com/cloud-drive/api/internal/config"
	"example.com/cloud-drive/api/internal/domain"
	"example.com/cloud-drive/api/internal/storage"
	"example.com/cloud-drive/api/internal/store/postgres"
	"example.com/cloud-drive/api/internal/transport"
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
	clock := domain.SystemClock{}
	sessions, err := auth.NewManager(cfg.PasswordHash, cfg.SessionSecret, clock, rand.Reader)
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
	blobs, err := storage.NewFilesystem(cfg.DataRoot)
	if err != nil {
		return err
	}
	service := app.NewService(store, blobs, clock, cfg.ChunkSizeBytes, cfg.ReservedFreeBytes, cfg.UploadTTL, cfg.ShareAppOrigin)

	purgeContext, cancelPurge := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := service.PurgeExpired(purgeContext); err != nil {
		logger.Error("startup retention cleanup failed", "error", err)
	}
	cancelPurge()
	go runPurger(logger, service)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           transport.NewHTTPHandler(cfg, service, sessions, logger),
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

func runPurger(logger *slog.Logger, service *app.Service) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		context, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		if err := service.PurgeExpired(context); err != nil {
			logger.Error("retention cleanup failed", "error", err)
		}
		cancel()
	}
}
