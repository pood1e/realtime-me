package worker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"example.com/cloud-drive/api/internal/domain"
)

type failureDecision struct {
	Code    string
	Retry   bool
	RetryAt time.Time
}

type unsupportedProcessorError struct{ kind domain.ProcessingJobKind }

func (e *unsupportedProcessorError) Error() string {
	return fmt.Sprintf("unsupported processing job kind %q", e.kind)
}

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
		if value = cleanMetadata(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func cleanMetadata(value string) string {
	value = strings.ToValidUTF8(value, "")
	value = strings.ReplaceAll(value, "\x00", "")
	return strings.TrimSpace(value)
}

func leadingInteger(value string) int {
	value, _, _ = strings.Cut(value, "/")
	number, _ := strconv.Atoi(strings.TrimSpace(value))
	return number
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = cleanMetadata(value); value != "" {
			return value
		}
	}
	return ""
}

func errorCode(err error) string {
	var unsupported *unsupportedProcessorError
	if errors.As(err, &unsupported) {
		return "unsupported_job_kind"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, domain.ErrProviderReconnectRequired) {
		return "provider_reconnect_required"
	}
	if errors.Is(err, domain.ErrResourceExhausted) {
		return "storage_exhausted"
	}
	if errors.Is(err, domain.ErrRateLimited) {
		return "rate_limited"
	}
	return "processing_failed"
}

func classifyFailure(err error, attempts int, now time.Time) failureDecision {
	retry := true
	var unsupported *unsupportedProcessorError
	if errors.As(err, &unsupported) ||
		errors.Is(err, domain.ErrInvalidArgument) ||
		errors.Is(err, domain.ErrForbidden) ||
		errors.Is(err, domain.ErrConflict) ||
		errors.Is(err, domain.ErrProviderReconnectRequired) {
		retry = false
	}
	backoff := time.Duration(attempts*attempts) * 30 * time.Second
	if backoff < 30*time.Second {
		backoff = 30 * time.Second
	}
	if backoff > 30*time.Minute {
		backoff = 30 * time.Minute
	}
	if errors.Is(err, domain.ErrRateLimited) && backoff < 5*time.Minute {
		backoff = 5 * time.Minute
	}
	return failureDecision{Code: errorCode(err), Retry: retry, RetryAt: now.Add(backoff)}
}
