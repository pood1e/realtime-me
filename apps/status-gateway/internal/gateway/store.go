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
// and GitHub sync state. Pushed status (mobile/host/agent) is kept in memory
// only: it is re-pushed within seconds, so it is not persisted and never shown
// stale after a restart. Scrape targets and GitHub sync state are durable.
type StatusStore struct {
	stateFile string
	mutex     sync.Mutex

	mobile  *mev1.MobileState
	hosts   map[string]*mev1.DeviceState
	agents  map[string]*mev1.Agent
	targets map[string]*mev1.ScrapeTarget
	github  *mev1.GithubSyncDetail
}

func NewStatusStore(stateFile string) *StatusStore {
	return &StatusStore{
		stateFile: stateFile,
		hosts:     map[string]*mev1.DeviceState{},
		agents:    map[string]*mev1.Agent{},
		targets:   map[string]*mev1.ScrapeTarget{},
		github: &mev1.GithubSyncDetail{
			Configured: false,
			State:      mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED,
		},
	}
}

// persistedState is the durable on-disk shape. Proto values are stored as their
// canonical protojson encoding so the schema stays the single source of truth.
type persistedState struct {
	Targets []json.RawMessage `json:"targets,omitempty"`
	GitHub  json.RawMessage   `json:"github,omitempty"`
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
	for _, raw := range snapshot.Targets {
		target := &mev1.ScrapeTarget{}
		if err := protojson.Unmarshal(raw, target); err != nil {
			return err
		}
		store.targets[scrapeTargetKey(target)] = target
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

func (store *StatusStore) UpdateMobile(mobile *mev1.MobileState) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.mobile = mobile
}

func (store *StatusStore) UpdateHost(device *mev1.DeviceState) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.hosts[device.GetDeviceUid()] = device
}

func (store *StatusStore) UpdateAgent(agent *mev1.Agent) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.agents[agentStoreKey(agent)] = agent
}

func (store *StatusStore) RegisterTargets(targets []*mev1.ScrapeTarget) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	for _, target := range targets {
		store.targets[scrapeTargetKey(target)] = target
	}
	return store.saveLocked()
}

// PrometheusHTTPDiscovery returns the Prometheus HTTP service-discovery groups
// for a job, mapping the opaque device uid onto the stable Prometheus labels.
func (store *StatusStore) PrometheusHTTPDiscovery(job mev1.ScrapeJob) []PrometheusHTTPDiscoveryGroup {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	targets := make([]*mev1.ScrapeTarget, 0, len(store.targets))
	for _, target := range store.targets {
		if target.GetJob() == job {
			targets = append(targets, target)
		}
	}
	sort.Slice(targets, func(left, right int) bool {
		return scrapeTargetKey(targets[left]) < scrapeTargetKey(targets[right])
	})

	groups := make([]PrometheusHTTPDiscoveryGroup, 0, len(targets))
	for _, target := range targets {
		groups = append(groups, PrometheusHTTPDiscoveryGroup{
			Targets: []string{target.GetTarget()},
			Labels:  prometheusTargetLabels(target),
		})
	}
	return groups
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
	Hosts  []*mev1.DeviceState
	Agents []*mev1.Agent
	GitHub *mev1.GithubSyncDetail
}

func (store *StatusStore) Snapshot() StatusSnapshot {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	hosts := make([]*mev1.DeviceState, 0, len(store.hosts))
	for _, host := range store.hosts {
		hosts = append(hosts, host)
	}
	sort.Slice(hosts, func(left, right int) bool {
		return hosts[left].GetDeviceUid() < hosts[right].GetDeviceUid()
	})

	agents := make([]*mev1.Agent, 0, len(store.agents))
	for _, agent := range store.agents {
		agents = append(agents, agent)
	}
	sort.Slice(agents, func(left, right int) bool {
		if agents[left].GetDeviceUid() != agents[right].GetDeviceUid() {
			return agents[left].GetDeviceUid() < agents[right].GetDeviceUid()
		}
		return agents[left].GetKind() < agents[right].GetKind()
	})

	return StatusSnapshot{
		Mobile: store.mobile,
		Hosts:  hosts,
		Agents: agents,
		GitHub: cloneGithub(store.github),
	}
}

// saveLocked persists the durable state atomically (temp file + fsync + rename)
// so a crash mid-write cannot leave a truncated or corrupt state file.
func (store *StatusStore) saveLocked() error {
	snapshot := persistedState{}
	targets := make([]*mev1.ScrapeTarget, 0, len(store.targets))
	for _, target := range store.targets {
		targets = append(targets, target)
	}
	sort.Slice(targets, func(left, right int) bool {
		return scrapeTargetKey(targets[left]) < scrapeTargetKey(targets[right])
	})
	for _, target := range targets {
		raw, err := protojson.Marshal(target)
		if err != nil {
			return err
		}
		snapshot.Targets = append(snapshot.Targets, raw)
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

func agentStoreKey(agent *mev1.Agent) string {
	if agent.GetDeviceUid() == "" {
		return agent.GetKind()
	}
	return agent.GetDeviceUid() + "/" + agent.GetKind()
}

func scrapeTargetKey(target *mev1.ScrapeTarget) string {
	return scrapeJobString(target.GetJob()) + "/" + target.GetTarget()
}

func prometheusTargetLabels(target *mev1.ScrapeTarget) map[string]string {
	labels := map[string]string{}
	instance := firstString(target.GetDeviceUid(), target.GetTarget())
	if instance != "" {
		labels["instance"] = instance
	}
	if target.GetDeviceUid() != "" {
		labels["device_id"] = target.GetDeviceUid()
	}
	if target.GetDisplayName() != "" {
		labels["device_name"] = target.GetDisplayName()
	}
	if target.GetModel() != "" {
		labels["device_model"] = target.GetModel()
	}
	if kind := deviceKindString(target.GetKind()); kind != "" {
		labels["device_kind"] = kind
	}
	if role := deviceRoleString(target.GetRole()); role != "" {
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
