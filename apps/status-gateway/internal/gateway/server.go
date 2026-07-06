package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
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
	mux.HandleFunc("POST /api/ingest/host", server.ingestDevice)
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

func (server *Server) ingestDevice(writer http.ResponseWriter, request *http.Request) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input DeviceStatus
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil || !validateDevice(&input) {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
		return
	}
	if _, err := server.store.UpdateDevice(input, time.Now()); err != nil {
		slog.Error("failed to save device status", "error", err)
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "store_failed"})
		return
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
		"devices":    status.Devices,
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
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339)
	agents := publicAgents(runningAgents(snapshot.Agents, nowTime, time.Duration(server.config.AgentFreshSeconds)*time.Second))
	if len(agents) == 0 && server.config.PublicAgentPlaceholder {
		agents = []StoredAgentStatus{{
			AgentIngest: AgentIngest{
				AgentID: "agents",
				State:   "idle",
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
		Server:  mergeServerStatus(server.prometheus.ServerStatus(ctx), snapshot.Devices),
		Mobile:  snapshot.Mobile,
		Devices: mergePrometheusDevices(server.prometheus.VirtualMachineStatuses(ctx), nonServerDevices(snapshot.Devices)),
		Agents:  agents,
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

func runningAgents(agents []StoredAgentStatus, now time.Time, freshness time.Duration) []StoredAgentStatus {
	result := make([]StoredAgentStatus, 0, len(agents))
	for _, agent := range agents {
		if agent.State != "running" {
			continue
		}
		timestamp, err := time.Parse(time.RFC3339, firstString(agent.ReceivedAt, agent.UpdatedAt))
		if err != nil || now.Sub(timestamp) <= freshness {
			result = append(result, agent)
		}
	}
	return result
}

func publicAgents(agents []StoredAgentStatus) []StoredAgentStatus {
	result := make([]StoredAgentStatus, 0, len(agents))
	for _, agent := range agents {
		agent.Task = ""
		result = append(result, agent)
	}
	return result
}

func mergeServerStatus(base DeviceStatus, devices []StoredDeviceStatus) DeviceStatus {
	for _, device := range devices {
		if device.Role != "server" && device.DeviceID != "server" {
			continue
		}
		base.DeviceID = firstString(device.DeviceID, base.DeviceID)
		base.DeviceName = firstString(device.DeviceName, base.DeviceName)
		base.DeviceModel = firstString(device.DeviceModel, base.DeviceModel)
		base.Kind = firstString(device.Kind, base.Kind)
		base.Role = firstString(device.Role, base.Role)
		base.State = firstString(device.State, base.State)
		base.UpdatedAt = firstString(device.ReceivedAt, base.UpdatedAt)
		if len(device.Metrics) > 0 {
			base.Metrics = device.Metrics
		}
		break
	}
	if base.DeviceName == "" {
		base.DeviceName = "Server"
	}
	return base
}

func nonServerDevices(devices []StoredDeviceStatus) []StoredDeviceStatus {
	filtered := make([]StoredDeviceStatus, 0, len(devices))
	for _, device := range devices {
		if device.Role == "server" || device.DeviceID == "server" {
			continue
		}
		filtered = append(filtered, device)
	}
	return filtered
}

func mergePrometheusDevices(primary []DeviceStatus, stored []StoredDeviceStatus) []StoredDeviceStatus {
	merged := make(map[string]StoredDeviceStatus, len(primary)+len(stored))
	for _, device := range primary {
		merged[device.DeviceID] = StoredDeviceStatus{
			DeviceStatus: device,
			ReceivedAt:   device.UpdatedAt,
		}
	}
	for _, device := range stored {
		existing, ok := merged[device.DeviceID]
		if !ok {
			merged[device.DeviceID] = device
			continue
		}
		existing.DeviceName = firstString(device.DeviceName, existing.DeviceName)
		existing.DeviceModel = firstString(device.DeviceModel, existing.DeviceModel)
		existing.ReceivedAt = firstString(existing.ReceivedAt, device.ReceivedAt)
		merged[device.DeviceID] = existing
	}

	result := make([]StoredDeviceStatus, 0, len(merged))
	for _, device := range merged {
		result = append(result, device)
	}
	sortDeviceStatuses(result)
	return result
}

func sortDeviceStatuses(devices []StoredDeviceStatus) {
	sort.Slice(devices, func(left int, right int) bool {
		return devices[left].DeviceID < devices[right].DeviceID
	})
}

func firstString(primary string, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
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
