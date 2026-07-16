package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// PrometheusHTTPDiscoveryGroup is one Prometheus HTTP service-discovery group.
// It is Prometheus's own JSON discovery shape, not part of the wire contract.
type PrometheusHTTPDiscoveryGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// StatusStore holds the gateway's live status plus the durable scrape targets
// and GitHub sync state. The phone's pushed status is kept in memory only: it is
// re-pushed within seconds, so it is not persisted and never shown stale after a
// restart. Hosts and agents are never pushed — Prometheus scrapes their
// exporters — so the store holds no copy of them.
type StatusStore struct {
	stateFile string
	mutex     sync.Mutex

	mobile  *mev1.MobileState
	targets map[string][]*mev1.ScrapeTarget
	github  *mev1.GithubSyncDetail
}

func NewStatusStore(stateFile string) *StatusStore {
	return &StatusStore{
		stateFile: stateFile,
		targets:   map[string][]*mev1.ScrapeTarget{},
		github: &mev1.GithubSyncDetail{
			Configured: false,
			State:      mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED,
		},
	}
}

// persistedState is the durable on-disk shape. Proto values are stored as their
// canonical protojson encoding so the schema stays the single source of truth.
// Scrape targets are keyed by the device uid that owns them.
type persistedState struct {
	Targets map[string][]json.RawMessage `json:"targets,omitempty"`
	GitHub  json.RawMessage              `json:"github,omitempty"`
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

	var snapshot persistedState
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	for deviceUID, rawTargets := range snapshot.Targets {
		targets := make([]*mev1.ScrapeTarget, 0, len(rawTargets))
		for _, raw := range rawTargets {
			target := &mev1.ScrapeTarget{}
			if err := protojson.Unmarshal(raw, target); err != nil {
				return err
			}
			targets = append(targets, target)
		}
		store.targets[deviceUID] = targets
	}
	if len(snapshot.GitHub) > 0 {
		github := &mev1.GithubSyncDetail{}
		if err := protojson.Unmarshal(snapshot.GitHub, github); err != nil {
			return err
		}
		store.github = github
	}
	return nil
}

// UpdateMobile replaces the phone snapshot while retaining optional capabilities
// collected by another enrolled Android device. This lets a dedicated Nintendo
// collector and the owner's primary phone report independently without erasing
// each other's watch or Switch state.
func (store *StatusStore) UpdateMobile(mobile *mev1.MobileState) *mev1.MobileState {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	next := cloneMobile(mobile)
	if current := cloneMobile(store.mobile); current != nil {
		if next.Phone == nil {
			next.Phone = current.Phone
		}
		if next.Watch == nil {
			next.Watch = current.Watch
		}
		if next.SwitchPresence == nil {
			next.SwitchPresence = current.SwitchPresence
		}
	}
	store.mobile = next
	return cloneMobile(next)
}

// SetTargets replaces the device's entire target set. An empty set deregisters
// the device, so a decommissioned host leaves service discovery rather than
// lingering as a permanently-down target.
func (store *StatusStore) SetTargets(deviceUID string, targets []*mev1.ScrapeTarget) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if len(targets) == 0 {
		delete(store.targets, deviceUID)
	} else {
		store.targets[deviceUID] = targets
	}
	return store.saveLocked()
}

// PrometheusHTTPDiscovery returns the Prometheus HTTP service-discovery groups
// for a job. Every label describing a device is read from its enrollment, so a
// target can never claim an identity the gateway did not mint.
func (store *StatusStore) PrometheusHTTPDiscovery(job mev1.ScrapeJob, lookup func(string) (*EnrolledDevice, bool)) []PrometheusHTTPDiscoveryGroup {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	groups := make([]PrometheusHTTPDiscoveryGroup, 0, len(store.targets))
	for _, deviceUID := range sortedKeys(store.targets) {
		device, ok := lookup(deviceUID)
		if !ok {
			continue
		}
		for _, target := range store.targets[deviceUID] {
			if target.GetJob() != job {
				continue
			}
			groups = append(groups, PrometheusHTTPDiscoveryGroup{
				Targets: []string{target.GetTarget()},
				Labels:  prometheusTargetLabels(device),
			})
		}
	}
	return groups
}

func sortedKeys(targets map[string][]*mev1.ScrapeTarget) []string {
	keys := make([]string, 0, len(targets))
	for key := range targets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (store *StatusStore) UpdateGitHub(mutator func(*mev1.GithubSyncDetail)) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	mutator(store.github)
	return store.saveLocked()
}

func (store *StatusStore) GitHubSnapshot() *mev1.GithubSyncDetail {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return cloneGithub(store.github)
}

// StatusSnapshot is a consistent read of the live pushed status.
type StatusSnapshot struct {
	Mobile *mev1.MobileState
	GitHub *mev1.GithubSyncDetail
}

func (store *StatusStore) Snapshot() StatusSnapshot {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return StatusSnapshot{
		Mobile: cloneMobile(store.mobile),
		GitHub: cloneGithub(store.github),
	}
}

// saveLocked persists the durable state atomically (temp file + fsync + rename)
// so a crash mid-write cannot leave a truncated or corrupt state file.
func (store *StatusStore) saveLocked() error {
	snapshot := persistedState{}
	if len(store.targets) > 0 {
		snapshot.Targets = make(map[string][]json.RawMessage, len(store.targets))
	}
	for deviceUID, targets := range store.targets {
		for _, target := range targets {
			raw, err := protojson.Marshal(target)
			if err != nil {
				return err
			}
			snapshot.Targets[deviceUID] = append(snapshot.Targets[deviceUID], raw)
		}
	}
	if store.github != nil {
		raw, err := protojson.Marshal(store.github)
		if err != nil {
			return err
		}
		snapshot.GitHub = raw
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	directory := filepath.Dir(store.stateFile)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(directory, ".status-state-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempName, 0o600); err != nil {
		return err
	}
	return os.Rename(tempName, store.stateFile)
}

func prometheusTargetLabels(device *EnrolledDevice) map[string]string {
	labels := map[string]string{
		"instance":  device.UID,
		"device_id": device.UID,
	}
	if device.DisplayName != "" {
		labels["device_name"] = device.DisplayName
	}
	if device.Model != "" {
		labels["device_model"] = device.Model
	}
	if kind := deviceKindString(device.Kind); kind != "" {
		labels["device_kind"] = kind
	}
	if role := deviceRoleString(device.Role); role != "" {
		labels["device_role"] = role
	}
	return labels
}

// cloneGithub copies every field. A hand-written copy silently drops fields
// added to the proto later, so this defers to the schema instead.
func cloneGithub(github *mev1.GithubSyncDetail) *mev1.GithubSyncDetail {
	if github == nil {
		return nil
	}
	return proto.Clone(github).(*mev1.GithubSyncDetail)
}

func cloneMobile(mobile *mev1.MobileState) *mev1.MobileState {
	if mobile == nil {
		return nil
	}
	return proto.Clone(mobile).(*mev1.MobileState)
}
