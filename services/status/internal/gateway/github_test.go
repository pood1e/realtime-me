package gateway

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/services/status/internal/genproto/realtime/me/v1"
)

// cloneGithub once copied field by field and silently dropped last_signature and
// last_attempt_time, which disabled both the rate limit and the redundant-update
// skip: every report reached GitHub. Copy the whole message or not at all.
func TestCloneGithubPreservesEveryField(t *testing.T) {
	now := timestamppb.New(time.Unix(1_700_000_000, 0).UTC())
	original := &mev1.GithubSyncDetail{
		Configured:      true,
		State:           mev1.GithubSyncState_GITHUB_SYNC_STATE_OK,
		Emoji:           ":rocket:",
		Message:         "❤️72",
		LastSuccessTime: now,
		LastErrorTime:   now,
		LastError:       "boom",
		LastAttemptTime: now,
		LastSignature:   "❤️72|:rocket:",
	}

	clone := cloneGithub(original)

	if !proto.Equal(original, clone) {
		t.Fatalf("cloneGithub dropped fields:\n original = %v\n clone    = %v", original, clone)
	}
	if clone.GetLastSignature() == "" {
		t.Error("last_signature must survive the clone; the redundant-update skip depends on it")
	}
	if clone.GetLastAttemptTime() == nil {
		t.Error("last_attempt_time must survive the clone; the rate limit depends on it")
	}
	if cloneGithub(nil) != nil {
		t.Error("cloneGithub(nil) must stay nil")
	}
}

func TestWaitUntilStale(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()

	if got := waitUntilStale(nil, now, 10); got != 0 {
		t.Errorf("a never-attempted publish must not wait, got %v", got)
	}
	if got := waitUntilStale(timestamppb.New(now.Add(-10*time.Second)), now, 10); got != 0 {
		t.Errorf("an exactly-elapsed interval must not wait, got %v", got)
	}
	if got := waitUntilStale(timestamppb.New(now.Add(-30*time.Second)), now, 10); got != 0 {
		t.Errorf("a long-elapsed interval must not wait, got %v", got)
	}
	if got := waitUntilStale(timestamppb.New(now.Add(-4*time.Second)), now, 10); got != 6*time.Second {
		t.Errorf("waitUntilStale = %v, want 6s remaining", got)
	}
}

// A report inside the rate-limit window must ask to be retried, not be dropped.
func TestPublishOnceDefersRatherThanDropping(t *testing.T) {
	store := newTestStore(t)
	if err := store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
		sync.LastAttemptTime = timestamppb.New(time.Now().UTC())
	}); err != nil {
		t.Fatalf("UpdateGitHub: %v", err)
	}

	publisher := NewGitHubStatusPublisher(Config{
		GitHubToken:                    "token",
		GitHubStatusMinIntervalSeconds: 60,
		GitHubStatusTTLMinutes:         20,
	}, store)
	// Any network call would be a bug: the rate limit must short-circuit first.
	publisher.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("publishOnce called GitHub while rate limited")
		return nil, io.ErrUnexpectedEOF
	})}

	wait, err := publisher.publishOnce(context.Background(), mobileWithHeartRate(72))
	if err != nil {
		t.Fatalf("publishOnce: %v", err)
	}
	if wait <= 0 || wait > 60*time.Second {
		t.Fatalf("publishOnce returned wait = %v, want a positive retry delay", wait)
	}
}

// The last report before a phone goes idle usually lands inside the rate-limit
// window. It must still reach GitHub once the window closes.
func TestRunPublishesTheTrailingReport(t *testing.T) {
	var calls atomic.Int32
	store := newTestStore(t)
	publisher := NewGitHubStatusPublisher(Config{
		GitHubToken:                    "token",
		GitHubStatusMinIntervalSeconds: 1,
		GitHubStatusTTLMinutes:         20,
	}, store)
	publisher.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"data":{}}`)),
			Header:     http.Header{},
		}, nil
	})}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go publisher.Run(ctx)

	publisher.Enqueue(mobileWithHeartRate(72))
	waitForCalls(t, &calls, 1, 2*time.Second, "the first report must publish immediately")

	// Lands inside the 1s window, and is the final report: nothing follows it.
	publisher.Enqueue(mobileWithHeartRate(120))
	time.Sleep(250 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("rate limit did not hold: got %d GitHub calls, want 1", got)
	}

	waitForCalls(t, &calls, 2, 3*time.Second, "the trailing report must publish once the window closes")

	if signature := store.GitHubSnapshot().GetLastSignature(); !strings.Contains(signature, "120") {
		t.Errorf("published signature = %q, want the trailing heart rate 120", signature)
	}
}

// With no token configured, the publisher records DISABLED and never dials out.
func TestPublishOnceWithoutTokenIsDisabled(t *testing.T) {
	store := newTestStore(t)
	publisher := NewGitHubStatusPublisher(Config{GitHubStatusMinIntervalSeconds: 10}, store)
	publisher.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("publishOnce dialed GitHub with no token configured")
		return nil, io.ErrUnexpectedEOF
	})}

	wait, err := publisher.publishOnce(context.Background(), mobileWithHeartRate(72))
	if err != nil || wait != 0 {
		t.Fatalf("publishOnce = (%v, %v), want (0, nil)", wait, err)
	}
	if state := store.GitHubSnapshot().GetState(); state != mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED {
		t.Errorf("state = %v, want DISABLED", state)
	}
}

func newTestStore(t *testing.T) *StatusStore {
	t.Helper()
	return NewStatusStore(filepath.Join(t.TempDir(), "state.json"))
}

func mobileWithHeartRate(beats int32) *mev1.MobileState {
	return &mev1.MobileState{
		Watch: &mev1.WatchSnapshot{
			HeartRate: &mev1.HeartRateSample{BeatsPerMinute: beats},
		},
	}
}

func waitForCalls(t *testing.T, calls *atomic.Int32, want int32, within time.Duration, reason string) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if calls.Load() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s: got %d GitHub calls after %v, want %d", reason, calls.Load(), within, want)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }
