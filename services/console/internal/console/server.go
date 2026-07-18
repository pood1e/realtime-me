// Package console implements the authenticated web BFF and static application host.
package console

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/pood1e/realtime-me/gen/go/realtime/me/auth/v1"
	authv1connect "github.com/pood1e/realtime-me/gen/go/realtime/me/auth/v1/authv1connect"
	"github.com/pood1e/realtime-me/libs/go/authn"
	"github.com/pood1e/realtime-me/services/console/internal/config"
	"github.com/pood1e/realtime-me/services/console/internal/session"
)

const (
	productionCookieName   = "__Host-realtime-me-console"
	developmentCookieName  = "realtime-me-console"
	productionLoginCookie  = "__Host-realtime-me-console-login"
	developmentLoginCookie = "realtime-me-console-login"
	maximumSessionAge      = 24 * time.Hour
)

type contextKey struct{}

// Server owns OIDC login, server-side sessions, upstream proxies, and static assets.
type Server struct {
	config        config.Config
	logger        *slog.Logger
	sessions      *session.Store
	oauth         oauth2.Config
	oidcVerifier  *oidc.IDTokenVerifier
	oidcContext   context.Context
	connectErrors *connect.ErrorWriter
	sessionPath   string
	sessionAPI    http.Handler
	proxies       map[string]*httputil.ReverseProxy
}

// NewServer discovers the OIDC provider and constructs the Console boundary.
func NewServer(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Server, error) {
	identityClient := boundedHTTPClient(10 * time.Second)
	oidcContext := oidc.ClientContext(context.Background(), identityClient)
	discoveryContext, cancel := context.WithTimeout(oidc.ClientContext(ctx, identityClient), 10*time.Second)
	defer cancel()
	provider, err := oidc.NewProvider(discoveryContext, cfg.OIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	server := &Server{
		config:       cfg,
		logger:       logger,
		sessions:     session.NewStore(),
		oidcContext:  oidcContext,
		oidcVerifier: provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID}),
		oauth: oauth2.Config{
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.CallbackURL(),
			Scopes:       cfg.OIDCScopes,
		},
		connectErrors: connect.NewErrorWriter(),
		proxies: map[string]*httputil.ReverseProxy{
			"/api/status/":  newProxy("/api/status", cfg.StatusUpstream, logger),
			"/api/library/": newProxy("/api/library", cfg.LibraryUpstream, logger),
			"/api/manager/": newProxy("/api/manager", cfg.ManagerUpstream, logger),
		},
	}
	server.sessionPath, server.sessionAPI = authv1connect.NewSessionServiceHandler(&sessionServer{server: server})
	return server, nil
}

// ServeHTTP routes authentication, authenticated proxies, and the Console SPA.
func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	applySecurityHeaders(writer)
	path := request.URL.Path
	switch {
	case path == "/healthz":
		server.health(writer, request)
	case path == "/auth/login":
		server.login(writer, request)
	case path == "/auth/callback":
		server.callback(writer, request)
	case strings.HasPrefix(path, server.sessionPath):
		server.serveAuthenticated(writer, request, server.sessionAPI)
	default:
		for prefix, proxy := range server.proxies {
			if strings.HasPrefix(path, prefix) {
				server.serveAuthenticated(writer, request, proxy)
				return
			}
		}
		server.static(writer, request)
	}
}

func (server *Server) health(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		writer.Header().Set("Allow", "GET, HEAD")
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(http.StatusNoContent)
}

func (server *Server) login(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", "GET")
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	pending, err := server.sessions.Begin(returnPath(request.URL.Query().Get("return_to")))
	if err != nil {
		server.logger.Error("start OIDC login", "error", err)
		http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	loginAge := max(1, int(time.Until(pending.ExpireTime).Seconds()))
	http.SetCookie(writer, server.loginCookie(pending.State, loginAge))
	target := server.oauth.AuthCodeURL(
		pending.State,
		oauth2.AccessTypeOffline,
		oidc.Nonce(pending.Nonce),
		oauth2.S256ChallengeOption(pending.Verifier),
	)
	http.Redirect(writer, request, target, http.StatusSeeOther)
}

func (server *Server) callback(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", "GET")
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	state := request.URL.Query().Get("state")
	loginCookie, cookieError := request.Cookie(server.loginCookieName())
	if cookieError != nil || !session.NonceMatches(loginCookie.Value, state) {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	http.SetCookie(writer, server.loginCookie("", -1))
	pending, err := server.sessions.Consume(state)
	if err != nil || request.URL.Query().Get("error") != "" {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	exchangeContext, cancel := context.WithTimeout(oidc.ClientContext(request.Context(), boundedHTTPClient(10*time.Second)), 10*time.Second)
	defer cancel()
	token, err := server.oauth.Exchange(exchangeContext, request.URL.Query().Get("code"), oauth2.VerifierOption(pending.Verifier))
	if err != nil || token.AccessToken == "" || token.Expiry.IsZero() {
		server.logger.Warn("OIDC code exchange failed", "error", err)
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	idToken, err := server.oidcVerifier.Verify(exchangeContext, rawIDToken)
	if err != nil {
		server.logger.Warn("OIDC ID token rejected", "error", err)
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if idToken.AccessTokenHash != "" {
		if err := idToken.VerifyAccessToken(token.AccessToken); err != nil {
			server.logger.Warn("OIDC access token hash rejected", "error", err)
			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}
	var claims struct {
		Nonce             string   `json:"nonce"`
		Name              string   `json:"name"`
		PreferredUsername string   `json:"preferred_username"`
		Permissions       []string `json:"permissions"`
	}
	subject := strings.TrimSpace(idToken.Subject)
	if err := idToken.Claims(&claims); err != nil || subject == "" || !session.NonceMatches(pending.Nonce, claims.Nonce) {
		http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(claims.PreferredUsername)
	}
	if displayName == "" {
		displayName = subject
	}
	sessionID, err := server.sessions.Create(
		session.Identity{
			Subject:     subject,
			DisplayName: displayName,
			Permissions: authn.ParsePermissionNames(claims.Permissions),
		},
		server.oauth.TokenSource(server.oidcContext, token),
	)
	if err != nil {
		server.logger.Error("create Console session", "error", err)
		http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	http.SetCookie(writer, server.sessionCookie(sessionID, int(maximumSessionAge.Seconds())))
	http.Redirect(writer, request, pending.ReturnPath, http.StatusSeeOther)
}

func (server *Server) serveAuthenticated(writer http.ResponseWriter, request *http.Request, next http.Handler) {
	if stateChanging(request) && request.Header.Get("Origin") != server.config.PublicOrigin.String() {
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	cookie, err := request.Cookie(server.cookieName())
	if err != nil {
		server.writeUnauthenticated(writer, request)
		return
	}
	resolved, err := server.sessions.Resolve(request.Context(), cookie.Value)
	if err != nil {
		server.writeUnauthenticated(writer, request)
		return
	}
	request = request.WithContext(context.WithValue(request.Context(), contextKey{}, resolved))
	next.ServeHTTP(writer, request)
}

func (server *Server) writeUnauthenticated(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("WWW-Authenticate", "Bearer")
	err := connect.NewError(connect.CodeUnauthenticated, session.ErrUnauthenticated)
	if server.connectErrors.IsSupported(request) {
		_ = server.connectErrors.Write(writer, request, err)
		return
	}
	http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (server *Server) static(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if server.config.WebDirectory == "" {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	cleanPath := strings.TrimPrefix(filepath.Clean("/"+request.URL.Path), "/")
	filePath := filepath.Join(server.config.WebDirectory, cleanPath)
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		if strings.HasPrefix(request.URL.Path, "/assets/") {
			writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeFile(writer, request, filePath)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	http.ServeFile(writer, request, filepath.Join(server.config.WebDirectory, "index.html"))
}

func (server *Server) cookieName() string {
	if server.config.SecureCookies() {
		return productionCookieName
	}
	return developmentCookieName
}

func (server *Server) loginCookieName() string {
	if server.config.SecureCookies() {
		return productionLoginCookie
	}
	return developmentLoginCookie
}

func (server *Server) sessionCookie(value string, maxAge int) *http.Cookie {
	return server.cookie(server.cookieName(), value, maxAge)
}

func (server *Server) loginCookie(value string, maxAge int) *http.Cookie {
	return server.cookie(server.loginCookieName(), value, maxAge)
}

func (server *Server) cookie(name, value string, maxAge int) *http.Cookie {
	expires := time.Now().UTC().Add(time.Duration(maxAge) * time.Second)
	if maxAge < 0 {
		expires = time.Unix(1, 0).UTC()
	}
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expires,
		Secure:   server.config.SecureCookies(),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

type sessionServer struct {
	server *Server
}

func (service *sessionServer) GetSession(ctx context.Context, _ *connect.Request[authv1.GetSessionRequest]) (*connect.Response[authv1.GetSessionResponse], error) {
	resolved, ok := ctx.Value(contextKey{}).(session.Resolved)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, session.ErrUnauthenticated)
	}
	return connect.NewResponse(&authv1.GetSessionResponse{Session: &authv1.Session{
		Subject:     resolved.Identity.Subject,
		DisplayName: resolved.Identity.DisplayName,
		ExpireTime:  timestamppb.New(resolved.ExpireTime),
		Permissions: resolved.Identity.Permissions,
	}}), nil
}

func (service *sessionServer) Logout(ctx context.Context, _ *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	resolved, ok := ctx.Value(contextKey{}).(session.Resolved)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, session.ErrUnauthenticated)
	}
	service.server.sessions.Delete(resolved.ID)
	response := connect.NewResponse(&authv1.LogoutResponse{})
	response.Header().Add("Set-Cookie", service.server.sessionCookie("", -1).String())
	return response, nil
}

func newProxy(prefix string, target *url.URL, logger *slog.Logger) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(request *http.Request) {
		director(request)
		request.URL.Path = strings.TrimPrefix(request.URL.Path, prefix)
		if request.URL.Path == "" {
			request.URL.Path = "/"
		}
		request.Host = target.Host
		request.Header.Del("Cookie")
		if resolved, ok := request.Context().Value(contextKey{}).(session.Resolved); ok {
			request.Header.Set("Authorization", "Bearer "+resolved.AccessToken)
		}
	}
	proxy.Transport = boundedTransport()
	proxy.FlushInterval = -1
	proxy.ModifyResponse = func(response *http.Response) error {
		response.Header.Del("Set-Cookie")
		response.Header.Set("Cache-Control", "no-store")
		return nil
	}
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		logger.Error("Console upstream failed", "path", request.URL.Path, "error", err)
		http.Error(writer, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
	}
	return proxy
}

func boundedHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Transport: boundedTransport(), Timeout: timeout}
}

func boundedTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}
}

func stateChanging(request *http.Request) bool {
	if strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
		return true
	}
	return request.Method != http.MethodGet && request.Method != http.MethodHead && request.Method != http.MethodOptions
}

func returnPath(candidate string) string {
	parsed, err := url.Parse(strings.TrimSpace(candidate))
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "/"
	}
	return parsed.RequestURI()
}

func applySecurityHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Content-Security-Policy", strings.Join([]string{
		"default-src 'self'",
		"base-uri 'none'",
		"frame-ancestors 'none'",
		"form-action 'self'",
		"object-src 'none'",
		"script-src 'self' https://sdk.scdn.co",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: blob: https://*.gtimg.cn https://*.qq.com https://*.music.126.net https://*.music.163.com https://*.scdn.co",
		"media-src 'self' blob: https://*.qqmusic.qq.com https://*.music.126.net https://*.music.163.com https://*.scdn.co",
		"connect-src 'self' https://api.spotify.com https://*.spotify.com https://*.scdn.co wss://*.spotify.com",
	}, "; "))
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("X-Frame-Options", "DENY")
}
