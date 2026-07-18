import type { Agent, InternalStatus } from "@realtime-me/status-contracts";
import { AgentCard, agentKey, EmptyAgentCard } from "@realtime-me/status-web/components/AgentCard";
import { BrandIcon } from "@realtime-me/status-web/components/brand";
import { InternalDeviceCard } from "@realtime-me/status-web/components/DeviceCard";
import { GitHubDetails } from "@realtime-me/status-web/components/GitHubDetails";
import { githubStatusTitle } from "@realtime-me/status-web/components/github";
import {
  EmptyCard,
  ErrorCard,
  LoadingCard,
  PageFooter,
  StatusSection,
  SummaryCard,
} from "@realtime-me/status-web/components/layout";
import { MetricsExplorer } from "@realtime-me/status-web/components/MetricsExplorer";
import { MobileDeviceCards } from "@realtime-me/status-web/components/MobileCards";
import { usePolling } from "@realtime-me/status-web/hooks/usePolling";
import { formatTime } from "@realtime-me/status-web/lib/format";
import { deviceCounts, hostDevices, isVirtualMachine } from "@realtime-me/status-web/lib/status";
import type { StatusApi } from "@realtime-me/status-web/lib/transport";
import { POLL_INTERVAL_MS } from "@realtime-me/status-web/lib/transport";
import { ConsolePage } from "@realtime-me/web-shell";
import { Badge } from "@realtime-me/web-ui/badge";
import { Button } from "@realtime-me/web-ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@realtime-me/web-ui/tabs";
import { Tooltip, TooltipContent, TooltipTrigger } from "@realtime-me/web-ui/tooltip";
import {
  AlertTriangle,
  Bot,
  Box,
  Gauge,
  Laptop,
  LineChart as LineChartIcon,
  RefreshCw,
  Server,
  ShieldCheck,
  Watch,
} from "lucide-react";
import { useCallback } from "react";
import { siGithub } from "simple-icons/icons";

export function InternalStatusApp({ api }: { api: StatusApi }) {
  const fetchInternal = useCallback(
    async (signal: AbortSignal) => (await api.status.getInternalStatus({}, { signal })).status,
    [api],
  );

  const {
    data: status,
    error,
    refresh,
  } = usePolling(fetchInternal, { intervalMs: POLL_INTERVAL_MS });
  const failed = error !== null;

  return (
    <ConsolePage
      title="Status"
      subtitle="Devices, agents, metrics, and GitHub presence"
      actions={
        <>
          <Badge variant={failed ? "destructive" : status ? "default" : "secondary"}>
            {failed ? <AlertTriangle /> : <ShieldCheck />}
            {failed ? "API offline" : status ? "Connected" : "Connecting"}
          </Badge>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="secondary"
                size="icon"
                aria-label="Refresh Status"
                title="Refresh Status"
                onClick={refresh}
              >
                <RefreshCw />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Refresh</TooltipContent>
          </Tooltip>
        </>
      }
    >
      {failed && !status ? (
        <ErrorCard text="Cannot reach the Status service." retry={refresh} />
      ) : status ? (
        <InternalDashboard status={status} api={api} />
      ) : (
        <LoadingCard />
      )}
      <PageFooter updatedAt={status?.updateTime} />
    </ConsolePage>
  );
}

function InternalDashboard({ status, api }: { status: InternalStatus; api: StatusApi }) {
  return (
    <Tabs defaultValue="overview" className="gap-5">
      <TabsList className="flex-wrap">
        <TabsTrigger value="overview">
          <Gauge />
          Overview
        </TabsTrigger>
        <TabsTrigger value="devices">
          <Server />
          Devices
        </TabsTrigger>
        <TabsTrigger value="metrics">
          <LineChartIcon />
          Metrics
        </TabsTrigger>
        <TabsTrigger value="agents">
          <Bot />
          Agents
        </TabsTrigger>
        <TabsTrigger value="github">
          <BrandIcon icon={siGithub} />
          GitHub
        </TabsTrigger>
      </TabsList>
      <TabsContent value="overview">
        <InternalOverview status={status} />
      </TabsContent>
      <TabsContent value="devices">
        <InternalDevices status={status} />
      </TabsContent>
      <TabsContent value="metrics">
        <MetricsExplorer status={status} api={api} />
      </TabsContent>
      <TabsContent value="agents">
        <InternalAgents agents={status.agents} />
      </TabsContent>
      <TabsContent value="github">
        {status.github ? (
          <GitHubDetails github={status.github} />
        ) : (
          <EmptyCard text="No GitHub status" />
        )}
      </TabsContent>
    </Tabs>
  );
}

function InternalOverview({ status }: { status: InternalStatus }) {
  const { online, total } = deviceCounts(status);
  const watchMobile = status.mobiles.find((mobile) => mobile.watch);
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <SummaryCard icon={<Server />} title="Devices" value={`${online}/${total}`} detail="online" />
      <SummaryCard
        icon={<Bot />}
        title="Agents"
        value={`${status.agents.length}`}
        detail="working"
      />
      <SummaryCard
        icon={<BrandIcon icon={siGithub} />}
        title="GitHub"
        value={githubStatusTitle(status.github?.state)}
        detail={formatTime(status.github?.lastSuccessTime)}
      />
      <SummaryCard
        icon={<Watch />}
        title="Watch"
        value={watchMobile ? "live" : "—"}
        detail={watchMobile ? formatTime(watchMobile.updateTime) : "waiting"}
      />
    </div>
  );
}

function InternalDevices({ status }: { status: InternalStatus }) {
  const virtualMachines = status.devices.filter(isVirtualMachine);
  const hosts = hostDevices(status).filter((device) => !isVirtualMachine(device));
  return (
    <div className="grid gap-6">
      <StatusSection
        title="Hosts"
        icon={<Server className="size-4" />}
        columns="md:grid-cols-2 xl:grid-cols-3"
      >
        {hosts.map((device) => (
          <InternalDeviceCard
            key={device.deviceUid}
            device={device}
            icon={<Laptop className="size-4" />}
          />
        ))}
      </StatusSection>
      <StatusSection
        title="Virtual machines"
        icon={<Box className="size-4" />}
        columns="md:grid-cols-2 xl:grid-cols-3"
      >
        {virtualMachines.length === 0 ? (
          <EmptyCard text="No VM metrics" />
        ) : (
          virtualMachines.map((device) => (
            <InternalDeviceCard
              key={device.deviceUid}
              device={device}
              icon={<Box className="size-4" />}
            />
          ))
        )}
      </StatusSection>
      <StatusSection
        title="Personal devices"
        icon={<Watch className="size-4" />}
        columns="md:grid-cols-2 xl:grid-cols-3"
      >
        <MobileDeviceCards mobiles={status.mobiles} githubState={status.github?.state} />
      </StatusSection>
    </div>
  );
}

function InternalAgents({ agents }: { agents: Agent[] }) {
  return (
    <StatusSection
      title="Agents"
      icon={<Bot className="size-4" />}
      columns="md:grid-cols-2 xl:grid-cols-4"
    >
      {agents.length === 0 ? (
        <EmptyAgentCard />
      ) : (
        agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)
      )}
    </StatusSection>
  );
}
