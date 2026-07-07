package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1/mev1connect"
)

type Server struct {
	config     Config
	store      *StatusStore
	identity   *IdentityStore
	prometheus *PrometheusClient
	github     *GitHubStatusPublisher
	profile    *ProfileService
}

func NewServer(config Config, store *StatusStore, identity *IdentityStore, prometheus *PrometheusClient, github *GitHubStatusPublisher, profile *ProfileService) *Server {
	return &Server{
		config:     config,
		store:      store,
		identity:   identity,
		prometheus: prometheus,
		github:     github,
		profile:    profile,
	}
}

// Handler builds the Gin engine. Gin owns routing, CORS, and recovery; the
// ConnectRPC services are mounted onto it, and the REST/Prometheus/static
// endpoints are registered as Gin routes.
func (server *Server) Handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), corsMiddleware())

	router.GET("/healthz", gin.WrapF(server.health))
	router.POST("/api/ingest/mobile", gin.WrapF(server.ingestMobile))
	router.POST("/api/ingest/host", gin.WrapF(server.ingestDevice))
	router.POST("/api/ingest/agent", gin.WrapF(server.ingestAgent))
	router.POST("/api/prometheus/register", gin.WrapF(server.registerPrometheusTargets))
	router.GET("/api/prometheus/http-sd/:job", server.prometheusHTTPDiscovery)
	router.GET("/api/public-status", gin.WrapF(server.publicStatus))
	router.GET("/api/profile", gin.WrapF(server.profilePage))
	router.GET("/api/internal/status", gin.WrapF(server.internalStatus))
	router.GET("/api/internal/metrics/query", gin.WrapF(server.internalMetricQuery))
	router.GET("/api/internal/metrics/query_range", gin.WrapF(server.internalMetricQueryRange))
	router.GET("/metrics", gin.WrapF(server.metrics))

	server.mountConnectServices(router)

	if server.config.StaticDir != "" {
		router.NoRoute(gin.WrapF(server.static))
	}
	return router
}

// mountConnectServices mounts the ConnectRPC handlers onto the Gin router. The
// EnrollmentService is authenticated because minting a device identity requires
// a valid ingest token.
func (server *Server) mountConnectServices(router *gin.Engine) {
	enrollPath, enrollHandler := mev1connect.NewEnrollmentServiceHandler(
		NewEnrollmentServer(server.identity),
		connect.WithInterceptors(NewAuthInterceptor(server.config)),
	)
	router.Any(enrollPath+"*any", gin.WrapH(enrollHandler))
}

func corsMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		context.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
		if context.Request.Method == http.MethodOptions {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}
		context.Next()
	}
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

func (server *Server) registerPrometheusTargets(writer http.ResponseWriter, request *http.Request) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input PrometheusTargetRegistration
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil || !validatePrometheusRegistration(&input) {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid_payload"})
		return
	}
	if err := server.store.RegisterPrometheusTargets(input.Targets, time.Now()); err != nil {
		slog.Error("failed to save prometheus scrape targets", "error", err)
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "store_failed"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
}

func (server *Server) prometheusHTTPDiscovery(context *gin.Context) {
	job := context.Param("job")
	if !validPrometheusJob(job) {
		writeJSON(context.Writer, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	writeJSON(context.Writer, http.StatusOK, server.store.PrometheusHTTPDiscovery(job))
}

func (server *Server) publicStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, server.status(request.Context()))
}

func (server *Server) profilePage(writer http.ResponseWriter, _ *http.Request) {
	writeProtoJSON(writer, http.StatusOK, server.profile.Page(time.Now()))
}

func (server *Server) internalStatus(writer http.ResponseWriter, request *http.Request) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	writeJSON(writer, http.StatusOK, server.internalSnapshot(request.Context()))
}

func (server *Server) internalMetricQuery(writer http.ResponseWriter, request *http.Request) {
	server.prometheusProxy(writer, request, "/api/v1/query", []string{"query", "time", "timeout"})
}

func (server *Server) internalMetricQueryRange(writer http.ResponseWriter, request *http.Request) {
	server.prometheusProxy(writer, request, "/api/v1/query_range", []string{"query", "start", "end", "step", "timeout"})
}

func (server *Server) prometheusProxy(writer http.ResponseWriter, request *http.Request, path string, allowedParams []string) {
	if !server.config.Authorized(request.Header.Get("Authorization")) {
		writeJSON(writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	params, ok := prometheusParams(request.URL.Query(), allowedParams)
	if !ok {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid_query"})
		return
	}
	body, status, err := server.prometheus.Proxy(request.Context(), path, params)
	if err != nil {
		slog.Error("prometheus proxy failed", "error", err)
		writeJSON(writer, http.StatusBadGateway, map[string]string{"error": "prometheus_unavailable"})
		return
	}

	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_, _ = writer.Write(body)
}

func (server *Server) metrics(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = writer.Write([]byte(RenderMetrics(server.store.Snapshot())))
}

func (server *Server) status(ctx context.Context) PublicStatus {
	snapshot := server.store.Snapshot()
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339)
	agents := publicAgents(mergeAgentStatuses(
		server.prometheus.AgentStatuses(ctx),
		runningAgents(snapshot.Agents, nowTime, time.Duration(server.config.AgentFreshSeconds)*time.Second),
	))
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
		Devices: mergePrometheusDevices(server.prometheus.NodeExporterStatuses(ctx), nonServerDevices(snapshot.Devices)),
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

func (server *Server) internalSnapshot(ctx context.Context) InternalStatus {
	public := server.status(ctx)
	snapshot := server.store.Snapshot()
	github := snapshot.GitHub
	if server.config.GitHubToken == "" {
		github.Configured = false
		github.State = GitHubSyncDisabled
	} else if !github.Configured {
		github.State = GitHubSyncPending
	}
	return InternalStatus{
		Server:    public.Server,
		Mobile:    public.Mobile,
		Devices:   public.Devices,
		Agents:    public.Agents,
		GitHub:    github,
		UpdatedAt: public.UpdatedAt,
	}
}

func mergeAgentStatuses(primary []StoredAgentStatus, fallback []StoredAgentStatus) []StoredAgentStatus {
	merged := make(map[string]StoredAgentStatus, len(primary)+len(fallback))
	for _, agent := range fallback {
		merged[agentStoreKey(agent.AgentIngest)] = agent
	}
	for _, agent := range primary {
		merged[agentStoreKey(agent.AgentIngest)] = agent
	}

	result := make([]StoredAgentStatus, 0, len(merged))
	for _, agent := range merged {
		result = append(result, agent)
	}
	sort.Slice(result, func(left int, right int) bool {
		if result[left].DeviceID != result[right].DeviceID {
			return result[left].DeviceID < result[right].DeviceID
		}
		return result[left].AgentID < result[right].AgentID
	})
	return result
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
		if device.Media != nil {
			base.Media = device.Media
		}
		if len(device.Accessories) > 0 {
			base.Accessories = device.Accessories
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
		if device.Media != nil {
			existing.Media = device.Media
		}
		if len(device.Accessories) > 0 {
			existing.Accessories = device.Accessories
		}
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

func prometheusParams(query url.Values, allowed []string) (url.Values, bool) {
	if len(query.Get("query")) == 0 || len(query.Get("query")) > 4096 {
		return nil, false
	}
	filtered := url.Values{}
	for _, key := range allowed {
		value := query.Get(key)
		if len(value) > 512 {
			return nil, false
		}
		if value != "" {
			filtered.Set(key, value)
		}
	}
	return filtered, true
}

func (server *Server) static(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/api/") {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}

	relativePath := strings.TrimPrefix(filepath.Clean("/"+request.URL.Path), "/")
	if relativePath == "." {
		relativePath = ""
	}
	filePath := filepath.Join(server.config.StaticDir, relativePath)
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		http.ServeFile(writer, request, filePath)
		return
	}
	http.ServeFile(writer, request, filepath.Join(server.config.StaticDir, "index.html"))
}

func firstString(primary string, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

// writeProtoJSON serializes a protobuf message using the canonical proto3 JSON
// mapping (camelCase field names, RFC 3339 timestamps, string enums) so the
// generated client types decode it without any manual mapping.
func writeProtoJSON(writer http.ResponseWriter, status int, message proto.Message) {
	data, err := protojson.Marshal(message)
	if err != nil {
		slog.Error("failed to encode protobuf response", "error", err)
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": "encode_failed"})
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_, _ = writer.Write(data)
}
