import { Activity, Box, Cpu, HardDrive, MemoryStick } from 'lucide-react';
import type { ReactElement } from 'react';
import type { DeviceState } from '@/gen/realtime/me/v1/status_pb';
import { OnlineState } from '@/gen/realtime/me/v1/status_types_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { deviceIcon } from '@/components/brand';
import {
  AccessoryBadges,
  DeviceModel,
  MediaBadge,
  MetricBadge,
  MetricBadges,
  NoMetrics,
  ProgressMetric,
  RingGauge,
  StatCell,
  StatusBadge,
  accessoryCount,
} from '@/components/badges';
import { InlineTime } from '@/components/layout';
import {
  CPU_CORES,
  CPU_USAGE,
  cpuCoreText,
  cpuText,
  diskValues,
  hasMetric,
  memoryValues,
  metricPercent,
} from '@/lib/metrics';
import { deviceDisplayName, onlineStateLabel } from '@/lib/status';

export function DeviceCard({ device, title, icon, showChildren = true }: {
  device: DeviceState | null;
  title: string;
  icon: ReactElement;
  showChildren?: boolean;
}) {
  const displayName = deviceDisplayName(device, title);
  const memory = memoryValues(device);
  const disk = diskValues(device);
  const cpuUsage = metricPercent(device, CPU_USAGE);
  const hasCpuCores = hasMetric(device, CPU_CORES);
  const hasMemory = memory.percent !== undefined;
  const hasDisk = disk.percent !== undefined;
  const showCpuBadge = hasCpuCores && cpuUsage === undefined;
  const hasAnyMetric = hasCpuCores || cpuUsage !== undefined || hasMemory || hasDisk || !!device?.media?.title || accessoryCount(device?.accessories) > 0;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">{deviceIcon(device, icon)}{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={device?.updateTime} />
          <StatusBadge state={device?.state} />
        </CardAction>
      </CardHeader>
      <CardContent className="flex h-full flex-col gap-4">
        <DeviceModel model={device?.model} />
        {(cpuUsage !== undefined || hasMemory || hasDisk || showCpuBadge) && (
          <div className="flex flex-wrap items-start justify-around gap-x-2 gap-y-3 py-1">
            {cpuUsage !== undefined && <RingGauge value={cpuUsage} label="CPU" detail={cpuText(device)} />}
            {hasMemory && <RingGauge value={memory.percent} label="Mem" detail={memory.text} />}
            {hasDisk && <RingGauge value={disk.percent} label="Disk" detail={disk.text} />}
            {showCpuBadge && cpuUsage === undefined && <StatCell icon={<Cpu />} value={cpuCoreText(device)} label="CPU" />}
          </div>
        )}
        {(device?.media?.title || accessoryCount(device?.accessories) > 0) && (
          <MetricBadges>
            {device?.media?.title && <MediaBadge media={device.media} />}
            <AccessoryBadges accessories={device?.accessories} />
          </MetricBadges>
        )}
        {!hasAnyMetric && <NoMetrics />}
        {showChildren && <ChildDevices devices={device?.children ?? []} />}
      </CardContent>
    </Card>
  );
}

export function InternalDeviceCard({ device, icon }: { device: DeviceState; icon: ReactElement }) {
  const memory = memoryValues(device);
  const disk = diskValues(device);
  const cpuUsage = metricPercent(device, CPU_USAGE);
  const metricCount = device.metrics.length;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">{deviceIcon(device, icon)}<span className="truncate">{deviceDisplayName(device, 'Device')}</span></CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={device.updateTime} />
          <StatusBadge state={device.state} />
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <DeviceModel model={device.model} />
        <MetricBadges>
          {device.media?.title && <MediaBadge media={device.media} />}
          <AccessoryBadges accessories={device.accessories} />
          {hasMetric(device, CPU_CORES) && <MetricBadge icon={<Cpu />} value={cpuCoreText(device)} title="CPU cores" variant="secondary" />}
          <MetricBadge icon={<Activity />} value={`${metricCount}`} title="Metrics" variant="outline" />
        </MetricBadges>
        {cpuUsage !== undefined && <ProgressMetric label="CPU" value={cpuUsage} valueText={cpuText(device)} />}
        {memory.percent !== undefined && <ProgressMetric label="Mem" value={memory.percent} valueText={memory.text} />}
        {disk.percent !== undefined && <ProgressMetric label="Disk" value={disk.percent} valueText={disk.text} />}
        {metricCount === 0 && !device.media?.title && accessoryCount(device.accessories) === 0 && <NoMetrics />}
      </CardContent>
    </Card>
  );
}

function ChildDevices({ devices }: { devices: DeviceState[] }) {
  if (devices.length === 0) return null;
  return (
    <div className="grid gap-2">
      {devices.map((device) => (
        <div key={device.deviceUid} className="grid gap-1 rounded-md bg-muted/40 p-2 text-sm">
          <div className="flex items-center justify-between gap-2">
            <span className="flex min-w-0 items-center gap-2 truncate text-muted-foreground"><Box className="size-3.5" />{deviceDisplayName(device, 'Device')}</span>
            <Badge variant={device.state === OnlineState.ONLINE ? 'default' : 'secondary'}>{onlineStateLabel(device.state)}</Badge>
          </div>
          <MetricBadges>
            {device.media?.title && <MediaBadge media={device.media} maxLength={28} />}
            <AccessoryBadges accessories={device.accessories} maxLength={28} />
            {hasMetric(device, CPU_CORES) && <MetricBadge icon={<Cpu />} value={cpuCoreText(device)} title="CPU cores" variant="secondary" />}
            {memoryValues(device).percent !== undefined && <MetricBadge icon={<MemoryStick />} value={memoryValues(device).text} title="Memory" />}
            {diskValues(device).percent !== undefined && <MetricBadge icon={<HardDrive />} value={diskValues(device).text} title="Disk" />}
          </MetricBadges>
        </div>
      ))}
    </div>
  );
}
