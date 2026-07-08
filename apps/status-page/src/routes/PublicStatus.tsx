import { Bot, Box, Laptop, MonitorSmartphone, Server } from 'lucide-react';
import { useOutletContext } from 'react-router-dom';
import { DeviceRole } from '@/gen/realtime/me/v1/status_types_pb';
import { AgentCard, EmptyAgentCard, agentKey } from '@/components/AgentCard';
import type { ShellContext } from '@/components/AppShell';
import { DeviceCard } from '@/components/DeviceCard';
import { StatusSection } from '@/components/layout';
import { PhoneCard, WatchCard } from '@/components/MobileCards';
import { isVirtualMachine } from '@/lib/status';

export function PublicStatusApp() {
  const { status } = useOutletContext<ShellContext>();
  const server = status?.server ?? null;
  const devices = status?.devices ?? [];
  const agents = status?.agents ?? [];
  const virtualMachines = devices.filter(isVirtualMachine);
  const personalDevices = devices.filter((device) => !isVirtualMachine(device));

  return (
    <div className="grid gap-9">
      <section className="grid gap-3.5">
        <h2 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-muted-foreground">
          <span className="text-primary"><MonitorSmartphone className="size-4" /></span>
          Devices
        </h2>
        <div className="grid items-stretch gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <DeviceCard device={server} title="Server" icon={<Server className="size-4" />} showChildren={false} />
          {virtualMachines.map((device) => (
            <DeviceCard key={device.deviceUid} device={device} title="VM" icon={<Box className="size-4" />} showChildren={false} />
          ))}
          {personalDevices.map((device) => (
            <DeviceCard key={device.deviceUid} device={device} title={device.role === DeviceRole.DESKTOP ? 'Mac' : 'Device'} icon={<Laptop className="size-4" />} />
          ))}
          <PhoneCard mobile={status?.mobile ?? null} />
          <WatchCard mobile={status?.mobile ?? null} githubState={status?.github?.state} />
        </div>
      </section>

      <StatusSection title="Agents" icon={<Bot className="size-4" />}>
        {agents.length === 0 ? <EmptyAgentCard /> : agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)}
      </StatusSection>
    </div>
  );
}
