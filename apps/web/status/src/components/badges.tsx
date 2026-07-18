import type { Accessory, MediaStatus } from "@realtime-me/status-contracts";
import { OnlineState } from "@realtime-me/status-contracts";
import { Badge } from "@realtime-me/web-ui/badge";
import { CardDescription } from "@realtime-me/web-ui/card";
import { Progress } from "@realtime-me/web-ui/progress";
import { Tooltip, TooltipContent, TooltipTrigger } from "@realtime-me/web-ui/tooltip";
import { AlertTriangle, CheckCircle2, Headphones, Music } from "lucide-react";
import type { ReactElement, ReactNode } from "react";
import { compactText, formatPercent } from "@/lib/format";
import { onlineStateLabel } from "@/lib/status";

type MetricBadgeVariant = "default" | "secondary" | "destructive" | "outline";

export function MetricBadges({ children }: { children: ReactNode }) {
  return <div className="flex flex-wrap gap-2">{children}</div>;
}

export function MetricBadge({
  icon,
  value,
  title,
  variant = "outline",
}: {
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
  const text = mediaText(media);
  return (
    <MetricBadge
      icon={<Music />}
      value={compactText(text, maxLength)}
      title={`Playing: ${text}`}
      variant="secondary"
    />
  );
}

export function AccessoryBadges({
  accessories,
  maxLength = 30,
}: {
  accessories?: Accessory[];
  maxLength?: number;
}) {
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

export function RingGauge({
  value,
  label,
  detail,
}: {
  value: number | null | undefined;
  label: string;
  detail?: string;
}) {
  const radius = 15.5;
  const circumference = 2 * Math.PI * radius;
  const pct = Math.max(0, Math.min(100, value ?? 0));
  const offset = circumference * (1 - pct / 100);
  const tone = pct >= 90 ? "stroke-destructive" : pct >= 75 ? "stroke-warning" : "stroke-primary";
  return (
    <div
      className="flex flex-col items-center gap-1.5"
      title={detail ? `${label}: ${detail}` : label}
    >
      <div className="relative size-[3.25rem]">
        <svg aria-hidden="true" viewBox="0 0 40 40" className="size-full -rotate-90">
          <circle
            cx="20"
            cy="20"
            r={radius}
            fill="none"
            strokeWidth="3.5"
            className="stroke-muted"
          />
          <circle
            cx="20"
            cy="20"
            r={radius}
            fill="none"
            strokeWidth="3.5"
            strokeLinecap="round"
            className={`${tone} transition-[stroke-dashoffset] duration-700`}
            style={{ strokeDasharray: circumference, strokeDashoffset: offset }}
          />
        </svg>
        <span className="absolute inset-0 flex items-center justify-center text-xs font-semibold tabular-nums">
          {Math.round(pct)}
          <span className="text-[9px] text-muted-foreground">%</span>
        </span>
      </div>
      <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </span>
    </div>
  );
}

export function StatCell({
  icon,
  value,
  label,
}: {
  icon: ReactElement;
  value: string;
  label: string;
}) {
  return (
    <div className="flex flex-col items-center gap-1.5" title={label}>
      <div className="flex size-[3.25rem] items-center justify-center rounded-full border border-border/70 bg-muted/40 text-primary [&_svg]:size-5">
        {icon}
      </div>
      <span className="text-center text-[11px] font-semibold tabular-nums leading-tight">
        {value}
      </span>
    </div>
  );
}

export function ProgressMetric({
  label,
  value,
  valueText,
}: {
  label: string;
  value: number | null | undefined;
  valueText?: string;
}) {
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
    <Badge
      variant={offline ? "destructive" : online ? "default" : "outline"}
      aria-label={onlineStateLabel(state)}
    >
      {offline ? <AlertTriangle /> : online ? <CheckCircle2 /> : "—"}
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
  return accessory.batteryPercent === undefined
    ? accessory.displayName
    : `${accessory.displayName} · ${accessory.batteryPercent}%`;
}

function accessoryTitle(accessory: Accessory): string {
  return [
    accessory.displayName,
    accessory.model,
    accessory.batteryPercent === undefined ? "" : `${accessory.batteryPercent}%`,
  ]
    .filter(Boolean)
    .join(" · ");
}
