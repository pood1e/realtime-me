package gateway

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	"realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1/mev1connect"
)

type Server struct {
	config     Config
	store      *StatusStore
	identity   *IdentityStore
	prometheus *PrometheusClient
	github     *GitHubStatusPublisher
	profile    *ProfileService
	metrics    http.Handler
}

func NewServer(config Config, store *StatusStore, identity *IdentityStore, prometheus *PrometheusClient, github *GitHubStatusPublisher, profile *ProfileService, metrics http.Handler) *Server {
	return &Server{
		config:     config,
		store:      store,
		identity:   identity,
		prometheus: prometheus,
		github:     github,
		profile:    profile,
		metrics:    metrics,
	}
}

// Handler builds the Gin engine. Gin owns routing, CORS, and recovery; the
// ConnectRPC services are mounted onto it, and the Prometheus/static endpoints
// remain plain HTTP routes because Prometheus does not speak the Connect
// protocol.
func (server *Server) Handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), corsMiddleware())

	router.GET("/healthz", gin.WrapF(server.health))
	router.GET("/api/prometheus/http-sd/:job", server.prometheusHTTPDiscovery)
	router.GET("/metrics", gin.WrapH(server.metrics))

	server.mountConnectServices(router)

	if server.config.StaticDir != "" {
		router.NoRoute(gin.WrapF(server.static))
	}
	return router
}

// mountConnectServices mounts the ConnectRPC handlers. Enrollment and ingest
// require an ingest token; the profile and public status are unauthenticated,
// while internal status requires a token.
func (server *Server) mountConnectServices(router *gin.Engine) {
	mount := func(path string, handler http.Handler) {
		router.Any(path+"*any", gin.WrapH(handler))
	}

	mount(mev1connect.NewEnrollmentServiceHandler(
		NewEnrollmentServer(server.identity),
		connect.WithInterceptors(NewAuthInterceptor(server.config.IngestTokens)),
	))
	mount(mev1connect.NewIngestServiceHandler(
		NewIngestServer(server.store, server.identity, server.github),
		connect.WithInterceptors(NewAuthInterceptor(server.config.IngestTokens)),
	))
	mount(mev1connect.NewStatusServiceHandler(
		NewStatusServer(server.store, server.prometheus, server.config),
		connect.WithInterceptors(NewAuthInterceptor(server.config.QueryTokens, mev1connect.StatusServiceGetPublicStatusProcedure)),
	))
	mount(mev1connect.NewMetricsServiceHandler(
		NewMetricsServer(server.prometheus),
		connect.WithInterceptors(NewAuthInterceptor(server.config.QueryTokens)),
	))
	mount(mev1connect.NewProfileServiceHandler(
		NewProfileServer(server.profile),
	))
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

// prometheusHTTPDiscovery is gated behind the read token because it enumerates
// every device's name, model, and LAN scrape address.
func (server *Server) prometheusHTTPDiscovery(context *gin.Context) {
	if !server.config.AuthorizedQuery(context.Request.Header.Get("Authorization")) {
		writeJSON(context.Writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	job, ok := parseScrapeJob(context.Param("job"))
	if !ok {
		writeJSON(context.Writer, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	writeJSON(context.Writer, http.StatusOK, server.store.PrometheusHTTPDiscovery(job, server.identity.Lookup))
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

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
