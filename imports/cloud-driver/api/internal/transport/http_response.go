package transport

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

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

func writeDomainError(writer http.ResponseWriter, err error, hidePrivate bool) {
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
	case errors.Is(err, domain.ErrRateLimited):
		status = http.StatusTooManyRequests
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusPreconditionFailed
	case errors.Is(err, domain.ErrUnavailable):
		status = http.StatusServiceUnavailable
	}
	if hidePrivate && (status == http.StatusForbidden || status == http.StatusPreconditionFailed) {
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
	start, startErr := strconv.ParseInt(startText, 10, 64)
	end, endErr := strconv.ParseInt(endText, 10, 64)
	total, totalErr := strconv.ParseInt(totalPart, 10, 64)
	if startErr != nil || endErr != nil || totalErr != nil || start < 0 || end < start || total < 0 || end >= total {
		return 0, 0, 0, errors.New("invalid content range")
	}
	return start, end, total, nil
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
	for _, raw := range rawSegments {
		segment, err := url.PathUnescape(raw)
		if err != nil || segment == "" || strings.Contains(segment, "/") {
			return nil
		}
		segments = append(segments, segment)
	}
	return segments
}

func getOrHead(writer http.ResponseWriter, request *http.Request) bool {
	if request.Method == http.MethodGet || request.Method == http.MethodHead {
		return true
	}
	methodNotAllowed(writer, http.MethodGet, http.MethodHead)
	return false
}

func methodNotAllowed(writer http.ResponseWriter, methods ...string) {
	writer.Header().Set("Allow", strings.Join(methods, ", "))
	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func requestHost(host string) string {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		return parsed
	}
	return host
}
