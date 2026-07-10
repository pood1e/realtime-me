package transport

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

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
		logger.Info("request complete", "method", request.Method, "path", redactedPath(request.URL.Path), "status", status, "duration", time.Since(started))
	})
}

func redactedPath(path string) string {
	for _, prefix := range []string{"/v1/shares/", "/i/"} {
		if strings.HasPrefix(path, prefix) {
			remainder := strings.TrimPrefix(path, prefix)
			if separator := strings.IndexByte(remainder, '/'); separator >= 0 {
				return prefix + "[redacted]" + remainder[separator:]
			}
			return prefix + "[redacted]"
		}
	}
	return path
}
