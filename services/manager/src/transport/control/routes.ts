import { create } from "@bufbuild/protobuf";
import { Code, ConnectError, type ConnectRouter } from "@connectrpc/connect";
import type { ExecutionCoordinator } from "../../application/execution-coordinator.js";
import type { PairingAuthority } from "../../application/pairing-authority.js";
import type { RuntimeRegistry } from "../../application/runtime-registry.js";
import type { TerminalManager } from "../../application/terminal-manager.js";
import type { WorkspaceRegistry } from "../../application/workspace-registry.js";
import {
  DeleteDeviceResponseSchema,
  DeviceService,
  ListDevicesResponseSchema,
  PairDeviceResponseSchema,
  PairingService,
} from "../../gen/realtime/me/manager/control/v1/device_pb.js";
import {
  CancelExecutionResponseSchema,
  ExecutionService,
  GetExecutionResponseSchema,
  ListExecutionsResponseSchema,
  SteerExecutionResponseSchema,
} from "../../gen/realtime/me/manager/control/v1/execution_pb.js";
import {
  GetRuntimeQuotaResponseSchema,
  GetRuntimeResponseSchema,
  ListRuntimesResponseSchema,
  RuntimeService,
} from "../../gen/realtime/me/manager/control/v1/runtime_pb.js";
import {
  CreateTerminalSessionResponseSchema,
  DeleteTerminalSessionResponseSchema,
  GetTerminalSessionResponseSchema,
  ListTerminalSessionsResponseSchema,
  TerminalService,
} from "../../gen/realtime/me/manager/control/v1/terminal_pb.js";
import {
  CreateThreadResponseSchema,
  DeleteThreadResponseSchema,
  GetThreadResponseSchema,
  ListThreadsResponseSchema,
  ThreadService,
} from "../../gen/realtime/me/manager/control/v1/thread_pb.js";
import {
  CreateWorkspaceResponseSchema,
  DeleteWorkspaceResponseSchema,
  GetWorkspaceResponseSchema,
  ListWorkspacesResponseSchema,
  WorkspaceService,
} from "../../gen/realtime/me/manager/control/v1/workspace_pb.js";
import type { QuotaRecord, ResourceStore } from "../../infrastructure/resource-store.js";
import {
  toDevice,
  toExecution,
  toQuota,
  toRuntime,
  toTerminalSession,
  toThread,
  toWorkspace,
} from "./converters.js";

export interface ControlRouteDependencies {
  readonly store: ResourceStore;
  readonly workspaces: WorkspaceRegistry;
  readonly runtimes: RuntimeRegistry;
  readonly executions: ExecutionCoordinator;
  readonly terminals: TerminalManager;
  readonly pairing: PairingAuthority;
}

export function createControlRoutes(dependencies: ControlRouteDependencies) {
  return (router: ConnectRouter): void => {
    registerRuntimeRoutes(router, dependencies);
    registerWorkspaceRoutes(router, dependencies);
    registerThreadRoutes(router, dependencies);
    registerExecutionRoutes(router, dependencies);
    registerTerminalRoutes(router, dependencies);
    registerDeviceRoutes(router, dependencies);
  };
}

function registerRuntimeRoutes(
  router: ConnectRouter,
  { runtimes, store }: ControlRouteDependencies,
): void {
  router.service(RuntimeService, {
    getRuntime: (request) =>
      rpc(async () => {
        const runtime = runtimes.get(request.uid);
        if (!runtime) {
          throw notFound("Runtime");
        }
        return create(GetRuntimeResponseSchema, { runtime: toRuntime(runtime) });
      }),
    listRuntimes: () =>
      rpc(async () =>
        create(ListRuntimesResponseSchema, { runtimes: runtimes.list().map(toRuntime) }),
      ),
    getRuntimeQuota: (request) =>
      rpc(async () => {
        if (!runtimes.get(request.runtimeUid)) {
          throw notFound("Runtime");
        }
        const stored = store.getQuota(request.runtimeUid);
        const quota = stored
          ? withCurrentFreshness(stored)
          : {
              runtimeUid: request.runtimeUid,
              freshness: "UNAVAILABLE" as const,
              usedRatio: null,
              resetTime: null,
              observeTime: new Date(),
              source: "unavailable",
            };
        return create(GetRuntimeQuotaResponseSchema, { quotaSnapshot: toQuota(quota) });
      }),
  });
}

function registerWorkspaceRoutes(
  router: ConnectRouter,
  dependencies: ControlRouteDependencies,
): void {
  const { store, workspaces, terminals } = dependencies;
  router.service(WorkspaceService, {
    createWorkspace: (request) =>
      rpc(async () => {
        const workspace = await workspaces.register(request.displayName, request.path);
        return create(CreateWorkspaceResponseSchema, { workspace: toWorkspace(workspace) });
      }),
    getWorkspace: (request) =>
      rpc(async () => {
        const workspace = store.getWorkspace(request.uid);
        if (!workspace) {
          throw notFound("Workspace");
        }
        return create(GetWorkspaceResponseSchema, { workspace: toWorkspace(workspace) });
      }),
    listWorkspaces: (request) =>
      rpc(async () => {
        const page = store.listWorkspaces(request.pageSize, request.pageToken);
        return create(ListWorkspacesResponseSchema, {
          workspaces: page.items.map(toWorkspace),
          nextPageToken: page.nextPageToken,
        });
      }),
    deleteWorkspace: (request) =>
      rpc(async () => {
        const workspace = store.getWorkspace(request.uid);
        if (!workspace) {
          throw notFound("Workspace");
        }
        if (workspace.activeExecutionUid) {
          throw new ConnectError("Workspace has an active execution", Code.FailedPrecondition);
        }
        const sessions = await terminals.list(workspace.uid);
        if (sessions.some((session) => session.state === "RUNNING")) {
          throw new ConnectError("Workspace has a running terminal", Code.FailedPrecondition);
        }
        store.deleteWorkspace(workspace.uid);
        return create(DeleteWorkspaceResponseSchema);
      }),
  });
}

function registerThreadRoutes(router: ConnectRouter, dependencies: ControlRouteDependencies): void {
  const { store, runtimes } = dependencies;
  router.service(ThreadService, {
    createThread: (request) =>
      rpc(async () => {
        if (!store.getWorkspace(request.workspaceUid)) {
          throw notFound("Workspace");
        }
        runtimes.requireAvailable(request.runtimeUid);
        const thread = store.createThread(
          request.workspaceUid,
          request.runtimeUid,
          request.displayName,
        );
        return create(CreateThreadResponseSchema, { thread: toThread(thread) });
      }),
    getThread: (request) =>
      rpc(async () => {
        const thread = store.getThread(request.uid);
        if (!thread) {
          throw notFound("Thread");
        }
        return create(GetThreadResponseSchema, { thread: toThread(thread) });
      }),
    listThreads: (request) =>
      rpc(async () => {
        if (!store.getWorkspace(request.workspaceUid)) {
          throw notFound("Workspace");
        }
        const page = store.listThreads(request.workspaceUid, request.pageSize, request.pageToken);
        return create(ListThreadsResponseSchema, {
          threads: page.items.map(toThread),
          nextPageToken: page.nextPageToken,
        });
      }),
    deleteThread: (request) =>
      rpc(async () => {
        const thread = store.getThread(request.uid);
        if (!thread) {
          throw notFound("Thread");
        }
        if (store.getActiveExecutionForThread(thread.uid)) {
          throw new ConnectError("Thread has an active execution", Code.FailedPrecondition);
        }
        store.deleteThread(thread.uid);
        return create(DeleteThreadResponseSchema);
      }),
  });
}

function registerExecutionRoutes(
  router: ConnectRouter,
  { store, executions }: ControlRouteDependencies,
): void {
  router.service(ExecutionService, {
    getExecution: (request) =>
      rpc(async () => {
        const execution = store.getExecution(request.uid);
        if (!execution) {
          throw notFound("Execution");
        }
        return create(GetExecutionResponseSchema, { execution: toExecution(execution) });
      }),
    listExecutions: (request) =>
      rpc(async () => {
        if (!store.getThread(request.threadUid)) {
          throw notFound("Thread");
        }
        const page = store.listExecutions(request.threadUid, request.pageSize, request.pageToken);
        return create(ListExecutionsResponseSchema, {
          executions: page.items.map(toExecution),
          nextPageToken: page.nextPageToken,
        });
      }),
    cancelExecution: (request) =>
      rpc(async () => {
        if (!store.getExecution(request.uid)) {
          throw notFound("Execution");
        }
        return create(CancelExecutionResponseSchema, {
          execution: toExecution(await executions.cancel(request.uid)),
        });
      }),
    steerExecution: (request) =>
      rpc(async () => {
        if (!store.getExecution(request.uid)) {
          throw notFound("Execution");
        }
        return create(SteerExecutionResponseSchema, {
          execution: toExecution(await executions.steer(request.uid, request.instruction)),
        });
      }),
  });
}

function registerTerminalRoutes(
  router: ConnectRouter,
  dependencies: ControlRouteDependencies,
): void {
  const { store, terminals } = dependencies;
  router.service(TerminalService, {
    createTerminalSession: (request) =>
      rpc(async () => {
        const workspace = store.getWorkspace(request.workspaceUid);
        if (!workspace) {
          throw notFound("Workspace");
        }
        const session = await terminals.create({
          workspaceUid: workspace.uid,
          displayName: request.displayName,
          cwd: workspace.path,
          columns: request.columns,
          rows: request.rows,
        });
        return create(CreateTerminalSessionResponseSchema, {
          terminalSession: toTerminalSession(session),
        });
      }),
    getTerminalSession: (request) =>
      rpc(async () => {
        const session = await terminals.get(request.uid);
        if (!session) {
          throw notFound("Terminal session");
        }
        return create(GetTerminalSessionResponseSchema, {
          terminalSession: toTerminalSession(session),
        });
      }),
    listTerminalSessions: (request) =>
      rpc(async () => {
        if (!store.getWorkspace(request.workspaceUid)) {
          throw notFound("Workspace");
        }
        const sessions = await terminals.list(request.workspaceUid);
        return create(ListTerminalSessionsResponseSchema, {
          terminalSessions: sessions.map(toTerminalSession),
        });
      }),
    deleteTerminalSession: (request) =>
      rpc(async () => {
        if (!(await terminals.delete(request.uid))) {
          throw notFound("Terminal session");
        }
        return create(DeleteTerminalSessionResponseSchema);
      }),
  });
}

function registerDeviceRoutes(
  router: ConnectRouter,
  { store, pairing }: ControlRouteDependencies,
): void {
  router.service(PairingService, {
    pairDevice: (request) =>
      rpc(async () => {
        const credentials = await pairing.pair(request.pairingSecret, request.displayName);
        return create(PairDeviceResponseSchema, {
          device: toDevice(credentials.device),
          devicePkcs12: credentials.pkcs12,
          pkcs12Password: credentials.pkcs12Password,
          deviceToken: credentials.deviceToken,
          caCertificatePem: credentials.caCertificatePem,
        });
      }),
  });
  router.service(DeviceService, {
    listDevices: () =>
      rpc(async () =>
        create(ListDevicesResponseSchema, { devices: store.listDevices().map(toDevice) }),
      ),
    deleteDevice: (request) =>
      rpc(async () => {
        if (!store.revokeDevice(request.uid)) {
          throw notFound("Active device");
        }
        return create(DeleteDeviceResponseSchema);
      }),
  });
}

async function rpc<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation();
  } catch (error) {
    if (error instanceof ConnectError) {
      throw error;
    }
    throw mapError(error);
  }
}

function notFound(resource: string): ConnectError {
  return new ConnectError(`${resource} not found`, Code.NotFound);
}

function mapError(error: unknown): ConnectError {
  const message = error instanceof Error ? error.message : "Operation failed";
  if (/unique|already belongs|already has/i.test(message)) {
    return new ConnectError(message, Code.AlreadyExists);
  }
  if (/limit|disk space/i.test(message)) {
    return new ConnectError(message, Code.ResourceExhausted);
  }
  if (/not available|not active|waiting|live|failed precondition/i.test(message)) {
    return new ConnectError(message, Code.FailedPrecondition);
  }
  if (/invalid|must|outside|exceeds|out of range|empty|match/i.test(message)) {
    return new ConnectError(message, Code.InvalidArgument);
  }
  return new ConnectError(message.slice(0, 512), Code.Internal);
}

function withCurrentFreshness(quota: QuotaRecord): QuotaRecord {
  const stale = Date.now() - quota.observeTime.getTime() > 15 * 60 * 1000;
  return stale && quota.freshness === "FRESH" ? { ...quota, freshness: "STALE" } : quota;
}
