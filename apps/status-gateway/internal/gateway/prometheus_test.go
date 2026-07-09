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
	subagents := samplesByAgent([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1"}, Value: 3},
	})
	budgets := samplesByAgent([]prometheusSample{
		{Metric: map[string]string{"agent_kind": "claude", "device_uid": "dev_1"}, Value: 0.42},
	})
	if subagents["dev_1/claude"] != 3 {
		t.Errorf("sub-agent count = %v, want 3", subagents["dev_1/claude"])
	}
	if budgets["dev_1/claude"] != 0.42 {
		t.Errorf("budget = %v, want 0.42", budgets["dev_1/claude"])
	}
}
