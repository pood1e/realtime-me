package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	cloud_drivev1 "example.com/cloud-drive/api/gen/cloud/drive/v1"
	cloud_drivev1connect "example.com/cloud-drive/api/gen/cloud/drive/v1/drivev1connect"
	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/auth"
	"example.com/cloud-drive/api/internal/config"
	"example.com/cloud-drive/api/internal/domain"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const maxChunkBodyBytes = 16 << 20

// NewHTTPHandler creates strict host-separated private and public API routes.
func NewHTTPHandler(cfg config.Config, service *app.Service, sessions *auth.Manager, logger *slog.Logger) http.Handler {
	connectServer := NewConnectServer(service, sessions)
	privateMux := http.NewServeMux()
	publicMux := http.NewServeMux()
	mountPrivateConnect(privateMux, connectServer)
	mountPublicConnect(publicMux, connectServer)

	handler := &httpRouter{
		config:        cfg,
		service:       service,
		sessions:      sessions,
		logger:        logger,
		connectErrors: connect.NewErrorWriter(),
		privateMux:    privateMux,
		publicMux:     publicMux,
	}
	return requestLogger(logger, handler)
}

func mountPrivateConnect(mux *http.ServeMux, server *ConnectServer) {
	path, handler := cloud_drivev1connect.NewHealthServiceHandler(server)
	mux.Handle(path, handler)
	path, handler = cloud_drivev1connect.NewSessionServiceHandler(server)
	mux.Handle(path, http.MaxBytesHandler(handler, 4<<10))
	path, handler = cloud_drivev1connect.NewDriveServiceHandler(server)
	mux.Handle(path, handler)
	path, handler = cloud_drivev1connect.NewUploadServiceHandler(server)
	mux.Handle(path, handler)
	path, handler = cloud_drivev1connect.NewShareServiceHandler(server)
	mux.Handle(path, handler)
}

func mountPublicConnect(mux *http.ServeMux, server *ConnectServer) {
	path, handler := cloud_drivev1connect.NewShareServiceHandler(server)
	publicAllowed := map[string]struct{}{
		cloud_drivev1connect.ShareServiceResolveShareProcedure:      {},
		cloud_drivev1connect.ShareServiceListSharedItemsProcedure:   {},
		cloud_drivev1connect.ShareServiceGetSharedDownloadProcedure: {},
	}
	mux.Handle(path, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, allowed := publicAllowed[request.URL.Path]; !allowed {
			http.NotFound(writer, request)
			return
		}
		handler.ServeHTTP(writer, request)
	}))
}

type httpRouter struct {
	config        config.Config
	service       *app.Service
	sessions      *auth.Manager
	logger        *slog.Logger
	connectErrors *connect.ErrorWriter
	privateMux    http.Handler
	publicMux     http.Handler
}

func (router *httpRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/healthz" {
		router.health(writer, request)
		return
	}
	host := requestHost(request.Host)
	switch host {
	case router.config.PrivateAPIHost:
		router.servePrivate(writer, request)
	case router.config.ShareAPIHost:
		router.servePublic(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

func (router *httpRouter) health(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		writer.Header().Set("Allow", "GET, HEAD")
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	context, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()
	if err := router.service.Ping(context); err != nil {
		http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(http.StatusNoContent)
}

func (router *httpRouter) servePrivate(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Cache-Control", "no-store")
	if !applyCORS(writer, request, router.config.PrivateAppOrigin, "GET, POST, PUT, OPTIONS") {
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

func (router *httpRouter) writeUnauthenticated(writer http.ResponseWriter, request *http.Request) {
	err := connect.NewError(connect.CodeUnauthenticated, auth.ErrUnauthenticated)
	if router.connectErrors.IsSupported(request) {
		_ = router.connectErrors.Write(writer, request, err)
		return
	}
	http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func privatePublicProcedure(path string) bool {
	return path == cloud_drivev1connect.HealthServiceCheckProcedure ||
		path == cloud_drivev1connect.SessionServiceLoginProcedure
}

func (router *httpRouter) servePublic(writer http.ResponseWriter, request *http.Request) {
	if !applyCORS(writer, request, router.config.ShareAppOrigin, "GET, OPTIONS") {
		return
	}
	if request.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasPrefix(request.URL.Path, "/v1/") {
		router.servePublicRaw(writer, request)
		return
	}
	router.publicMux.ServeHTTP(writer, request)
}

func (router *httpRouter) servePrivateRaw(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/v1/uploads" {
		if request.Method != http.MethodPost {
			methodNotAllowed(writer, http.MethodPost)
			return
		}
		var message cloud_drivev1.StartUploadRequest
		if err := decodeProtoJSON(writer, request, &message); err != nil {
			return
		}
		upload, err := router.service.StartUpload(request.Context(), stringPointer(message.GetParentUid()), message.GetFileName(), message.GetContentType(), message.GetTotalSizeBytes())
		if err != nil {
			writeDomainError(writer, err, false)
			return
		}
		writeProtoJSON(writer, http.StatusOK, &cloud_drivev1.StartUploadResponse{Upload: uploadProto(upload), ChunkUrl: "/v1/uploads/" + upload.UID + "/chunks"})
		return
	}
	if strings.HasPrefix(request.URL.Path, "/v1/uploads/") {
		router.servePrivateUpload(writer, request)
		return
	}
	if strings.HasPrefix(request.URL.Path, "/v1/items/") {
		router.servePrivateDownload(writer, request)
		return
	}
	http.NotFound(writer, request)
}

func (router *httpRouter) servePrivateUpload(writer http.ResponseWriter, request *http.Request) {
	segments := pathSegments(request.URL.Path, "/v1/uploads/")
	if len(segments) == 1 && request.Method == http.MethodGet {
		upload, err := router.service.GetUpload(request.Context(), segments[0])
		if err != nil {
			writeDomainError(writer, err, false)
			return
		}
		writeProtoJSON(writer, http.StatusOK, &cloud_drivev1.GetUploadResponse{Upload: uploadProto(upload)})
		return
	}
	if len(segments) == 2 && segments[1] == "chunks" && request.Method == http.MethodPut {
		router.writeRawChunk(writer, request, segments[0])
		return
	}
	if len(segments) == 2 && segments[1] == "complete" && request.Method == http.MethodPost {
		item, err := router.service.CompleteUpload(request.Context(), segments[0])
		if err != nil {
			writeDomainError(writer, err, false)
			return
		}
		writeProtoJSON(writer, http.StatusOK, &cloud_drivev1.CompleteUploadResponse{Item: itemProto(item)})
		return
	}
	methodNotAllowed(writer, http.MethodGet, http.MethodPut, http.MethodPost)
}

func (router *httpRouter) writeRawChunk(writer http.ResponseWriter, request *http.Request, uploadUID string) {
	startOffset, endOffset, totalSizeBytes, err := parseContentRange(request.Header.Get("Content-Range"))
	if err != nil {
		writeDomainError(writer, fmt.Errorf("%w: invalid Content-Range", domain.ErrInvalidArgument), false)
		return
	}
	if request.ContentLength >= 0 && request.ContentLength != endOffset-startOffset+1 {
		writeDomainError(writer, fmt.Errorf("%w: Content-Length does not match range", domain.ErrInvalidArgument), false)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxChunkBodyBytes+1)
	data, err := io.ReadAll(request.Body)
	if err != nil || int64(len(data)) != endOffset-startOffset+1 {
		writeDomainError(writer, fmt.Errorf("%w: chunk body does not match range", domain.ErrInvalidArgument), false)
		return
	}
	upload, err := router.service.WriteUploadChunk(request.Context(), uploadUID, startOffset, totalSizeBytes, data)
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	writeProtoJSON(writer, http.StatusOK, &cloud_drivev1.WriteUploadChunkResponse{Upload: uploadProto(upload)})
}

func (router *httpRouter) servePrivateDownload(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		methodNotAllowed(writer, http.MethodGet, http.MethodHead)
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/items/")
	if len(segments) != 2 || segments[1] != "content" {
		http.NotFound(writer, request)
		return
	}
	item, err := router.service.GetItem(request.Context(), segments[0])
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	router.serveFile(writer, request, item)
}

func (router *httpRouter) servePublicRaw(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		methodNotAllowed(writer, http.MethodGet, http.MethodHead)
		return
	}
	segments := pathSegments(request.URL.Path, "/v1/shares/")
	if len(segments) == 1 {
		share, item, err := router.service.ResolveShare(request.Context(), segments[0])
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		writeProtoJSON(writer, http.StatusOK, &cloud_drivev1.ResolveShareResponse{ShareLink: shareProto(share), Target: itemProto(item)})
		return
	}
	if len(segments) == 2 && segments[1] == "items" {
		parentUID := request.URL.Query().Get("parentUid")
		pageSize, err := pageSizeFromRequest(request)
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		page, err := router.service.ListSharedItems(request.Context(), segments[0], stringPointer(parentUID), pageSize, request.URL.Query().Get("pageToken"))
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		items, nextToken := pageProto(page)
		writeProtoJSON(writer, http.StatusOK, &cloud_drivev1.ListSharedItemsResponse{Items: items, NextPageToken: nextToken})
		return
	}
	if len(segments) == 4 && segments[1] == "items" && segments[3] == "content" {
		_, item, err := router.service.SharedItem(request.Context(), segments[0], segments[2])
		if err != nil {
			writeDomainError(writer, err, true)
			return
		}
		router.serveFile(writer, request, item)
		return
	}
	http.NotFound(writer, request)
}

func (router *httpRouter) serveFile(writer http.ResponseWriter, request *http.Request, item domain.Item) {
	file, err := router.service.OpenFile(request.Context(), item)
	if err != nil {
		writeDomainError(writer, err, false)
		return
	}
	defer file.Close()
	contentType := item.ContentType
	if contentType == "" {
		sample := make([]byte, 512)
		read, _ := file.ReadAt(sample, 0)
		contentType = http.DetectContentType(sample[:read])
	}
	writer.Header().Set("Content-Type", contentType)
	disposition := "attachment"
	if request.URL.Query().Get("download") != "1" && safeInlineContentType(contentType) {
		disposition = "inline"
	}
	writer.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": item.Name}))
	writer.Header().Set("Cache-Control", "private, no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	http.ServeContent(writer, request, item.Name, item.UpdateTime, file)
}

func safeInlineContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	mediaType = strings.ToLower(mediaType)
	if mediaType == "application/pdf" || mediaType == "text/plain" {
		return true
	}
	if strings.HasPrefix(mediaType, "image/") && mediaType != "image/svg+xml" {
		return true
	}
	return false
}

func applyCORS(writer http.ResponseWriter, request *http.Request, allowedOrigin, methods string) bool {
	origin := request.Header.Get("Origin")
	if origin != "" && origin != allowedOrigin {
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return false
	}
	if origin == "" {
		return true
	}
	writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	writer.Header().Set("Access-Control-Allow-Credentials", "true")
	writer.Header().Set("Access-Control-Allow-Methods", methods)
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, Range, Content-Range")
	writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
	writer.Header().Add("Vary", "Origin")
	return true
}

func decodeProtoJSON(writer http.ResponseWriter, request *http.Request, message proto.Message) error {
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return err
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: false}).Unmarshal(body, message); err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return err
	}
	return nil
}

func writeProtoJSON(writer http.ResponseWriter, status int, message proto.Message) {
	body, err := (protojson.MarshalOptions{UseProtoNames: false}).Marshal(message)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_, _ = writer.Write(body)
}

func writeDomainError(writer http.ResponseWriter, err error, publicShare bool) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, domain.ErrResourceExhausted):
		status = http.StatusInsufficientStorage
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusPreconditionFailed
	case errors.Is(err, domain.ErrUnavailable):
		status = http.StatusServiceUnavailable
	}
	if publicShare && (status == http.StatusForbidden || status == http.StatusPreconditionFailed) {
		status = http.StatusNotFound
	}
	http.Error(writer, http.StatusText(status), status)
}

func parseContentRange(value string) (int64, int64, int64, error) {
	parts := strings.Fields(value)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, 0, 0, errors.New("invalid content range")
	}
	rangePart, totalPart, found := strings.Cut(parts[1], "/")
	if !found || totalPart == "*" {
		return 0, 0, 0, errors.New("invalid content range")
	}
	startText, endText, found := strings.Cut(rangePart, "-")
	if !found {
		return 0, 0, 0, errors.New("invalid content range")
	}
	startOffset, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || startOffset < 0 {
		return 0, 0, 0, errors.New("invalid content range")
	}
	endOffset, err := strconv.ParseInt(endText, 10, 64)
	if err != nil || endOffset < startOffset || endOffset == int64(^uint64(0)>>1) {
		return 0, 0, 0, errors.New("invalid content range")
	}
	totalSize, err := strconv.ParseInt(totalPart, 10, 64)
	if err != nil || totalSize < 0 || endOffset >= totalSize {
		return 0, 0, 0, errors.New("invalid content range")
	}
	return startOffset, endOffset, totalSize, nil
}

func pathSegments(path, prefix string) []string {
	if !strings.HasPrefix(path, prefix) {
		return nil
	}
	remainder := strings.TrimPrefix(path, prefix)
	if remainder == "" || strings.HasPrefix(remainder, "/") || strings.HasSuffix(remainder, "/") {
		return nil
	}
	rawSegments := strings.Split(remainder, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, rawSegment := range rawSegments {
		segment, err := url.PathUnescape(rawSegment)
		if err != nil || segment == "" || strings.Contains(segment, "/") {
			return nil
		}
		segments = append(segments, segment)
	}
	return segments
}

func pageSizeFromRequest(request *http.Request) (int, error) {
	value := request.URL.Query().Get("pageSize")
	if value == "" {
		return 0, nil
	}
	pageSize, err := strconv.Atoi(value)
	if err != nil || pageSize < 1 {
		return 0, fmt.Errorf("%w: invalid page size", domain.ErrInvalidArgument)
	}
	return pageSize, nil
}

func methodNotAllowed(writer http.ResponseWriter, methods ...string) {
	writer.Header().Set("Allow", strings.Join(methods, ", "))
	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func requestHost(host string) string {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return host
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *responseRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *responseRecorder) Write(body []byte) (int, error) {
	if recorder.status == 0 {
		recorder.status = http.StatusOK
	}
	return recorder.ResponseWriter.Write(body)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &responseRecorder{ResponseWriter: writer}
		next.ServeHTTP(recorder, request)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Info("request complete",
			"method", request.Method,
			"path", redactedPath(request.URL.Path),
			"status", status,
			"duration", time.Since(started),
		)
	})
}

func redactedPath(path string) string {
	const prefix = "/v1/shares/"
	if !strings.HasPrefix(path, prefix) {
		return path
	}
	remainder := strings.TrimPrefix(path, prefix)
	if remainder == "" {
		return path
	}
	if separator := strings.IndexByte(remainder, '/'); separator >= 0 {
		return prefix + "[redacted]" + remainder[separator:]
	}
	return prefix + "[redacted]"
}
