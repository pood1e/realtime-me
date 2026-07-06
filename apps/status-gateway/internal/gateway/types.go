package gateway

type ChargeState string

type WristState string

const (
	ChargeUnknown     ChargeState = "unknown"
	ChargeCharging    ChargeState = "charging"
	ChargeNotCharging ChargeState = "not_charging"

	WristUnknown  WristState = "unknown"
	WristOnWrist  WristState = "on_wrist"
	WristOffWrist WristState = "off_wrist"
)

type MobileIngest struct {
	DeviceID  string       `json:"device_id"`
	UpdatedAt string       `json:"updated_at,omitempty"`
	Phone     *PhoneStatus `json:"phone,omitempty"`
	Watch     *WatchStatus `json:"watch,omitempty"`
}

type PhoneStatus struct {
	BatteryPercent *int        `json:"battery_percent,omitempty"`
	ChargeState    ChargeState `json:"charge_state,omitempty"`
	Network        string      `json:"network,omitempty"`
}

type WatchStatus struct {
	HeartRate      *int        `json:"heart_rate,omitempty"`
	Steps          *int        `json:"steps,omitempty"`
	BatteryPercent *int        `json:"battery_percent,omitempty"`
	ChargeState    ChargeState `json:"charge_state,omitempty"`
	WristState     WristState  `json:"wrist_state,omitempty"`
}

type AgentIngest struct {
	AgentID                string `json:"agent_id"`
	UpdatedAt              string `json:"updated_at,omitempty"`
	State                  string `json:"state,omitempty"`
	Task                   string `json:"task,omitempty"`
	BudgetRemainingPercent *int   `json:"budget_remaining_percent,omitempty"`
}

type StoredMobileStatus struct {
	MobileIngest
	ReceivedAt string `json:"received_at"`
}

type StoredAgentStatus struct {
	AgentIngest
	ReceivedAt string `json:"received_at"`
}

type GitHubSyncState string

const (
	GitHubSyncDisabled GitHubSyncState = "disabled"
	GitHubSyncPending  GitHubSyncState = "pending"
	GitHubSyncOK       GitHubSyncState = "ok"
	GitHubSyncError    GitHubSyncState = "error"
)

type GitHubSyncStatus struct {
	Configured    bool            `json:"configured"`
	State         GitHubSyncState `json:"state"`
	LastSignature string          `json:"last_signature,omitempty"`
	LastAttemptAt string          `json:"last_attempt_at,omitempty"`
	LastSuccessAt string          `json:"last_success_at,omitempty"`
	LastErrorAt   string          `json:"last_error_at,omitempty"`
	LastError     string          `json:"last_error,omitempty"`
	Message       string          `json:"message,omitempty"`
	Emoji         string          `json:"emoji,omitempty"`
}

type GatewayStateSnapshot struct {
	Mobile *StoredMobileStatus `json:"mobile"`
	Agents []StoredAgentStatus `json:"agents"`
	GitHub GitHubSyncStatus    `json:"github"`
}

type ServerSummary struct {
	Online        bool     `json:"online"`
	CPUPercent    *float64 `json:"cpu_percent"`
	MemoryPercent *float64 `json:"memory_percent"`
	DiskPercent   *float64 `json:"disk_percent"`
	UpdatedAt     string   `json:"updated_at"`
}

type PublicGitHubStatus struct {
	Enabled   bool            `json:"enabled"`
	State     GitHubSyncState `json:"state"`
	UpdatedAt string          `json:"updated_at,omitempty"`
	Emoji     string          `json:"emoji,omitempty"`
	Message   string          `json:"message,omitempty"`
}

type PublicStatus struct {
	Server    ServerSummary       `json:"server"`
	Mobile    *StoredMobileStatus `json:"mobile"`
	Agents    []StoredAgentStatus `json:"agents"`
	GitHub    PublicGitHubStatus  `json:"github"`
	UpdatedAt string              `json:"updated_at"`
}
