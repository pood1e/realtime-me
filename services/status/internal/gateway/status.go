package gateway

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mev1 "github.com/pood1e/realtime-me/gen/go/realtime/me/status/v1"
)

// mobileStaleAfter bounds how long the phone's last push keeps being served.
// Everything else on the status document comes from Prometheus, which expires a
// series that stops being scraped; the pushed phone state has no such clock, so
// without this a phone that lost power would report "charging" indefinitely.
const mobileStaleAfter = 15 * time.Minute

// derivedStatusTTL bounds how often the status document is rebuilt from
// Prometheus. GetPublicStatus is unauthenticated and unmetered, and one call
// fans out across every host, agent, media and accessory query; without this a
// loop over the public RPC is cheap for the caller and a query storm for us.
// It is well under the page's own 10s poll, so the page never shows stale data
// it would not have shown anyway.
const derivedStatusTTL = 2 * time.Second

// derivedStatusTimeout bounds one refresh, well under the server write timeout.
const derivedStatusTimeout = 10 * time.Second

// derivedStatus is everything on the status document that comes from Prometheus.
type derivedStatus struct {
	server  *mev1.DeviceState
	devices []*mev1.DeviceState
	agents  []*mev1.Agent
}

// prometheusCache serves one Prometheus fan-out to every caller within the TTL,
// and performs at most one fan-out at a time.
type prometheusCache struct {
	mutex    sync.Mutex
	value    derivedStatus
	cachedAt time.Time
}

// buildPublicStatus assembles the unauthenticated status document. Hosts, VMs,
// and agents come from Prometheus, which scrapes their exporters. Only the
// phones are pushed, because they cannot be scraped.
func (server *StatusServer) buildPublicStatus(ctx context.Context) *mev1.PublicStatus {
	snapshot := server.store.Snapshot()
	now := time.Now().UTC()
	derived := server.derivedStatus(ctx, now)

	return &mev1.PublicStatus{
		Server:     derived.server,
		Mobiles:    freshMobiles(snapshot.Mobiles, now),
		Devices:    derived.devices,
		Agents:     derived.agents,
		Github:     publicGithubStatus(snapshot.GitHub, server.config.GitHubToken),
		UpdateTime: timestamppb.New(now),
	}
}

// derivedStatus reads the Prometheus-derived half of the document, from the
// cache when it is fresh and by fanning out concurrently when it is not.
func (server *StatusServer) derivedStatus(ctx context.Context, now time.Time) derivedStatus {
	server.cache.mutex.Lock()
	defer server.cache.mutex.Unlock()
	if !server.cache.cachedAt.IsZero() && time.Since(server.cache.cachedAt) < derivedStatusTTL {
		return server.cache.value
	}

	// A client that hangs up mid-refresh must not cancel the queries and leave a
	// truncated document in the cache for everyone else.
	queryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), derivedStatusTimeout)
	defer cancel()

	var media map[string]*mev1.MediaStatus
	var accessories map[string][]*mev1.Accessory
	parallel(
		func() { media = server.prometheus.DeviceMediaStatuses(queryCtx) },
		func() { accessories = server.prometheus.DeviceAccessoryStatuses(queryCtx) },
	)

	var fresh derivedStatus
	parallel(
		func() { fresh.server = namedServer(server.prometheus.ServerStatus(queryCtx, media, accessories)) },
		func() { fresh.devices = server.prometheus.NodeExporterStatuses(queryCtx, media, accessories) },
		func() { fresh.agents = server.visibleAgents(queryCtx, now) },
	)

	server.cache.value = fresh
	// Stamped now the fan-out is over, not when the request arrived: a rebuild
	// that took longer than the TTL would otherwise be born expired, and every
	// caller after it would rebuild again behind the same lock.
	server.cache.cachedAt = time.Now()
	return fresh
}

// buildInternalStatus reuses the public assembly but exposes the full GitHub
// sync diagnostics.
func (server *StatusServer) buildInternalStatus(ctx context.Context) *mev1.InternalStatus {
	public := server.buildPublicStatus(ctx)
	github := internalGithubDetail(server.store.GitHubSnapshot(), server.config.GitHubToken)
	return &mev1.InternalStatus{
		Server:     public.GetServer(),
		Mobiles:    public.GetMobiles(),
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
			Uid:        agentUID("", "agents", "", 0),
			Kind:       "agents",
			State:      mev1.AgentState_AGENT_STATE_IDLE,
			UpdateTime: timestamppb.New(now),
		}}
	}
	return agents
}

// freshMobiles drops each phone independently when it stops refreshing.
func freshMobiles(mobiles []*mev1.MobileState, now time.Time) []*mev1.MobileState {
	fresh := make([]*mev1.MobileState, 0, len(mobiles))
	for _, mobile := range mobiles {
		if mobile = freshMobile(mobile, now); mobile != nil {
			fresh = append(fresh, mobile)
		}
	}
	return fresh
}

func freshMobile(mobile *mev1.MobileState, now time.Time) *mev1.MobileState {
	updateTime := mobile.GetUpdateTime()
	if updateTime == nil {
		return nil
	}
	if now.Sub(updateTime.AsTime()) > mobileStaleAfter {
		return nil
	}
	return mobile
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
