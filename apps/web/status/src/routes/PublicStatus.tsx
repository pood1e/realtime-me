import type { Agent } from "@realtime-me/status-contracts";
import { AgentState, DeviceRole } from "@realtime-me/status-contracts";
import { Box, Laptop, Server } from "lucide-react";
import { useOutletContext } from "react-router-dom";
import type { ShellContext } from "@/components/AppShell";
import { DeviceCard } from "@/components/DeviceCard";
import { ErrorCard, SkeletonCard } from "@/components/layout";
import { MobileDeviceCards } from "@/components/MobileCards";
import { WorkingNow } from "@/components/WorkingNow";
import { isVirtualMachine } from "@/lib/status";

const SKELETON_CARDS = 5;

export function PublicStatusApp() {
  const { status, statusFailed } = useOutletContext<ShellContext>();

  if (statusFailed && status == null) {
    return (
      <ErrorCard text="Cannot reach the status API. This says nothing about the devices themselves." />
    );
  }

  return (
    <div className="grid gap-4">
      {status != null && <WorkingNow agents={status.agents.filter(isWorking)} />}
      <div className="grid items-stretch gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {status == null ? <LoadingCards /> : <DeviceCards status={status} />}
      </div>
    </div>
  );
}

function LoadingCards() {
  return Array.from({ length: SKELETON_CARDS }, (_, index) => <SkeletonCard key={index} />);
}

function DeviceCards({ status }: { status: NonNullable<ShellContext["status"]> }) {
  const virtualMachines = status.devices.filter(isVirtualMachine);
  const personalDevices = status.devices.filter((device) => !isVirtualMachine(device));

  return (
    <>
      <DeviceCard
        device={status.server ?? null}
        title="Server"
        icon={<Server className="size-4" />}
      />
      {virtualMachines.map((device) => (
        <DeviceCard
          key={device.deviceUid}
          device={device}
          title="VM"
          icon={<Box className="size-4" />}
        />
      ))}
      {personalDevices.map((device) => (
        <DeviceCard
          key={device.deviceUid}
          device={device}
          title={device.role === DeviceRole.DESKTOP ? "Mac" : "Device"}
          icon={<Laptop className="size-4" />}
        />
      ))}
      <MobileDeviceCards mobiles={status.mobiles} githubState={status.github?.state} />
    </>
  );
}

function isWorking(agent: Agent): boolean {
  return agent.state === AgentState.RUNNING;
}
