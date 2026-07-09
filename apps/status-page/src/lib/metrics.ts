import type { DeviceState } from '@/gen/realtime/me/v1/status_pb';
import { formatGigabytes, formatPercent } from '@/lib/format';

export const CPU_CORES = 'system.cpu.logical.count';
export const CPU_USAGE = 'system.cpu.utilization';
export const MEMORY_USAGE = 'system.memory.usage';
export const MEMORY_LIMIT = 'system.memory.limit';
export const FILESYSTEM_USAGE = 'system.filesystem.usage';
export const FILESYSTEM_LIMIT = 'system.filesystem.limit';
export const FILESYSTEM_UTILIZATION = 'system.filesystem.utilization';

type Device = DeviceState | null | undefined;

export function metricValue(device: Device, name: string): number | undefined {
  return device?.metrics?.find((metric) => metric.name === name)?.value;
}

export function metricPercent(device: Device, name: string): number | undefined {
  const value = metricValue(device, name);
  return value === undefined ? undefined : value * 100;
}

export function hasMetric(device: Device, name: string): boolean {
  return metricValue(device, name) !== undefined;
}

export function hasDiskMetric(device: Device): boolean {
  return hasMetric(device, FILESYSTEM_UTILIZATION) || hasMetric(device, FILESYSTEM_USAGE);
}

export function cpuCoreText(device: Device): string {
  const cores = metricValue(device, CPU_CORES);
  return cores === undefined ? '—' : `${Math.round(cores)}`;
}

export function cpuText(device: Device): string {
  const percent = formatPercent(metricPercent(device, CPU_USAGE));
  const cores = cpuCoreText(device);
  return cores === '—' ? percent : `${percent} · ${cores} cores`;
}

export function memoryValues(device: Device): { text: string; percent?: number } {
  const used = metricValue(device, MEMORY_USAGE);
  const total = metricValue(device, MEMORY_LIMIT);
  if (used === undefined || total === undefined || total <= 0) return { text: '—' };
  return { text: `${formatGigabytes(used)}/${formatGigabytes(total)}`, percent: (used * 100) / total };
}

export function diskValues(device: Device): { text: string; percent?: number } {
  const used = metricValue(device, FILESYSTEM_USAGE);
  const total = metricValue(device, FILESYSTEM_LIMIT);
  if (used !== undefined && total !== undefined && total > 0) {
    const percent = (used * 100) / total;
    return { text: `${formatGigabytes(used)}/${formatGigabytes(total)} · ${formatPercent(percent)}`, percent };
  }
  const directPercent = metricPercent(device, FILESYSTEM_UTILIZATION);
  if (directPercent !== undefined) return { text: formatPercent(directPercent), percent: directPercent };
  return { text: '—' };
}
