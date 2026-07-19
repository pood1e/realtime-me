package gateway

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

// enrolledDevices builds the identity lookup discovery joins against.
func enrolledDevices(devices ...*EnrolledDevice) func(string) (*EnrolledDevice, bool) {
	byUID := map[string]*EnrolledDevice{}
	for _, device := range devices {
		byUID[device.UID] = device
	}
	return func(uid string) (*EnrolledDevice, bool) {
		device, ok := byUID[uid]
		return device, ok
	}
}

func TestPrometheusHTTPDiscoveryFiltersByJobAndStampsEnrolledLabels(t *testing.T) {
	store := NewStatusStore(filepath.Join(t.TempDir(), "state.json"))
	lookup := enrolledDevices(
		&EnrolledDevice{UID: "dev_aaaa", Kind: mev1.DeviceKind_DEVICE_KIND_HOST},
		&EnrolledDevice{
			UID:         "dev_bbbb",
			DisplayName: "Studio Mac",
			Model:       "Mac16,1",
			Kind:        mev1.DeviceKind_DEVICE_KIND_HOST,
			Role:        mev1.DeviceRole_DEVICE_ROLE_DESKTOP,
		},
	)

	if err := store.SetTargets("dev_bbbb", []*mev1.ScrapeTarget{
		{Job: mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, Target: "10.0.0.5:9100"},
		{Job: mev1.ScrapeJob_SCRAPE_JOB_AGENT_EXPORTER, Target: "10.0.0.5:18082"},
	}); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}
	if err := store.SetTargets("dev_aaaa", []*mev1.ScrapeTarget{
		{Job: mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, Target: "10.0.0.4:9100"},
	}); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}

	groups := store.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, lookup)
	if len(groups) != 2 {
		t.Fatalf("got %d node-exporter groups, want 2 (the agent target must not leak in)", len(groups))
	}

	// Groups are sorted by device uid, so dev_aaaa precedes dev_bbbb.
	if got := groups[0].Targets; len(got) != 1 || got[0] != "10.0.0.4:9100" {
		t.Errorf("first group targets = %v", got)
	}

	labels := groups[1].Labels
	for key, want := range map[string]string{
		"device_id":    "dev_bbbb",
		"device_name":  "Studio Mac",
		"device_model": "Mac16,1",
	} {
		if labels[key] != want {
			t.Errorf("label %q = %q, want %q", key, labels[key], want)
		}
	}

	if len(store.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_DEVICE_EXPORTER, lookup)) != 0 {
		t.Error("a job with no targets must yield no groups")
	}
}

// A target for a uid the gateway never minted must not reach Prometheus: the
// labels come from the enrollment, so an unenrolled device has none.
func TestPrometheusHTTPDiscoverySkipsUnenrolledDevices(t *testing.T) {
	store := NewStatusStore(filepath.Join(t.TempDir(), "state.json"))
	if err := store.SetTargets("dev_ghost", []*mev1.ScrapeTarget{
		{Job: mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, Target: "10.0.0.9:9100"},
	}); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}
	if got := store.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, enrolledDevices()); len(got) != 0 {
		t.Fatalf("discovery exposed %d groups for an unenrolled device, want 0", len(got))
	}
}

// Re-running register-device.py is idempotent, and declaring a smaller set
// removes the targets left out of it.
func TestSetTargetsReplacesTheDeviceSet(t *testing.T) {
	store := NewStatusStore(filepath.Join(t.TempDir(), "state.json"))
	lookup := enrolledDevices(&EnrolledDevice{UID: "dev_aaaa"})
	full := []*mev1.ScrapeTarget{
		{Job: mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, Target: "10.0.0.5:9100"},
		{Job: mev1.ScrapeJob_SCRAPE_JOB_DEVICE_EXPORTER, Target: "10.0.0.5:18083"},
	}

	for range 3 {
		if err := store.SetTargets("dev_aaaa", full); err != nil {
			t.Fatalf("SetTargets: %v", err)
		}
	}
	if got := len(store.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, lookup)); got != 1 {
		t.Fatalf("got %d groups after re-declaring the same targets, want 1", got)
	}

	if err := store.SetTargets("dev_aaaa", full[:1]); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}
	if got := len(store.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_DEVICE_EXPORTER, lookup)); got != 0 {
		t.Fatalf("a target left out of the declared set is still discovered (%d groups)", got)
	}
}

// A decommissioned device leaves discovery instead of failing forever.
func TestSetTargetsWithAnEmptySetDeregistersTheDevice(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "state.json")
	store := NewStatusStore(stateFile)
	lookup := enrolledDevices(&EnrolledDevice{UID: "dev_aaaa"})

	if err := store.SetTargets("dev_aaaa", []*mev1.ScrapeTarget{
		{Job: mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, Target: "10.0.0.5:9100"},
	}); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}
	if err := store.SetTargets("dev_aaaa", nil); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}
	if got := len(store.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, lookup)); got != 0 {
		t.Fatalf("a deregistered device still yields %d groups", got)
	}

	restarted := NewStatusStore(stateFile)
	if err := restarted.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := len(restarted.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, lookup)); got != 0 {
		t.Fatalf("deregistration did not persist: %d groups after restart", got)
	}
}

// Scrape targets and GitHub state are durable; a restart must not lose them.
func TestStoreRoundTripsDurableStateAcrossRestart(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "state.json")

	original := NewStatusStore(stateFile)
	if err := original.SetTargets("dev_aaaa", []*mev1.ScrapeTarget{{
		Job:    mev1.ScrapeJob_SCRAPE_JOB_AGENT_EXPORTER,
		Target: "10.0.0.5:18082",
	}}); err != nil {
		t.Fatalf("SetTargets: %v", err)
	}
	if err := original.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
		sync.Configured = true
		sync.Emoji = ":rocket:"
		sync.LastSignature = "sig-1"
	}); err != nil {
		t.Fatalf("UpdateGitHub: %v", err)
	}

	restarted := NewStatusStore(stateFile)
	if err := restarted.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	lookup := enrolledDevices(&EnrolledDevice{UID: "dev_aaaa"})
	if got := len(restarted.PrometheusHTTPDiscovery(mev1.ScrapeJob_SCRAPE_JOB_AGENT_EXPORTER, lookup)); got != 1 {
		t.Fatalf("got %d agent targets after restart, want 1", got)
	}
	github := restarted.GitHubSnapshot()
	if github.GetEmoji() != ":rocket:" || github.GetLastSignature() != "sig-1" {
		t.Errorf("GitHub sync state did not survive restart: %+v", github)
	}
}

// Loading a missing state file is a first boot, not an error.
func TestStoreLoadTreatsMissingFileAsEmpty(t *testing.T) {
	store := NewStatusStore(filepath.Join(t.TempDir(), "absent.json"))
	if err := store.Load(); err != nil {
		t.Fatalf("Load on a missing state file must succeed, got %v", err)
	}
}

// GitHubSnapshot must hand out a clone: internalGithubDetail mutates what it
// receives, and must not be able to reach into the store.
func TestGitHubSnapshotIsACloneNotAnAlias(t *testing.T) {
	store := NewStatusStore(filepath.Join(t.TempDir(), "state.json"))
	if err := store.UpdateGitHub(func(sync *mev1.GithubSyncDetail) {
		sync.Configured = true
		sync.Emoji = ":rocket:"
	}); err != nil {
		t.Fatalf("UpdateGitHub: %v", err)
	}

	snapshot := store.GitHubSnapshot()
	snapshot.Configured = false
	snapshot.Emoji = ":skull:"

	if fresh := store.GitHubSnapshot(); !fresh.GetConfigured() || fresh.GetEmoji() != ":rocket:" {
		t.Fatalf("mutating a snapshot corrupted the store: %+v", fresh)
	}
}

func TestParseScrapeJob(t *testing.T) {
	got, ok := parseScrapeJob("probe-agent")
	if !ok || got != mev1.ScrapeJob_SCRAPE_JOB_PROBE {
		t.Errorf("parseScrapeJob(probe-agent) = %v, %v", got, ok)
	}
	for _, retired := range []string{"node-exporter", "vm-node-exporter", "device-exporter", "agent-exporter", "../../etc/passwd"} {
		if _, ok := parseScrapeJob(retired); ok {
			t.Errorf("retired or unknown job %q must be rejected", retired)
		}
	}
}

// Device uids are backend-owned, opaque, and unique per enrollment.
func TestEnrollMintsOpaqueUniqueUIDs(t *testing.T) {
	store := NewIdentityStore(filepath.Join(t.TempDir(), "identity.json"))
	now := time.Now()

	seen := map[string]bool{}
	for range 16 {
		device, err := store.Enroll(mev1.DeviceKind_DEVICE_KIND_HOST, mev1.DeviceRole_DEVICE_ROLE_DESKTOP, "Studio Mac", "Mac16,1", now)
		if err != nil {
			t.Fatalf("Enroll: %v", err)
		}
		if !strings.HasPrefix(device.UID, "dev_") || len(device.UID) != len("dev_")+24 {
			t.Fatalf("uid %q is not an opaque dev_<24 hex> identifier", device.UID)
		}
		if seen[device.UID] {
			t.Fatalf("Enroll minted a duplicate uid %q", device.UID)
		}
		seen[device.UID] = true
	}
}

func TestIdentityStoreRoundTripsAcrossRestart(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "identity.json")

	original := NewIdentityStore(stateFile)
	device, err := original.Enroll(mev1.DeviceKind_DEVICE_KIND_PHONE, mev1.DeviceRole_DEVICE_ROLE_DESKTOP, "Pixel", "komodo", time.Now())
	if err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	restarted := NewIdentityStore(stateFile)
	if err := restarted.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	enrolled, ok := restarted.Lookup(device.UID)
	if !ok {
		t.Fatalf("uid %q did not survive restart", device.UID)
	}
	if enrolled.DisplayName != "Pixel" || enrolled.Kind != mev1.DeviceKind_DEVICE_KIND_PHONE {
		t.Errorf("enrolled identity changed across restart: %+v", enrolled)
	}
	if _, ok := restarted.Lookup("dev_neverminted"); ok {
		t.Error("Lookup must reject a uid the gateway never minted")
	}
}
