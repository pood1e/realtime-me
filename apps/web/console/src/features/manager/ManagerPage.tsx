import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import {
  InlineError,
  LoadingIndicator,
  useDialog,
  useQuery,
  useToast,
} from "@realtime-me/library-web";
import {
  type Device,
  DeviceService,
  DeviceStatus,
  type Runtime,
  RuntimeAvailability,
  RuntimeKind,
  RuntimeService,
  type Workspace,
  WorkspaceService,
} from "@realtime-me/manager-contracts";
import { ConsolePage } from "@realtime-me/web-shell";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@realtime-me/web-ui";
import { Bot, FolderGit2, Laptop, RefreshCw, ShieldOff } from "lucide-react";
import { useState } from "react";
import { MANAGER_API_BASE } from "@/config";

const transport = createConnectTransport({
  baseUrl: MANAGER_API_BASE,
  useBinaryFormat: false,
  defaultTimeoutMs: 30_000,
  fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
});
const clients = {
  devices: createClient(DeviceService, transport),
  runtimes: createClient(RuntimeService, transport),
  workspaces: createClient(WorkspaceService, transport),
};

export function ManagerPage() {
  const devices = useQuery({
    queryKey: ["manager", "devices"],
    queryFn: ({ signal }) => clients.devices.listDevices({}, { signal }),
  });
  const runtimes = useQuery({
    queryKey: ["manager", "runtimes"],
    queryFn: ({ signal }) => clients.runtimes.listRuntimes({}, { signal }),
  });
  const workspaces = useQuery({
    queryKey: ["manager", "workspaces"],
    queryFn: ({ signal }) => clients.workspaces.listWorkspaces({ pageSize: 100 }, { signal }),
  });
  const [revoking, setRevoking] = useState("");
  const { confirm } = useDialog();
  const { showToast } = useToast();
  const loading = devices.isPending || runtimes.isPending || workspaces.isPending;
  const error = devices.error ?? runtimes.error ?? workspaces.error;
  const refresh = () => {
    void Promise.all([devices.refetch(), runtimes.refetch(), workspaces.refetch()]);
  };
  const revoke = async (device: Device) => {
    if (
      !(await confirm({
        title: "Revoke device",
        description: `${device.displayName} will immediately lose Manager access.`,
        confirmLabel: "Revoke",
        destructive: true,
      }))
    )
      return;
    setRevoking(device.uid);
    try {
      await clients.devices.deleteDevice({ uid: device.uid });
      await devices.refetch();
      showToast("Device access revoked");
    } catch (cause) {
      showToast(errorMessage(cause), "error");
    } finally {
      setRevoking("");
    }
  };

  return (
    <ConsolePage
      title="Manager"
      subtitle="Coding-agent runtimes, approved workspaces, and paired clients"
      actions={
        <Button variant="secondary" size="icon" aria-label="Refresh Manager" onClick={refresh}>
          <RefreshCw />
        </Button>
      }
    >
      {error ? <InlineError message={errorMessage(error)} onRetry={refresh} /> : null}
      {loading ? (
        <LoadingIndicator label="Loading Manager" />
      ) : (
        <div className="grid gap-6">
          <RuntimeSection runtimes={runtimes.data?.runtimes ?? []} />
          <WorkspaceSection workspaces={workspaces.data?.workspaces ?? []} />
          <DeviceSection
            devices={devices.data?.devices ?? []}
            revoking={revoking}
            onRevoke={revoke}
          />
        </div>
      )}
    </ConsolePage>
  );
}

function RuntimeSection({ runtimes }: { runtimes: readonly Runtime[] }) {
  return (
    <section>
      <SectionHeading icon={<Bot />} title="Agent runtimes" count={runtimes.length} />
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {runtimes.map((runtime) => (
          <Card key={runtime.uid} size="sm">
            <CardHeader>
              <CardTitle>{runtime.displayName || runtimeKind(runtime.kind)}</CardTitle>
              <CardDescription>{runtime.version || "Version unavailable"}</CardDescription>
            </CardHeader>
            <CardContent className="flex items-center justify-between gap-3">
              <AvailabilityBadge availability={runtime.availability} />
              <span className="text-xs text-muted-foreground">
                {runtime.capabilities.length} capabilities
              </span>
            </CardContent>
            {runtime.diagnostic ? (
              <CardContent className="text-xs leading-5 text-muted-foreground">
                {runtime.diagnostic}
              </CardContent>
            ) : null}
          </Card>
        ))}
        {runtimes.length === 0 ? <EmptyCard text="No coding-agent runtimes detected." /> : null}
      </div>
    </section>
  );
}

function WorkspaceSection({ workspaces }: { workspaces: readonly Workspace[] }) {
  return (
    <section>
      <SectionHeading icon={<FolderGit2 />} title="Approved workspaces" count={workspaces.length} />
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Path</TableHead>
                <TableHead>Lease</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {workspaces.map((workspace) => (
                <TableRow key={workspace.uid}>
                  <TableCell className="font-medium">{workspace.displayName}</TableCell>
                  <TableCell className="max-w-xl truncate font-mono text-xs text-muted-foreground">
                    {workspace.path}
                  </TableCell>
                  <TableCell>
                    <Badge variant={workspace.activeExecutionUid ? "default" : "secondary"}>
                      {workspace.activeExecutionUid ? "In use" : "Available"}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))}
              {workspaces.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={3} className="h-24 text-center text-muted-foreground">
                    No approved workspaces.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </section>
  );
}

function DeviceSection({
  devices,
  revoking,
  onRevoke,
}: {
  devices: readonly Device[];
  revoking: string;
  onRevoke: (device: Device) => Promise<void>;
}) {
  return (
    <section>
      <SectionHeading icon={<Laptop />} title="Paired clients" count={devices.length} />
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Device</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Certificate</TableHead>
                <TableHead className="text-right">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {devices.map((device) => (
                <TableRow key={device.uid}>
                  <TableCell className="font-medium">{device.displayName}</TableCell>
                  <TableCell>
                    <Badge
                      variant={device.status === DeviceStatus.ACTIVE ? "default" : "secondary"}
                    >
                      {device.status === DeviceStatus.ACTIVE ? "Active" : "Revoked"}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {device.certificateSerial || "—"}
                  </TableCell>
                  <TableCell className="text-right">
                    {device.status === DeviceStatus.ACTIVE ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={revoking === device.uid}
                        onClick={() => void onRevoke(device)}
                      >
                        <ShieldOff />
                        {revoking === device.uid ? "Revoking…" : "Revoke"}
                      </Button>
                    ) : null}
                  </TableCell>
                </TableRow>
              ))}
              {devices.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="h-24 text-center text-muted-foreground">
                    No paired clients.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </section>
  );
}

function SectionHeading({
  icon,
  title,
  count,
}: {
  icon: React.ReactNode;
  title: string;
  count: number;
}) {
  return (
    <div className="mb-3 flex items-center gap-2">
      <span className="text-primary [&>svg]:size-4">{icon}</span>
      <h2 className="font-heading text-lg font-medium">{title}</h2>
      <Badge variant="secondary">{count}</Badge>
    </div>
  );
}

function EmptyCard({ text }: { text: string }) {
  return (
    <Card size="sm" className="md:col-span-2 xl:col-span-3">
      <CardContent className="py-8 text-center text-muted-foreground">{text}</CardContent>
    </Card>
  );
}

function AvailabilityBadge({ availability }: { availability: RuntimeAvailability }) {
  const available = availability === RuntimeAvailability.AVAILABLE;
  return (
    <Badge variant={available ? "default" : "secondary"}>{availabilityLabel(availability)}</Badge>
  );
}

function runtimeKind(kind: RuntimeKind): string {
  if (kind === RuntimeKind.CODEX) return "Codex CLI";
  if (kind === RuntimeKind.CLAUDE_CODE) return "Claude Code";
  return "Unknown runtime";
}

function availabilityLabel(availability: RuntimeAvailability): string {
  const labels: Readonly<Record<RuntimeAvailability, string>> = {
    [RuntimeAvailability.UNSPECIFIED]: "Unknown",
    [RuntimeAvailability.AVAILABLE]: "Available",
    [RuntimeAvailability.NOT_INSTALLED]: "Not installed",
    [RuntimeAvailability.NOT_AUTHENTICATED]: "Sign-in required",
    [RuntimeAvailability.INCOMPATIBLE]: "Incompatible",
    [RuntimeAvailability.UNHEALTHY]: "Unhealthy",
  };
  return labels[availability];
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "Manager request failed";
}
