import {
  Activity,
  AlertTriangle,
  Battery,
  BatteryCharging,
  Bot,
  Box,
  CheckCircle2,
  CircleOff,
  Clock,
  Code,
  Cpu,
  ExternalLink,
  Footprints,
  Gauge,
  Globe,
  HardDrive,
  Headphones,
  HeartPulse,
  KeyRound,
  Laptop,
  LineChart as LineChartIcon,
  LoaderCircle,
  Lock,
  LogOut,
  Mail,
  MapPin,
  MemoryStick,
  Music,
  RefreshCw,
  Server,
  ShieldCheck,
  Star,
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
import { lazy, Suspense, useCallback, useEffect, useMemo, useState, type ReactElement, type ReactNode } from 'react';
import agentOrbitUrl from '@/assets/agents/agent-orbit.svg';
import blueberryLogoUrl from '@/assets/blueberry.svg';
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
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import type { ProfileJson, ProfileLinkJson, ProfilePageJson, ProjectJson } from '@/gen/realtime/me/v1/profile_pb';

type AgentMotionAsset = {
  src: string;
  durationMs: number;
};

type PublicStatus = {
  server: DeviceStatus;
  mobile: MobileStatus | null;
  devices: StoredDeviceStatus[];
  agents: AgentStatus[];
  github: PublicGitHubStatus;
  updated_at: string;
};

type InternalStatus = Omit<PublicStatus, 'github'> & {
  github: InternalGitHubStatus;
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
  media?: {
    title: string;
    artist?: string;
  };
  accessories?: AccessoryStatus[];
  children?: DeviceStatus[];
};

type AccessoryStatus = {
  kind: string;
  name: string;
  model?: string;
  battery_percent?: number;
};

type MetricSample = {
  name: string;
  unit?: string;
  value: number;
  attributes?: Record<string, string>;
};

type PublicGitHubStatus = {
  enabled: boolean;
  state: GitHubSyncState;
  updated_at?: string;
  emoji?: string;
  message?: string;
};

type InternalGitHubStatus = {
  configured: boolean;
  state: GitHubSyncState;
  last_signature?: string;
  last_attempt_at?: string;
  last_success_at?: string;
  last_error_at?: string;
  last_error?: string;
  message?: string;
  emoji?: string;
};

type GitHubSyncState = 'disabled' | 'pending' | 'ok' | 'error';

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
    accessories?: AccessoryStatus[];
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
  unit: 'percent' | 'bytes' | 'count' | 'rate';
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

type ChartPoint = {
  time: number;
  value: number;
};

const POLL_INTERVAL_MS = 10_000;
const INTERNAL_TOKEN_KEY = 'realtime-me.internalToken';
const CPU_CORES = 'system.cpu.logical.count';
const CPU_USAGE = 'system.cpu.utilization';
const MEMORY_USAGE = 'system.memory.usage';
const MEMORY_LIMIT = 'system.memory.limit';
const FILESYSTEM_USAGE = 'system.filesystem.usage';
const FILESYSTEM_LIMIT = 'system.filesystem.limit';
const FILESYSTEM_UTILIZATION = 'system.filesystem.utilization';
const AGENT_MOTION_MIN_VISIBLE_MS = 10_000;
const CHART_RANGES: ChartRange[] = [
  { id: '15m', label: '15m', durationMs: 15 * 60_000, step: '15s' },
  { id: '1h', label: '1h', durationMs: 60 * 60_000, step: '30s' },
  { id: '6h', label: '6h', durationMs: 6 * 60 * 60_000, step: '2m' },
  { id: '24h', label: '24h', durationMs: 24 * 60 * 60_000, step: '5m' },
];
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
const StatusChart = lazy(() => import('@/components/StatusChart'));

export function App() {
  const path = window.location.pathname;
  if (path.startsWith('/internal')) return <InternalStatusApp />;
  if (path.startsWith('/about')) return <ProfileApp />;
  return <PublicStatusApp />;
}

function PublicStatusApp() {
  const apiBaseUrl = useMemo(statusApiBaseUrl, []);
  const [status, setStatus] = useState<PublicStatus | null>(null);
  const [failed, setFailed] = useState(false);
  const server = status?.server ?? null;
  const devices = status?.devices ?? [];
  const agents = status?.agents ?? [];
  const virtualMachines = devices.filter((device) => isVirtualMachine(device));
  const personalDevices = devices.filter((device) => !isVirtualMachine(device));

  const refresh = useCallback(async () => {
    const next = await fetchJSON<PublicStatus>(`${apiBaseUrl}/api/public-status`);
    setFailed(next === null);
    if (next) setStatus(next);
  }, [apiBaseUrl]);

  useEffect(() => {
    void refresh();
    const interval = window.setInterval(() => void refresh(), POLL_INTERVAL_MS);
    return () => window.clearInterval(interval);
  }, [refresh]);

  return (
    <PageFrame>
      <header className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <SiteLogo />
          <NavLinks />
        </div>
        <HeaderActions failed={failed} refresh={refresh} />
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

      <PageFooter updatedAt={status?.updated_at} />
    </PageFrame>
  );
}

function ProfileApp() {
  const apiBaseUrl = useMemo(statusApiBaseUrl, []);
  const [page, setPage] = useState<ProfilePageJson | null>(null);
  const [failed, setFailed] = useState(false);

  const refresh = useCallback(async () => {
    const next = await fetchJSON<ProfilePageJson>(`${apiBaseUrl}/api/profile`);
    setFailed(next === null);
    if (next) setPage(next);
  }, [apiBaseUrl]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const projects = page?.projects ?? [];

  return (
    <PageFrame>
      <header className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <SiteLogo />
          <NavLinks />
        </div>
        <HeaderActions failed={failed} refresh={refresh} />
      </header>

      <ProfileIntro profile={page?.profile} loaded={page !== null} />

      <StatusSection title="Projects" icon={<BrandIcon icon={siGithub} />} columns="md:grid-cols-2 xl:grid-cols-3">
        {page === null ? (
          <LoadingCard />
        ) : projects.length === 0 ? (
          <EmptyCard text="No projects yet" />
        ) : (
          projects.map((project) => <ProjectCard key={project.uid ?? project.displayName} project={project} />)
        )}
      </StatusSection>

      <PageFooter updatedAt={page?.updateTime} />
    </PageFrame>
  );
}

function ProfileIntro({ profile, loaded }: { profile?: ProfileJson; loaded: boolean }) {
  if (!loaded) return <LoadingCard />;
  if (!profile) return <EmptyCard text="Profile not configured" />;
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-4">
          {profile.avatarUrl && (
            <img src={profile.avatarUrl} alt={profile.displayName ?? 'avatar'} className="size-16 rounded-full border border-border" width={64} height={64} />
          )}
          <div className="grid gap-1">
            <CardTitle className="text-2xl tracking-tight">{profile.displayName || '—'}</CardTitle>
            {profile.headline && <CardDescription>{profile.headline}</CardDescription>}
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-4">
        {profile.bio && <p className="whitespace-pre-line text-sm text-muted-foreground">{profile.bio}</p>}
        <div className="flex flex-wrap items-center gap-2">
          {profile.location && <Badge variant="outline"><MapPin />{profile.location}</Badge>}
          {profile.githubLogin && (
            <a href={`https://github.com/${profile.githubLogin}`} target="_blank" rel="noreferrer" aria-label="GitHub profile">
              <Badge variant="secondary"><BrandIcon icon={siGithub} />{profile.githubLogin}</Badge>
            </a>
          )}
        </div>
        <ProfileLinks links={profile.links ?? []} />
      </CardContent>
    </Card>
  );
}

function ProfileLinks({ links }: { links: ProfileLinkJson[] }) {
  if (links.length === 0) return null;
  return (
    <div className="flex flex-wrap gap-2">
      {links.map((link) => (
        <Button key={`${link.platform ?? ''}:${link.uri ?? ''}`} asChild variant="outline" size="sm">
          <a href={link.uri} target="_blank" rel="noreferrer">
            {linkIcon(link)}
            {link.label || link.platform || link.uri}
          </a>
        </Button>
      ))}
    </div>
  );
}

function ProjectCard({ project }: { project: ProjectJson }) {
  const isPrivate = project.visibility === 'PROJECT_VISIBILITY_PRIVATE';
  const blurb = project.summary || project.description;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">
          <span className="truncate">{project.displayName || '—'}</span>
        </CardTitle>
        <CardAction>
          <Badge variant={isPrivate ? 'secondary' : 'outline'} title={isPrivate ? 'Private' : 'Public'}>
            {isPrivate ? <Lock /> : <BrandIcon icon={siGithub} />}
            {isPrivate ? 'Private' : 'Public'}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-3">
        {blurb && <p className="text-sm text-muted-foreground">{blurb}</p>}
        {project.topics && project.topics.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {project.topics.map((topic) => <Badge key={topic} variant="outline">{topic}</Badge>)}
          </div>
        )}
        <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
          {project.primaryLanguage && <span className="flex items-center gap-1"><Code className="size-3.5" />{project.primaryLanguage}</span>}
          {!!project.starCount && <span className="flex items-center gap-1"><Star className="size-3.5" />{project.starCount}</span>}
          {project.lastPushTime && <span className="flex items-center gap-1"><Clock className="size-3.5" />{formatDateTime(project.lastPushTime)}</span>}
        </div>
        <ProjectLinks project={project} />
      </CardContent>
    </Card>
  );
}

function ProjectLinks({ project }: { project: ProjectJson }) {
  if (!project.repositoryUrl && !project.homepageUrl) return null;
  return (
    <div className="flex flex-wrap gap-2">
      {project.repositoryUrl && (
        <Button asChild variant="outline" size="sm">
          <a href={project.repositoryUrl} target="_blank" rel="noreferrer"><BrandIcon icon={siGithub} />Repository</a>
        </Button>
      )}
      {project.homepageUrl && (
        <Button asChild variant="secondary" size="sm">
          <a href={project.homepageUrl} target="_blank" rel="noreferrer"><ExternalLink />Homepage</a>
        </Button>
      )}
    </div>
  );
}

function linkIcon(link: ProfileLinkJson): ReactElement {
  const platform = (link.platform ?? '').toLowerCase();
  if (platform === 'github') return <BrandIcon icon={siGithub} />;
  if (platform === 'email' || (link.uri ?? '').startsWith('mailto:')) return <Mail />;
  return <Globe />;
}

function NavLinks() {
  const onAbout = window.location.pathname.startsWith('/about');
  return (
    <nav className="flex items-center gap-1 text-sm">
      <a href="/" className={navLinkClass(!onAbout)}>Status</a>
      <a href="/about" className={navLinkClass(onAbout)}>About</a>
    </nav>
  );
}

function navLinkClass(active: boolean): string {
  return `rounded-md px-2.5 py-1 font-medium transition-colors ${active ? 'bg-muted text-foreground' : 'text-muted-foreground hover:text-foreground'}`;
}

function InternalStatusApp() {
  const apiBaseUrl = useMemo(statusApiBaseUrl, []);
  const [accessToken, setAccessToken] = useState(() => localStorage.getItem(INTERNAL_TOKEN_KEY) ?? '');
  const [draftToken, setDraftToken] = useState('');
  const [status, setStatus] = useState<InternalStatus | null>(null);
  const [failed, setFailed] = useState(false);
  const [authFailed, setAuthFailed] = useState(false);
  const token = accessToken.trim();

  const refresh = useCallback(async () => {
    if (!token) return;
    const response = await fetch(`${apiBaseUrl}/api/internal/status`, {
      cache: 'no-store',
      headers: { Authorization: `Bearer ${token}` },
    }).catch(() => null);
    if (!response) {
      setFailed(true);
      return;
    }
    if (response.status === 401) {
      setAuthFailed(true);
      setFailed(false);
      return;
    }
    if (!response.ok) {
      setFailed(true);
      return;
    }
    const next = await response.json() as InternalStatus;
    setStatus(next);
    setFailed(false);
    setAuthFailed(false);
  }, [apiBaseUrl, token]);

  useEffect(() => {
    void refresh();
    const interval = window.setInterval(() => void refresh(), POLL_INTERVAL_MS);
    return () => window.clearInterval(interval);
  }, [refresh]);

  function authorize() {
    const next = draftToken.trim();
    if (!next) return;
    localStorage.setItem(INTERNAL_TOKEN_KEY, next);
    setAccessToken(next);
    setDraftToken('');
    setAuthFailed(false);
  }

  function clearToken() {
    localStorage.removeItem(INTERNAL_TOKEN_KEY);
    setAccessToken('');
    setStatus(null);
    setAuthFailed(false);
  }

  return (
    <PageFrame maxWidth="max-w-[90rem]">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <SiteLogo />
        <div className="flex items-center gap-2">
          <Badge variant={authFailed || failed ? 'destructive' : status ? 'default' : 'secondary'}>
            {authFailed || failed ? <AlertTriangle /> : status ? <ShieldCheck /> : <KeyRound />}
            {authFailed ? 'Unauthorized' : failed ? 'API offline' : status ? 'Connected' : 'Locked'}
          </Badge>
          {token && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="secondary" size="icon" aria-label="Refresh" title="Refresh" onClick={() => void refresh()}>
                  <RefreshCw />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Refresh</TooltipContent>
            </Tooltip>
          )}
          {token && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="icon" aria-label="Lock" title="Lock" onClick={clearToken}>
                  <LogOut />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Lock</TooltipContent>
            </Tooltip>
          )}
        </div>
      </header>

      {!token || authFailed ? (
        <InternalAccessCard token={draftToken} setToken={setDraftToken} authorize={authorize} authFailed={authFailed} />
      ) : status ? (
        <InternalDashboard status={status} apiBaseUrl={apiBaseUrl} token={token} />
      ) : (
        <LoadingCard />
      )}

      <PageFooter updatedAt={status?.updated_at} />
    </PageFrame>
  );
}

function InternalAccessCard({ token, setToken, authorize, authFailed }: {
  token: string;
  setToken: (value: string) => void;
  authorize: () => void;
  authFailed: boolean;
}) {
  return (
    <Card className="mx-auto w-full max-w-md">
      <CardHeader>
        <CardTitle className="flex items-center gap-2"><KeyRound className="size-4" />Internal access</CardTitle>
        <CardDescription>{authFailed ? 'The saved token was rejected.' : 'Use the LAN status access token.'}</CardDescription>
      </CardHeader>
      <CardContent className="flex gap-2">
        <Input
          aria-label="Internal access token"
          autoComplete="current-password"
          placeholder="Access token"
          type="password"
          value={token}
          onChange={(event) => setToken(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter') authorize();
          }}
        />
        <Button aria-label="Unlock" title="Unlock" onClick={authorize}>
          <KeyRound />
        </Button>
      </CardContent>
    </Card>
  );
}

function InternalDashboard({ status, apiBaseUrl, token }: { status: InternalStatus; apiBaseUrl: string; token: string }) {
  return (
    <Tabs defaultValue="overview" className="gap-5">
      <TabsList className="flex-wrap">
        <TabsTrigger value="overview"><Gauge />Overview</TabsTrigger>
        <TabsTrigger value="devices"><Server />Devices</TabsTrigger>
        <TabsTrigger value="metrics"><LineChartIcon />Metrics</TabsTrigger>
        <TabsTrigger value="agents"><Bot />Agents</TabsTrigger>
        <TabsTrigger value="github"><BrandIcon icon={siGithub} />GitHub</TabsTrigger>
      </TabsList>
      <TabsContent value="overview"><InternalOverview status={status} /></TabsContent>
      <TabsContent value="devices"><InternalDevices status={status} /></TabsContent>
      <TabsContent value="metrics"><MetricsExplorer status={status} apiBaseUrl={apiBaseUrl} token={token} /></TabsContent>
      <TabsContent value="agents"><InternalAgents agents={status.agents} /></TabsContent>
      <TabsContent value="github"><GitHubDetails github={status.github} /></TabsContent>
    </Tabs>
  );
}

function InternalOverview({ status }: { status: InternalStatus }) {
  const resources = allDevices(status);
  const online = resources.filter((device) => device.state === 'online' || device.state === 'running').length;
  const watch = status.mobile?.watch;
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <SummaryCard icon={<Server />} title="Devices" value={`${online}/${resources.length}`} detail="online" />
      <SummaryCard icon={<Bot />} title="Agents" value={`${status.agents.length}`} detail="working" />
      <SummaryCard icon={<BrandIcon icon={siGithub} />} title="GitHub" value={githubStatusTitle(status.github.state)} detail={status.github.last_success_at ? formatTime(status.github.last_success_at) : '—'} />
      <SummaryCard icon={<Watch />} title="Watch" value={watch ? 'live' : '—'} detail={status.mobile?.received_at ? formatTime(status.mobile.received_at) : 'waiting'} />
    </div>
  );
}

function SummaryCard({ icon, title, value, detail }: { icon: ReactElement; title: string; value: string; detail: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-sm text-muted-foreground">{icon}{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-semibold tracking-tight">{value}</div>
        <div className="mt-1 text-xs text-muted-foreground">{detail}</div>
      </CardContent>
    </Card>
  );
}

function InternalDevices({ status }: { status: InternalStatus }) {
  const virtualMachines = status.devices.filter((device) => isVirtualMachine(device));
  const hosts = [status.server, ...status.devices.filter((device) => !isVirtualMachine(device))];
  return (
    <div className="grid gap-6">
      <StatusSection title="Hosts" icon={<Server className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-3">
        {hosts.map((device) => (
          <InternalDeviceCard key={device.device_id} device={device} icon={<Laptop className="size-4" />} />
        ))}
      </StatusSection>
      <StatusSection title="Virtual machines" icon={<Box className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-3">
        {virtualMachines.length === 0 ? <EmptyCard text="No VM metrics" /> : virtualMachines.map((device) => <InternalDeviceCard key={device.device_id} device={device} icon={<Box className="size-4" />} />)}
      </StatusSection>
      <StatusSection title="Personal devices" icon={<Watch className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-3">
        <PhoneCard mobile={status.mobile} />
        <WatchCard mobile={status.mobile} github={publicGitHub(status.github)} />
      </StatusSection>
    </div>
  );
}

function InternalDeviceCard({ device, icon }: { device: DeviceStatus; icon: ReactElement }) {
  const memory = memoryValues(device);
  const disk = diskValues(device);
  const cpuUsage = metricPercent(device, CPU_USAGE);
  const metricCount = device.metrics?.length ?? 0;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex min-w-0 items-center gap-2">{deviceIcon(device, icon)}<span className="truncate">{deviceDisplayName(device, 'Device')}</span></CardTitle>
        <CardAction className="flex items-center gap-2">
          <InlineTime value={device.updated_at} />
          <StatusBadge state={device.state} />
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <DeviceModel model={device.device_model} />
        <MetricBadges>
          {device.media?.title && <MediaBadge media={device.media} />}
          <AccessoryBadges accessories={device.accessories} />
          {hasMetric(device, CPU_CORES) && <MetricBadge icon={<Cpu />} value={cpuCoreText(device)} title="CPU cores" variant="secondary" />}
          <MetricBadge icon={<Activity />} value={`${metricCount}`} title="Metrics" variant="outline" />
        </MetricBadges>
        {cpuUsage !== undefined && <ProgressMetric label="CPU" value={cpuUsage} valueText={cpuText(device)} />}
        {memory.percent !== undefined && <ProgressMetric label="Mem" value={memory.percent} valueText={memory.text} />}
        {disk.percent !== undefined && <ProgressMetric label="Disk" value={disk.percent} valueText={disk.text} />}
        {metricCount === 0 && !device.media?.title && accessoryCount(device.accessories) === 0 && <CardDescription>No metrics yet</CardDescription>}
      </CardContent>
    </Card>
  );
}

function MetricsExplorer({ status, apiBaseUrl, token }: { status: InternalStatus; apiBaseUrl: string; token: string }) {
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
          <TimeSeriesCard key={chart.id} chart={chart} range={range} apiBaseUrl={apiBaseUrl} token={token} />
        ))}
      </div>
    </div>
  );
}

function TimeSeriesCard({ chart, range, apiBaseUrl, token }: { chart: ChartDefinition; range: ChartRange; apiBaseUrl: string; token: string }) {
  const { data, failed } = usePrometheusRange(apiBaseUrl, token, chart.query, range);
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

function usePrometheusRange(apiBaseUrl: string, token: string, query: string, range: ChartRange): { data: ChartPoint[]; failed: boolean } {
  const [data, setData] = useState<ChartPoint[]>([]);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    const end = Math.floor(Date.now() / 1000);
    const start = Math.floor((Date.now() - range.durationMs) / 1000);
    const params = new URLSearchParams({ query, start: `${start}`, end: `${end}`, step: range.step });
    fetch(`${apiBaseUrl}/api/internal/metrics/query_range?${params.toString()}`, {
      cache: 'no-store',
      headers: { Authorization: `Bearer ${token}` },
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
  }, [apiBaseUrl, query, range.durationMs, range.step, token]);

  return { data, failed };
}

function prometheusPoints(payload: PrometheusRangeResponse): ChartPoint[] {
  const values = payload.data?.result?.[0]?.values ?? [];
  return values
    .map(([time, value]) => ({ time: Number(time) * 1000, value: Number(value) }))
    .filter((point) => Number.isFinite(point.time) && Number.isFinite(point.value));
}

function InternalAgents({ agents }: { agents: AgentStatus[] }) {
  return (
    <StatusSection title="Agents" icon={<Bot className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-4">
      {agents.length === 0 ? <EmptyAgentCard /> : agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)}
    </StatusSection>
  );
}

function GitHubDetails({ github }: { github: InternalGitHubStatus }) {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><BrandIcon icon={siGithub} />GitHub status sync</CardTitle>
          <CardAction><Badge variant={githubBadgeVariant(github.state)}>{githubIcon(github.state)}</Badge></CardAction>
        </CardHeader>
        <CardContent className="grid gap-3 text-sm">
          <DetailRow label="State" value={githubStatusTitle(github.state)} />
          <DetailRow label="Last success" value={github.last_success_at ? formatDateTime(github.last_success_at) : '—'} />
          <DetailRow label="Last attempt" value={github.last_attempt_at ? formatDateTime(github.last_attempt_at) : '—'} />
          <DetailRow label="Emoji" value={github.emoji ?? '—'} />
          <DetailRow label="Message" value={github.message ?? '—'} />
        </CardContent>
      </Card>
      {github.last_error && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-destructive"><AlertTriangle className="size-4" />Last error</CardTitle>
            <CardAction><InlineTime value={github.last_error_at} /></CardAction>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">{github.last_error}</CardContent>
        </Card>
      )}
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border/50 pb-2 last:border-0 last:pb-0">
      <span className="text-muted-foreground">{label}</span>
      <span className="min-w-0 truncate text-right font-medium">{value}</span>
    </div>
  );
}

function SiteLogo() {
  return (
    <img
      src={blueberryLogoUrl}
      alt="pood1e"
      className="size-14 rounded-2xl drop-shadow-sm"
      width={56}
      height={56}
    />
  );
}

function PageFrame({ maxWidth = 'max-w-7xl', children }: { maxWidth?: string; children: ReactNode }) {
  return (
    <TooltipProvider>
      <main className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(20,184,166,0.18),transparent_35rem),radial-gradient(circle_at_top_right,rgba(59,130,246,0.18),transparent_30rem)]">
        <div className={`mx-auto flex min-h-screen w-full ${maxWidth} flex-col gap-6 px-5 py-8`}>{children}</div>
      </main>
    </TooltipProvider>
  );
}

function HeaderActions({ failed, refresh }: { failed: boolean; refresh: () => Promise<void> }) {
  return (
    <div className="flex items-center gap-2">
      <Badge variant={failed ? 'destructive' : 'default'} aria-label={failed ? 'API offline' : 'Online'} title={failed ? 'API offline' : 'Online'}>
        {failed ? <AlertTriangle /> : <CheckCircle2 />}
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
          <InlineTime value={device?.updated_at} />
          <StatusBadge state={device?.state} />
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4">
        <DeviceModel model={device?.device_model} />
        {showCpuBadge && (
          <MetricBadges>
            {device?.media?.title && <MediaBadge media={device.media} />}
            <AccessoryBadges accessories={device?.accessories} />
            <MetricBadge icon={<Cpu />} value={cpuCoreText(device)} title="CPU cores" variant="secondary" />
          </MetricBadges>
        )}
        {!showCpuBadge && (device?.media?.title || accessoryCount(device?.accessories) > 0) && (
          <MetricBadges>
            {device?.media?.title && <MediaBadge media={device.media} />}
            <AccessoryBadges accessories={device?.accessories} />
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
          <AccessoryBadges accessories={phone?.accessories} />
          <MetricBadge icon={<Battery />} value={formatBattery(phone?.battery_percent)} title="Battery" />
          {phone?.charge_state === 'charging' && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          <MetricBadge icon={<Wifi />} value={phone?.network ?? '—'} title="Network" variant="secondary" />
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

function WatchCard({ mobile, github }: { mobile: MobileStatus | null; github: PublicGitHubStatus | null }) {
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
          <MetricBadge icon={<Footprints />} value={formatSteps(watch)} title="Steps" variant="secondary" />
          <MetricBadge icon={<Battery />} value={formatBattery(watch?.battery_percent)} title="Battery" />
          {watch?.charge_state === 'charging' && <MetricBadge icon={<BatteryCharging />} value="" title="Charging" />}
          {offWrist && <MetricBadge icon={<CircleOff />} value="" title="Off wrist" variant="secondary" />}
        </MetricBadges>
      </CardContent>
    </Card>
  );
}

function GitHubStatusBadge({ github }: { github: PublicGitHubStatus | null }) {
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
  return <EmptyCard text="No active agents" />;
}

function EmptyCard({ text }: { text: string }) {
  return (
    <Card>
      <CardContent>
        <CardDescription>{text}</CardDescription>
      </CardContent>
    </Card>
  );
}

function LoadingCard() {
  return (
    <Card>
      <CardContent className="flex items-center gap-2 text-sm text-muted-foreground">
        <LoaderCircle className="size-4 animate-spin" />Loading
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
    }, agentMotionDelayMs(asset));
    return () => window.clearTimeout(timeout);
  }, [asset, assets.length, index]);

  return (
    <div className="agent-motion">
      <img key={asset.src} className={imageClassName} src={asset.src} alt={`${agentName(agent.agent_id)} working`} />
    </div>
  );
}

function chartDefinitions(status: InternalStatus): ChartDefinition[] {
  const definitions: ChartDefinition[] = [];
  const devices = [status.server, ...status.devices];
  for (const device of devices) {
    definitions.push(...hostChartDefinitions(device));
  }
  if (status.mobile) {
    definitions.push(...mobileChartDefinitions(status.mobile));
  }
  for (const agent of status.agents) {
    if (agent.budget_remaining_percent !== undefined) definitions.push(agentBudgetChart(agent));
  }
  return definitions;
}

function hostChartDefinitions(device: DeviceStatus): ChartDefinition[] {
  const identity = deviceDisplayName(device, 'Device');
  const queries = hostQueries(device);
  const definitions: ChartDefinition[] = [];
  if (queries.cpu && hasMetric(device, CPU_USAGE)) definitions.push({ id: `${device.device_id}:cpu`, title: `${identity} CPU`, query: queries.cpu, unit: 'percent', icon: <Cpu className="size-4" /> });
  if (queries.memory && hasMetric(device, MEMORY_USAGE)) definitions.push({ id: `${device.device_id}:mem`, title: `${identity} memory`, query: queries.memory, unit: 'bytes', icon: <MemoryStick className="size-4" /> });
  if (queries.disk && hasDiskMetric(device)) definitions.push({ id: `${device.device_id}:disk`, title: `${identity} disk`, query: queries.disk, unit: 'percent', icon: <HardDrive className="size-4" /> });
  definitions.push(...accessoryBatteryCharts(device.device_id, identity, device.accessories));
  return definitions;
}

function mobileChartDefinitions(mobile: MobileStatus): ChartDefinition[] {
  const definitions: ChartDefinition[] = [];
  if (mobile.phone?.battery_percent !== undefined) {
    definitions.push({ id: `${mobile.device_id}:phone-battery`, title: `${mobile.device_name || 'Phone'} battery`, query: `realtime_device_battery_level_ratio{device_id=${promLabel(mobile.device_id)},device_type="phone"} * 100`, unit: 'percent', icon: <Battery className="size-4" /> });
  }
  definitions.push(...accessoryBatteryCharts(mobile.device_id, mobile.device_name || 'Phone', mobile.phone?.accessories, 'phone'));
  const watch = mobile.watch;
  if (!watch) return definitions;
  const watchName = watch.device_name || 'Watch';
  if (watch.wrist_state !== 'off_wrist' && watch.heart_rate !== undefined) {
    definitions.push({ id: `${mobile.device_id}:watch-hr`, title: `${watchName} heart rate`, query: `realtime_watch_heart_rate_beats_per_minute{device_id=${promLabel(mobile.device_id)}}`, unit: 'rate', icon: <HeartPulse className="size-4" /> });
  }
  if (watch.steps !== undefined) {
    definitions.push({ id: `${mobile.device_id}:watch-steps`, title: `${watchName} steps`, query: `realtime_watch_steps{device_id=${promLabel(mobile.device_id)}}`, unit: 'count', icon: <Footprints className="size-4" /> });
  }
  if (watch.battery_percent !== undefined) {
    definitions.push({ id: `${mobile.device_id}:watch-battery`, title: `${watchName} battery`, query: `realtime_device_battery_level_ratio{device_id=${promLabel(mobile.device_id)},device_type="watch"} * 100`, unit: 'percent', icon: <Battery className="size-4" /> });
  }
  return definitions;
}

function accessoryBatteryCharts(deviceId: string, deviceName: string, accessories: AccessoryStatus[] | undefined, deviceType?: string): ChartDefinition[] {
  return (accessories ?? [])
    .filter((accessory) => accessory.name && accessory.battery_percent !== undefined)
    .map((accessory) => {
      const labels = [
        `device_id=${promLabel(deviceId)}`,
        `accessory_kind=${promLabel(accessory.kind)}`,
        `accessory_name=${promLabel(accessory.name)}`,
      ];
      if (deviceType) labels.push(`device_type=${promLabel(deviceType)}`);
      return {
        id: `${deviceId}:${accessory.kind}:${accessory.name}:battery`,
        title: `${deviceName} ${accessory.name}`,
        query: `realtime_device_accessory_battery_level_ratio{${labels.join(',')}} * 100`,
        unit: 'percent',
        icon: <Headphones className="size-4" />,
      };
    });
}

function agentBudgetChart(agent: AgentStatus): ChartDefinition {
  const labels = [`agent_id=${promLabel(agent.agent_id)}`];
  if (agent.device_id) labels.push(`device_id=${promLabel(agent.device_id)}`);
  return {
    id: `${agentKey(agent)}:budget`,
    title: `${agentName(agent.agent_id)} budget`,
    query: `realtime_agent_budget_remaining_ratio{${labels.join(',')}} * 100`,
    unit: 'percent',
    icon: agentIcon(agent.agent_id),
  };
}

function hostQueries(device: DeviceStatus): { cpu?: string; memory?: string; disk?: string } {
  if (device.role === 'server' || device.device_id === 'server') return nodeExporterQueries('node-exporter', 'server');
  if (isVirtualMachine(device)) return nodeExporterQueries('vm-node-exporter', device.device_id);
  const label = `device_id=${promLabel(device.device_id)}`;
  return {
    cpu: `realtime_host_cpu_usage_ratio{${label}} * 100`,
    memory: `realtime_host_memory_usage_bytes{${label}}`,
    disk: `realtime_host_filesystem_usage_ratio{${label}} * 100`,
  };
}

function nodeExporterQueries(job: string, instance: string): { cpu: string; memory: string; disk: string } {
  const base = `job=${promLabel(job)},instance=${promLabel(instance)}`;
  const diskBase = `${base},mountpoint="/",fstype!~"tmpfs|overlay|squashfs"`;
  return {
    cpu: `100 * (1 - avg(rate(node_cpu_seconds_total{${base},mode="idle"}[2m])))`,
    memory: `node_memory_MemTotal_bytes{${base}} - node_memory_MemAvailable_bytes{${base}}`,
    disk: `100 * (1 - node_filesystem_avail_bytes{${diskBase}} / node_filesystem_size_bytes{${diskBase}})`,
  };
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

function agentMotionDelayMs(asset: AgentMotionAsset): number {
  if (asset.durationMs <= 0) return AGENT_MOTION_MIN_VISIBLE_MS;
  return Math.ceil(AGENT_MOTION_MIN_VISIBLE_MS / asset.durationMs) * asset.durationMs;
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
  const device = displayLabel('', agent.device_name, agent.device_id);
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
            <span className="flex min-w-0 items-center gap-2 truncate text-muted-foreground"><Box className="size-3.5" />{deviceDisplayName(device, 'Device')}</span>
            <Badge variant={device.state === 'running' ? 'default' : 'secondary'}>{device.state ?? 'unknown'}</Badge>
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

function MediaBadge({ media, maxLength = 36 }: { media: NonNullable<DeviceStatus['media']>; maxLength?: number }) {
  const text = mediaText(media);
  return <MetricBadge icon={<Music />} value={compactText(text, maxLength)} title={`Playing: ${text}`} variant="secondary" />;
}

function AccessoryBadges({ accessories, maxLength = 30 }: { accessories?: AccessoryStatus[]; maxLength?: number }) {
  const connected = accessories?.filter((accessory) => accessory.name) ?? [];
  if (connected.length === 0) return null;
  return (
    <>
      {connected.map((accessory) => (
        <MetricBadge
          key={`${accessory.kind}:${accessory.name}:${accessory.model ?? ''}`}
          icon={<Headphones />}
          value={compactText(accessoryText(accessory), maxLength)}
          title={accessoryTitle(accessory)}
          variant="secondary"
        />
      ))}
    </>
  );
}

function mediaText(media: NonNullable<DeviceStatus['media']>): string {
  return media.artist ? `${media.title} · ${media.artist}` : media.title;
}

function accessoryText(accessory: AccessoryStatus): string {
  return accessory.battery_percent === undefined ? accessory.name : `${accessory.name} · ${accessory.battery_percent}%`;
}

function accessoryTitle(accessory: AccessoryStatus): string {
  return [accessory.name, accessory.model, accessory.battery_percent === undefined ? '' : `${accessory.battery_percent}%`]
    .filter(Boolean)
    .join(' · ');
}

function accessoryCount(accessories: AccessoryStatus[] | undefined): number {
  return accessories?.filter((accessory) => accessory.name).length ?? 0;
}

function InlineTime({ value }: { value?: string }) {
  return <span className="text-xs text-muted-foreground">{value ? formatTime(value) : '—'}</span>;
}

function PageFooter({ updatedAt }: { updatedAt?: string }) {
  return (
    <footer className="flex items-center gap-2 text-xs text-muted-foreground">
      <Clock className="size-3.5" />
      <span>{updatedAt ? `Updated ${formatTime(updatedAt)}` : 'Waiting for first status'}</span>
    </footer>
  );
}

function githubIcon(state: GitHubSyncState | undefined): ReactElement {
  if (state === 'error') return <AlertTriangle />;
  if (state === 'pending') return <LoaderCircle className="animate-spin" />;
  if (state === 'ok') return <CheckCircle2 />;
  return <CircleOff />;
}

function githubBadgeVariant(state: GitHubSyncState | undefined): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (state === 'error') return 'destructive';
  if (state === 'ok') return 'default';
  if (state === 'pending') return 'secondary';
  return 'outline';
}

function githubStatusTitle(state: GitHubSyncState | undefined): string {
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
  const used = metricValue(device, FILESYSTEM_USAGE);
  const total = metricValue(device, FILESYSTEM_LIMIT);
  if (used !== undefined && total !== undefined && total > 0) {
    const percent = used * 100 / total;
    return { text: `${formatGigabytes(used)}/${formatGigabytes(total)} · ${formatPercent(percent)}`, percent };
  }
  const directPercent = metricPercent(device, FILESYSTEM_UTILIZATION);
  if (directPercent !== undefined) return { text: formatPercent(directPercent), percent: directPercent };
  return { text: '—' };
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

function hasDiskMetric(device: DeviceStatus): boolean {
  return hasMetric(device, FILESYSTEM_UTILIZATION) || hasMetric(device, FILESYSTEM_USAGE);
}

function isVirtualMachine(device: DeviceStatus): boolean {
  return device.kind === 'virtual_machine' || device.role === 'vm';
}

function allDevices(status: InternalStatus): DeviceStatus[] {
  const devices: DeviceStatus[] = [status.server, ...status.devices];
  if (status.mobile) {
    devices.push({
      device_id: status.mobile.device_id,
      device_name: status.mobile.device_name,
      device_model: status.mobile.device_model,
      kind: 'phone',
      state: 'online',
      updated_at: status.mobile.received_at,
    });
  }
  if (status.mobile?.watch) {
    devices.push({
      device_id: `${status.mobile.device_id}:watch`,
      device_name: status.mobile.watch.device_name,
      device_model: status.mobile.watch.device_model,
      kind: 'watch',
      state: 'online',
      updated_at: status.mobile.received_at,
    });
  }
  return devices;
}

function publicGitHub(github: InternalGitHubStatus): PublicGitHubStatus {
  return {
    enabled: github.configured,
    state: github.state,
    updated_at: github.last_success_at,
    emoji: github.emoji,
    message: github.message,
  };
}

function fetchJSON<T>(url: string): Promise<T | null> {
  return fetch(url, { cache: 'no-store' })
    .then((response) => (response.ok ? response.json() as Promise<T> : null))
    .catch(() => null);
}

function formatBattery(value: number | undefined): string {
  return value === undefined ? '—' : `${value}%`;
}

function formatSteps(watch: MobileStatus['watch'] | undefined): string {
  if (!watch) return '—';
  return (watch.steps ?? 0).toLocaleString();
}

function deviceDisplayName(device: DeviceStatus | null | undefined, fallback: string): string {
  return displayLabel(fallback, device?.device_name, device?.device_id);
}

function displayLabel(fallback: string, ...values: Array<string | undefined>): string {
  return values.find((value) => value && !isPrivateIPv4(value)) ?? fallback;
}

function isPrivateIPv4(value: string): boolean {
  const octets = value.split('.').map((part) => Number(part));
  if (octets.length !== 4 || octets.some((part) => !Number.isInteger(part) || part < 0 || part > 255)) return false;
  const [first, second] = octets;
  return first === 10 ||
    first === 127 ||
    (first === 172 && second >= 16 && second <= 31) ||
    (first === 192 && second === 168) ||
    (first === 169 && second === 254);
}

function formatPercent(value: number | null | undefined): string {
  return value === null || value === undefined ? '—' : `${Math.round(value)}%`;
}

function formatGigabytes(value: number): string {
  return `${(value / 1024 / 1024 / 1024).toFixed(1)}GB`;
}

function compactText(value: string, maxLength: number): string {
  return value.length <= maxLength ? value : `${value.slice(0, maxLength - 1).trim()}…`;
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(new Date(value));
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  const now = new Date();
  const sameDay = date.toDateString() === now.toDateString();
  if (sameDay) return formatTime(value);
  return new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
}

function promLabel(value: string): string {
  return `"${value.replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/\n/g, '\\n')}"`;
}

function statusApiBaseUrl(): string {
  const configured = import.meta.env.VITE_STATUS_API_BASE_URL as string | undefined;
  if (configured?.trim()) return configured.replace(/\/+$/, '');

  const { protocol, hostname } = window.location;
  if (hostname === 'localhost' || hostname === '127.0.0.1') return 'http://localhost:18080';
  if (hostname.startsWith('status.')) return `${protocol}//api-${hostname}`;
  return window.location.origin;
}
