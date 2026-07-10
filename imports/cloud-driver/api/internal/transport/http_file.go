package transport

import (
	"mime"
	"net/http"
	"os"
	"strings"
	"time"
)

func serveFile(writer http.ResponseWriter, request *http.Request, file *os.File, name, contentType string, modified time.Time, cacheControl string, forceAttachment bool) {
	defer file.Close()
	if contentType == "" {
		sample := make([]byte, 512)
		read, _ := file.ReadAt(sample, 0)
		contentType = http.DetectContentType(sample[:read])
	}
	disposition := "attachment"
	if !forceAttachment && request.URL.Query().Get("download") != "1" && safeInlineContentType(contentType) {
		disposition = "inline"
	}
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name}))
	writer.Header().Set("Cache-Control", cacheControl)
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	http.ServeContent(writer, request, name, modified, file)
}

func safeInlineContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	mediaType = strings.ToLower(mediaType)
	return mediaType == "application/pdf" || mediaType == "text/plain" || strings.HasPrefix(mediaType, "audio/") || (strings.HasPrefix(mediaType, "image/") && mediaType != "image/svg+xml")
}

func applyCORS(writer http.ResponseWriter, request *http.Request, allowedOrigins map[string]struct{}, methods string, credentials bool) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if _, allowed := allowedOrigins[origin]; !allowed {
		http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return false
	}
	writer.Header().Set("Access-Control-Allow-Origin", origin)
	if credentials {
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	writer.Header().Set("Access-Control-Allow-Methods", methods)
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, Range, Content-Range")
	writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges")
	writer.Header().Add("Vary", "Origin")
	return true
}
