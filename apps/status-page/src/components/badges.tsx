import { AlertTriangle, CheckCircle2, Headphones, Music } from 'lucide-react';
import type { ReactElement, ReactNode } from 'react';
import type { Accessory, MediaStatus } from '@/gen/realtime/me/v1/status_types_pb';
import { OnlineState } from '@/gen/realtime/me/v1/status_types_pb';
import { Badge } from '@/components/ui/badge';
import { CardDescription } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { compactText, formatPercent } from '@/lib/format';
import { onlineStateLabel } from '@/lib/status';

type MetricBadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline';

export function MetricBadges({ children }: { children: ReactNode }) {
  return <div className="flex flex-wrap gap-2">{children}</div>;
}

export function MetricBadge({ icon, value, title, variant = 'outline' }: {
  icon: ReactElement;
  value: string;
  title: string;
  variant?: MetricBadgeVariant;
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant={variant} title={title} aria-label={value ? `${title}: ${value}` : title}>
          {icon}
          {value}
        </Badge>
      </TooltipTrigger>
      <TooltipContent>{title}</TooltipContent>
    </Tooltip>
  );
}

export function MediaBadge({ media, maxLength = 36 }: { media: MediaStatus; maxLength?: number }) {
  const cover = media.coverUrl;
  const text = mediaText(media);
  const icon = cover ? (
    <img src={cover} alt="" className="-ml-0.5 size-4 rounded-sm object-cover" width={16} height={16} />
  ) : (
    <Music />
  );
  return <MetricBadge icon={icon} value={compactText(text, maxLength)} title={`Playing: ${text}`} variant="secondary" />;
}

export function AccessoryBadges({ accessories, maxLength = 30 }: { accessories?: Accessory[]; maxLength?: number }) {
  const connected = accessories?.filter((accessory) => accessory.displayName) ?? [];
  if (connected.length === 0) return null;
  return (
    <>
      {connected.map((accessory) => (
        <MetricBadge
          key={`${accessory.kind}:${accessory.displayName}:${accessory.model}`}
          icon={<Headphones />}
          value={compactText(accessoryText(accessory), maxLength)}
          title={accessoryTitle(accessory)}
          variant="secondary"
        />
      ))}
    </>
  );
}

export function ProgressMetric({ label, value, valueText }: { label: string; value: number | null | undefined; valueText?: string }) {
  const safeValue = Math.max(0, Math.min(100, value ?? 0));
  return (
    <div className="grid gap-2">
      <div className="flex items-center justify-between gap-3 text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="truncate text-right font-medium">{valueText ?? formatPercent(value)}</span>
      </div>
      <Progress value={safeValue} />
    </div>
  );
}

export function DeviceModel({ model }: { model?: string }) {
  if (!model) return null;
  return <p className="text-xs text-muted-foreground">{model}</p>;
}

export function NoMetrics() {
  return <CardDescription>No metrics yet</CardDescription>;
}

export function StatusBadge({ state }: { state?: OnlineState }) {
  const online = state === OnlineState.ONLINE;
  const offline = state === OnlineState.OFFLINE;
  return (
    <Badge variant={offline ? 'destructive' : online ? 'default' : 'outline'} aria-label={onlineStateLabel(state)}>
      {offline ? <AlertTriangle /> : online ? <CheckCircle2 /> : '—'}
    </Badge>
  );
}

export function accessoryCount(accessories: Accessory[] | undefined): number {
  return accessories?.filter((accessory) => accessory.displayName).length ?? 0;
}

function mediaText(media: MediaStatus): string {
  return media.artist ? `${media.title} · ${media.artist}` : media.title;
}

function accessoryText(accessory: Accessory): string {
  return accessory.batteryPercent === undefined ? accessory.displayName : `${accessory.displayName} · ${accessory.batteryPercent}%`;
}

function accessoryTitle(accessory: Accessory): string {
  return [accessory.displayName, accessory.model, accessory.batteryPercent === undefined ? '' : `${accessory.batteryPercent}%`]
    .filter(Boolean)
    .join(' · ');
}
