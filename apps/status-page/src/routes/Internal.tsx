import { AlertTriangle, Bot, Box, Gauge, KeyRound, Laptop, LineChart as LineChartIcon, LogOut, RefreshCw, Server, ShieldCheck, Watch } from 'lucide-react';
import { useCallback, useState } from 'react';
import { siGithub } from 'simple-icons/icons';
import type { Agent, InternalStatus } from '@/gen/realtime/me/v1/status_pb';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { AgentCard, EmptyAgentCard, agentKey } from '@/components/AgentCard';
import { BrandIcon } from '@/components/brand';
import { InternalDeviceCard } from '@/components/DeviceCard';
import { GitHubDetails } from '@/components/GitHubDetails';
import { githubStatusTitle } from '@/components/github';
import { EmptyCard, LoadingCard, PageFooter, PageFrame, SiteLogo, StatusSection, SummaryCard } from '@/components/layout';
import { MetricsExplorer } from '@/components/MetricsExplorer';
import { PhoneCard, WatchCard } from '@/components/MobileCards';
import { usePolling } from '@/hooks/usePolling';
import { formatTime } from '@/lib/format';
import { deviceCounts, hostDevices, isVirtualMachine } from '@/lib/status';
import { POLL_INTERVAL_MS, authHeaders, isUnauthorized, statusClient } from '@/lib/transport';

const INTERNAL_TOKEN_KEY = 'realtime-me.internalToken';

export function InternalStatusApp() {
  const [accessToken, setAccessToken] = useState(() => localStorage.getItem(INTERNAL_TOKEN_KEY) ?? '');
  const [draftToken, setDraftToken] = useState('');
  const token = accessToken.trim();

  const fetchInternal = useCallback(async (signal: AbortSignal) => {
    return (await statusClient.getInternalStatus({}, { headers: authHeaders(token), signal })).status;
  }, [token]);

  const { data: status, error, refresh } = usePolling(fetchInternal, { intervalMs: POLL_INTERVAL_MS, enabled: token !== '' });
  const authFailed = isUnauthorized(error);
  const failed = error !== null && !authFailed;

  function authorize() {
    const next = draftToken.trim();
    if (!next) return;
    localStorage.setItem(INTERNAL_TOKEN_KEY, next);
    setAccessToken(next);
    setDraftToken('');
  }

  function clearToken() {
    localStorage.removeItem(INTERNAL_TOKEN_KEY);
    setAccessToken('');
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
                <Button variant="secondary" size="icon" aria-label="Refresh" title="Refresh" onClick={refresh}>
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
        <InternalDashboard status={status} token={token} />
      ) : (
        <LoadingCard />
      )}

      <PageFooter updatedAt={status?.updateTime} />
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

function InternalDashboard({ status, token }: { status: InternalStatus; token: string }) {
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
      <TabsContent value="metrics"><MetricsExplorer status={status} token={token} /></TabsContent>
      <TabsContent value="agents"><InternalAgents agents={status.agents} /></TabsContent>
      <TabsContent value="github">{status.github ? <GitHubDetails github={status.github} /> : <EmptyCard text="No GitHub status" />}</TabsContent>
    </Tabs>
  );
}

function InternalOverview({ status }: { status: InternalStatus }) {
  const { online, total } = deviceCounts(status);
  const watch = status.mobile?.watch;
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <SummaryCard icon={<Server />} title="Devices" value={`${online}/${total}`} detail="online" />
      <SummaryCard icon={<Bot />} title="Agents" value={`${status.agents.length}`} detail="working" />
      <SummaryCard icon={<BrandIcon icon={siGithub} />} title="GitHub" value={githubStatusTitle(status.github?.state)} detail={formatTime(status.github?.lastSuccessTime)} />
      <SummaryCard icon={<Watch />} title="Watch" value={watch ? 'live' : '—'} detail={status.mobile ? formatTime(status.mobile.updateTime) : 'waiting'} />
    </div>
  );
}

function InternalDevices({ status }: { status: InternalStatus }) {
  const virtualMachines = status.devices.filter(isVirtualMachine);
  const hosts = hostDevices(status).filter((device) => !isVirtualMachine(device));
  return (
    <div className="grid gap-6">
      <StatusSection title="Hosts" icon={<Server className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-3">
        {hosts.map((device) => (
          <InternalDeviceCard key={device.deviceUid} device={device} icon={<Laptop className="size-4" />} />
        ))}
      </StatusSection>
      <StatusSection title="Virtual machines" icon={<Box className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-3">
        {virtualMachines.length === 0 ? <EmptyCard text="No VM metrics" /> : virtualMachines.map((device) => <InternalDeviceCard key={device.deviceUid} device={device} icon={<Box className="size-4" />} />)}
      </StatusSection>
      <StatusSection title="Personal devices" icon={<Watch className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-3">
        <PhoneCard mobile={status.mobile ?? null} />
        <WatchCard mobile={status.mobile ?? null} githubState={status.github?.state} />
      </StatusSection>
    </div>
  );
}

function InternalAgents({ agents }: { agents: Agent[] }) {
  return (
    <StatusSection title="Agents" icon={<Bot className="size-4" />} columns="md:grid-cols-2 xl:grid-cols-4">
      {agents.length === 0 ? <EmptyAgentCard /> : agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)}
    </StatusSection>
  );
}
