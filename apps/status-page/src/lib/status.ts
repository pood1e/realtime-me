import type { Agent, DeviceState, InternalStatus, Subagent } from '@/gen/realtime/me/v1/status_pb';
import { DeviceKind, DeviceRole, NetworkState, OnlineState } from '@/gen/realtime/me/v1/status_types_pb';

export function isVirtualMachine(device: DeviceState): boolean {
  return device.kind === DeviceKind.VIRTUAL_MACHINE || device.role === DeviceRole.VM;
}

// hostDevices is the server plus every reporting device, skipping an absent server.
export function hostDevices(status: InternalStatus): DeviceState[] {
  return [status.server, ...status.devices].filter((device): device is DeviceState => device !== undefined);
}

// deviceCounts reports how many devices are online against the total, counting the
// phone and watch as always-online personal devices when present.
export function deviceCounts(status: InternalStatus): { online: number; total: number } {
  const hosts = hostDevices(status);
  let total = hosts.length;
  let online = hosts.filter((device) => device.state === OnlineState.ONLINE).length;
  if (status.mobile) {
    total += 1;
    online += 1;
  }
  if (status.mobile?.watch) {
    total += 1;
    online += 1;
  }
  return { online, total };
}

export function deviceDisplayName(device: DeviceState | null | undefined, fallback: string): string {
  return humanLabel(device?.displayName) ?? fallback;
}

export function agentDeviceLabel(agent: Agent): string {
  return humanLabel(agent.displayName) ?? '';
}

// The sub-agents an agent has out, grouped by the model each runs, busiest first.
// A sub-agent need not run the model that spawned it, which is why they are
// grouped by model rather than counted as heads.
export function subagentModelCounts(subagents: Subagent[]): Array<{ model: string; count: number }> {
  const counts = new Map<string, number>();
  for (const subagent of subagents) counts.set(subagent.model, (counts.get(subagent.model) ?? 0) + 1);
  return [...counts]
    .map(([model, count]) => ({ model, count }))
    .sort((left, right) => right.count - left.count || left.model.localeCompare(right.model));
}

// A sub-agent whose model the exporter could not read is left nameable only by
// its number, which is still worth showing.
export function subagentLabel(model: string, count: number): string {
  if (model) return `${count} × ${model}`;
  return count === 1 ? '1 sub-agent' : `${count} sub-agents`;
}

// humanLabel keeps only a name a person would recognise. A LAN address is not
// one; neither is the device uid, which is an internal identifier and never a
// candidate here however empty the name.
function humanLabel(value: string | undefined): string | undefined {
  return value && !isPrivateIPv4(value) ? value : undefined;
}

export function onlineStateLabel(state: OnlineState | undefined): string {
  if (state === OnlineState.ONLINE) return 'online';
  if (state === OnlineState.OFFLINE) return 'offline';
  return 'unknown';
}

export function networkLabel(network: NetworkState): string {
  switch (network) {
    case NetworkState.WIFI:
      return 'Wi-Fi';
    case NetworkState.CELLULAR:
      return 'Cellular';
    case NetworkState.VPN:
      return 'VPN';
    case NetworkState.ONLINE:
      return 'Online';
    case NetworkState.OFFLINE:
      return 'Offline';
    default:
      return '—';
  }
}

function isPrivateIPv4(value: string): boolean {
  const octets = value.split('.').map((part) => Number(part));
  if (octets.length !== 4 || octets.some((part) => !Number.isInteger(part) || part < 0 || part > 255)) return false;
  const [first, second] = octets;
  return first === 10 ||
    first === 127 ||
    (first === 172 && second >= 16 && second <= 31) ||
    (first === 192 && second === 168) ||
    (first === 169 && second === 254);
}
