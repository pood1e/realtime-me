package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	mev1 "realtime-me/apps/status-gateway/internal/genproto/realtime/me/v1"
)

// agentUID derives a stable, opaque agent identifier from the host, the agent
// kind, the model it runs and its ordinal among the agents of that kind, so the
// public surface never exposes an internal session or thread id. A host can run
// several agents of one kind at once, and the ordinal is what tells them apart:
// it is stable while they are, and no agent's own identity ever leaves the host.
func agentUID(deviceUID string, kind string, model string, ordinal int) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "realtime.me.agent:%s/%s/%s/%d", deviceUID, kind, model, ordinal))
	return "agt_" + hex.EncodeToString(sum[:8])
}

// The Prometheus label contract uses lowercase string values for kind/role/
// state/job. These helpers translate between those strings and the proto enums
// so the wire contract stays proto while Prometheus labels stay stable.

func deviceKindString(kind mev1.DeviceKind) string {
	switch kind {
	case mev1.DeviceKind_DEVICE_KIND_HOST:
		return "host"
	case mev1.DeviceKind_DEVICE_KIND_VIRTUAL_MACHINE:
		return "virtual_machine"
	case mev1.DeviceKind_DEVICE_KIND_PHONE:
		return "phone"
	case mev1.DeviceKind_DEVICE_KIND_WATCH:
		return "watch"
	default:
		return ""
	}
}

func parseDeviceKind(value string) mev1.DeviceKind {
	switch value {
	case "host":
		return mev1.DeviceKind_DEVICE_KIND_HOST
	case "virtual_machine":
		return mev1.DeviceKind_DEVICE_KIND_VIRTUAL_MACHINE
	case "phone":
		return mev1.DeviceKind_DEVICE_KIND_PHONE
	case "watch":
		return mev1.DeviceKind_DEVICE_KIND_WATCH
	default:
		return mev1.DeviceKind_DEVICE_KIND_UNSPECIFIED
	}
}

func deviceRoleString(role mev1.DeviceRole) string {
	switch role {
	case mev1.DeviceRole_DEVICE_ROLE_SERVER:
		return "server"
	case mev1.DeviceRole_DEVICE_ROLE_DESKTOP:
		return "desktop"
	case mev1.DeviceRole_DEVICE_ROLE_VM:
		return "vm"
	default:
		return ""
	}
}

func parseDeviceRole(value string) mev1.DeviceRole {
	switch value {
	case "server":
		return mev1.DeviceRole_DEVICE_ROLE_SERVER
	case "desktop":
		return mev1.DeviceRole_DEVICE_ROLE_DESKTOP
	case "vm":
		return mev1.DeviceRole_DEVICE_ROLE_VM
	default:
		return mev1.DeviceRole_DEVICE_ROLE_UNSPECIFIED
	}
}

func onlineState(up bool) mev1.OnlineState {
	if up {
		return mev1.OnlineState_ONLINE_STATE_ONLINE
	}
	return mev1.OnlineState_ONLINE_STATE_OFFLINE
}

func networkStateString(state mev1.NetworkState) string {
	switch state {
	case mev1.NetworkState_NETWORK_STATE_OFFLINE:
		return "offline"
	case mev1.NetworkState_NETWORK_STATE_WIFI:
		return "wifi"
	case mev1.NetworkState_NETWORK_STATE_CELLULAR:
		return "cellular"
	case mev1.NetworkState_NETWORK_STATE_VPN:
		return "vpn"
	case mev1.NetworkState_NETWORK_STATE_ONLINE:
		return "online"
	default:
		return "unknown"
	}
}

func githubStateString(state mev1.GithubSyncState) string {
	switch state {
	case mev1.GithubSyncState_GITHUB_SYNC_STATE_DISABLED:
		return "disabled"
	case mev1.GithubSyncState_GITHUB_SYNC_STATE_PENDING:
		return "pending"
	case mev1.GithubSyncState_GITHUB_SYNC_STATE_OK:
		return "ok"
	case mev1.GithubSyncState_GITHUB_SYNC_STATE_ERROR:
		return "error"
	default:
		return "disabled"
	}
}

func parseScrapeJob(value string) (mev1.ScrapeJob, bool) {
	switch value {
	case "node-exporter":
		return mev1.ScrapeJob_SCRAPE_JOB_NODE_EXPORTER, true
	case "vm-node-exporter":
		return mev1.ScrapeJob_SCRAPE_JOB_VM_NODE_EXPORTER, true
	case "device-exporter":
		return mev1.ScrapeJob_SCRAPE_JOB_DEVICE_EXPORTER, true
	case "agent-exporter":
		return mev1.ScrapeJob_SCRAPE_JOB_AGENT_EXPORTER, true
	default:
		return mev1.ScrapeJob_SCRAPE_JOB_UNSPECIFIED, false
	}
}

func firstString(primary string, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
