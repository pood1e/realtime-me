import type {
  Agent,
  DeviceState,
  InternalStatus,
  MobileState,
  Subagent,
} from "@realtime-me/status-contracts";
import { DeviceKind, DeviceRole, NetworkState, OnlineState } from "@realtime-me/status-contracts";

export function isVirtualMachine(device: DeviceState): boolean {
  return device.kind === DeviceKind.VIRTUAL_MACHINE || device.role === DeviceRole.VM;
}

// hostDevices is the server plus every reporting device, skipping an absent server.
export function hostDevices(status: InternalStatus): DeviceState[] {
  return [status.server, ...status.devices].filter(
    (device): device is DeviceState => device !== undefined,
  );
}

// deviceCounts reports how many devices are online against the total, counting the
// phones and watches as always-online personal devices when present.
export function deviceCounts(status: InternalStatus): { online: number; total: number } {
  const hosts = hostDevices(status);
  let total = hosts.length;
  let online = hosts.filter((device) => device.state === OnlineState.ONLINE).length;

  for (const mobile of status.mobiles) {
    total += 1;
    online += 1;
    if (mobile.watch) {
      total += 1;
      online += 1;
    }
    if (mobile.switchPresence) {
      total += 1;
      if (mobile.switchPresence.state === OnlineState.ONLINE) online += 1;
    }
  }
  return { online, total };
}

export function isPlayingOnSwitch(mobile: MobileState): boolean {
  return (
    mobile.switchPresence?.state === OnlineState.ONLINE && Boolean(mobile.switchPresence.gameName)
  );
}

export function deviceDisplayName(
  device: DeviceState | null | undefined,
  fallback: string,
): string {
  return humanLabel(device?.displayName) ?? fallback;
}

export function agentDeviceLabel(agent: Agent): string {
  return humanLabel(agent.displayName) ?? "";
}

// The sub-agents an agent has out, grouped by the model each runs, busiest first.
// A sub-agent need not run the model that spawned it, which is why they are
// grouped by model rather than counted as heads.
function subagentModelCounts(subagents: Subagent[]): Array<{ model: string; count: number }> {
  const counts = new Map<string, number>();
  for (const subagent of subagents)
    counts.set(subagent.model, (counts.get(subagent.model) ?? 0) + 1);
  return [...counts]
    .map(([model, count]) => ({ model, count }))
    .sort((left, right) => right.count - left.count || left.model.localeCompare(right.model));
}

// How many sub-agents an agent has out. The models they run are the detail, not
// the headline: a sub-agent usually runs the model that spawned it, and naming
// it on the badge only repeats the agent's own model back to the reader.
export function subagentCountLabel(count: number): string {
  return count === 1 ? "1 sub-agent" : `${count} sub-agents`;
}

// The models behind that count, for the badge's tooltip. Empty when the exporter
// could read no model at all, which is the one case the count must stand alone.
export function subagentModelSummary(subagents: Subagent[]): string {
  return subagentModelCounts(subagents)
    .filter(({ model }) => model)
    .map(({ model, count }) => (count === 1 ? model : `${count} × ${model}`))
    .join(" · ");
}

// humanLabel keeps only a name a person would recognise. A LAN address is not
// one; neither is the device uid, which is an internal identifier and never a
// candidate here however empty the name.
function humanLabel(value: string | undefined): string | undefined {
  return value && !isPrivateIPv4(value) ? value : undefined;
}

export function onlineStateLabel(state: OnlineState | undefined): string {
  if (state === OnlineState.ONLINE) return "online";
  if (state === OnlineState.OFFLINE) return "offline";
  return "unknown";
}

export function networkLabel(network: NetworkState): string {
  switch (network) {
    case NetworkState.WIFI:
      return "Wi-Fi";
    case NetworkState.CELLULAR:
      return "Cellular";
    case NetworkState.VPN:
      return "VPN";
    case NetworkState.ONLINE:
      return "Online";
    case NetworkState.OFFLINE:
      return "Offline";
    default:
      return "—";
  }
}

function isPrivateIPv4(value: string): boolean {
  const octets = value.split(".").map((part) => Number(part));
  if (
    octets.length !== 4 ||
    octets.some((part) => !Number.isInteger(part) || part < 0 || part > 255)
  )
    return false;
  const [first, second] = octets;
  return (
    first === 10 ||
    first === 127 ||
    (first === 172 && second >= 16 && second <= 31) ||
    (first === 192 && second === 168) ||
    (first === 169 && second === 254)
  );
}
