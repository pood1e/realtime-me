package gateway

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// buildPublicStatus assembles the unauthenticated status document. Hosts, VMs,
// and agents come from Prometheus, which scrapes their exporters. Only the
// phone is pushed, because it cannot be scraped.
func (server *StatusServer) buildPublicStatus(ctx context.Context) *mev1.PublicStatus {
	snapshot := server.store.Snapshot()
	now := time.Now().UTC()

	return &mev1.PublicStatus{
		Server:     namedServer(server.prometheus.ServerStatus(ctx)),
		Mobile:     snapshot.Mobile,
		Devices:    server.prometheus.NodeExporterStatuses(ctx),
		Agents:     server.visibleAgents(ctx, now),
		Github:     publicGithubStatus(snapshot.GitHub, server.config.GitHubToken),
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

// visibleAgents returns the agents Prometheus currently sees, or a single idle
// placeholder when the deployment prefers to show one rather than nothing.
func (server *StatusServer) visibleAgents(ctx context.Context, now time.Time) []*mev1.Agent {
	agents := server.prometheus.AgentStatuses(ctx)
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

// namedServer gives the always-on server a label when node_exporter supplies no
// hostname, so the card never renders untitled.
func namedServer(server *mev1.DeviceState) *mev1.DeviceState {
	if server != nil && server.GetDisplayName() == "" {
		server.DisplayName = "Server"
	}
	return server
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
