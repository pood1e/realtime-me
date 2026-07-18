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
	"github.com/pood1e/realtime-me/services/status/internal/gateway"
)

func main() {
	config, err := gateway.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	if err := config.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	access, err := authn.NewVerifier(config.AccessConfig())
	if err != nil {
		slog.Error("failed to initialize OIDC verifier", "error", err)
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

	// The curated projects are data, not settings, and they live in their own file.
	// This process also carries phone ingest and Prometheus scrape discovery, so a
	// file it cannot read must not take the metrics pipeline down with it: say so
	// here, and let ProjectsService answer unavailable while the rest serve on.
	projectsConfig, projectsErr := gateway.LoadProjectsConfig(config.ProjectsFile)
	if projectsErr != nil {
		slog.Error("failed to load projects file", "path", config.ProjectsFile, "error", projectsErr)
	}

	metrics, err := gateway.NewMetricsExporter(store)
	if err != nil {
		slog.Error("failed to initialize metrics exporter", "error", err)
		os.Exit(1)
	}

	prometheus := gateway.NewPrometheusClient(config.PrometheusURL)
	github := gateway.NewGitHubStatusPublisher(config, store)
	profile := gateway.NewProfileService(config.Profile)
	projects := gateway.NewProjectsService(
		projectsConfig,
		projectsErr,
		gateway.NewGitHubProjectsClient(config.GitHubProjectsTokens),
		config.GitHubProjectsRefreshHours,
	)
	server := gateway.NewServer(config, store, identity, prometheus, github, profile, projects, metrics.Handler(), access)

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
