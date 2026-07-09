package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

type GitHubStatusPublisher struct {
	config  Config
	store   *StatusStore
	client  *http.Client
	latest  atomic.Pointer[mev1.MobileState]
	trigger chan struct{}
}

type gitHubStatus struct {
	Message   string
	Emoji     string
	ExpiresAt string
	Signature string
}

func NewGitHubStatusPublisher(config Config, store *StatusStore) *GitHubStatusPublisher {
	return &GitHubStatusPublisher{
		config:  config,
		store:   store,
		client:  &http.Client{Timeout: 10 * time.Second},
		trigger: make(chan struct{}, 1),
	}
}

// Enqueue records the latest mobile status and wakes the publish worker without
// blocking the ingest request. Rapid reports coalesce to a single publish of the
// most recent status.
func (publisher *GitHubStatusPublisher) Enqueue(mobile *mev1.MobileState) {
	publisher.latest.Store(mobile)
	select {
	case publisher.trigger <- struct{}{}:
	default:
	}
}

// Run is the single publish worker. It owns all GitHub updates so overlapping
// reports never issue concurrent updates, and it stops when ctx is cancelled.
//
// A report that arrives inside the minimum-interval window cannot simply be
// dropped: the trigger channel holds one slot, so dropping it loses the report
// entirely if no further report follows. The last report before a phone goes
// idle is exactly the one most likely to land in that window, so the worker
// re-arms a timer and publishes the newest state once the window closes.
func (publisher *GitHubStatusPublisher) Run(ctx context.Context) {
	retry := time.NewTimer(0)
	if !retry.Stop() {
		<-retry.C
	}
	defer retry.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-publisher.trigger:
		case <-retry.C:
		}

		mobile := publisher.latest.Load()
		if mobile == nil {
			continue
		}

		retry.Stop()
		wait, err := publisher.publishOnce(ctx, mobile)
		if err != nil {
			slog.Error("failed to publish github status", "error", err)
		}
		if wait > 0 {
			retry.Reset(wait)
		}
	}
}

// publishOnce publishes the status unless the rate limit forbids it. It returns
// how long the caller must wait before retrying, which is zero when the status
// was published or when there is nothing left to publish.
func (publisher *GitHubStatusPublisher) publishOnce(ctx context.Context, mobile *mev1.MobileState) (time.Duration, error) {
	if publisher.config.GitHubToken == "" {
		return 0, publisher.store.UpdateGitHub(func(status *mev1.GithubSyncDetail) {
			status.Configured = false
			status.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED
		})
	}

	now := time.Now().UTC()
	status := formatGitHubStatus(mobile, now, publisher.config.GitHubStatusTTLMinutes)
	current := publisher.store.GitHubSnapshot()
	interval := publisher.config.GitHubStatusMinIntervalSeconds

	// Too soon to call GitHub again, but this status is still unpublished:
	// ask the worker to come back once the window closes.
	if wait := waitUntilStale(current.GetLastAttemptTime(), now, interval); wait > 0 {
		return wait, nil
	}
	// Nothing changed since the last successful publish, so there is nothing to
	// retry: a later report will bring a new signature and its own trigger.
	if current.GetLastSignature() == status.Signature && waitUntilStale(current.GetLastSuccessTime(), now, interval) > 0 {
		return 0, nil
	}

	if err := publisher.store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
		sync.Configured = true
		sync.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_PENDING
		sync.LastAttemptTime = timestamppb.New(now)
		sync.Message = status.Message
		sync.Emoji = status.Emoji
	}); err != nil {
		return 0, err
	}

	result := publisher.changeStatus(ctx, status)
	if result.OK {
		return 0, publisher.store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
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

	return 0, publisher.store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
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
	decodeErr := json.NewDecoder(response.Body).Decode(&payload)
	accepted := response.StatusCode >= 200 && response.StatusCode <= 299
	if accepted && decodeErr != nil {
		// A 2xx whose body is not the GraphQL answer never carried the mutation's
		// result, and reading it as one would record a publish that never happened
		// -- signature and all, which is what suppresses the retry it needs.
		return gitHubUpdateResult{OK: false, Retryable: true, Message: "Unreadable response from GitHub"}
	}
	if accepted && len(payload.Errors) == 0 {
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
	segments := make([]string, 0, 3)
	if watch.GetHeartRate().GetBeatsPerMinute() > 0 {
		segments = append(segments, fmt.Sprintf("❤️%d", watch.GetHeartRate().GetBeatsPerMinute()))
	}
	if watch.GetActivityTotals().GetSteps() > 0 {
		segments = append(segments, "👣"+compactCount(int(watch.GetActivityTotals().GetSteps())))
	}
	if state.GetBatteryPercent() > 0 {
		segments = append(segments, fmt.Sprintf("%s%d%%", batteryEmoji(state), state.GetBatteryPercent()))
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

// waitUntilStale reports how long remains before value is at least
// intervalSeconds old. Zero means the interval has already elapsed.
func waitUntilStale(value *timestamppb.Timestamp, now time.Time, intervalSeconds int) time.Duration {
	if value == nil {
		return 0
	}
	interval := time.Duration(intervalSeconds) * time.Second
	elapsed := now.Sub(value.AsTime())
	if elapsed >= interval {
		return 0
	}
	return interval - elapsed
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
