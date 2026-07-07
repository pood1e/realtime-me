package main

import (
	"log/slog"
	"net/http"

	"realtime-me/apps/status-gateway/internal/gateway"
)

func main() {
	config := gateway.LoadConfig()
	store := gateway.NewStatusStore(config.StateFile)
	if err := store.Load(); err != nil {
		slog.Error("failed to load state", "error", err)
	}

	profileConfig, err := gateway.LoadProfileConfig(config.ProfileConfigFile)
	if err != nil {
		slog.Error("failed to load profile config", "error", err)
	}

	prometheus := gateway.NewPrometheusClient(config.PrometheusURL)
	github := gateway.NewGitHubStatusPublisher(config, store)
	profile := gateway.NewProfileService(profileConfig)
	server := gateway.NewServer(config, store, prometheus, github, profile)

	address := ":" + config.Port
	slog.Info("status-gateway listening", "address", address)
	if err := http.ListenAndServe(address, server.Handler()); err != nil {
		slog.Error("status-gateway stopped", "error", err)
	}
}
