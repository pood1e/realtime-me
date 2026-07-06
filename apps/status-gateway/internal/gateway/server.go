package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	config     Config
	store      *StatusStore
	prometheus *PrometheusClient
	github     *GitHubStatusPublisher
}

func NewServer(config Config, store *StatusStore, prometheus *PrometheusClient, github *GitHubStatusPublisher) *Server {
	return &Server{
		config:     config,
		store:      store,
		prometheus: prometheus,
		github:     github,
	}
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.health)
	mux.HandleFunc("POST /api/ingest/mobile", server.ingestMobile)
	mux.HandleFunc("POST /api/ingest/agent", server.ingestAgent)
	mux.HandleFunc("GET /api/public-status", server.publicStatus)
	mux.HandleFunc("GET /api/internal-status", server.internalStatus)
	mux.HandleFunc("GET /metrics", server.metrics)
	return cors(mux)
}

func (server *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
}

func (server *Server) ingestMobile(writer http.ResponseWriter, request *http.Request) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input MobileIngest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil || !validateMobile(&input) {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
		return
	}
	normalizeMobile(&input)
	mobile, err := server.store.UpdateMobile(input, time.Now())
	if err != nil {
		slog.Error("failed to save mobile status", "error", err)
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "store_failed"})
		return
	}
	if err := server.github.Publish(request.Context(), mobile); err != nil {
		slog.Error("failed to publish github status", "error", err)
	}
	writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
}

func (server *Server) ingestAgent(writer http.ResponseWriter, request *http.Request) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input AgentIngest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil || !validateAgent(&input) {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
		return
	}
	if _, err := server.store.UpdateAgent(input, time.Now()); err != nil {
		slog.Error("failed to save agent status", "error", err)
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "store_failed"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
}

func (server *Server) publicStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, server.status(request.Context()))
}

func (server *Server) internalStatus(writer http.ResponseWriter, request *http.Request) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	status := server.status(request.Context())
	writeJSON(writer, http.StatusOK, map[string]any{
		"server":     status.Server,
		"mobile":     status.Mobile,
		"agents":     status.Agents,
		"github":     status.GitHub,
		"updated_at": status.UpdatedAt,
		"raw":        server.store.Snapshot(),
	})
}

func (server *Server) metrics(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = writer.Write([]byte(RenderMetrics(server.store.Snapshot())))
}

func (server *Server) status(ctx context.Context) PublicStatus {
	snapshot := server.store.Snapshot()
	now := time.Now().UTC().Format(time.RFC3339)
	agents := snapshot.Agents
	if len(agents) == 0 && server.config.PublicAgentPlaceholder {
		agents = []StoredAgentStatus{{
			AgentIngest: AgentIngest{
				AgentID: "agents",
				State:   "idle",
				Task:    "not connected",
			},
			ReceivedAt: now,
		}}
		agents[0].UpdatedAt = now
	}

	githubState := snapshot.GitHub.State
	if server.config.GitHubToken == "" {
		githubState = GitHubSyncDisabled
	} else if !snapshot.GitHub.Configured {
		githubState = GitHubSyncPending
	}

	return PublicStatus{
		Server: server.prometheus.ServerSummary(ctx),
		Mobile: snapshot.Mobile,
		Agents: agents,
		GitHub: PublicGitHubStatus{
			Enabled:   server.config.GitHubToken != "",
			State:     githubState,
			UpdatedAt: snapshot.GitHub.LastSuccessAt,
			Emoji:     snapshot.GitHub.Emoji,
			Message:   snapshot.GitHub.Message,
		},
		UpdatedAt: now,
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
