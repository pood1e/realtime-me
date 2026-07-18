import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import type {
  DeviceRecord,
  ExecutionRecord,
  TerminalSessionRecord,
  ThreadRecord,
  WorkspaceRecord,
} from "../../domain/records.js";
import type { RuntimeStatus } from "../../domain/runtime.js";
import {
  type Device,
  DeviceSchema,
  DeviceStatus,
} from "../../gen/super_manager/control/v1/device_pb.js";
import {
  type Execution,
  ExecutionSchema,
  ExecutionState,
} from "../../gen/super_manager/control/v1/execution_pb.js";
import {
  QuotaFreshness,
  type QuotaSnapshot,
  QuotaSnapshotSchema,
  type Runtime,
  RuntimeAvailability,
  RuntimeCapability,
  RuntimeKind,
  RuntimeSchema,
} from "../../gen/super_manager/control/v1/runtime_pb.js";
import {
  type TerminalSession,
  TerminalSessionSchema,
  TerminalSessionState,
} from "../../gen/super_manager/control/v1/terminal_pb.js";
import {
  type Thread,
  ThreadSchema,
  ThreadState,
} from "../../gen/super_manager/control/v1/thread_pb.js";
import {
  type Workspace,
  WorkspaceSchema,
} from "../../gen/super_manager/control/v1/workspace_pb.js";
import type { QuotaRecord } from "../../infrastructure/resource-store.js";

export function toWorkspace(record: WorkspaceRecord): Workspace {
  return create(WorkspaceSchema, {
    uid: record.uid,
    displayName: record.displayName,
    path: record.path,
    activeExecutionUid: record.activeExecutionUid ?? "",
    createTime: timestampFromDate(record.createTime),
  });
}

export function toThread(record: ThreadRecord): Thread {
  return create(ThreadSchema, {
    uid: record.uid,
    workspaceUid: record.workspaceUid,
    runtimeUid: record.runtimeUid,
    displayName: record.displayName,
    state: THREAD_STATES[record.state],
    createTime: timestampFromDate(record.createTime),
    updateTime: timestampFromDate(record.updateTime),
  });
}

export function toExecution(record: ExecutionRecord): Execution {
  return create(ExecutionSchema, {
    uid: record.uid,
    threadUid: record.threadUid,
    runId: record.runId,
    state: EXECUTION_STATES[record.state],
    startTime: timestampFromDate(record.startTime),
    ...(record.endTime ? { endTime: timestampFromDate(record.endTime) } : {}),
  });
}

export function toTerminalSession(record: TerminalSessionRecord): TerminalSession {
  return create(TerminalSessionSchema, {
    uid: record.uid,
    workspaceUid: record.workspaceUid,
    displayName: record.displayName,
    cwd: record.cwd,
    state: record.state === "RUNNING" ? TerminalSessionState.RUNNING : TerminalSessionState.CLOSED,
    createTime: timestampFromDate(record.createTime),
  });
}

export function toDevice(record: DeviceRecord): Device {
  return create(DeviceSchema, {
    uid: record.uid,
    displayName: record.displayName,
    status: record.status === "ACTIVE" ? DeviceStatus.ACTIVE : DeviceStatus.REVOKED,
    certificateSerial: record.certificateSerial,
    createTime: timestampFromDate(record.createTime),
    expireTime: timestampFromDate(record.expireTime),
  });
}

export function toRuntime(status: RuntimeStatus): Runtime {
  return create(RuntimeSchema, {
    uid: status.uid,
    kind: status.kind === "codex" ? RuntimeKind.CODEX : RuntimeKind.CLAUDE_CODE,
    displayName: status.displayName,
    version: status.version,
    availability: RUNTIME_AVAILABILITY[status.availability],
    capabilities: runtimeCapabilities(status),
    diagnostic: status.diagnostic.slice(0, 512),
    updateTime: timestampFromDate(status.updateTime),
  });
}

export function toQuota(record: QuotaRecord): QuotaSnapshot {
  return create(QuotaSnapshotSchema, {
    runtimeUid: record.runtimeUid,
    freshness: QUOTA_FRESHNESS[record.freshness],
    ...(record.usedRatio === null ? {} : { usedRatio: record.usedRatio }),
    ...(record.resetTime ? { resetTime: timestampFromDate(record.resetTime) } : {}),
    observeTime: timestampFromDate(record.observeTime),
    source: record.source,
  });
}

const THREAD_STATES = {
  IDLE: ThreadState.IDLE,
  RUNNING: ThreadState.RUNNING,
  INPUT_REQUIRED: ThreadState.INPUT_REQUIRED,
  LOST: ThreadState.LOST,
} as const;

const EXECUTION_STATES = {
  RUNNING: ExecutionState.RUNNING,
  INPUT_REQUIRED: ExecutionState.INPUT_REQUIRED,
  SUCCEEDED: ExecutionState.SUCCEEDED,
  FAILED: ExecutionState.FAILED,
  CANCELED: ExecutionState.CANCELED,
  LOST: ExecutionState.LOST,
} as const;

const RUNTIME_AVAILABILITY = {
  AVAILABLE: RuntimeAvailability.AVAILABLE,
  NOT_INSTALLED: RuntimeAvailability.NOT_INSTALLED,
  NOT_AUTHENTICATED: RuntimeAvailability.NOT_AUTHENTICATED,
  INCOMPATIBLE: RuntimeAvailability.INCOMPATIBLE,
  UNHEALTHY: RuntimeAvailability.UNHEALTHY,
} as const;

const QUOTA_FRESHNESS = {
  FRESH: QuotaFreshness.FRESH,
  STALE: QuotaFreshness.STALE,
  UNAVAILABLE: QuotaFreshness.UNAVAILABLE,
} as const;

function runtimeCapabilities(status: RuntimeStatus): RuntimeCapability[] {
  const capabilities = [
    RuntimeCapability.TEXT_STREAMING,
    RuntimeCapability.TOOL_STREAMING,
    RuntimeCapability.STRUCTURED_QUESTIONS,
    RuntimeCapability.CANCEL,
    RuntimeCapability.QUOTA,
  ];
  if (status.kind === "codex") {
    capabilities.push(RuntimeCapability.STEER, RuntimeCapability.REASONING_SUMMARIES);
  }
  return capabilities;
}
