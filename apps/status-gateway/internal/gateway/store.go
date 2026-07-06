package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type StatusStore struct {
	stateFile string
	mutex     sync.Mutex
	mobile    *StoredMobileStatus
	agents    map[string]StoredAgentStatus
	devices   map[string]StoredDeviceStatus
	github    GitHubSyncStatus
}

func NewStatusStore(stateFile string) *StatusStore {
	return &StatusStore{
		stateFile: stateFile,
		agents:    map[string]StoredAgentStatus{},
		devices:   map[string]StoredDeviceStatus{},
		github: GitHubSyncStatus{
			Configured: false,
			State:      GitHubSyncDisabled,
		},
	}
}

func (store *StatusStore) Load() error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	data, err := os.ReadFile(store.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var snapshot GatewayStateSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	store.mobile = snapshot.Mobile
	for _, agent := range snapshot.Agents {
		store.agents[agent.AgentID] = agent
	}
	for _, device := range snapshot.Devices {
		store.devices[device.DeviceID] = device
	}
	if snapshot.GitHub.State != "" {
		store.github = snapshot.GitHub
	}
	return nil
}

func (store *StatusStore) UpdateMobile(input MobileIngest, receivedAt time.Time) (StoredMobileStatus, error) {
	if input.UpdatedAt == "" {
		input.UpdatedAt = receivedAt.UTC().Format(time.RFC3339)
	}
	status := StoredMobileStatus{
		MobileIngest: input,
		ReceivedAt:   receivedAt.UTC().Format(time.RFC3339),
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.mobile = &status
	return status, store.saveLocked()
}

func (store *StatusStore) UpdateDevice(input DeviceStatus, receivedAt time.Time) (StoredDeviceStatus, error) {
	if input.UpdatedAt == "" {
		input.UpdatedAt = receivedAt.UTC().Format(time.RFC3339)
	}
	status := StoredDeviceStatus{
		DeviceStatus: input,
		ReceivedAt:   receivedAt.UTC().Format(time.RFC3339),
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.devices[status.DeviceID] = status
	return status, store.saveLocked()
}

func (store *StatusStore) UpdateAgent(input AgentIngest, receivedAt time.Time) (StoredAgentStatus, error) {
	if input.UpdatedAt == "" {
		input.UpdatedAt = receivedAt.UTC().Format(time.RFC3339)
	}
	status := StoredAgentStatus{
		AgentIngest: input,
		ReceivedAt:  receivedAt.UTC().Format(time.RFC3339),
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.agents[status.AgentID] = status
	return status, store.saveLocked()
}

func (store *StatusStore) UpdateGitHub(mutator func(*GitHubSyncStatus)) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	mutator(&store.github)
	return store.saveLocked()
}

func (store *StatusStore) GitHubSnapshot() GitHubSyncStatus {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.github
}

func (store *StatusStore) Snapshot() GatewayStateSnapshot {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.snapshotLocked()
}

func (store *StatusStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(store.stateFile), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(store.snapshotLocked())
	if err != nil {
		return err
	}
	return os.WriteFile(store.stateFile, data, 0o600)
}

func (store *StatusStore) snapshotLocked() GatewayStateSnapshot {
	agents := make([]StoredAgentStatus, 0, len(store.agents))
	for _, agent := range store.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(left, right int) bool {
		return agents[left].AgentID < agents[right].AgentID
	})
	devices := make([]StoredDeviceStatus, 0, len(store.devices))
	for _, device := range store.devices {
		devices = append(devices, device)
	}
	sort.Slice(devices, func(left, right int) bool {
		return devices[left].DeviceID < devices[right].DeviceID
	})
	return GatewayStateSnapshot{
		Mobile:  store.mobile,
		Agents:  agents,
		Devices: devices,
		GitHub:  store.github,
	}
}
