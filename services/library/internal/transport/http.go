package transport

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	authv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/auth/v1/authv1connect"
	booksv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/books/v1/booksv1connect"
	contentv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/content/v1/contentv1connect"
	drivev1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/drive/v1/drivev1connect"
	imagesv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/images/v1/imagesv1connect"
	musicv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/music/v1/musicv1connect"
	systemv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/system/v1/systemv1connect"
	wallpapersv1connect "github.com/pood1e/realtime-me/services/library/gen/cloud/wallpapers/v1/wallpapersv1connect"
	"github.com/pood1e/realtime-me/services/library/internal/app"
	"github.com/pood1e/realtime-me/services/library/internal/auth"
	"github.com/pood1e/realtime-me/services/library/internal/config"
	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

const maxChunkBodyBytes = 16 << 20

const providerCallbackPrefix = "/v1/music/providers/"

// NewHTTPHandler creates strict host-separated private and public routes.
func NewHTTPHandler(cfg config.Config, suite *app.Suite, sessions *auth.Manager, logger *slog.Logger) http.Handler {
	privateMux := http.NewServeMux()
	publicMux := http.NewServeMux()
	mountPrivateConnect(privateMux, cfg, suite, sessions)
	mountPublicConnect(publicMux, suite)
	router := &httpRouter{config: cfg, suite: suite, sessions: sessions, connectErrors: connect.NewErrorWriter(), privateMux: privateMux, publicMux: publicMux}
	return requestLogger(logger, router)
}

func mountPrivateConnect(mux *http.ServeMux, cfg config.Config, suite *app.Suite, sessions *auth.Manager) {
	path, handler := authv1connect.NewSessionServiceHandler(&authServer{sessions: sessions, validator: cfg})
	registerConnect(mux, path, handler)
	path, handler = systemv1connect.NewHealthServiceHandler(&systemServer{service: suite.System})
	registerConnect(mux, path, handler)
	path, handler = contentv1connect.NewContentUploadServiceHandler(&contentServer{service: suite.Content})
	registerConnect(mux, path, handler)
	path, handler = drivev1connect.NewDriveServiceHandler(&driveServer{service: suite.Drive})
	registerConnect(mux, path, handler)
	path, handler = drivev1connect.NewShareServiceHandler(&shareServer{service: suite.Drive})
	registerConnect(mux, path, handler)
	path, handler = booksv1connect.NewBookServiceHandler(&bookServer{service: suite.Books})
	registerConnect(mux, path, handler)
	path, handler = musicv1connect.NewMusicLibraryServiceHandler(&musicLibraryServer{service: suite.Music.Library})
	registerConnect(mux, path, handler)
	path, handler = musicv1connect.NewMusicProviderServiceHandler(&musicProviderServer{service: suite.Music.Providers})
	registerConnect(mux, path, handler)
	path, handler = musicv1connect.NewMusicPlaylistServiceHandler(&musicPlaylistServer{service: suite.Music.Playlists})
	registerConnect(mux, path, handler)
	path, handler = imagesv1connect.NewImageServiceHandler(&imageServer{service: suite.Images})
	registerConnect(mux, path, handler)
	path, handler = wallpapersv1connect.NewWallpaperAdminServiceHandler(&wallpaperAdminServer{service: suite.Wallpapers})
	registerConnect(mux, path, handler)
}

func mountPublicConnect(mux *http.ServeMux, suite *app.Suite) {
	sharePath, shareHandler := drivev1connect.NewShareServiceHandler(&shareServer{service: suite.Drive})
	allowedShareProcedures := map[string]struct{}{
		drivev1connect.ShareServiceResolveShareProcedure:      {},
		drivev1connect.ShareServiceListSharedItemsProcedure:   {},
		drivev1connect.ShareServiceGetSharedDownloadProcedure: {},
	}
	mux.Handle(sharePath, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, allowed := allowedShareProcedures[request.URL.Path]; !allowed {
			http.NotFound(writer, request)
			return
		}
		shareHandler.ServeHTTP(writer, request)
	}))
	path, handler := wallpapersv1connect.NewWallpaperPublicServiceHandler(&wallpaperPublicServer{service: suite.Wallpapers})
	registerConnect(mux, path, handler)
}

func registerConnect(mux *http.ServeMux, path string, handler http.Handler) {
	mux.Handle(path, handler)
}

type httpRouter struct {
	config        config.Config
	suite         *app.Suite
	sessions      *auth.Manager
	connectErrors *connect.ErrorWriter
	privateMux    http.Handler
	publicMux     http.Handler
}

func (router *httpRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/healthz" {
		router.health(writer, request)
		return
	}
	switch requestHost(request.Host) {
	case router.config.PrivateAPIHost:
		router.servePrivate(writer, request)
	case router.config.PublicAPIHost:
		router.servePublic(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

func (router *httpRouter) health(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		methodNotAllowed(writer, http.MethodGet, http.MethodHead)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()
	healthy, _, _, _, err := router.suite.System.Check(ctx)
	if err != nil || !healthy {
		http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(http.StatusNoContent)
}

func (router *httpRouter) servePrivate(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Cache-Control", "no-store")
	if strings.HasPrefix(request.URL.Path, providerCallbackPrefix) {
		segments := pathSegments(request.URL.Path, providerCallbackPrefix)
		if len(segments) != 2 || segments[1] != "callback" {
			http.NotFound(writer, request)
			return
		}
		provider := domain.MusicProvider(segments[0])
		if !domain.ValidMusicProviderID(provider) || provider == domain.MusicProviderLocal {
			http.NotFound(writer, request)
			return
		}
		router.serveProviderCallback(writer, request, provider)
		return
	}
	if !applyCORS(writer, request, router.config.PrivateAppOrigins, "GET, POST, PUT, DELETE, OPTIONS", true) {
		return
	}
	if request.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	if privatePublicProcedure(request.URL.Path) {
		router.privateMux.ServeHTTP(writer, request)
		return
	}
	session, err := router.sessions.Authenticate(request)
	if err != nil {
		router.writeUnauthenticated(writer, request)
		return
	}
	request = request.WithContext(auth.ContextWithSession(request.Context(), session))
	if strings.HasPrefix(request.URL.Path, "/v1/") {
		router.servePrivateRaw(writer, request)
		return
	}
	router.privateMux.ServeHTTP(writer, request)
}

func (router *httpRouter) serveProviderCallback(writer http.ResponseWriter, request *http.Request, provider domain.MusicProvider) {
	if request.Method != http.MethodGet {
		methodNotAllowed(writer, http.MethodGet)
		return
	}
	if router.config.MusicAppOrigin == "" {
		http.NotFound(writer, request)
		return
	}
	status := "failed"
	if request.URL.Query().Get("error") == "" {
		err := router.suite.Music.Providers.CompleteRedirectConnection(
			request.Context(), provider, request.URL.Query().Get("state"), request.URL.Query().Get("code"),
		)
		if err == nil {
			status = "connected"
		}
	}
	query := url.Values{"provider": {string(provider)}, "connection": {status}}
	target := router.config.MusicAppOrigin + "/?" + query.Encode()
	http.Redirect(writer, request, target, http.StatusSeeOther)
}

func (router *httpRouter) servePublic(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/i/") {
		if publicAssetPreflight(writer, request) {
			return
		}
		router.servePublicImage(writer, request)
		return
	}
	if strings.HasPrefix(request.URL.Path, "/v1/wallpapers/") {
		if publicAssetPreflight(writer, request) {
			return
		}
		router.serveWallpaperFile(writer, request)
		return
	}
	if !applyCORS(writer, request, router.config.PublicAppOrigins, "GET, POST, OPTIONS", false) {
		return
	}
	if request.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasPrefix(request.URL.Path, "/v1/shares/") {
		router.servePublicShareRaw(writer, request)
		return
	}
	if strings.HasPrefix(request.URL.Path, "/cloud.wallpapers.v1.WallpaperPublicService/") {
		writer.Header().Set("Cache-Control", "public, max-age=300, s-maxage=3600")
	} else {
		writer.Header().Set("Cache-Control", "no-store")
	}
	router.publicMux.ServeHTTP(writer, request)
}

func (router *httpRouter) writeUnauthenticated(writer http.ResponseWriter, request *http.Request) {
	err := connect.NewError(connect.CodeUnauthenticated, auth.ErrUnauthenticated)
	if router.connectErrors.IsSupported(request) {
		_ = router.connectErrors.Write(writer, request, err)
		return
	}
	http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func privatePublicProcedure(path string) bool {
	return path == authv1connect.SessionServiceLoginProcedure
}
