import { Bot, Box, Laptop, MonitorSmartphone, Server } from 'lucide-react';
import { useOutletContext } from 'react-router-dom';
import { DeviceRole } from '@/gen/realtime/me/v1/status_types_pb';
import { AgentCard, EmptyAgentCard, agentKey } from '@/components/AgentCard';
import type { ShellContext } from '@/components/AppShell';
import { DeviceCard } from '@/components/DeviceCard';
import { ErrorCard, SkeletonCard, StatusSection } from '@/components/layout';
import { PhoneCard, WatchCard } from '@/components/MobileCards';
import { isVirtualMachine } from '@/lib/status';

const SKELETON_CARDS = 5;

export function PublicStatusApp() {
  const { status, statusFailed } = useOutletContext<ShellContext>();

  if (statusFailed && status == null) {
    return <ErrorCard text="Cannot reach the status API. This says nothing about the devices themselves." />;
  }

  return (
    <div className="grid gap-9">
      <section className="grid gap-3.5">
        <h2 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-muted-foreground">
          <span className="text-primary"><MonitorSmartphone className="size-4" /></span>
          Devices
        </h2>
        <div className="grid items-stretch gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {status == null ? <LoadingCards /> : <DeviceCards status={status} />}
        </div>
      </section>

      <StatusSection title="Agents" icon={<Bot className="size-4" />}>
        {status == null ? <SkeletonCard /> : <AgentCards status={status} />}
      </StatusSection>
    </div>
  );
}

function LoadingCards() {
  return Array.from({ length: SKELETON_CARDS }, (_, index) => <SkeletonCard key={index} />);
}

function DeviceCards({ status }: { status: NonNullable<ShellContext['status']> }) {
  const devices = status.devices;
  const virtualMachines = devices.filter(isVirtualMachine);
  const personalDevices = devices.filter((device) => !isVirtualMachine(device));

  return (
    <>
      <DeviceCard device={status.server ?? null} title="Server" icon={<Server className="size-4" />} />
      {virtualMachines.map((device) => (
        <DeviceCard key={device.deviceUid} device={device} title="VM" icon={<Box className="size-4" />} />
      ))}
      {personalDevices.map((device) => (
        <DeviceCard key={device.deviceUid} device={device} title={device.role === DeviceRole.DESKTOP ? 'Mac' : 'Device'} icon={<Laptop className="size-4" />} />
      ))}
      <PhoneCard mobile={status.mobile ?? null} />
      <WatchCard mobile={status.mobile ?? null} githubState={status.github?.state} />
    </>
  );
}

function AgentCards({ status }: { status: NonNullable<ShellContext['status']> }) {
  if (status.agents.length === 0) return <EmptyAgentCard />;
  return status.agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />);
}
