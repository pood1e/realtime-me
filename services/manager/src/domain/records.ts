export type ThreadState = "IDLE" | "RUNNING" | "INPUT_REQUIRED" | "LOST";
export type ExecutionState =
  | "RUNNING"
  | "INPUT_REQUIRED"
  | "SUCCEEDED"
  | "FAILED"
  | "CANCELED"
  | "LOST";
export type TerminalState = "RUNNING" | "CLOSED";
export type DeviceState = "ACTIVE" | "REVOKED";
export type InterruptState = "PENDING" | "RESOLVED" | "CANCELED" | "LOST";
export type RunState = "RUNNING" | "SUCCEEDED" | "INTERRUPTED" | "FAILED" | "CANCELED";

export interface WorkspaceRecord {
  readonly uid: string;
  readonly displayName: string;
  readonly path: string;
  readonly activeExecutionUid: string | null;
  readonly createTime: Date;
}

export interface ThreadRecord {
  readonly uid: string;
  readonly workspaceUid: string;
  readonly runtimeUid: string;
  readonly displayName: string;
  readonly state: ThreadState;
  readonly providerSessionId: string | null;
  readonly createTime: Date;
  readonly updateTime: Date;
}

export interface ExecutionRecord {
  readonly uid: string;
  readonly threadUid: string;
  readonly workspaceUid: string;
  readonly runId: string;
  readonly state: ExecutionState;
  readonly nativeTurnId: string | null;
  readonly startTime: Date;
  readonly endTime: Date | null;
}

export interface RunRecord {
  readonly runId: string;
  readonly executionUid: string;
  readonly threadUid: string;
  readonly parentRunId: string | null;
  readonly state: RunState;
  readonly createTime: Date;
  readonly endTime: Date | null;
}

export interface TerminalSessionRecord {
  readonly uid: string;
  readonly workspaceUid: string;
  readonly displayName: string;
  readonly cwd: string;
  readonly tmuxName: string;
  readonly columns: number;
  readonly rows: number;
  readonly state: TerminalState;
  readonly createTime: Date;
}

export interface DeviceRecord {
  readonly uid: string;
  readonly displayName: string;
  readonly status: DeviceState;
  readonly certificateSerial: string;
  readonly createTime: Date;
  readonly expireTime: Date;
}

export interface PendingInterruptRecord {
  readonly uid: string;
  readonly executionUid: string;
  readonly threadUid: string;
  readonly runId: string;
  readonly providerRequestId: string;
  readonly toolCallId: string;
  readonly input: unknown;
  readonly inputHash: string;
  readonly state: InterruptState;
  readonly createTime: Date;
}

export interface StoredEvent {
  readonly threadUid: string;
  readonly sequence: number;
  readonly runId: string;
  readonly event: unknown;
  readonly createTime: Date;
}
