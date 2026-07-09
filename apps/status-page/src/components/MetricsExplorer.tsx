import { AlertTriangle, Battery, Cpu, Footprints, HardDrive, Headphones, HeartPulse, LineChart as LineChartIcon, MemoryStick } from 'lucide-react';
import { lazy, Suspense, useEffect, useMemo, useState, type ReactElement } from 'react';
import type { Agent, DeviceState, InternalStatus, MobileState } from '@/gen/realtime/me/v1/status_pb';
import { DeviceRole } from '@/gen/realtime/me/v1/status_types_pb';
import type { Accessory } from '@/gen/realtime/me/v1/status_types_pb';
import { WristState } from '@/gen/realtime/me/v1/watch_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { agentIcon, agentName } from '@/components/AgentCard';
import { EmptyCard } from '@/components/layout';
import type { ChartPoint, ChartUnit } from '@/lib/format';
import { CPU_USAGE, MEMORY_USAGE, hasDiskMetric, hasMetric, promLabel } from '@/lib/metrics';
import { deviceDisplayName, hostDevices } from '@/lib/status';
import { apiBaseUrl, authHeaders } from '@/lib/transport';

type ChartRange = {
  id: string;
  label: string;
  durationMs: number;
  step: string;
};

type ChartDefinition = {
  id: string;
  title: string;
  query: string;
  unit: ChartUnit;
  icon: ReactElement;
};

type PrometheusRangeResponse = {
  status: 'success' | 'error';
  data?: {
    result?: Array<{
      metric: Record<string, string>;
      values: Array<[number, string]>;
    }>;
  };
  error?: string;
};

const CHART_RANGES: ChartRange[] = [
  { id: '15m', label: '15m', durationMs: 15 * 60_000, step: '15s' },
  { id: '1h', label: '1h', durationMs: 60 * 60_000, step: '30s' },
  { id: '6h', label: '6h', durationMs: 6 * 60 * 60_000, step: '2m' },
  { id: '24h', label: '24h', durationMs: 24 * 60 * 60_000, step: '5m' },
];

// The always-on server is not enrolled, so it has no minted uid. Prometheus
// labels its static node-exporter target with this fixed instance instead.
const SERVER_INSTANCE = 'server';

const StatusChart = lazy(() => import('@/components/StatusChart'));

export function MetricsExplorer({ status, token }: { status: InternalStatus; token: string }) {
  const [rangeId, setRangeId] = useState(CHART_RANGES[1].id);
  const range = CHART_RANGES.find((item) => item.id === rangeId) ?? CHART_RANGES[1];
  const charts = useMemo(() => chartDefinitions(status), [status]);
  return (
    <div className="grid gap-4">
      <div className="flex items-center justify-between gap-3">
        <h2 className="flex items-center gap-2 text-lg font-semibold tracking-tight"><LineChartIcon className="size-4" />Metrics</h2>
        <Select value={range.id} onValueChange={setRangeId}>
          <SelectTrigger size="sm" aria-label="Metric range">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {CHART_RANGES.map((item) => <SelectItem key={item.id} value={item.id}>{item.label}</SelectItem>)}
          </SelectContent>
        </Select>
      </div>
      <div className="grid gap-4 lg:grid-cols-2">
        {charts.length === 0 ? <EmptyCard text="No chartable metrics" /> : charts.map((chart) => (
          <TimeSeriesCard key={chart.id} chart={chart} range={range} token={token} />
        ))}
      </div>
    </div>
  );
}

function TimeSeriesCard({ chart, range, token }: { chart: ChartDefinition; range: ChartRange; token: string }) {
  const { data, failed } = usePrometheusRange(token, chart.query, range);
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">{chart.icon}{chart.title}</CardTitle>
        <CardAction>{failed ? <Badge variant="destructive"><AlertTriangle /></Badge> : <Badge variant="outline">{range.label}</Badge>}</CardAction>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <CardDescription>No samples</CardDescription>
        ) : (
          <Suspense fallback={<CardDescription>Loading chart</CardDescription>}>
            <StatusChart data={data} unit={chart.unit} />
          </Suspense>
        )}
      </CardContent>
    </Card>
  );
}

function usePrometheusRange(token: string, query: string, range: ChartRange): { data: ChartPoint[]; failed: boolean } {
  const [data, setData] = useState<ChartPoint[]>([]);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    const end = Math.floor(Date.now() / 1000);
    const start = Math.floor((Date.now() - range.durationMs) / 1000);
    const params = new URLSearchParams({ query, start: `${start}`, end: `${end}`, step: range.step });
    fetch(`${apiBaseUrl}/api/internal/metrics/query_range?${params.toString()}`, {
      cache: 'no-store',
      headers: authHeaders(token),
      signal: controller.signal,
    })
      .then((response) => response.ok ? response.json() as Promise<PrometheusRangeResponse> : null)
      .then((payload) => {
        if (!payload || payload.status !== 'success') {
          setFailed(true);
          return;
        }
        setData(prometheusPoints(payload));
        setFailed(false);
      })
      .catch((error: unknown) => {
        if ((error as DOMException).name !== 'AbortError') setFailed(true);
      });
    return () => controller.abort();
  }, [query, range.durationMs, range.step, token]);

  return { data, failed };
}

function prometheusPoints(payload: PrometheusRangeResponse): ChartPoint[] {
  const values = payload.data?.result?.[0]?.values ?? [];
  return values
    .map(([time, value]) => ({ time: Number(time) * 1000, value: Number(value) }))
    .filter((point) => Number.isFinite(point.time) && Number.isFinite(point.value));
}

function chartDefinitions(status: InternalStatus): ChartDefinition[] {
  const definitions: ChartDefinition[] = [];
  for (const device of hostDevices(status)) {
    definitions.push(...hostChartDefinitions(device));
  }
  if (status.mobile) {
    definitions.push(...mobileChartDefinitions(status.mobile));
  }
  for (const agent of status.agents) {
    if (agent.budgetRemainingPercent !== undefined) definitions.push(agentBudgetChart(agent));
  }
  return definitions;
}

function hostChartDefinitions(device: DeviceState): ChartDefinition[] {
  const identity = deviceDisplayName(device, 'Device');
  const queries = hostQueries(device);
  const definitions: ChartDefinition[] = [];
  if (queries.cpu && hasMetric(device, CPU_USAGE)) definitions.push({ id: `${device.deviceUid}:cpu`, title: `${identity} CPU`, query: queries.cpu, unit: 'percent', icon: <Cpu className="size-4" /> });
  if (queries.memory && hasMetric(device, MEMORY_USAGE)) definitions.push({ id: `${device.deviceUid}:mem`, title: `${identity} memory`, query: queries.memory, unit: 'bytes', icon: <MemoryStick className="size-4" /> });
  if (queries.disk && hasDiskMetric(device)) definitions.push({ id: `${device.deviceUid}:disk`, title: `${identity} disk`, query: queries.disk, unit: 'percent', icon: <HardDrive className="size-4" /> });
  definitions.push(...accessoryBatteryCharts(device.deviceUid, identity, device.accessories));
  return definitions;
}

function mobileChartDefinitions(mobile: MobileState): ChartDefinition[] {
  const definitions: ChartDefinition[] = [];
  const phoneName = mobile.displayName || 'Phone';
  if (mobile.phone?.batteryPercent !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:phone-battery`, title: `${phoneName} battery`, query: `realtime_device_battery_level_ratio{device_id=${promLabel(mobile.deviceUid)},device_type="phone"} * 100`, unit: 'percent', icon: <Battery className="size-4" /> });
  }
  definitions.push(...accessoryBatteryCharts(mobile.deviceUid, phoneName, mobile.phone?.accessories, 'phone'));
  const watch = mobile.watch;
  if (!watch) return definitions;
  const watchName = watch.deviceInfo?.displayName || 'Watch';
  if (watch.watchState?.wristState !== WristState.OFF_WRIST && watch.heartRate !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:watch-hr`, title: `${watchName} heart rate`, query: `realtime_watch_heart_rate_beats_per_minute{device_id=${promLabel(mobile.deviceUid)}}`, unit: 'rate', icon: <HeartPulse className="size-4" /> });
  }
  if (watch.activityTotals !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:watch-steps`, title: `${watchName} steps`, query: `realtime_watch_steps{device_id=${promLabel(mobile.deviceUid)}}`, unit: 'count', icon: <Footprints className="size-4" /> });
  }
  if (watch.watchState !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:watch-battery`, title: `${watchName} battery`, query: `realtime_device_battery_level_ratio{device_id=${promLabel(mobile.deviceUid)},device_type="watch"} * 100`, unit: 'percent', icon: <Battery className="size-4" /> });
  }
  return definitions;
}

function accessoryBatteryCharts(deviceId: string, deviceName: string, accessories: Accessory[] | undefined, deviceType?: string): ChartDefinition[] {
  return (accessories ?? [])
    .filter((accessory) => accessory.displayName && accessory.batteryPercent !== undefined)
    .map((accessory) => {
      const labels = [
        `device_id=${promLabel(deviceId)}`,
        `accessory_kind=${promLabel(accessory.kind)}`,
        `accessory_name=${promLabel(accessory.displayName)}`,
      ];
      if (deviceType) labels.push(`device_type=${promLabel(deviceType)}`);
      return {
        id: `${deviceId}:${accessory.kind}:${accessory.displayName}:battery`,
        title: `${deviceName} ${accessory.displayName}`,
        query: `realtime_device_accessory_battery_level_ratio{${labels.join(',')}} * 100`,
        unit: 'percent' as ChartUnit,
        icon: <Headphones className="size-4" />,
      };
    });
}

function agentBudgetChart(agent: Agent): ChartDefinition {
  const labels = [`agent_id=${promLabel(agent.uid)}`];
  if (agent.deviceUid) labels.push(`device_id=${promLabel(agent.deviceUid)}`);
  return {
    id: `${agent.uid}:budget`,
    title: `${agentName(agent.kind)} budget`,
    query: `realtime_agent_budget_remaining_ratio{${labels.join(',')}} * 100`,
    unit: 'percent',
    icon: agentIcon(agent.kind),
  };
}

// Every host, VM, and the server itself runs node_exporter. Service discovery
// sets `instance` to the device uid, and the always-on server's static config
// sets it to "server", so one query shape covers them all.
function hostQueries(device: DeviceState): { cpu: string; memory: string; disk: string } {
  const isServer = device.role === DeviceRole.SERVER || device.deviceUid === SERVER_INSTANCE;
  return nodeExporterQueries(isServer ? SERVER_INSTANCE : device.deviceUid);
}

function nodeExporterQueries(instance: string): { cpu: string; memory: string; disk: string } {
  const base = `job=~"node-exporter|vm-node-exporter",instance=${promLabel(instance)}`;
  const diskBase = `${base},mountpoint="/",fstype!~"tmpfs|overlay|squashfs"`;
  return {
    cpu: `100 * (1 - avg(rate(node_cpu_seconds_total{${base},mode="idle"}[2m])))`,
    memory: `node_memory_MemTotal_bytes{${base}} - node_memory_MemAvailable_bytes{${base}}`,
    disk: `100 * (1 - node_filesystem_avail_bytes{${diskBase}} / node_filesystem_size_bytes{${diskBase}})`,
  };
}
