package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
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

func (publisher *GitHubStatusPublisher) Publish(ctx context.Context, mobile *mev1.MobileState) error {
	if publisher.config.GitHubToken == "" {
		return publisher.store.UpdateGitHub(func(status *mev1.GithubSyncDetail) {
			status.Configured = false
			status.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED
		})
	}

	now := time.Now().UTC()
	status := formatGitHubStatus(mobile, now, publisher.config.GitHubStatusTTLMinutes)
	current := publisher.store.GitHubSnapshot()
	if tooSoon(current.GetLastAttemptTime(), now, publisher.config.GitHubStatusMinIntervalSeconds) {
		return nil
	}
	if current.GetLastSignature() == status.Signature && tooSoon(current.GetLastSuccessTime(), now, publisher.config.GitHubStatusMinIntervalSeconds) {
		return nil
	}

	if err := publisher.store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
		sync.Configured = true
		sync.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_PENDING
		sync.LastAttemptTime = timestamppb.New(now)
		sync.Message = status.Message
		sync.Emoji = status.Emoji
	}); err != nil {
		return err
	}

	result := publisher.changeStatus(ctx, status)
	if result.OK {
		return publisher.store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
			sync.Configured = true
			sync.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_OK
			sync.LastSignature = status.Signature
			sync.LastSuccessTime = timestamppb.New(now)
			sync.LastErrorTime = nil
			sync.LastError = ""
			sync.Message = status.Message
			sync.Emoji = status.Emoji
		})
	}

	return publisher.store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
		sync.Configured = true
		sync.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_ERROR
		sync.LastErrorTime = timestamppb.New(time.Now().UTC())
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

func formatGitHubStatus(mobile *mev1.MobileState, now time.Time, ttlMinutes int) gitHubStatus {
	watch := mobile.GetWatch()
	state := watch.GetWatchState()
	offWrist := state.GetWristState() == mev1.WristState_WRIST_STATE_OFF_WRIST
	segments := make([]string, 0, 4)
	if !offWrist && watch.GetHeartRate().GetBeatsPerMinute() > 0 {
		segments = append(segments, fmt.Sprintf("❤️%d", watch.GetHeartRate().GetBeatsPerMinute()))
	}
	if watch.GetActivityTotals().GetSteps() > 0 {
		segments = append(segments, "👣"+compactCount(int(watch.GetActivityTotals().GetSteps())))
	}
	if state.GetBatteryPercent() > 0 {
		segments = append(segments, fmt.Sprintf("%s%d%%", batteryEmoji(state), state.GetBatteryPercent()))
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

func batteryEmoji(state *mev1.WatchState) string {
	if state.GetChargeState() == mev1.ChargeState_CHARGE_STATE_CHARGING {
		return "🔌"
	}
	if battery := state.GetBatteryPercent(); battery > 0 && battery < 15 {
		return "🪫"
	}
	return "🔋"
}

func statusEmoji(watch *mev1.WatchSnapshot) string {
	state := watch.GetWatchState()
	if watch == nil || state == nil {
		return "⌚"
	}
	if state.GetWristState() == mev1.WristState_WRIST_STATE_OFF_WRIST {
		return "💤"
	}
	if state.GetChargeState() == mev1.ChargeState_CHARGE_STATE_CHARGING {
		return "🔌"
	}
	if battery := state.GetBatteryPercent(); battery > 0 && battery < 15 {
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

func tooSoon(value *timestamppb.Timestamp, now time.Time, intervalSeconds int) bool {
	if value == nil {
		return false
	}
	return now.Sub(value.AsTime()) < time.Duration(intervalSeconds)*time.Second
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
