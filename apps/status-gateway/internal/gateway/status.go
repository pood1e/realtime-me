package gateway

import (
	"context"
	"sort"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// buildPublicStatus assembles the unauthenticated status document by merging the
// pulled Prometheus view with the pushed in-memory status.
func (server *StatusServer) buildPublicStatus(ctx context.Context) *mev1.PublicStatus {
	snapshot := server.store.Snapshot()
	now := time.Now().UTC()

	agents := server.mergedAgents(ctx, snapshot, now)
	github := snapshot.GitHub

	return &mev1.PublicStatus{
		Server:     mergeServerDevice(server.prometheus.ServerStatus(ctx), snapshot.Hosts),
		Mobile:     snapshot.Mobile,
		Devices:    mergeDevices(server.prometheus.NodeExporterStatuses(ctx), nonServerHosts(snapshot.Hosts)),
		Agents:     agents,
		Github:     publicGithubStatus(github, server.config.GitHubToken),
		UpdateTime: timestamppb.New(now),
	}
}

// buildInternalStatus reuses the public assembly but exposes the full GitHub
// sync diagnostics.
func (server *StatusServer) buildInternalStatus(ctx context.Context) *mev1.InternalStatus {
	public := server.buildPublicStatus(ctx)
	github := internalGithubDetail(server.store.GitHubSnapshot(), server.config.GitHubToken)
	return &mev1.InternalStatus{
		Server:     public.GetServer(),
		Mobile:     public.GetMobile(),
		Devices:    public.GetDevices(),
		Agents:     public.GetAgents(),
		Github:     github,
		UpdateTime: public.GetUpdateTime(),
	}
}

func (server *StatusServer) mergedAgents(ctx context.Context, snapshot StatusSnapshot, now time.Time) []*mev1.Agent {
	fresh := runningAgents(snapshot.Agents, now, time.Duration(server.config.AgentFreshSeconds)*time.Second)
	agents := mergeAgents(server.prometheus.AgentStatuses(ctx), fresh)
	if len(agents) == 0 && server.config.PublicAgentPlaceholder {
		return []*mev1.Agent{{
			Uid:        agentUID("", "agents"),
			Kind:       "agents",
			State:      mev1.AgentState_AGENT_STATE_IDLE,
			UpdateTime: timestamppb.New(now),
		}}
	}
	return agents
}

func mergeServerDevice(base *mev1.DeviceState, hosts []*mev1.DeviceState) *mev1.DeviceState {
	for _, host := range hosts {
		if host.GetRole() != mev1.DeviceRole_DEVICE_ROLE_SERVER && host.GetDeviceUid() != "server" {
			continue
		}
		base.DeviceUid = firstString(host.GetDeviceUid(), base.GetDeviceUid())
		base.DisplayName = firstString(host.GetDisplayName(), base.GetDisplayName())
		base.Model = firstString(host.GetModel(), base.GetModel())
		if host.GetState() != mev1.OnlineState_ONLINE_STATE_UNSPECIFIED {
			base.State = host.GetState()
		}
		base.UpdateTime = host.GetUpdateTime()
		if len(host.GetMetrics()) > 0 {
			base.Metrics = host.GetMetrics()
		}
		if host.GetMedia() != nil {
			base.Media = host.GetMedia()
		}
		if len(host.GetAccessories()) > 0 {
			base.Accessories = host.GetAccessories()
		}
		break
	}
	if base.GetDisplayName() == "" {
		base.DisplayName = "Server"
	}
	return base
}

func nonServerHosts(hosts []*mev1.DeviceState) []*mev1.DeviceState {
	filtered := make([]*mev1.DeviceState, 0, len(hosts))
	for _, host := range hosts {
		if host.GetRole() == mev1.DeviceRole_DEVICE_ROLE_SERVER || host.GetDeviceUid() == "server" {
			continue
		}
		filtered = append(filtered, host)
	}
	return filtered
}

func mergeDevices(primary []*mev1.DeviceState, stored []*mev1.DeviceState) []*mev1.DeviceState {
	merged := make(map[string]*mev1.DeviceState, len(primary)+len(stored))
	for _, device := range primary {
		merged[device.GetDeviceUid()] = device
	}
	for _, device := range stored {
		existing, ok := merged[device.GetDeviceUid()]
		if !ok {
			merged[device.GetDeviceUid()] = device
			continue
		}
		existing.DisplayName = firstString(device.GetDisplayName(), existing.GetDisplayName())
		existing.Model = firstString(device.GetModel(), existing.GetModel())
		if device.GetMedia() != nil {
			existing.Media = device.GetMedia()
		}
		if len(device.GetAccessories()) > 0 {
			existing.Accessories = device.GetAccessories()
		}
	}

	result := make([]*mev1.DeviceState, 0, len(merged))
	for _, device := range merged {
		result = append(result, device)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].GetDeviceUid() < result[right].GetDeviceUid()
	})
	return result
}

func mergeAgents(primary []*mev1.Agent, fallback []*mev1.Agent) []*mev1.Agent {
	merged := make(map[string]*mev1.Agent, len(primary)+len(fallback))
	for _, agent := range fallback {
		merged[agent.GetDeviceUid()+"/"+agent.GetKind()] = agent
	}
	for _, agent := range primary {
		merged[agent.GetDeviceUid()+"/"+agent.GetKind()] = agent
	}

	result := make([]*mev1.Agent, 0, len(merged))
	for _, agent := range merged {
		result = append(result, agent)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].GetDeviceUid() != result[right].GetDeviceUid() {
			return result[left].GetDeviceUid() < result[right].GetDeviceUid()
		}
		return result[left].GetKind() < result[right].GetKind()
	})
	return result
}

func runningAgents(agents []*mev1.Agent, now time.Time, freshness time.Duration) []*mev1.Agent {
	result := make([]*mev1.Agent, 0, len(agents))
	for _, agent := range agents {
		if agent.GetState() != mev1.AgentState_AGENT_STATE_RUNNING {
			continue
		}
		update := agent.GetUpdateTime()
		if update == nil || now.Sub(update.AsTime()) <= freshness {
			result = append(result, agent)
		}
	}
	return result
}

func publicGithubStatus(detail *mev1.GithubSyncDetail, token string) *mev1.GithubStatus {
	state := detail.GetState()
	if token == "" {
		state = mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED
	} else if !detail.GetConfigured() {
		state = mev1.GithubSyncState_GITHUB_SYNC_STATE_PENDING
	}
	return &mev1.GithubStatus{
		Enabled:    token != "",
		State:      state,
		Emoji:      detail.GetEmoji(),
		Message:    detail.GetMessage(),
		UpdateTime: detail.GetLastSuccessTime(),
	}
}

func internalGithubDetail(detail *mev1.GithubSyncDetail, token string) *mev1.GithubSyncDetail {
	if token == "" {
		detail.Configured = false
		detail.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED
	} else if !detail.GetConfigured() {
		detail.State = mev1.GithubSyncState_GITHUB_SYNC_STATE_PENDING
	}
	return detail
}
