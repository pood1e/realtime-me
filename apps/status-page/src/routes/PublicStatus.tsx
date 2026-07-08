import { Bot, Box, Laptop, Server, Smartphone } from 'lucide-react';
import { useOutletContext } from 'react-router-dom';
import { DeviceRole } from '@/gen/realtime/me/v1/status_types_pb';
import { AgentCard, EmptyAgentCard, agentKey } from '@/components/AgentCard';
import type { ShellContext } from '@/components/AppShell';
import { DeviceCard } from '@/components/DeviceCard';
import { EmptyCard } from '@/components/layout';
import { PhoneCard, WatchCard } from '@/components/MobileCards';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { isVirtualMachine } from '@/lib/status';

const GRID = 'grid items-start gap-4 sm:grid-cols-2 lg:grid-cols-3';

export function PublicStatusApp() {
  const { status } = useOutletContext<ShellContext>();
  const server = status?.server ?? null;
  const devices = status?.devices ?? [];
  const agents = status?.agents ?? [];
  const virtualMachines = devices.filter(isVirtualMachine);
  const personalDevices = devices.filter((device) => !isVirtualMachine(device));

  return (
    <Tabs defaultValue="infrastructure" className="gap-6">
      <TabsList variant="line" className="h-auto flex-wrap gap-4">
        <TabsTrigger value="infrastructure"><Server />Infrastructure</TabsTrigger>
        <TabsTrigger value="devices"><Laptop />Devices</TabsTrigger>
        <TabsTrigger value="mobile"><Smartphone />Mobile</TabsTrigger>
        <TabsTrigger value="agents"><Bot />Agents</TabsTrigger>
      </TabsList>

      <TabsContent value="infrastructure" className={GRID}>
        <DeviceCard device={server} title="Server" icon={<Server className="size-4" />} showChildren={false} />
        {virtualMachines.map((device) => (
          <DeviceCard key={device.deviceUid} device={device} title="VM" icon={<Box className="size-4" />} showChildren={false} />
        ))}
      </TabsContent>

      <TabsContent value="devices" className={GRID}>
        {personalDevices.length === 0 ? (
          <EmptyCard text="No devices" />
        ) : (
          personalDevices.map((device) => (
            <DeviceCard key={device.deviceUid} device={device} title={device.role === DeviceRole.DESKTOP ? 'Mac' : 'Device'} icon={<Laptop className="size-4" />} />
          ))
        )}
      </TabsContent>

      <TabsContent value="mobile" className="grid items-start gap-4 sm:grid-cols-2">
        <PhoneCard mobile={status?.mobile ?? null} />
        <WatchCard mobile={status?.mobile ?? null} githubState={status?.github?.state} />
      </TabsContent>

      <TabsContent value="agents" className={GRID}>
        {agents.length === 0 ? <EmptyAgentCard /> : agents.map((agent) => <AgentCard key={agentKey(agent)} agent={agent} />)}
      </TabsContent>
    </Tabs>
  );
}
