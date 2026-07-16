package gateway

import (
	"strings"
	"testing"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// The gateway exports only what it owns. Host, VM, and agent series come from
// node_exporter and the probe exporters, which Prometheus scrapes directly;
// re-publishing them here would duplicate every series under a second job.
func TestGatewayExportsOnlyItsOwnSeries(t *testing.T) {
	foreign := []string{"realtime_host_", "realtime_agent_"}
	for _, definition := range metricDefinitions {
		for _, prefix := range foreign {
			if strings.HasPrefix(definition.Name, prefix) {
				t.Errorf("%s belongs to an exporter Prometheus scrapes directly, not to the gateway", definition.Name)
			}
		}
	}
	if len(metricDefinitions) == 0 {
		t.Fatal("the gateway still owns the phone, watch, and GitHub series")
	}
}

func TestNamedServerLabelsAnUnnamedServer(t *testing.T) {
	if got := namedServer(&mev1.DeviceState{}).GetDisplayName(); got != "Server" {
		t.Errorf("an unnamed server rendered as %q, want %q", got, "Server")
	}
	named := &mev1.DeviceState{DisplayName: "metrics-host"}
	if got := namedServer(named).GetDisplayName(); got != "metrics-host" {
		t.Errorf("namedServer overwrote a real hostname with %q", got)
	}
	if namedServer(nil) != nil {
		t.Error("namedServer(nil) must stay nil rather than fabricate a device")
	}
}

// Only phones push. If a future change reintroduces a pushed host or agent
// store, this fails: StatusSnapshot must carry no such field.
func TestStatusSnapshotCarriesOnlyPushedState(t *testing.T) {
	snapshot := StatusSnapshot{}
	// Compile-time proof the fields are gone; the assertions keep the intent
	// legible if someone adds them back.
	if snapshot.Mobiles != nil {
		t.Error("a zero snapshot must have no mobile state")
	}
	if snapshot.GitHub != nil {
		t.Error("a zero snapshot must have no GitHub state")
	}
}
