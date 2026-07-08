import { Battery, BatteryCharging, CheckCircle2, CircleOff, Footprints, HeartPulse, Wifi } from 'lucide-react';
import { siAndroid, siWearos } from 'simple-icons/icons';
import type { MobileState } from '@/gen/realtime/me/v1/status_pb';
import type { GithubSyncState } from '@/gen/realtime/me/v1/status_pb';
import type { WatchSnapshot } from '@/gen/realtime/me/v1/watch_pb';
import { ChargeState, WristState } from '@/gen/realtime/me/v1/watch_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AccessoryBadges, DeviceModel, MetricBadge, MetricBadges } from '@/components/badges';
import { BrandIcon } from '@/components/brand';
import { GitHubStatusBadge } from '@/components/github';
import { InlineTime } from '@/components/layout';
import { formatBattery } from '@/lib/format';
import { networkLabel } from '@/lib/status';

export function PhoneCard({ mobile }: { mobile: MobileState | null }) {
  const phone = mobile?.phone;
  const displayName = mobile?.displayName || 'Phone';
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><BrandIcon icon={siAndroid} />{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={mobile?.updateTime} />
          {mobile ? <Badge aria-label="Online"><CheckCircle2 /></Badge> : <Badge variant="outline" aria-label="No data">—</Badge>}
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        <DeviceModel model={mobile?.model} />
        <MetricBadges>
          <AccessoryBadges accessories={phone?.accessories} />
          <MetricBadge icon={<Battery />} value={formatBattery(phone?.batteryPercent)} title="Battery" />
          {phone?.chargeState === ChargeState.CHARGING && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          <MetricBadge icon={<Wifi />} value={phone ? networkLabel(phone.network) : '—'} title="Network" variant="secondary" />
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

export function WatchCard({ mobile, githubState }: { mobile: MobileState | null; githubState?: GithubSyncState }) {
  const watch = mobile?.watch;
  const info = watch?.deviceInfo;
  const watchState = watch?.watchState;
  const offWrist = watchState?.wristState === WristState.OFF_WRIST;
  const heartRate = watch?.heartRate?.beatsPerMinute;
  const displayName = info?.displayName || 'Watch';
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><BrandIcon icon={siWearos} />{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={watch?.recordTime ?? mobile?.updateTime} />
          <GitHubStatusBadge state={githubState} />
          <WatchStatusBadge watch={watch} offWrist={offWrist} />
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        <DeviceModel model={info?.model} />
        <MetricBadges>
          {!offWrist && <MetricBadge icon={<HeartPulse />} value={heartRate ? `${heartRate}` : '—'} title="Heart rate" />}
          <MetricBadge icon={<Footprints />} value={formatSteps(watch)} title="Steps" variant="secondary" />
          <MetricBadge icon={<Battery />} value={formatBattery(watchState?.batteryPercent)} title="Battery" />
          {watchState?.chargeState === ChargeState.CHARGING && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          {offWrist && <MetricBadge icon={<CircleOff />} value="" title="Off wrist" variant="secondary" />}
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

function WatchStatusBadge({ watch, offWrist }: { watch?: WatchSnapshot; offWrist: boolean }) {
  if (!watch) return <Badge variant="outline" aria-label="No data">—</Badge>;
  return offWrist
    ? <Badge variant="secondary" aria-label="Off wrist"><CircleOff /></Badge>
    : <Badge aria-label="Online"><CheckCircle2 /></Badge>;
}

function formatSteps(watch: WatchSnapshot | undefined): string {
  if (!watch) return '—';
  return (watch.activityTotals?.steps ?? 0).toLocaleString();
}
