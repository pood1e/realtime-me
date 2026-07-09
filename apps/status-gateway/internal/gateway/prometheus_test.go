package gateway

import "testing"

// The model an agent runs rides on the labels of its info series, not its value,
// because a name cannot be a sample. Reading it back has to key on the same
// device/kind pair the other agent series use, or the join silently drops it.
func TestAgentModelsKeyOnDeviceAndKind(t *testing.T) {
	models := agentModels([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1", "model": "claude-opus-4-8"}, Value: 1},
		{Metric: map[string]string{"agent_kind": "codex", "device_uid": "dev_1", "model": "gpt-5-codex"}, Value: 1},
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_2", "model": "claude-haiku-4-5"}, Value: 1},
	})
	for key, want := range map[string]string{
		"dev_1/claude": "claude-opus-4-8",
		"dev_1/codex":  "gpt-5-codex",
		"dev_2/claude": "claude-haiku-4-5",
	} {
		if got := models[key]; got != want {
			t.Errorf("agentModels[%q] = %q, want %q", key, got, want)
		}
	}
	if len(models) != 3 {
		t.Errorf("agentModels returned %d entries, want 3", len(models))
	}
}

// An exporter that cannot name the model omits the label rather than exporting
// an empty one, and a sample without a kind cannot be joined to any agent.
func TestAgentModelsSkipsUnusableSamples(t *testing.T) {
	models := agentModels([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1"}, Value: 1},
		{Metric: map[string]string{"device_uid": "dev_1", "model": "gpt-5-codex"}, Value: 1},
		{Metric: map[string]string{"agent_kind": "codex", "device_uid": "dev_1", "model": ""}, Value: 1},
	})
	if len(models) != 0 {
		t.Errorf("agentModels kept %v, want nothing usable", models)
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
