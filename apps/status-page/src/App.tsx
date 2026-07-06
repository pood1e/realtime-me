import {
  AlertTriangle,
  Battery,
  BatteryCharging,
  Bot,
  Box,
  CheckCircle2,
  CircleOff,
  Clock,
  Cpu,
  HardDrive,
  HeartPulse,
  Footprints,
  Laptop,
  LoaderCircle,
  MemoryStick,
  RefreshCw,
  Server,
  Smartphone,
  Watch,
  Wifi,
} from 'lucide-react';
import {
  siAlpinelinux,
  siAndroid,
  siApple,
  siArchlinux,
  siClaude,
  siCentos,
  siDebian,
  siFedora,
  siGithub,
  siKalilinux,
  siLinux,
  siLinuxmint,
  siPopos,
  siRedhat,
  siUbuntu,
  siWearos,
  siZorin,
} from 'simple-icons/icons';
import type { SimpleIcon } from 'simple-icons';
import { useEffect, useMemo, useState, type ReactElement, type ReactNode } from 'react';
import agentOrbitUrl from '@/assets/agents/agent-orbit.svg';
import clawdBuildingUrl from '@/assets/agents/clawd-working-building.gif';
import clawdDebuggerUrl from '@/assets/agents/clawd-working-debugger.gif';
import clawdJugglingUrl from '@/assets/agents/clawd-working-juggling.gif';
import clawdSweepingUrl from '@/assets/agents/clawd-working-sweeping.gif';
import clawdThinkingUrl from '@/assets/agents/clawd-working-thinking.gif';
import clawdTypingUrl from '@/assets/agents/clawd-working-typing.gif';
import codexOrbitUrl from '@/assets/agents/codex-orbit.svg';
import codexRibbonsUrl from '@/assets/agents/codex-ribbons.svg';
import codexSparksUrl from '@/assets/agents/codex-sparks.svg';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

type AgentMotionAsset = {
  src: string;
  durationMs: number;
};

const POLL_INTERVAL_MS = 10_000;
const CPU_CORES = 'system.cpu.logical.count';
const CPU_USAGE = 'system.cpu.utilization';
const MEMORY_USAGE = 'system.memory.usage';
const MEMORY_LIMIT = 'system.memory.limit';
const FILESYSTEM_USAGE = 'system.filesystem.usage';
const FILESYSTEM_LIMIT = 'system.filesystem.limit';
const FILESYSTEM_UTILIZATION = 'system.filesystem.utilization';
const CLAWD_MOTION_ASSETS: AgentMotionAsset[] = [
  { src: clawdTypingUrl, durationMs: 1_440 },
  { src: clawdBuildingUrl, durationMs: 960 },
  { src: clawdDebuggerUrl, durationMs: 2_880 },
  { src: clawdThinkingUrl, durationMs: 3_840 },
  { src: clawdSweepingUrl, durationMs: 1_440 },
  { src: clawdJugglingUrl, durationMs: 1_120 },
];
const CODEX_MOTION_ASSETS: AgentMotionAsset[] = [
  { src: codexOrbitUrl, durationMs: 4_000 },
  { src: codexRibbonsUrl, durationMs: 4_000 },
  { src: codexSparksUrl, durationMs: 4_000 },
];
const DEFAULT_MOTION_ASSETS: AgentMotionAsset[] = [{ src: agentOrbitUrl, durationMs: 4_000 }];

type PublicStatus = {
  server: DeviceStatus;
  mobile: MobileStatus | null;
  devices: StoredDeviceStatus[];
  agents: AgentStatus[];
  github: GitHubStatus;
  updated_at: string;
};

type StoredDeviceStatus = DeviceStatus & {
  received_at: string;
};

type DeviceStatus = {
  device_id: string;
  device_name?: string;
  device_model?: string;
  kind?: string;
  role?: string;
  state?: string;
  updated_at?: string;
  metrics?: MetricSample[];
  children?: DeviceStatus[];
};

type MetricSample = {
  name: string;
  unit?: string;
  value: number;
  attributes?: Record<string, string>;
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
  device_name?: string;
  device_model?: string;
  updated_at: string;
  received_at: string;
  phone?: {
    battery_percent?: number;
    charge_state?: ChargeState;
    network?: string;
  };
  watch?: {
    device_name?: string;
    device_model?: string;
    heart_rate?: number;
    steps?: number;
    battery_percent?: number;
    charge_state?: ChargeState;
    wrist_state?: 'unknown' | 'on_wrist' | 'off_wrist';
  };
};

type AgentStatus = {
  agent_id: string;
  device_id?: string;
  device_name?: string;
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
  const server = status?.server ?? null;
  const devices = status?.devices ?? [];
  const agents = status?.agents ?? [];
  const virtualMachines = devices.filter((device) => device.kind === 'virtual_machine' || device.role === 'vm');
  const personalDevices = devices.filter((device) => device.kind !== 'virtual_machine' && device.role !== 'vm');

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
        <div className="mx-auto flex min-h-screen w-full max-w-7xl flex-col gap-6 px-5 py-8">
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

          <StatusSection title="Infrastructure" icon={<Server className="size-4" />}>
            <DeviceCard device={server} title="Server" icon={<Server className="size-4" />} showChildren={false} />
            {virtualMachines.map((device) => (
              <DeviceCard key={device.device_id} device={device} title="VM" icon={<Box className="size-4" />} showChildren={false} />
            ))}
          </StatusSection>

          <StatusSection title="Devices" icon={<Watch className="size-4" />}>
            {personalDevices.map((device) => (
              <DeviceCard key={device.device_id} device={device} title={device.role === 'desktop' ? 'Mac' : 'Device'} icon={<Laptop className="size-4" />} />
            ))}
            <PhoneCard mobile={status?.mobile ?? null} />
            <WatchCard mobile={status?.mobile ?? null} github={status?.github ?? null} />
          </StatusSection>

          <StatusSection title="Agents" icon={<Bot className="size-4" />}>
            {agents.length === 0 ? <EmptyAgentCard /> : agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)}
          </StatusSection>

          <footer className="flex items-center gap-2 text-xs text-muted-foreground">
            <Clock className="size-3.5" />
            <span>{status ? `Updated ${formatTime(status.updated_at)}` : 'Waiting for first status'}</span>
          </footer>
        </div>
      </main>
    </TooltipProvider>
  );
}

function StatusSection({ title, icon, columns = 'md:grid-cols-2 xl:grid-cols-4', children }: {
  title: string;
  icon: ReactElement;
  columns?: string;
  children: ReactNode;
}) {
  return (
    <section className="grid gap-3">
      <div className="flex items-end justify-between gap-3">
        <h2 className="flex items-center gap-2 text-lg font-semibold tracking-tight">{icon}{title}</h2>
      </div>
      <div className={`grid gap-4 ${columns}`}>{children}</div>
    </section>
  );
}

function DeviceCard({ device, title, icon, showChildren = true }: { device: DeviceStatus | null; title: string; icon: ReactElement; showChildren?: boolean }) {
  const displayName = device?.device_name ?? title;
  const memory = memoryValues(device);
  const disk = diskValues(device);
  const cpuUsage = metricPercent(device, CPU_USAGE);
  const hasCpuCores = hasMetric(device, CPU_CORES);
  const hasMemory = memory.percent !== undefined;
  const hasDisk = disk.percent !== undefined;
  const showCpuBadge = hasCpuCores && cpuUsage === undefined;
  const hasAnyMetric = hasCpuCores || cpuUsage !== undefined || hasMemory || hasDisk;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">{deviceIcon(device, icon)}{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={device?.updated_at} />
          <StatusBadge state={device?.state} />
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <DeviceModel model={device?.device_model} />
        {showCpuBadge && (
          <MetricBadges>
            <MetricBadge icon={<Cpu />} value={cpuCoreText(device)} title="CPU cores" variant="secondary" />
          </MetricBadges>
        )}
        {cpuUsage !== undefined && <ProgressMetric label="CPU" value={cpuUsage} valueText={cpuText(device)} />}
        {hasMemory && <ProgressMetric label="Mem" value={memory.percent} valueText={memory.text} />}
        {hasDisk && <ProgressMetric label="Disk" value={disk.percent} valueText={disk.text} />}
        {!hasAnyMetric && <CardDescription>No metrics yet</CardDescription>}
        {showChildren && <ChildDevices devices={device?.children ?? []} />}
      </CardContent>
    </Card>
  );
}

function PhoneCard({ mobile }: { mobile: MobileStatus | null }) {
  const phone = mobile?.phone;
  const displayName = mobile?.device_name ?? 'Phone';
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><BrandIcon icon={siAndroid} />{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={mobile?.received_at} />
          {mobile ? <Badge><CheckCircle2 /></Badge> : <Badge variant="outline">—</Badge>}
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        <DeviceModel model={mobile?.device_model} />
        <MetricBadges>
          <MetricBadge icon={<Battery />} value={formatBattery(phone?.battery_percent)} title="Battery" />
          {phone?.charge_state === 'charging' && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          <MetricBadge icon={<Wifi />} value={phone?.network ?? '—'} title="Network" variant="secondary" />
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

function WatchCard({ mobile, github }: { mobile: MobileStatus | null; github: GitHubStatus | null }) {
  const watch = mobile?.watch;
  const offWrist = watch?.wrist_state === 'off_wrist';
  const displayName = watch?.device_name ?? 'Watch';
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><BrandIcon icon={siWearos} />{displayName}</CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={mobile?.received_at} />
          <GitHubStatusBadge github={github} />
          <WatchStatusBadge watch={watch} offWrist={offWrist} />
        </CardAction>
      </CardHeader>
      <CardContent className="space-y-4">
        <DeviceModel model={watch?.device_model} />
        <MetricBadges>
          {!offWrist && <MetricBadge icon={<HeartPulse />} value={watch?.heart_rate ? `${watch.heart_rate}` : '—'} title="Heart rate" />}
          <MetricBadge icon={<Footprints />} value={watch?.steps?.toLocaleString() ?? '—'} title="Steps" variant="secondary" />
          <MetricBadge icon={<Battery />} value={formatBattery(watch?.battery_percent)} title="Battery" />
          {watch?.charge_state === 'charging' && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          {offWrist && <MetricBadge icon={<CircleOff />} value="" title="Off wrist" variant="secondary" />}
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

function GitHubStatusBadge({ github }: { github: GitHubStatus | null }) {
  const state = github?.state;
  const title = githubStatusTitle(state);
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant={githubBadgeVariant(state)} title={title}>
          <BrandIcon icon={siGithub} />
          {githubIcon(state)}
        </Badge>
      </TooltipTrigger>
      <TooltipContent>{title}</TooltipContent>
    </Tooltip>
  );
}

function EmptyAgentCard() {
  return (
    <Card>
      <CardContent>
        <CardDescription>No active agents</CardDescription>
      </CardContent>
    </Card>
  );
}

function AgentCard({ agent }: { agent: AgentStatus }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">
          {agentIcon(agent.agent_id)}
          <span className="truncate">{agentName(agent.agent_id)}</span>
        </CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={agent.updated_at || agent.received_at} />
          <Badge variant={agentBadgeVariant(agent.state)} title={agent.state}>{agentStateIcon(agent.state)}</Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <AgentMotion agent={agent} />
        <AgentDeviceBadge agent={agent} />
        {agent.budget_remaining_percent !== undefined && (
          <div className="grid gap-2">
            <div className="flex items-center justify-between gap-3 text-xs text-muted-foreground">
              <span>Budget</span>
              <span>{agent.budget_remaining_percent}%</span>
            </div>
            <Progress value={agent.budget_remaining_percent} />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function AgentMotion({ agent }: { agent: AgentStatus }) {
  const assets = agentMotionAssets(agent.agent_id);
  const initialIndex = hashString(`${agent.device_id ?? ''}:${agent.agent_id}`) % assets.length;
  const [index, setIndex] = useState(initialIndex);
  const asset = assets[index % assets.length];
  const imageClassName = isClaudeAgent(agent.agent_id) ? 'agent-motion-image agent-motion-image-pixel' : 'agent-motion-image';

  useEffect(() => {
    setIndex(initialIndex);
  }, [initialIndex]);

  useEffect(() => {
    if (assets.length <= 1) return;
    const timeout = window.setTimeout(() => {
      setIndex((current) => (current + 1) % assets.length);
    }, asset.durationMs);
    return () => window.clearTimeout(timeout);
  }, [asset.durationMs, assets.length, index]);

  return (
    <div className="agent-motion">
      <img key={asset.src} className={imageClassName} src={asset.src} alt={`${agentName(agent.agent_id)} working`} />
    </div>
  );
}

function agentName(agentId: string): string {
  if (isClaudeAgent(agentId)) return 'Claude Code';
  if (agentId === 'codex') return 'Codex';
  if (agentId.startsWith('codex:')) return `Codex · ${agentId.slice('codex:'.length)}`;
  return agentId;
}

function agentKey(agent: AgentStatus): string {
  return `${agent.device_id ?? ''}/${agent.agent_id}`;
}

function agentIcon(agentId: string): ReactElement {
  if (isClaudeAgent(agentId)) return <BrandIcon icon={siClaude} />;
  if (agentId === 'codex' || agentId.startsWith('codex:')) return <CodexIcon />;
  return <Bot className="size-4" />;
}

function agentMotionAssets(agentId: string): AgentMotionAsset[] {
  if (isClaudeAgent(agentId)) return CLAWD_MOTION_ASSETS;
  if (agentId === 'codex' || agentId.startsWith('codex:')) return CODEX_MOTION_ASSETS;
  return DEFAULT_MOTION_ASSETS;
}

function isClaudeAgent(agentId: string): boolean {
  return agentId === 'claude-code';
}

function hashString(value: string): number {
  let hash = 0;
  for (const character of value) {
    hash = Math.imul(31, hash) + character.charCodeAt(0);
  }
  return Math.abs(hash);
}

function CodexIcon({ className = 'size-4' }: { className?: string }) {
  return (
    <svg aria-label="Codex" className={`${className} shrink-0`} fill="currentColor" fillRule="evenodd" role="img" viewBox="0 0 24 24">
      <title>Codex</title>
      <path
        clipRule="evenodd"
        d="M8.086.457a6.105 6.105 0 013.046-.415c1.333.153 2.521.72 3.564 1.7a.117.117 0 00.107.029c1.408-.346 2.762-.224 4.061.366l.063.03.154.076c1.357.703 2.33 1.77 2.918 3.198.278.679.418 1.388.421 2.126a5.655 5.655 0 01-.18 1.631.167.167 0 00.04.155 5.982 5.982 0 011.578 2.891c.385 1.901-.01 3.615-1.183 5.14l-.182.22a6.063 6.063 0 01-2.934 1.851.162.162 0 00-.108.102c-.255.736-.511 1.364-.987 1.992-1.199 1.582-2.962 2.462-4.948 2.451-1.583-.008-2.986-.587-4.21-1.736a.145.145 0 00-.14-.032c-.518.167-1.04.191-1.604.185a5.924 5.924 0 01-2.595-.622 6.058 6.058 0 01-2.146-1.781c-.203-.269-.404-.522-.551-.821a7.74 7.74 0 01-.495-1.283 6.11 6.11 0 01-.017-3.064.166.166 0 00.008-.074.115.115 0 00-.037-.064 5.958 5.958 0 01-1.38-2.202 5.196 5.196 0 01-.333-1.589 6.915 6.915 0 01.188-2.132c.45-1.484 1.309-2.648 2.577-3.493.282-.188.55-.334.802-.438.286-.12.573-.22.861-.304a.129.129 0 00.087-.087A6.016 6.016 0 015.635 2.31C6.315 1.464 7.132.846 8.086.457zm-.804 7.85a.848.848 0 00-1.473.842l1.694 2.965-1.688 2.848a.849.849 0 001.46.864l1.94-3.272a.849.849 0 00.007-.854l-1.94-3.393zm5.446 6.24a.849.849 0 000 1.695h4.848a.849.849 0 000-1.696h-4.848z"
      />
    </svg>
  );
}

function AgentDeviceBadge({ agent }: { agent: AgentStatus }) {
  const device = agent.device_name || agent.device_id;
  if (!device) return null;
  return (
    <Badge variant="outline" className="min-w-0 shrink" title={device}>
      <Laptop />
      <span className="truncate">{device}</span>
    </Badge>
  );
}

function agentStateIcon(state: AgentStatus['state']): ReactElement {
  if (state === 'failed') return <AlertTriangle />;
  if (state === 'running') return <CheckCircle2 />;
  return <CircleOff />;
}

function agentBadgeVariant(state: AgentStatus['state']): 'default' | 'secondary' | 'destructive' {
  if (state === 'failed') return 'destructive';
  if (state === 'running') return 'default';
  return 'secondary';
}

function deviceIcon(device: DeviceStatus | null, fallback: ReactElement): ReactElement {
  const icon = osIcon(device?.device_model ?? device?.device_name ?? '');
  return icon ? <BrandIcon icon={icon} /> : fallback;
}

function osIcon(value: string): SimpleIcon | null {
  const text = value.toLowerCase();
  if (text.includes('wear os')) return siWearos;
  if (text.includes('android')) return siAndroid;
  if (text.includes('macos') || text.includes('darwin')) return siApple;
  if (text.includes('kali')) return siKalilinux;
  if (text.includes('ubuntu')) return siUbuntu;
  if (text.includes('debian')) return siDebian;
  if (text.includes('fedora')) return siFedora;
  if (text.includes('arch')) return siArchlinux;
  if (text.includes('centos')) return siCentos;
  if (text.includes('red hat') || text.includes('rhel')) return siRedhat;
  if (text.includes('alpine')) return siAlpinelinux;
  if (text.includes('linux mint')) return siLinuxmint;
  if (text.includes('pop!_os') || text.includes('pop! os')) return siPopos;
  if (text.includes('zorin')) return siZorin;
  if (text.includes('linux')) return siLinux;
  return null;
}

function BrandIcon({ icon, className = 'size-4' }: { icon: SimpleIcon; className?: string }) {
  return (
    <svg aria-label={icon.title} className={`${className} shrink-0`} role="img" style={{ color: `#${icon.hex}` }} viewBox="0 0 24 24">
      <title>{icon.title}</title>
      <path d={icon.path} fill="currentColor" />
    </svg>
  );
}

function DeviceModel({ model }: { model?: string }) {
  if (!model) return null;
  return <p className="text-xs text-muted-foreground">{model}</p>;
}

function ChildDevices({ devices }: { devices: DeviceStatus[] }) {
  if (devices.length === 0) return null;
  return (
    <div className="grid gap-2">
      {devices.map((device) => (
        <div key={device.device_id} className="grid gap-1 rounded-md bg-muted/40 p-2 text-sm">
          <div className="flex items-center justify-between gap-2">
            <span className="flex min-w-0 items-center gap-2 truncate text-muted-foreground"><Box className="size-3.5" />{device.device_name || device.device_id}</span>
            <Badge variant={device.state === 'running' ? 'default' : 'secondary'}>{device.state ?? 'unknown'}</Badge>
          </div>
          <MetricBadges>
            {hasMetric(device, CPU_CORES) && <MetricBadge icon={<Cpu />} value={cpuCoreText(device)} title="CPU cores" variant="secondary" />}
            {memoryValues(device).percent !== undefined && <MetricBadge icon={<MemoryStick />} value={memoryValues(device).text} title="Memory" />}
            {diskValues(device).percent !== undefined && <MetricBadge icon={<HardDrive />} value={diskValues(device).text} title="Disk" />}
          </MetricBadges>
        </div>
      ))}
    </div>
  );
}

function StatusBadge({ state }: { state?: string }) {
  const online = state === 'online' || state === 'running';
  const offline = state === 'offline' || state === 'failed';
  return (
    <Badge variant={offline ? 'destructive' : online ? 'default' : 'outline'}>
      {offline ? <AlertTriangle /> : online ? <CheckCircle2 /> : '—'}
    </Badge>
  );
}

function WatchStatusBadge({ watch, offWrist }: { watch?: MobileStatus['watch']; offWrist: boolean }) {
  if (!watch) return <Badge variant="outline">—</Badge>;
  return offWrist ? <Badge variant="secondary"><CircleOff /></Badge> : <Badge><CheckCircle2 /></Badge>;
}

function ProgressMetric({ label, value, valueText }: { label: string; value: number | null | undefined; valueText?: string }) {
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

function InlineTime({ value }: { value?: string }) {
  return <span className="text-xs text-muted-foreground">{value ? formatTime(value) : '—'}</span>;
}

function githubIcon(state: GitHubStatus['state'] | undefined): ReactElement {
  if (state === 'error') return <AlertTriangle />;
  if (state === 'pending') return <LoaderCircle className="animate-spin" />;
  if (state === 'ok') return <CheckCircle2 />;
  return <CircleOff />;
}

function githubBadgeVariant(state: GitHubStatus['state'] | undefined): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (state === 'error') return 'destructive';
  if (state === 'ok') return 'default';
  if (state === 'pending') return 'secondary';
  return 'outline';
}

function githubStatusTitle(state: GitHubStatus['state'] | undefined): string {
  if (state === 'ok') return 'Connected';
  if (state === 'pending') return 'Connecting';
  if (state === 'error') return 'Sync failed';
  return 'Disconnected';
}

function cpuCoreText(device: DeviceStatus | null | undefined): string {
  const cores = metricValue(device, CPU_CORES);
  return cores === undefined ? '—' : `${Math.round(cores)}`;
}

function cpuText(device: DeviceStatus | null | undefined): string {
  const percent = formatPercent(metricPercent(device, CPU_USAGE));
  const cores = cpuCoreText(device);
  return cores === '—' ? percent : `${percent} · ${cores} cores`;
}

function memoryValues(device: DeviceStatus | null | undefined): { text: string; percent?: number } {
  const used = metricValue(device, MEMORY_USAGE);
  const total = metricValue(device, MEMORY_LIMIT);
  if (used === undefined || total === undefined || total <= 0) return { text: '—' };
  return { text: `${formatGigabytes(used)}/${formatGigabytes(total)}`, percent: used * 100 / total };
}

function diskValues(device: DeviceStatus | null | undefined): { text: string; percent?: number } {
  const directPercent = metricPercent(device, FILESYSTEM_UTILIZATION);
  if (directPercent !== undefined) return { text: formatPercent(directPercent), percent: directPercent };
  const used = metricValue(device, FILESYSTEM_USAGE);
  const total = metricValue(device, FILESYSTEM_LIMIT);
  if (used === undefined || total === undefined || total <= 0) return { text: '—' };
  const percent = used * 100 / total;
  return { text: formatPercent(percent), percent };
}

function metricPercent(device: DeviceStatus | null | undefined, name: string): number | undefined {
  const value = metricValue(device, name);
  return value === undefined ? undefined : value * 100;
}

function metricValue(device: DeviceStatus | null | undefined, name: string): number | undefined {
  return device?.metrics?.find((metric) => metric.name === name)?.value;
}

function hasMetric(device: DeviceStatus | null | undefined, name: string): boolean {
  return metricValue(device, name) !== undefined;
}

function formatBattery(value: number | undefined): string {
  return value === undefined ? '—' : `${value}%`;
}

function formatPercent(value: number | null | undefined): string {
  return value === null || value === undefined ? '—' : `${Math.round(value)}%`;
}

function formatGigabytes(value: number): string {
  return `${(value / 1024 / 1024 / 1024).toFixed(1)}GB`;
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
