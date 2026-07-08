import { Bot, Box, Laptop, Server, Smartphone } from 'lucide-react';
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
    <div className="grid gap-8">
      <StatusSection title="Infrastructure" icon={<Server className="size-4" />}>
        <DeviceCard device={server} title="Server" icon={<Server className="size-4" />} showChildren={false} />
        {virtualMachines.map((device) => (
          <DeviceCard key={device.deviceUid} device={device} title="VM" icon={<Box className="size-4" />} showChildren={false} />
        ))}
      </StatusSection>

      {personalDevices.length > 0 && (
        <StatusSection title="Devices" icon={<Laptop className="size-4" />}>
          {personalDevices.map((device) => (
            <DeviceCard key={device.deviceUid} device={device} title={device.role === DeviceRole.DESKTOP ? 'Mac' : 'Device'} icon={<Laptop className="size-4" />} />
          ))}
        </StatusSection>
      )}

      <StatusSection title="Mobile" icon={<Smartphone className="size-4" />} columns="sm:grid-cols-2">
        <PhoneCard mobile={status?.mobile ?? null} />
        <WatchCard mobile={status?.mobile ?? null} githubState={status?.github?.state} />
      </StatusSection>

      <StatusSection title="Agents" icon={<Bot className="size-4" />}>
        {agents.length === 0 ? <EmptyAgentCard /> : agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)}
      </StatusSection>
    </div>
  );
}
