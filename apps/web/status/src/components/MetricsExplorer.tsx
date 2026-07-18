import { AlertTriangle, Battery, Cpu, Footprints, HardDrive, Headphones, HeartPulse, LineChart as LineChartIcon, MemoryStick } from 'lucide-react';
import { lazy, Suspense, useEffect, useState, type ReactElement } from 'react';
import { create } from '@bufbuild/protobuf';
import { timestampFromDate } from '@bufbuild/protobuf/wkt';
import { GetMetricRangeRequestSchema, MetricSeries } from '@/gen/realtime/me/v1/metrics_pb';
import type { GetMetricRangeRequest } from '@/gen/realtime/me/v1/metrics_pb';
import type { Agent, DeviceState, InternalStatus, MobileState } from '@/gen/realtime/me/v1/status_pb';
import type { Accessory } from '@/gen/realtime/me/v1/status_types_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { agentIcon, agentName } from '@/components/AgentCard';
import { EmptyCard } from '@/components/layout';
import type { ChartPoint, ChartUnit } from '@/lib/format';
import { CPU_USAGE, MEMORY_USAGE, hasDiskMetric, hasMetric } from '@/lib/metrics';
import { deviceDisplayName, hostDevices } from '@/lib/status';
import { authHeaders, metricsClient } from '@/lib/transport';

type ChartRange = {
  id: string;
  label: string;
  durationMs: number;
  stepSeconds: number;
};

// A chart names the series it wants and the entity it belongs to. The gateway
// owns every metric name, label, and query expression.
type ChartDefinition = {
  id: string;
  title: string;
  unit: ChartUnit;
  icon: ReactElement;
  series: MetricSeries;
  deviceUid?: string;
  agentKind?: string;
  accessory?: { kind: string; displayName: string };
};

const CHART_RANGES: ChartRange[] = [
  { id: '15m', label: '15m', durationMs: 15 * 60_000, stepSeconds: 15 },
  { id: '1h', label: '1h', durationMs: 60 * 60_000, stepSeconds: 30 },
  { id: '6h', label: '6h', durationMs: 6 * 60 * 60_000, stepSeconds: 120 },
  { id: '24h', label: '24h', durationMs: 24 * 60 * 60_000, stepSeconds: 300 },
];

const StatusChart = lazy(() => import('@/components/StatusChart'));

export function MetricsExplorer({ status, token }: { status: InternalStatus; token: string }) {
  const [rangeId, setRangeId] = useState(CHART_RANGES[1].id);
  const range = CHART_RANGES.find((item) => item.id === rangeId) ?? CHART_RANGES[1];
  // Not memoised on `status`: each poll yields a fresh object, so the memo never
  // hit. Building the definitions is cheap, and every chart fetches its own data.
  const charts = chartDefinitions(status);
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
  const { data, failed } = useMetricRange(token, chart, range);
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

// The window is recomputed on every refresh, so a chart advances with the
// dashboard's own poll instead of freezing at the moment it first rendered.
function useMetricRange(token: string, chart: ChartDefinition, range: ChartRange): { data: ChartPoint[]; failed: boolean } {
  const [data, setData] = useState<ChartPoint[]>([]);
  const [failed, setFailed] = useState(false);
  const { series, deviceUid, agentKind, accessory } = chart;

  useEffect(() => {
    const controller = new AbortController();
    let active = true;

    async function load() {
      try {
        const response = await metricsClient.getMetricRange(
          metricRangeRequest({ series, deviceUid, agentKind, accessory }, range),
          { headers: authHeaders(token), signal: controller.signal },
        );
        if (!active) return;
        setData(response.points.map(chartPoint).filter((point): point is ChartPoint => point !== null));
        setFailed(false);
      } catch {
        if (active && !controller.signal.aborted) setFailed(true);
      }
    }

    void load();
    const timer = window.setInterval(() => void load(), POLL_METRICS_MS);
    return () => {
      active = false;
      controller.abort();
      window.clearInterval(timer);
    };
  }, [series, deviceUid, agentKind, accessory?.kind, accessory?.displayName, range.durationMs, range.stepSeconds, token]);

  return { data, failed };
}

const POLL_METRICS_MS = 30_000;

function metricRangeRequest(
  chart: Pick<ChartDefinition, 'series' | 'deviceUid' | 'agentKind' | 'accessory'>,
  range: ChartRange,
): GetMetricRangeRequest {
  const end = new Date();
  const start = new Date(end.getTime() - range.durationMs);
  return create(GetMetricRangeRequestSchema, {
    series: chart.series,
    deviceUid: chart.deviceUid ?? '',
    agentKind: chart.agentKind ?? '',
    accessory: chart.accessory ? { kind: chart.accessory.kind, displayName: chart.accessory.displayName } : undefined,
    startTime: timestampFromDate(start),
    endTime: timestampFromDate(end),
    step: { seconds: BigInt(range.stepSeconds), nanos: 0 },
  });
}

function chartPoint(point: { time?: { seconds: bigint }; value: number }): ChartPoint | null {
  if (!point.time) return null;
  return { time: Number(point.time.seconds) * 1000, value: point.value };
}

function chartDefinitions(status: InternalStatus): ChartDefinition[] {
  const definitions: ChartDefinition[] = [];
  for (const device of hostDevices(status)) {
    definitions.push(...hostChartDefinitions(device));
  }
  for (const mobile of status.mobiles) {
    definitions.push(...mobileChartDefinitions(mobile));
  }
  for (const agent of status.agents) {
    if (agent.budgetRemainingPercent !== undefined) definitions.push(agentBudgetChart(agent));
  }
  return definitions;
}

function hostChartDefinitions(device: DeviceState): ChartDefinition[] {
  const identity = deviceDisplayName(device, 'Device');
  const definitions: ChartDefinition[] = [];
  if (hasMetric(device, CPU_USAGE)) {
    definitions.push({ id: `${device.deviceUid}:cpu`, title: `${identity} CPU`, unit: 'percent', icon: <Cpu className="size-4" />, series: MetricSeries.HOST_CPU_UTILIZATION, deviceUid: device.deviceUid });
  }
  if (hasMetric(device, MEMORY_USAGE)) {
    definitions.push({ id: `${device.deviceUid}:mem`, title: `${identity} memory`, unit: 'bytes', icon: <MemoryStick className="size-4" />, series: MetricSeries.HOST_MEMORY_USAGE, deviceUid: device.deviceUid });
  }
  if (hasDiskMetric(device)) {
    definitions.push({ id: `${device.deviceUid}:disk`, title: `${identity} disk`, unit: 'percent', icon: <HardDrive className="size-4" />, series: MetricSeries.HOST_FILESYSTEM_UTILIZATION, deviceUid: device.deviceUid });
  }
  definitions.push(...accessoryBatteryCharts(device.deviceUid, identity, device.accessories));
  return definitions;
}

function mobileChartDefinitions(mobile: MobileState): ChartDefinition[] {
  const definitions: ChartDefinition[] = [];
  const phoneName = mobile.displayName || 'Phone';
  if (mobile.phone?.batteryPercent !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:phone-battery`, title: `${phoneName} battery`, unit: 'percent', icon: <Battery className="size-4" />, series: MetricSeries.PHONE_BATTERY_LEVEL, deviceUid: mobile.deviceUid });
  }
  definitions.push(...accessoryBatteryCharts(mobile.deviceUid, phoneName, mobile.phone?.accessories));

  const watch = mobile.watch;
  if (!watch) return definitions;
  const watchName = watch.deviceInfo?.displayName || 'Watch';
  if (watch.heartRate !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:watch-hr`, title: `${watchName} heart rate`, unit: 'rate', icon: <HeartPulse className="size-4" />, series: MetricSeries.WATCH_HEART_RATE, deviceUid: mobile.deviceUid });
  }
  if (watch.activityTotals !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:watch-steps`, title: `${watchName} steps`, unit: 'count', icon: <Footprints className="size-4" />, series: MetricSeries.WATCH_STEPS, deviceUid: mobile.deviceUid });
  }
  if (watch.watchState !== undefined) {
    definitions.push({ id: `${mobile.deviceUid}:watch-battery`, title: `${watchName} battery`, unit: 'percent', icon: <Battery className="size-4" />, series: MetricSeries.WATCH_BATTERY_LEVEL, deviceUid: mobile.deviceUid });
  }
  return definitions;
}

function accessoryBatteryCharts(deviceUid: string, deviceName: string, accessories: Accessory[] | undefined): ChartDefinition[] {
  return (accessories ?? [])
    .filter((accessory) => accessory.displayName && accessory.batteryPercent !== undefined)
    .map((accessory) => ({
      id: `${deviceUid}:${accessory.kind}:${accessory.displayName}:battery`,
      title: `${deviceName} ${accessory.displayName}`,
      unit: 'percent' as ChartUnit,
      icon: <Headphones className="size-4" />,
      series: MetricSeries.ACCESSORY_BATTERY_LEVEL,
      deviceUid,
      accessory: { kind: accessory.kind, displayName: accessory.displayName },
    }));
}

function agentBudgetChart(agent: Agent): ChartDefinition {
  return {
    id: `${agent.uid}:budget`,
    title: `${agentName(agent.kind)} budget`,
    unit: 'percent',
    icon: agentIcon(agent.kind),
    series: MetricSeries.AGENT_BUDGET_REMAINING,
    agentKind: agent.kind,
    deviceUid: agent.deviceUid,
  };
}
