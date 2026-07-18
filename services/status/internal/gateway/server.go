package gateway

import (
	"encoding/json"
	"net/http"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"

	authv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/auth/v1"
	sitev1connect "github.com/pood1e/realtime-me/gen/go/realtime/me/site/v1/sitev1connect"
	statusv1connect "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1/statusv1connect"
	"github.com/pood1e/realtime-me/libs/go/authn"
)

type Server struct {
	config     Config
	store      *StatusStore
	identity   *IdentityStore
	prometheus *PrometheusClient
	github     *GitHubStatusPublisher
	profile    *ProfileService
	projects   *ProjectsService
	metrics    http.Handler
	access     *authn.Verifier
}

func NewServer(config Config, store *StatusStore, identity *IdentityStore, prometheus *PrometheusClient, github *GitHubStatusPublisher, profile *ProfileService, projects *ProjectsService, metrics http.Handler, access *authn.Verifier) *Server {
	return &Server{
		config:     config,
		store:      store,
		identity:   identity,
		prometheus: prometheus,
		github:     github,
		profile:    profile,
		projects:   projects,
		metrics:    metrics,
		access:     access,
	}
}

// Handler builds the Gin engine. Gin owns routing and recovery; the
// ConnectRPC services are mounted onto it, and the Prometheus/static endpoints
// remain plain HTTP routes because Prometheus does not speak the Connect
// protocol.
func (server *Server) Handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/healthz", gin.WrapF(server.health))
	router.GET("/api/prometheus/http-sd/:job", server.prometheusHTTPDiscovery)
	router.GET("/metrics", server.prometheusMetrics)

	server.mountConnectServices(router)

	return router
}

// mountConnectServices mounts the ConnectRPC handlers. Enrollment and ingest
// require an ingest token; profile, projects, and public status are
// unauthenticated, while internal status and metrics require owner OIDC permission.
func (server *Server) mountConnectServices(router *gin.Engine) {
	mount := func(path string, handler http.Handler) {
		router.Any(path+"*any", gin.WrapH(handler))
	}

	mount(statusv1connect.NewEnrollmentServiceHandler(
		NewEnrollmentServer(server.identity),
		connect.WithInterceptors(NewTokenAuthInterceptor(server.config.IngestTokens)),
	))
	mount(statusv1connect.NewIngestServiceHandler(
		NewIngestServer(server.store, server.identity, server.github),
		connect.WithInterceptors(NewTokenAuthInterceptor(server.config.IngestTokens)),
	))
	mount(statusv1connect.NewStatusServiceHandler(
		NewStatusServer(server.store, server.prometheus, server.config),
		connect.WithInterceptors(NewPermissionInterceptor(server.access, authv1.Permission_PERMISSION_STATUS_INTERNAL_READ, statusv1connect.StatusServiceGetPublicStatusProcedure)),
	))
	mount(statusv1connect.NewMetricsServiceHandler(
		NewMetricsServer(server.prometheus),
		connect.WithInterceptors(NewPermissionInterceptor(server.access, authv1.Permission_PERMISSION_STATUS_INTERNAL_READ)),
	))
	mount(sitev1connect.NewProfileServiceHandler(
		NewProfileServer(server.profile),
	))
	mount(sitev1connect.NewProjectsServiceHandler(
		NewProjectsServer(server.projects),
	))
}

func (server *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
}

// prometheusHTTPDiscovery is gated behind the workload discovery token because it enumerates
// every device's name, model, and LAN scrape address.
func (server *Server) prometheusHTTPDiscovery(context *gin.Context) {
	if !server.config.AuthorizedDiscovery(context.Request.Header.Get("Authorization")) {
		context.Header("WWW-Authenticate", "Bearer")
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

// prometheusMetrics keeps the raw process scrape endpoint on the Prometheus workload boundary.
// Human metric reads use the OIDC-protected MetricsService instead.
func (server *Server) prometheusMetrics(context *gin.Context) {
	if !server.config.AuthorizedDiscovery(context.Request.Header.Get("Authorization")) {
		context.Header("WWW-Authenticate", "Bearer")
		writeJSON(context.Writer, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	server.metrics.ServeHTTP(context.Writer, context.Request)
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
