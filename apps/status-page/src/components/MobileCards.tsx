import { Battery, BatteryCharging, CheckCircle2, Footprints, Gamepad2, HeartPulse, Wifi } from 'lucide-react';
import { siAndroid, siWearos } from 'simple-icons/icons';
import type { MobileState } from '@/gen/realtime/me/v1/status_pb';
import type { GithubSyncState } from '@/gen/realtime/me/v1/status_pb';
import type { WatchSnapshot } from '@/gen/realtime/me/v1/watch_pb';
import { ChargeState } from '@/gen/realtime/me/v1/watch_pb';
import { OnlineState } from '@/gen/realtime/me/v1/status_types_pb';
import { Badge } from '@/components/ui/badge';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AccessoryBadges, MetricBadges, StatCell } from '@/components/badges';
import { BrandIcon } from '@/components/brand';
import { GitHubStatusBadge } from '@/components/github';
import { InlineTime } from '@/components/layout';
import { SwitchArtwork } from '@/components/SwitchArtwork';
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
      <CardContent className="flex h-full flex-col gap-4">
        <div className="flex flex-wrap items-start justify-around gap-x-2 gap-y-3 py-1">
          <StatCell
            icon={phone?.chargeState === ChargeState.CHARGING ? <BatteryCharging /> : <Battery />}
            value={formatBattery(phone?.batteryPercent)}
            label={phone?.chargeState === ChargeState.CHARGING ? 'Charging' : 'Battery'}
          />
          <StatCell icon={<Wifi />} value={phone ? networkLabel(phone.network) : '—'} label="Network" />
        </div>
        {(phone?.accessories?.length ?? 0) > 0 && (
          <MetricBadges>
            <AccessoryBadges accessories={phone?.accessories} />
          </MetricBadges>
        )}
      </CardContent>
    </Card>
  );
}

export function WatchCard({ mobile, githubState }: { mobile: MobileState | null; githubState?: GithubSyncState }) {
  const watch = mobile?.watch;
  const info = watch?.deviceInfo;
  const watchState = watch?.watchState;
  const heartRate = watch?.heartRate?.beatsPerMinute;
  const displayName = info?.displayName || 'Watch';
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><BrandIcon icon={siWearos} />{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={watch?.recordTime ?? mobile?.updateTime} />
          <GitHubStatusBadge state={githubState} />
          <WatchStatusBadge watch={watch} />
        </CardAction>
      </CardHeader>
      <CardContent className="flex h-full flex-col gap-4">
        <div className="flex flex-wrap items-start justify-around gap-x-2 gap-y-3 py-1">
          <StatCell icon={<HeartPulse />} value={heartRate ? `${heartRate}` : '—'} label="BPM" />
          <StatCell icon={<Footprints />} value={formatSteps(watch)} label="Steps" />
          <StatCell
            icon={watchState?.chargeState === ChargeState.CHARGING ? <BatteryCharging /> : <Battery />}
            value={formatBattery(watchState?.batteryPercent)}
            label={watchState?.chargeState === ChargeState.CHARGING ? 'Charging' : 'Battery'}
          />
        </div>
      </CardContent>
    </Card>
  );
}

export function SwitchCard({ mobile }: { mobile: MobileState | null }) {
  const presence = mobile?.switchPresence;
  const online = presence?.state === OnlineState.ONLINE;
  const gameName = online ? presence?.gameName : undefined;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><Gamepad2 className="size-4" />Nintendo Switch</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={presence?.fetchTime ?? mobile?.updateTime} />
          <SwitchStatusBadge mobile={mobile} />
        </CardAction>
      </CardHeader>
      <CardContent className="flex h-full flex-col gap-4">
        {!!gameName && (
          <div className="flex min-w-0 items-center gap-3">
            <SwitchArtwork imageUri={presence?.imageUri} className="size-14 rounded-lg border border-border/70" />
            <div className="min-w-0">
              <p className="text-xs text-muted-foreground">Playing now</p>
              <p className="line-clamp-2 text-sm font-semibold leading-snug">{gameName}</p>
            </div>
          </div>
        )}
        <div className="flex flex-wrap items-start justify-around gap-x-2 gap-y-3 py-1">
          {!gameName && <StatCell icon={<Gamepad2 />} value={online ? 'Online' : '—'} label="Presence" />}
          <StatCell icon={<CheckCircle2 />} value={presence ? onlineStateLabel(presence.state) : '—'} label="State" />
        </div>
        {!!gameName && !!presence?.titleId && (
          <MetricBadges>
            <Badge variant="secondary" title={presence.titleId}>Title {presence.titleId}</Badge>
          </MetricBadges>
        )}
      </CardContent>
    </Card>
  );
}

function WatchStatusBadge({ watch }: { watch?: WatchSnapshot }) {
  if (!watch) return <Badge variant="outline" aria-label="No data">—</Badge>;
  return <Badge aria-label="Online"><CheckCircle2 /></Badge>;
}

function SwitchStatusBadge({ mobile }: { mobile: MobileState | null }) {
  const presence = mobile?.switchPresence;
  if (!presence) return <Badge variant="outline" aria-label="No data">—</Badge>;
  if (presence.state === OnlineState.ONLINE) return <Badge aria-label="Online"><CheckCircle2 /></Badge>;
  return <Badge variant="outline" aria-label="Offline">Offline</Badge>;
}

function formatSteps(watch: WatchSnapshot | undefined): string {
  if (!watch) return '—';
  return (watch.activityTotals?.steps ?? 0).toLocaleString();
}

function onlineStateLabel(state: OnlineState): string {
  switch (state) {
    case OnlineState.ONLINE:
      return 'Online';
    case OnlineState.OFFLINE:
      return 'Offline';
    default:
      return '—';
  }
}
