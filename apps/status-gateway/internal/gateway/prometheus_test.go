package gateway

import (
	"strings"
	"testing"
)

// A host can drive several agents of one kind at once. Each is its own card, so
// each needs its own opaque identity -- and one that survives the next scrape,
// because the page keys a mascot on it and remounts the clip when it changes.
func TestAgentUIDDistinguishesConcurrentAgents(t *testing.T) {
	first := agentUID("dev_1", "codex", "gpt-5-codex", 0)
	second := agentUID("dev_1", "codex", "gpt-5-codex", 1)
	if first == second {
		t.Errorf("two codex agents on one host share the uid %q", first)
	}
	if again := agentUID("dev_1", "codex", "gpt-5-codex", 0); again != first {
		t.Errorf("agentUID is not stable: %q then %q", first, again)
	}
	for _, uid := range []string{first, second} {
		if len(uid) != len("agt_")+16 || uid[:4] != "agt_" {
			t.Errorf("uid %q is not an opaque agt_ handle", uid)
		}
	}
}

// The ordinal alone must not collide two agents that differ only by model, and
// neither the host nor the kind may leak into the handle.
func TestAgentUIDSeparatesModelsAndKinds(t *testing.T) {
	distinct := map[string]struct{}{}
	for _, uid := range []string{
		agentUID("dev_1", "codex", "gpt-5-codex", 0),
		agentUID("dev_1", "codex", "gpt-5.5", 0),
		agentUID("dev_1", "claude", "gpt-5-codex", 0),
		agentUID("dev_2", "codex", "gpt-5-codex", 0),
	} {
		distinct[uid] = struct{}{}
	}
	if len(distinct) != 4 {
		t.Errorf("agentUID collided: %v", distinct)
	}
	if uid := agentUID("dev_1", "codex", "gpt-5-codex", 0); strings.Contains(uid, "dev_1") || strings.Contains(uid, "codex") {
		t.Errorf("uid %q leaks the host or the kind", uid)
	}
}

// samplesByAgent builds the join key every agent series shares, so the sub-agent
// gauge lands on the same agent as its budget.
func TestSamplesByAgentSharesTheAgentKey(t *testing.T) {
	budgets := samplesByAgent([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1"}, Value: 0.42},
	})
	if budgets["dev_1/claude"] != 0.42 {
		t.Errorf("budget = %v, want 0.42", budgets["dev_1/claude"])
	}
}

// A sub-agent may run a different model from the agent that spawned it, so the
// exporter counts them per model and the gateway expands one worker per count.
func TestAgentSubagentsExpandCountsPerModel(t *testing.T) {
	subagents := agentSubagents([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1", "model": "claude-opus-4-8"}, Value: 2},
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1", "model": "claude-haiku-4-5"}, Value: 1},
		{Metric: map[string]string{"agent_kind": "codex", "device_uid": "dev_2", "model": "gpt-5-codex"}, Value: 1},
	})
	claude := subagents["dev_1/claude"]
	if len(claude) != 3 {
		t.Fatalf("claude sub-agents = %d, want 3", len(claude))
	}
	// sorted by model, so haiku precedes opus
	for index, want := range []string{"claude-haiku-4-5", "claude-opus-4-8", "claude-opus-4-8"} {
		if got := claude[index].GetModel(); got != want {
			t.Errorf("sub-agent %d model = %q, want %q", index, got, want)
		}
	}
	if len(subagents["dev_2/codex"]) != 1 {
		t.Errorf("codex sub-agents = %d, want 1", len(subagents["dev_2/codex"]))
	}
}

// The exporter keeps the series alive at zero when nothing works, and that zero
// must never become a sub-agent nobody is running.
func TestAgentSubagentsIgnoreZeroAndNegativeCounts(t *testing.T) {
	subagents := agentSubagents([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1"}, Value: 0},
		{Metric: map[string]string{"agent_kind": "codex", "device_uid": "dev_1", "model": "gpt-5-codex"}, Value: -1},
		{Metric: map[string]string{"device_uid": "dev_1", "model": "gpt-5-codex"}, Value: 2},
	})
	if len(subagents) != 0 {
		t.Errorf("agentSubagents built %v, want nothing", subagents)
	}
}

// A sub-agent whose model the exporter could not name still counts as a worker.
func TestAgentSubagentsKeepAnUnnamedModel(t *testing.T) {
	subagents := agentSubagents([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1"}, Value: 2},
	})
	if got := len(subagents["dev_1/claude"]); got != 2 {
		t.Fatalf("sub-agents = %d, want 2", got)
	}
	if got := subagents["dev_1/claude"][0].GetModel(); got != "" {
		t.Errorf("model = %q, want empty", got)
	}
}
