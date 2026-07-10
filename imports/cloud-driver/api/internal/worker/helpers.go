package worker

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

func extensionForContent(contentType string) string {
	return map[string]string{
		"application/pdf":      ".pdf",
		"application/epub+zip": ".epub",
		"audio/mpeg":           ".mp3",
		"audio/mp4":            ".m4a",
		"audio/aac":            ".aac",
		"audio/flac":           ".flac",
		"audio/ogg":            ".ogg",
		"audio/opus":           ".opus",
		"audio/wav":            ".wav",
		"audio/x-wav":          ".wav",
		"image/jpeg":           ".jpg",
		"image/png":            ".png",
		"image/webp":           ".webp",
		"image/gif":            ".gif",
		"image/svg+xml":        ".svg",
	}[contentType]
}

func contentTypeForExtension(extension string) string {
	switch strings.ToLower(extension) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func splitArtists(value string) []string {
	value = strings.ReplaceAll(value, ";", ",")
	return normalized(strings.Split(value, ","))
}

func normalized(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func leadingInteger(value string) int {
	value, _, _ = strings.Cut(value, "/")
	number, _ := strconv.Atoi(strings.TrimSpace(value))
	return number
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func errorCode(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	return "processing_failed"
}
