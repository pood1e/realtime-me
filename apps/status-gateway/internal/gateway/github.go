package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GitHubStatusPublisher struct {
	config Config
	store  *StatusStore
	client *http.Client
}

type gitHubStatus struct {
	Message   string
	Emoji     string
	ExpiresAt string
	Signature string
}

func NewGitHubStatusPublisher(config Config, store *StatusStore) *GitHubStatusPublisher {
	return &GitHubStatusPublisher{
		config: config,
		store:  store,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (publisher *GitHubStatusPublisher) Publish(ctx context.Context, mobile StoredMobileStatus) error {
	if publisher.config.GitHubToken == "" {
		return publisher.store.UpdateGitHub(func(status *GitHubSyncStatus) {
			status.Configured = false
			status.State = GitHubSyncDisabled
		})
	}

	now := time.Now().UTC()
	status := formatGitHubStatus(mobile, now, publisher.config.GitHubStatusTTLMinutes)
	current := publisher.store.GitHubSnapshot()
	if tooSoon(current.LastAttemptAt, now, publisher.config.GitHubStatusMinIntervalSeconds) {
		return nil
	}
	if current.LastSignature == status.Signature && tooSoon(current.LastSuccessAt, now, publisher.config.GitHubStatusMinIntervalSeconds) {
		return nil
	}

	if err := publisher.store.UpdateGitHub(func(sync *GitHubSyncStatus) {
		sync.Configured = true
		sync.State = GitHubSyncPending
		sync.LastAttemptAt = now.Format(time.RFC3339)
		sync.Message = status.Message
		sync.Emoji = status.Emoji
	}); err != nil {
		return err
	}

	result := publisher.changeStatus(ctx, status)
	if result.OK {
		return publisher.store.UpdateGitHub(func(sync *GitHubSyncStatus) {
			sync.Configured = true
			sync.State = GitHubSyncOK
			sync.LastSignature = status.Signature
			sync.LastSuccessAt = now.Format(time.RFC3339)
			sync.LastErrorAt = ""
			sync.LastError = ""
			sync.Message = status.Message
			sync.Emoji = status.Emoji
		})
	}

	return publisher.store.UpdateGitHub(func(sync *GitHubSyncStatus) {
		sync.Configured = true
		sync.State = GitHubSyncError
		sync.LastErrorAt = time.Now().UTC().Format(time.RFC3339)
		sync.LastError = result.Message
		sync.Message = status.Message
		sync.Emoji = status.Emoji
	})
}

func (publisher *GitHubStatusPublisher) changeStatus(ctx context.Context, status gitHubStatus) gitHubUpdateResult {
	lastFailure := gitHubUpdateResult{OK: false, Message: "GitHub update failed"}
	for attempt := range maxAttempts {
		result := publisher.sendRequest(ctx, status)
		if result.OK {
			return result
		}
		lastFailure = result
		if !result.Retryable || attempt == maxAttempts-1 {
			return result
		}
		time.Sleep(time.Duration(attempt+1) * retryBackoff)
	}
	return lastFailure
}

func (publisher *GitHubStatusPublisher) sendRequest(ctx context.Context, status gitHubStatus) gitHubUpdateResult {
	body, err := json.Marshal(map[string]any{
		"query": changeStatusMutation,
		"variables": map[string]any{
			"input": map[string]any{
				"message":             status.Message,
				"emoji":               status.Emoji,
				"expiresAt":           status.ExpiresAt,
				"limitedAvailability": false,
			},
		},
	})
	if err != nil {
		return gitHubUpdateResult{OK: false, Message: "Invalid GitHub status payload"}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, gitHubGraphQLURL, bytes.NewReader(body))
	if err != nil {
		return gitHubUpdateResult{OK: false, Retryable: true, Message: "Could not create GitHub request"}
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+publisher.config.GitHubToken)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("User-Agent", "realtime-me-status-gateway")

	response, err := publisher.client.Do(request)
	if err != nil {
		return gitHubUpdateResult{OK: false, Retryable: true, Message: "Network error while updating GitHub status"}
	}
	defer response.Body.Close()

	var payload gitHubGraphQLResponse
	_ = json.NewDecoder(response.Body).Decode(&payload)
	if response.StatusCode >= 200 && response.StatusCode <= 299 && len(payload.Errors) == 0 {
		return gitHubUpdateResult{OK: true}
	}
	return gitHubUpdateResult{
		OK:        false,
		Retryable: response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= http.StatusInternalServerError,
		Message:   gitHubRejectionMessage(payload, response.StatusCode),
	}
}

func formatGitHubStatus(mobile StoredMobileStatus, now time.Time, ttlMinutes int) gitHubStatus {
	watch := mobile.Watch
	offWrist := watch != nil && watch.WristState == WristOffWrist
	segments := make([]string, 0, 4)
	if !offWrist && watch != nil && watch.HeartRate != nil && *watch.HeartRate > 0 {
		segments = append(segments, fmt.Sprintf("❤️%d", *watch.HeartRate))
	}
	if watch != nil && watch.Steps != nil && *watch.Steps > 0 {
		segments = append(segments, "👣"+compactCount(*watch.Steps))
	}
	if watch != nil && watch.BatteryPercent != nil && *watch.BatteryPercent > 0 {
		segments = append(segments, fmt.Sprintf("%s%d%%", batteryEmoji(*watch), *watch.BatteryPercent))
	}
	if offWrist {
		segments = append(segments, "💤")
	}

	message := "⌚synced"
	if len(segments) > 0 {
		message = strings.Join(segments, " · ")
	}
	if len([]rune(message)) > maxGitHubMessageLength {
		message = string([]rune(message)[:maxGitHubMessageLength])
	}
	emoji := statusEmoji(watch)
	return gitHubStatus{
		Message:   message,
		Emoji:     emoji,
		ExpiresAt: now.Add(time.Duration(ttlMinutes) * time.Minute).Format(time.RFC3339),
		Signature: message + "|" + emoji,
	}
}

func batteryEmoji(watch WatchStatus) string {
	if watch.ChargeState == ChargeCharging {
		return "🔌"
	}
	if watch.BatteryPercent != nil && *watch.BatteryPercent > 0 && *watch.BatteryPercent < 15 {
		return "🪫"
	}
	return "🔋"
}

func statusEmoji(watch *WatchStatus) string {
	if watch == nil {
		return "⌚"
	}
	if watch.WristState == WristOffWrist {
		return "💤"
	}
	if watch.ChargeState == ChargeCharging {
		return "🔌"
	}
	if watch.BatteryPercent != nil && *watch.BatteryPercent > 0 && *watch.BatteryPercent < 15 {
		return "🪫"
	}
	return "⌚"
}

func compactCount(value int) string {
	if value < 1000 {
		return fmt.Sprintf("%d", value)
	}
	formatted := fmt.Sprintf("%.1f", float64(value)/1000)
	return strings.TrimSuffix(formatted, ".0") + "k"
}

func gitHubRejectionMessage(payload gitHubGraphQLResponse, statusCode int) string {
	if len(payload.Errors) > 0 && payload.Errors[0].Message != "" {
		return "GitHub rejected the status update: " + payload.Errors[0].Message
	}
	return fmt.Sprintf("GitHub rejected the status update (HTTP %d)", statusCode)
}

func tooSoon(value string, now time.Time, intervalSeconds int) bool {
	if value == "" {
		return false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return err == nil && now.Sub(parsed) < time.Duration(intervalSeconds)*time.Second
}

type gitHubGraphQLResponse struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type gitHubUpdateResult struct {
	OK        bool
	Retryable bool
	Message   string
}

const (
	gitHubGraphQLURL       = "https://api.github.com/graphql"
	maxAttempts            = 2
	maxGitHubMessageLength = 80
)

const retryBackoff = 500 * time.Millisecond

const changeStatusMutation = `
mutation ChangeUserStatus($input: ChangeUserStatusInput!) {
  changeUserStatus(input: $input) {
    status {
      message
      emoji
      expiresAt
    }
  }
}
`
