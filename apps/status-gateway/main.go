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

	"realtime-me/apps/status-gateway/internal/gateway"
)

func main() {
	config := gateway.LoadConfig()
	if err := config.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	store := gateway.NewStatusStore(config.StateFile)
	if err := store.Load(); err != nil {
		// A corrupt or unreadable state file must not be silently overwritten
		// with an empty snapshot on the next update; fail loudly instead.
		slog.Error("failed to load state", "error", err)
		os.Exit(1)
	}

	identity := gateway.NewIdentityStore(config.IdentityStateFile)
	if err := identity.Load(); err != nil {
		slog.Error("failed to load identity state", "error", err)
		os.Exit(1)
	}

	// This process also carries phone ingest and Prometheus scrape discovery, so a
	// document it cannot read must not take the metrics pipeline down with it. Say
	// so here, and let that one service answer unavailable; the rest serve on.
	profileConfig, profileErr := gateway.LoadProfileConfig(config.ProfileConfigFile)
	if profileErr != nil {
		slog.Error("failed to load profile config", "path", config.ProfileConfigFile, "error", profileErr)
	}

	projectsConfig, projectsErr := gateway.LoadProjectsConfig(config.ProjectsConfigFile)
	if projectsErr != nil {
		slog.Error("failed to load projects config", "path", config.ProjectsConfigFile, "error", projectsErr)
	}

	metrics, err := gateway.NewMetricsExporter(store)
	if err != nil {
		slog.Error("failed to initialize metrics exporter", "error", err)
		os.Exit(1)
	}

	prometheus := gateway.NewPrometheusClient(config.PrometheusURL)
	github := gateway.NewGitHubStatusPublisher(config, store)
	profile := gateway.NewProfileService(profileConfig, profileErr)
	projects := gateway.NewProjectsService(
		projectsConfig,
		projectsErr,
		gateway.NewGitHubProjectsClient(config.GitHubProjectsToken),
		config.GitHubProjectsRefreshHours,
	)
	server := gateway.NewServer(config, store, identity, prometheus, github, profile, projects, metrics.Handler())

	httpServer := &http.Server{
		Addr:              ":" + config.Port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go github.Run(ctx)
	go projects.Run(ctx)

	go func() {
		slog.Info("status-gateway listening", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("status-gateway stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("status-gateway shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("status-gateway shutdown error", "error", err)
	}
}
