import {
  Activity,
  AlertTriangle,
  Battery,
  BatteryCharging,
  Bot,
  GitBranch,
  CheckCircle2,
  Clock,
  CircleOff,
  HeartPulse,
  LoaderCircle,
  RefreshCw,
  Server,
  Smartphone,
  Watch,
  Wifi,
} from 'lucide-react';
import { useEffect, useMemo, useState, type ReactElement, type ReactNode } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

const POLL_INTERVAL_MS = 10_000;

type PublicStatus = {
  server: {
    online: boolean;
    cpu_percent: number | null;
    memory_percent: number | null;
    disk_percent: number | null;
    updated_at: string;
  };
  mobile: MobileStatus | null;
  agents: AgentStatus[];
  github: GitHubStatus;
  updated_at: string;
};

type GitHubStatus = {
  enabled: boolean;
  state: 'disabled' | 'pending' | 'ok' | 'error';
  updated_at?: string;
  emoji?: string;
  message?: string;
};

type MobileStatus = {
  device_id: string;
  updated_at: string;
  received_at: string;
  phone?: {
    battery_percent?: number;
    charge_state?: ChargeState;
    network?: string;
  };
  watch?: {
    heart_rate?: number;
    steps?: number;
    battery_percent?: number;
    charge_state?: ChargeState;
    wrist_state?: 'unknown' | 'on_wrist' | 'off_wrist';
  };
};

type AgentStatus = {
  agent_id: string;
  state: 'idle' | 'running' | 'failed';
  task?: string;
  budget_remaining_percent?: number;
  updated_at: string;
  received_at: string;
};

type ChargeState = 'unknown' | 'charging' | 'not_charging';

type MetricBadgeProps = {
  icon: ReactElement;
  value: string;
  title: string;
  variant?: 'default' | 'secondary' | 'destructive' | 'outline';
};

export function App() {
  const apiBaseUrl = useMemo(statusApiBaseUrl, []);
  const [status, setStatus] = useState<PublicStatus | null>(null);
  const [failed, setFailed] = useState(false);

  async function refresh() {
    const next = await fetch(`${apiBaseUrl}/api/public-status`, { cache: 'no-store' })
      .then((response) => (response.ok ? response.json() as Promise<PublicStatus> : null))
      .catch(() => null);
    setFailed(next === null);
    if (next) setStatus(next);
  }

  useEffect(() => {
    void refresh();
    const interval = window.setInterval(() => void refresh(), POLL_INTERVAL_MS);
    return () => window.clearInterval(interval);
  }, [apiBaseUrl]);

  return (
    <TooltipProvider>
      <main className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(20,184,166,0.18),transparent_35rem),radial-gradient(circle_at_top_right,rgba(59,130,246,0.18),transparent_30rem)]">
        <div className="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-6 px-5 py-8">
          <header className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm text-muted-foreground">Realtime Me</p>
              <h1 className="text-3xl font-semibold tracking-tight">Live status</h1>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={failed ? 'destructive' : 'default'}>
                {failed ? <AlertTriangle /> : <CheckCircle2 />}
                {failed ? 'API offline' : 'Live'}
              </Badge>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="secondary" size="icon" aria-label="Refresh" title="Refresh" onClick={() => void refresh()}>
                    <RefreshCw />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Refresh</TooltipContent>
              </Tooltip>
            </div>
          </header>

          <section className="grid gap-4 md:grid-cols-3">
            <ServerCard status={status} />
            <PhoneCard mobile={status?.mobile ?? null} />
            <WatchCard mobile={status?.mobile ?? null} />
          </section>

          <section className="grid gap-4 md:grid-cols-2">
            <GitHubCard github={status?.github ?? null} />
            <AgentCard agents={status?.agents ?? []} />
          </section>

          <footer className="flex items-center gap-2 text-xs text-muted-foreground">
            <Clock className="size-3.5" />
            <span>{status ? `Updated ${formatTime(status.updated_at)}` : 'Waiting for first status'}</span>
          </footer>
        </div>
      </main>
    </TooltipProvider>
  );
}

function ServerCard({ status }: { status: PublicStatus | null }) {
  const server = status?.server;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><Server className="size-4" />Server</CardTitle>
        <CardAction>
          <Badge variant={server?.online ? 'default' : 'destructive'}>
            {server?.online ? <CheckCircle2 /> : <AlertTriangle />}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <ProgressMetric label="CPU" value={server?.cpu_percent} />
        <ProgressMetric label="Mem" value={server?.memory_percent} />
        <ProgressMetric label="Disk" value={server?.disk_percent} />
      </CardContent>
    </Card>
  );
}

function PhoneCard({ mobile }: { mobile: MobileStatus | null }) {
  const phone = mobile?.phone;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><Smartphone className="size-4" />Phone</CardTitle>
        <CardAction>{mobile ? <Badge><CheckCircle2 /></Badge> : <Badge variant="outline">—</Badge>}</CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        <MetricBadges>
          <MetricBadge icon={<Battery />} value={formatBattery(phone?.battery_percent)} title="Battery" />
          {phone?.charge_state === 'charging' && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          <MetricBadge icon={<Wifi />} value={phone?.network ?? '—'} title="Network" variant="secondary" />
        </MetricBadges>
        <MutedTime value={mobile?.received_at} fallback="No phone report" />
      </CardContent>
    </Card>
  );
}

function WatchCard({ mobile }: { mobile: MobileStatus | null }) {
  const watch = mobile?.watch;
  const offWrist = watch?.wrist_state === 'off_wrist';
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><Watch className="size-4" />Watch</CardTitle>
        <CardAction>{offWrist ? <Badge variant="secondary"><CircleOff /></Badge> : <Badge><CheckCircle2 /></Badge>}</CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        <MetricBadges>
          {!offWrist && <MetricBadge icon={<HeartPulse />} value={watch?.heart_rate ? `${watch.heart_rate}` : '—'} title="Heart rate" />}
          <MetricBadge icon={<Activity />} value={watch?.steps?.toLocaleString() ?? '—'} title="Steps" variant="secondary" />
          <MetricBadge icon={<Battery />} value={formatBattery(watch?.battery_percent)} title="Battery" />
          {watch?.charge_state === 'charging' && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          {offWrist && <MetricBadge icon={<CircleOff />} value="" title="Off wrist" variant="secondary" />}
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

function GitHubCard({ github }: { github: GitHubStatus | null }) {
  const icon = githubIcon(github?.state);
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><GitBranch />GitHub</CardTitle>
        <CardAction><Badge variant={githubBadgeVariant(github?.state)}>{icon}</Badge></CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        {github ? (
          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary">{github.emoji ?? '⌚'}</Badge>
            <Badge variant="outline">{github.message ?? github.state}</Badge>
          </div>
        ) : (
          <Skeleton className="h-5 w-32" />
        )}
        <MutedTime value={github?.updated_at} fallback="Waiting for status sync" />
      </CardContent>
    </Card>
  );
}

function AgentCard({ agents }: { agents: AgentStatus[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><Bot className="size-4" />Agents</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3">
        {agents.map((agent, index) => (
          <div key={agent.agent_id}>
            {index > 0 && <Separator className="mb-3" />}
            <div className="flex items-center justify-between gap-3">
              <span className="font-medium">{agent.agent_id}</span>
              <Badge variant={agent.state === 'failed' ? 'destructive' : agent.state === 'running' ? 'default' : 'secondary'}>{agent.state}</Badge>
            </div>
            <p className="mt-2 text-sm text-muted-foreground">{agent.task ?? '—'}</p>
            {agent.budget_remaining_percent !== undefined && (
              <div className="mt-3 grid gap-2">
                <Progress value={agent.budget_remaining_percent} />
                <span className="text-xs text-muted-foreground">Budget {agent.budget_remaining_percent}%</span>
              </div>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

function ProgressMetric({ label, value }: { label: string; value: number | null | undefined }) {
  const safeValue = Math.max(0, Math.min(100, value ?? 0));
  return (
    <div className="grid gap-2">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">{value === null || value === undefined ? '—' : `${Math.round(safeValue)}%`}</span>
      </div>
      <Progress value={safeValue} />
    </div>
  );
}

function MetricBadges({ children }: { children: ReactNode }) {
  return <div className="flex flex-wrap gap-2">{children}</div>;
}

function MetricBadge({ icon, value, title, variant = 'outline' }: MetricBadgeProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant={variant} title={title}>
          {icon}
          {value}
        </Badge>
      </TooltipTrigger>
      <TooltipContent>{title}</TooltipContent>
    </Tooltip>
  );
}

function MutedTime({ value, fallback }: { value?: string; fallback: string }) {
  return <p className="text-xs text-muted-foreground">{value ? formatTime(value) : fallback}</p>;
}

function githubIcon(state: GitHubStatus['state'] | undefined): ReactElement {
  if (state === 'error') return <AlertTriangle />;
  if (state === 'pending') return <LoaderCircle className="animate-spin" />;
  if (state === 'ok') return <CheckCircle2 />;
  return <GitBranch />;
}

function githubBadgeVariant(state: GitHubStatus['state'] | undefined): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (state === 'error') return 'destructive';
  if (state === 'ok') return 'default';
  if (state === 'pending') return 'secondary';
  return 'outline';
}

function formatBattery(value: number | undefined): string {
  return value === undefined ? '—' : `${value}%`;
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(new Date(value));
}

function statusApiBaseUrl(): string {
  const configured = import.meta.env.VITE_STATUS_API_BASE_URL as string | undefined;
  if (configured?.trim()) return configured.replace(/\/+$/, '');

  const { protocol, hostname } = window.location;
  if (hostname === 'localhost' || hostname === '127.0.0.1') return 'http://localhost:18080';
  if (hostname.startsWith('status.')) return `${protocol}//api-${hostname}`;
  return window.location.origin;
}
