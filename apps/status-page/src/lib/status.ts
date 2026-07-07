import type { Agent, DeviceState, InternalStatus } from '@/gen/realtime/me/v1/status_pb';
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
  return displayLabel(fallback, device?.displayName, device?.deviceUid);
}

export function agentDeviceLabel(agent: Agent): string {
  return displayLabel('', agent.displayName, agent.deviceUid);
}

export function displayLabel(fallback: string, ...values: Array<string | undefined>): string {
  return values.find((value) => value && !isPrivateIPv4(value)) ?? fallback;
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
