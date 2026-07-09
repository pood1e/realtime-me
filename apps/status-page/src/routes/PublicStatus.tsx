import { Box, Laptop, Server } from 'lucide-react';
import { useOutletContext } from 'react-router-dom';
import type { Agent } from '@/gen/realtime/me/v1/status_pb';
import { AgentState, DeviceRole } from '@/gen/realtime/me/v1/status_types_pb';
import type { ShellContext } from '@/components/AppShell';
import { DeviceCard } from '@/components/DeviceCard';
import { ErrorCard, SkeletonCard } from '@/components/layout';
import { PhoneCard, WatchCard } from '@/components/MobileCards';
import { isVirtualMachine } from '@/lib/status';

const SKELETON_CARDS = 5;

export function PublicStatusApp() {
  const { status, statusFailed } = useOutletContext<ShellContext>();

  if (statusFailed && status == null) {
    return <ErrorCard text="Cannot reach the status API. This says nothing about the devices themselves." />;
  }

  return (
    <div className="grid items-stretch gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {status == null ? <LoadingCards /> : <DeviceCards status={status} />}
    </div>
  );
}

function LoadingCards() {
  return Array.from({ length: SKELETON_CARDS }, (_, index) => <SkeletonCard key={index} />);
}

function DeviceCards({ status }: { status: NonNullable<ShellContext['status']> }) {
  const working = workingAgentsByDevice(status.agents);
  const virtualMachines = status.devices.filter(isVirtualMachine);
  const personalDevices = status.devices.filter((device) => !isVirtualMachine(device));

  return (
    <>
      <DeviceCard
        device={status.server ?? null}
        title="Server"
        icon={<Server className="size-4" />}
        agents={working.get(status.server?.deviceUid ?? '')}
      />
      {virtualMachines.map((device) => (
        <DeviceCard key={device.deviceUid} device={device} title="VM" icon={<Box className="size-4" />} agents={working.get(device.deviceUid)} />
      ))}
      {personalDevices.map((device) => (
        <DeviceCard
          key={device.deviceUid}
          device={device}
          title={device.role === DeviceRole.DESKTOP ? 'Mac' : 'Device'}
          icon={<Laptop className="size-4" />}
          agents={working.get(device.deviceUid)}
        />
      ))}
      <PhoneCard mobile={status.mobile ?? null} />
      <WatchCard mobile={status.mobile ?? null} githubState={status.github?.state} />
    </>
  );
}

// A working agent animates on the card of the machine it works on. Prometheus
// discovers an agent through its device's scrape target, and that device always
// registers a node exporter too, so every working agent has a card to land on.
function workingAgentsByDevice(agents: Agent[]): Map<string, Agent[]> {
  const byDevice = new Map<string, Agent[]>();
  for (const agent of agents) {
    if (agent.state !== AgentState.RUNNING) continue;
    const existing = byDevice.get(agent.deviceUid);
    if (existing) existing.push(agent);
    else byDevice.set(agent.deviceUid, [agent]);
  }
  return byDevice;
}
