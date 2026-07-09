package gateway

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// An ingest-token holder must not be able to point Prometheus at anything it
// likes: the target has to be a bare host:port.
func TestValidateScrapeTargetRejectsAnythingButHostPort(t *testing.T) {
	for name, target := range map[string]string{
		"a URL":                  "http://10.0.0.5:9100/metrics",
		"a bare host":            "10.0.0.5",
		"an empty host":          ":9100",
		"a missing port":         "10.0.0.5:",
		"a non-numeric port":     "10.0.0.5:metrics",
		"a port out of range":    "10.0.0.5:70000",
		"a zero port":            "10.0.0.5:0",
		"a path traversal":       "10.0.0.5:9100/../admin",
		"a shell injection":      "10.0.0.5:9100;curl evil.test",
		"an unbracketed IPv6":    "fd00::1:9100",
		"a hostname with spaces": "my host:9100",
	} {
		err := validateScrapeTarget(&mev1.ScrapeTarget{
			Job:    mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER,
			Target: target,
		})
		if err == nil {
			t.Errorf("%s (%q) was accepted", name, target)
		}
	}
}

func TestValidateScrapeTargetAcceptsRealTargets(t *testing.T) {
	for _, target := range []string{"10.0.0.5:9100", "localhost:18083", "probe-host.lan:18082", "[fd00::1]:9100"} {
		if err := validateScrapeTarget(&mev1.ScrapeTarget{
			Job:    mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER,
			Target: target,
		}); err != nil {
			t.Errorf("validateScrapeTarget(%q) = %v, want nil", target, err)
		}
	}
}

func TestValidateScrapeTargetRequiresAJob(t *testing.T) {
	err := validateScrapeTarget(&mev1.ScrapeTarget{Target: "10.0.0.5:9100"})
	if err == nil {
		t.Fatal("a target with no job was accepted; it would land in no discovery list")
	}
}

// A phone that lost power must stop being reported as its last known state.
func TestFreshMobileExpiresAStalePush(t *testing.T) {
	now := time.Now().UTC()
	mobile := func(age time.Duration) *mev1.MobileState {
		return &mev1.MobileState{
			DeviceUid:  "dev_aaaa",
			UpdateTime: timestamppb.New(now.Add(-age)),
		}
	}

	if freshMobile(mobile(time.Minute), now) == nil {
		t.Error("a report from a minute ago must still be served")
	}
	if freshMobile(mobile(mobileStaleAfter-time.Second), now) == nil {
		t.Error("a report inside the freshness window must still be served")
	}
	if freshMobile(mobile(mobileStaleAfter+time.Second), now) != nil {
		t.Error("a report past the freshness window must be dropped")
	}
	if freshMobile(nil, now) != nil {
		t.Error("no report at all is not a fresh report")
	}
	if freshMobile(&mev1.MobileState{DeviceUid: "dev_aaaa"}, now) != nil {
		t.Error("a report with no update time has no freshness to check")
	}
}
