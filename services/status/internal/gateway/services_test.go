package gateway

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

func testScrapeTargetPolicy(t *testing.T) ScrapeTargetPolicy {
	t.Helper()
	policy, err := NewScrapeTargetPolicy([]string{"10.40.0.0/16", "fd42::/64"}, 18082)
	if err != nil {
		t.Fatalf("NewScrapeTargetPolicy() = %v", err)
	}
	return policy
}

// An ingest-token holder cannot redirect Prometheus outside the explicitly
// routed probe networks, to another port, or through rebinding-prone DNS.
func TestScrapeTargetPolicyRejectsUnsafeTargets(t *testing.T) {
	policy := testScrapeTargetPolicy(t)
	for name, target := range map[string]string{
		"a URL":               "http://10.40.0.5:18082/metrics",
		"a bare host":         "10.40.0.5",
		"an empty host":       ":18082",
		"a hostname":          "probe-host.lan:18082",
		"a DNS localhost":     "localhost:18082",
		"an external address": "203.0.113.8:18082",
		"a loopback address":  "127.0.0.1:18082",
		"a wrong port":        "10.40.0.5:9100",
		"a missing port":      "10.40.0.5:",
		"a path traversal":    "10.40.0.5:18082/../admin",
		"an unbracketed IPv6": "fd42::1:18082",
		"a zoned IPv6":        "[fd42::1%eth0]:18082",
	} {
		err := policy.Validate(&mev1.ScrapeTarget{
			Job:    mev1.ScrapeJob_SCRAPE_JOB_PROBE,
			Target: target,
		})
		if err == nil {
			t.Errorf("%s (%q) was accepted", name, target)
		}
	}
}

func TestScrapeTargetPolicyAcceptsAllowedAddresses(t *testing.T) {
	policy := testScrapeTargetPolicy(t)
	for _, target := range []string{"10.40.0.5:18082", "[fd42::1]:18082"} {
		if err := policy.Validate(&mev1.ScrapeTarget{
			Job:    mev1.ScrapeJob_SCRAPE_JOB_PROBE,
			Target: target,
		}); err != nil {
			t.Errorf("ScrapeTargetPolicy.Validate(%q) = %v, want nil", target, err)
		}
	}
}

func TestScrapeTargetPolicyRequiresProbeJob(t *testing.T) {
	err := testScrapeTargetPolicy(t).Validate(&mev1.ScrapeTarget{Target: "10.40.0.5:18082"})
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
