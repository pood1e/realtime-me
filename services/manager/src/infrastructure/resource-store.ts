import { randomUUID } from "node:crypto";
import type {
  DeviceRecord,
  DeviceState,
  ExecutionRecord,
  ExecutionState,
  InterruptState,
  PendingInterruptRecord,
  RunRecord,
  RunState,
  TerminalSessionRecord,
  TerminalState,
  ThreadRecord,
  ThreadState,
  WorkspaceRecord,
} from "../domain/records.js";
import type { SqliteDatabase } from "./sqlite.js";

interface WorkspaceRow {
  uid: string;
  display_name: string;
  path: string;
  active_execution_uid: string | null;
  create_time_ms: number;
}

interface ThreadRow {
  uid: string;
  workspace_uid: string;
  runtime_uid: string;
  display_name: string;
  state: ThreadState;
  provider_session_id: string | null;
  create_time_ms: number;
  update_time_ms: number;
}

interface ExecutionRow {
  uid: string;
  thread_uid: string;
  workspace_uid: string;
  run_id: string;
  state: ExecutionState;
  native_turn_id: string | null;
  start_time_ms: number;
  end_time_ms: number | null;
}

interface RunRow {
  run_id: string;
  execution_uid: string;
  thread_uid: string;
  parent_run_id: string | null;
  state: RunState;
  create_time_ms: number;
  end_time_ms: number | null;
}

interface InterruptRow {
  uid: string;
  execution_uid: string;
  thread_uid: string;
  run_id: string;
  provider_request_id: string;
  tool_call_id: string;
  input_json: string;
  input_hash: string;
  state: InterruptState;
  create_time_ms: number;
}

interface TerminalRow {
  uid: string;
  workspace_uid: string;
  display_name: string;
  cwd: string;
  tmux_name: string;
  columns: number;
  rows: number;
  state: TerminalState;
  create_time_ms: number;
}

interface DeviceRow {
  uid: string;
  display_name: string;
  status: DeviceState;
  certificate_serial: string;
  create_time_ms: number;
  expire_time_ms: number;
}

export interface Page<T> {
  readonly items: readonly T[];
  readonly nextPageToken: string;
}

export interface QuotaRecord {
  readonly runtimeUid: string;
  readonly freshness: "FRESH" | "STALE" | "UNAVAILABLE";
  readonly usedRatio: number | null;
  readonly resetTime: Date | null;
  readonly observeTime: Date;
  readonly source: string;
}

export class ResourceStore {
  constructor(private readonly database: SqliteDatabase) {}

  createWorkspace(displayName: string, path: string): WorkspaceRecord {
    const uid = randomUUID();
    const createTime = new Date();
    this.database
      .prepare(
        "INSERT INTO workspaces(uid, display_name, path, create_time_ms) VALUES (?, ?, ?, ?)",
      )
      .run(uid, displayName, path, createTime.getTime());
    return { uid, displayName, path, activeExecutionUid: null, createTime };
  }

  getWorkspace(uid: string): WorkspaceRecord | null {
    const row = this.database
      .prepare(
        `SELECT w.*,
          (SELECT e.uid FROM executions e
           WHERE e.workspace_uid = w.uid AND e.state IN ('RUNNING', 'INPUT_REQUIRED')
           LIMIT 1) AS active_execution_uid
         FROM workspaces w WHERE w.uid = ?`,
      )
      .get(uid) as WorkspaceRow | undefined;
    return row ? mapWorkspace(row) : null;
  }

  listWorkspaces(pageSize: number, pageToken: string): Page<WorkspaceRecord> {
    const limit = normalizePageSize(pageSize);
    const offset = decodePageToken(pageToken);
    const rows = this.database
      .prepare(
        `SELECT w.*,
          (SELECT e.uid FROM executions e
           WHERE e.workspace_uid = w.uid AND e.state IN ('RUNNING', 'INPUT_REQUIRED')
           LIMIT 1) AS active_execution_uid
         FROM workspaces w ORDER BY w.create_time_ms DESC LIMIT ? OFFSET ?`,
      )
      .all(limit + 1, offset) as WorkspaceRow[];
    return page(rows.map(mapWorkspace), limit, offset);
  }

  deleteWorkspace(uid: string): boolean {
    return this.database.prepare("DELETE FROM workspaces WHERE uid = ?").run(uid).changes > 0;
  }

  createThread(workspaceUid: string, runtimeUid: string, displayName: string): ThreadRecord {
    const uid = randomUUID();
    const now = new Date();
    this.database
      .prepare(
        `INSERT INTO threads(
          uid, workspace_uid, runtime_uid, display_name, state, create_time_ms, update_time_ms
        ) VALUES (?, ?, ?, ?, 'IDLE', ?, ?)`,
      )
      .run(uid, workspaceUid, runtimeUid, displayName, now.getTime(), now.getTime());
    return {
      uid,
      workspaceUid,
      runtimeUid,
      displayName,
      state: "IDLE",
      providerSessionId: null,
      createTime: now,
      updateTime: now,
    };
  }

  getThread(uid: string): ThreadRecord | null {
    const row = this.database.prepare("SELECT * FROM threads WHERE uid = ?").get(uid) as
      | ThreadRow
      | undefined;
    return row ? mapThread(row) : null;
  }

  listThreads(workspaceUid: string, pageSize: number, pageToken: string): Page<ThreadRecord> {
    const limit = normalizePageSize(pageSize);
    const offset = decodePageToken(pageToken);
    const rows = this.database
      .prepare(
        `SELECT * FROM threads WHERE workspace_uid = ?
         ORDER BY update_time_ms DESC LIMIT ? OFFSET ?`,
      )
      .all(workspaceUid, limit + 1, offset) as ThreadRow[];
    return page(rows.map(mapThread), limit, offset);
  }

  deleteThread(uid: string): boolean {
    return this.database.prepare("DELETE FROM threads WHERE uid = ?").run(uid).changes > 0;
  }

  setThreadProviderSession(uid: string, providerSessionId: string): void {
    this.database
      .prepare("UPDATE threads SET provider_session_id = ?, update_time_ms = ? WHERE uid = ?")
      .run(providerSessionId, Date.now(), uid);
  }

  setThreadState(uid: string, state: ThreadState): void {
    this.database
      .prepare("UPDATE threads SET state = ?, update_time_ms = ? WHERE uid = ?")
      .run(state, Date.now(), uid);
  }

  createExecution(thread: ThreadRecord, runId: string): ExecutionRecord {
    const uid = randomUUID();
    const startTime = new Date();
    this.database
      .prepare(
        `INSERT INTO executions(
          uid, thread_uid, workspace_uid, run_id, state, start_time_ms
        ) VALUES (?, ?, ?, ?, 'RUNNING', ?)`,
      )
      .run(uid, thread.uid, thread.workspaceUid, runId, startTime.getTime());
    this.setThreadState(thread.uid, "RUNNING");
    return {
      uid,
      threadUid: thread.uid,
      workspaceUid: thread.workspaceUid,
      runId,
      state: "RUNNING",
      nativeTurnId: null,
      startTime,
      endTime: null,
    };
  }

  getExecution(uid: string): ExecutionRecord | null {
    const row = this.database.prepare("SELECT * FROM executions WHERE uid = ?").get(uid) as
      | ExecutionRow
      | undefined;
    return row ? mapExecution(row) : null;
  }

  getExecutionByRunId(runId: string): ExecutionRecord | null {
    const row = this.database.prepare("SELECT * FROM executions WHERE run_id = ?").get(runId) as
      | ExecutionRow
      | undefined;
    return row ? mapExecution(row) : null;
  }

  deleteExecution(uid: string): void {
    this.database.prepare("DELETE FROM executions WHERE uid = ?").run(uid);
  }

  getActiveExecutionForThread(threadUid: string): ExecutionRecord | null {
    const row = this.database
      .prepare(
        `SELECT * FROM executions
         WHERE thread_uid = ? AND state IN ('RUNNING', 'INPUT_REQUIRED')
         ORDER BY start_time_ms DESC LIMIT 1`,
      )
      .get(threadUid) as ExecutionRow | undefined;
    return row ? mapExecution(row) : null;
  }

  listExecutions(threadUid: string, pageSize: number, pageToken: string): Page<ExecutionRecord> {
    const limit = normalizePageSize(pageSize);
    const offset = decodePageToken(pageToken);
    const rows = this.database
      .prepare(
        `SELECT * FROM executions WHERE thread_uid = ?
         ORDER BY start_time_ms DESC LIMIT ? OFFSET ?`,
      )
      .all(threadUid, limit + 1, offset) as ExecutionRow[];
    return page(rows.map(mapExecution), limit, offset);
  }

  countActiveExecutions(): number {
    const row = this.database
      .prepare(
        "SELECT COUNT(*) AS count FROM executions WHERE state IN ('RUNNING', 'INPUT_REQUIRED')",
      )
      .get() as { count: number };
    return row.count;
  }

  setExecutionNativeTurn(uid: string, nativeTurnId: string): void {
    this.database
      .prepare("UPDATE executions SET native_turn_id = ? WHERE uid = ?")
      .run(nativeTurnId, uid);
  }

  setExecutionState(uid: string, state: ExecutionState): void {
    const terminal = ["SUCCEEDED", "FAILED", "CANCELED", "LOST"].includes(state);
    this.database
      .prepare("UPDATE executions SET state = ?, end_time_ms = ? WHERE uid = ?")
      .run(state, terminal ? Date.now() : null, uid);
  }

  createRun(
    runId: string,
    executionUid: string,
    threadUid: string,
    parentRunId: string | null,
  ): RunRecord {
    const createTime = new Date();
    this.database
      .prepare(
        `INSERT INTO runs(
          run_id, execution_uid, thread_uid, parent_run_id, state, create_time_ms
        ) VALUES (?, ?, ?, ?, 'RUNNING', ?)`,
      )
      .run(runId, executionUid, threadUid, parentRunId, createTime.getTime());
    return {
      runId,
      executionUid,
      threadUid,
      parentRunId,
      state: "RUNNING",
      createTime,
      endTime: null,
    };
  }

  getRun(runId: string): RunRecord | null {
    const row = this.database.prepare("SELECT * FROM runs WHERE run_id = ?").get(runId) as
      | RunRow
      | undefined;
    return row ? mapRun(row) : null;
  }

  getLatestRunForExecution(executionUid: string): RunRecord | null {
    const row = this.database
      .prepare("SELECT * FROM runs WHERE execution_uid = ? ORDER BY create_time_ms DESC LIMIT 1")
      .get(executionUid) as RunRow | undefined;
    return row ? mapRun(row) : null;
  }

  setRunState(runId: string, state: RunState): void {
    this.database
      .prepare("UPDATE runs SET state = ?, end_time_ms = ? WHERE run_id = ?")
      .run(state, state === "RUNNING" ? null : Date.now(), runId);
  }

  deleteRun(runId: string): void {
    this.database.prepare("DELETE FROM runs WHERE run_id = ?").run(runId);
  }

  recoverLostExecutions(): readonly ExecutionRecord[] {
    const active = this.database
      .prepare("SELECT * FROM executions WHERE state IN ('RUNNING', 'INPUT_REQUIRED')")
      .all() as ExecutionRow[];
    if (active.length === 0) {
      return [];
    }
    const now = Date.now();
    this.database.transaction(() => {
      this.database
        .prepare(
          "UPDATE executions SET state = 'LOST', end_time_ms = ? WHERE state IN ('RUNNING', 'INPUT_REQUIRED')",
        )
        .run(now);
      this.database
        .prepare("UPDATE pending_interrupts SET state = 'LOST' WHERE state = 'PENDING'")
        .run();
      this.database
        .prepare("UPDATE runs SET state = 'FAILED', end_time_ms = ? WHERE state = 'RUNNING'")
        .run(now);
      this.database
        .prepare(
          `UPDATE threads SET state = 'LOST', update_time_ms = ?
           WHERE uid IN (${active.map(() => "?").join(",")})`,
        )
        .run(now, ...active.map((row) => row.thread_uid));
    })();
    return active.map((row) => mapExecution({ ...row, state: "LOST", end_time_ms: now }));
  }

  createInterrupt(input: {
    executionUid: string;
    threadUid: string;
    runId: string;
    providerRequestId: string;
    toolCallId: string;
    providerInput: unknown;
    inputHash: string;
  }): PendingInterruptRecord {
    const uid = randomUUID();
    const createTime = new Date();
    this.database
      .prepare(
        `INSERT INTO pending_interrupts(
          uid, execution_uid, thread_uid, run_id, provider_request_id, tool_call_id,
          input_json, input_hash, state, create_time_ms
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'PENDING', ?)`,
      )
      .run(
        uid,
        input.executionUid,
        input.threadUid,
        input.runId,
        input.providerRequestId,
        input.toolCallId,
        JSON.stringify(input.providerInput),
        input.inputHash,
        createTime.getTime(),
      );
    return {
      uid,
      executionUid: input.executionUid,
      threadUid: input.threadUid,
      runId: input.runId,
      providerRequestId: input.providerRequestId,
      toolCallId: input.toolCallId,
      input: input.providerInput,
      inputHash: input.inputHash,
      state: "PENDING",
      createTime,
    };
  }

  listPendingInterrupts(executionUid: string): readonly PendingInterruptRecord[] {
    const rows = this.database
      .prepare(
        `SELECT * FROM pending_interrupts
         WHERE execution_uid = ? AND state = 'PENDING' ORDER BY create_time_ms`,
      )
      .all(executionUid) as InterruptRow[];
    return rows.map(mapInterrupt);
  }

  setInterruptState(uid: string, state: InterruptState): void {
    this.database.prepare("UPDATE pending_interrupts SET state = ? WHERE uid = ?").run(state, uid);
  }

  createTerminalSession(input: {
    workspaceUid: string;
    displayName: string;
    cwd: string;
    columns: number;
    rows: number;
  }): TerminalSessionRecord {
    const uid = randomUUID();
    const createTime = new Date();
    const tmuxName = `sm-${uid}`;
    this.database
      .prepare(
        `INSERT INTO terminal_sessions(
          uid, workspace_uid, display_name, cwd, tmux_name, columns, rows, state, create_time_ms
        ) VALUES (?, ?, ?, ?, ?, ?, ?, 'RUNNING', ?)`,
      )
      .run(
        uid,
        input.workspaceUid,
        input.displayName,
        input.cwd,
        tmuxName,
        input.columns,
        input.rows,
        createTime.getTime(),
      );
    return { uid, tmuxName, state: "RUNNING", createTime, ...input };
  }

  getTerminalSession(uid: string): TerminalSessionRecord | null {
    const row = this.database.prepare("SELECT * FROM terminal_sessions WHERE uid = ?").get(uid) as
      | TerminalRow
      | undefined;
    return row ? mapTerminal(row) : null;
  }

  listTerminalSessions(workspaceUid: string): readonly TerminalSessionRecord[] {
    const rows = this.database
      .prepare(
        "SELECT * FROM terminal_sessions WHERE workspace_uid = ? ORDER BY create_time_ms DESC",
      )
      .all(workspaceUid) as TerminalRow[];
    return rows.map(mapTerminal);
  }

  listRunningTerminalSessions(): readonly TerminalSessionRecord[] {
    const rows = this.database
      .prepare("SELECT * FROM terminal_sessions WHERE state = 'RUNNING' ORDER BY create_time_ms")
      .all() as TerminalRow[];
    return rows.map(mapTerminal);
  }

  countRunningTerminals(): number {
    const row = this.database
      .prepare("SELECT COUNT(*) AS count FROM terminal_sessions WHERE state = 'RUNNING'")
      .get() as { count: number };
    return row.count;
  }

  setTerminalState(uid: string, state: TerminalState): void {
    this.database.prepare("UPDATE terminal_sessions SET state = ? WHERE uid = ?").run(state, uid);
  }

  setTerminalSize(uid: string, columns: number, rows: number): void {
    this.database
      .prepare("UPDATE terminal_sessions SET columns = ?, rows = ? WHERE uid = ?")
      .run(columns, rows, uid);
  }

  deleteTerminalSession(uid: string): boolean {
    return (
      this.database.prepare("DELETE FROM terminal_sessions WHERE uid = ?").run(uid).changes > 0
    );
  }

  createPairingSecret(secretHash: string, expireTime: Date): void {
    this.database.transaction(() => {
      this.database
        .prepare("DELETE FROM pairing_secrets WHERE expire_time_ms <= ? OR use_time_ms IS NOT NULL")
        .run(Date.now());
      this.database
        .prepare("INSERT INTO pairing_secrets(secret_hash, expire_time_ms) VALUES (?, ?)")
        .run(secretHash, expireTime.getTime());
    })();
  }

  consumePairingSecret(secretHash: string, now = new Date()): boolean {
    const result = this.database
      .prepare(
        `UPDATE pairing_secrets SET use_time_ms = ?
         WHERE secret_hash = ? AND use_time_ms IS NULL AND expire_time_ms > ?`,
      )
      .run(now.getTime(), secretHash, now.getTime());
    return result.changes === 1;
  }

  createDevice(input: {
    uid: string;
    displayName: string;
    certificateSerial: string;
    tokenHash: string;
    expireTime: Date;
  }): DeviceRecord {
    const createTime = new Date();
    this.database
      .prepare(
        `INSERT INTO devices(
          uid, display_name, status, certificate_serial, token_hash, create_time_ms, expire_time_ms
        ) VALUES (?, ?, 'ACTIVE', ?, ?, ?, ?)`,
      )
      .run(
        input.uid,
        input.displayName,
        input.certificateSerial,
        input.tokenHash,
        createTime.getTime(),
        input.expireTime.getTime(),
      );
    return {
      uid: input.uid,
      displayName: input.displayName,
      status: "ACTIVE",
      certificateSerial: input.certificateSerial,
      createTime,
      expireTime: input.expireTime,
    };
  }

  findActiveDeviceByTokenHash(tokenHash: string, now = new Date()): DeviceRecord | null {
    const row = this.database
      .prepare(
        `SELECT uid, display_name, status, certificate_serial, create_time_ms, expire_time_ms
         FROM devices
         WHERE token_hash = ? AND status = 'ACTIVE' AND expire_time_ms > ?`,
      )
      .get(tokenHash, now.getTime()) as DeviceRow | undefined;
    return row ? mapDevice(row) : null;
  }

  listDevices(): readonly DeviceRecord[] {
    const rows = this.database
      .prepare(
        `SELECT uid, display_name, status, certificate_serial, create_time_ms, expire_time_ms
         FROM devices ORDER BY create_time_ms DESC`,
      )
      .all() as DeviceRow[];
    return rows.map(mapDevice);
  }

  revokeDevice(uid: string): boolean {
    return (
      this.database
        .prepare("UPDATE devices SET status = 'REVOKED' WHERE uid = ? AND status = 'ACTIVE'")
        .run(uid).changes === 1
    );
  }

  putQuota(record: QuotaRecord): void {
    this.database
      .prepare(
        `INSERT INTO quota_snapshots(
          runtime_uid, freshness, used_ratio, reset_time_ms, observe_time_ms, source
        ) VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(runtime_uid) DO UPDATE SET
          freshness = excluded.freshness,
          used_ratio = excluded.used_ratio,
          reset_time_ms = excluded.reset_time_ms,
          observe_time_ms = excluded.observe_time_ms,
          source = excluded.source`,
      )
      .run(
        record.runtimeUid,
        record.freshness,
        record.usedRatio,
        record.resetTime?.getTime() ?? null,
        record.observeTime.getTime(),
        record.source,
      );
  }

  getQuota(runtimeUid: string): QuotaRecord | null {
    const row = this.database
      .prepare("SELECT * FROM quota_snapshots WHERE runtime_uid = ?")
      .get(runtimeUid) as
      | {
          runtime_uid: string;
          freshness: QuotaRecord["freshness"];
          used_ratio: number | null;
          reset_time_ms: number | null;
          observe_time_ms: number;
          source: string;
        }
      | undefined;
    if (!row) {
      return null;
    }
    return {
      runtimeUid: row.runtime_uid,
      freshness: row.freshness,
      usedRatio: row.used_ratio,
      resetTime: row.reset_time_ms === null ? null : new Date(row.reset_time_ms),
      observeTime: new Date(row.observe_time_ms),
      source: row.source,
    };
  }

  appendRawDiagnostic(source: string, payload: unknown): void {
    const payloadJson = JSON.stringify(sanitizeDiagnostic(payload));
    this.database
      .prepare(
        `INSERT INTO raw_diagnostics(uid, source, payload_json, byte_size, create_time_ms)
         VALUES (?, ?, ?, ?, ?)`,
      )
      .run(randomUUID(), source, payloadJson, Buffer.byteLength(payloadJson), Date.now());
    this.trimRawDiagnostics();
  }

  private trimRawDiagnostics(): void {
    const sevenDaysAgo = Date.now() - 7 * 24 * 60 * 60 * 1000;
    this.database.prepare("DELETE FROM raw_diagnostics WHERE create_time_ms < ?").run(sevenDaysAgo);
    const total = this.database
      .prepare("SELECT COALESCE(SUM(byte_size), 0) AS bytes FROM raw_diagnostics")
      .get() as { bytes: number };
    const maximumBytes = 100 * 1024 * 1024;
    if (total.bytes <= maximumBytes) {
      return;
    }
    const count = this.database.prepare("SELECT COUNT(*) AS count FROM raw_diagnostics").get() as {
      count: number;
    };
    this.database
      .prepare(
        `DELETE FROM raw_diagnostics WHERE uid IN (
          SELECT uid FROM raw_diagnostics ORDER BY create_time_ms LIMIT ?
        )`,
      )
      .run(Math.max(Math.floor(count.count / 10), 1));
  }
}

function mapWorkspace(row: WorkspaceRow): WorkspaceRecord {
  return {
    uid: row.uid,
    displayName: row.display_name,
    path: row.path,
    activeExecutionUid: row.active_execution_uid,
    createTime: new Date(row.create_time_ms),
  };
}

function mapThread(row: ThreadRow): ThreadRecord {
  return {
    uid: row.uid,
    workspaceUid: row.workspace_uid,
    runtimeUid: row.runtime_uid,
    displayName: row.display_name,
    state: row.state,
    providerSessionId: row.provider_session_id,
    createTime: new Date(row.create_time_ms),
    updateTime: new Date(row.update_time_ms),
  };
}

function mapExecution(row: ExecutionRow): ExecutionRecord {
  return {
    uid: row.uid,
    threadUid: row.thread_uid,
    workspaceUid: row.workspace_uid,
    runId: row.run_id,
    state: row.state,
    nativeTurnId: row.native_turn_id,
    startTime: new Date(row.start_time_ms),
    endTime: row.end_time_ms === null ? null : new Date(row.end_time_ms),
  };
}

function mapRun(row: RunRow): RunRecord {
  return {
    runId: row.run_id,
    executionUid: row.execution_uid,
    threadUid: row.thread_uid,
    parentRunId: row.parent_run_id,
    state: row.state,
    createTime: new Date(row.create_time_ms),
    endTime: row.end_time_ms === null ? null : new Date(row.end_time_ms),
  };
}

function mapInterrupt(row: InterruptRow): PendingInterruptRecord {
  return {
    uid: row.uid,
    executionUid: row.execution_uid,
    threadUid: row.thread_uid,
    runId: row.run_id,
    providerRequestId: row.provider_request_id,
    toolCallId: row.tool_call_id,
    input: JSON.parse(row.input_json) as unknown,
    inputHash: row.input_hash,
    state: row.state,
    createTime: new Date(row.create_time_ms),
  };
}

function mapTerminal(row: TerminalRow): TerminalSessionRecord {
  return {
    uid: row.uid,
    workspaceUid: row.workspace_uid,
    displayName: row.display_name,
    cwd: row.cwd,
    tmuxName: row.tmux_name,
    columns: row.columns,
    rows: row.rows,
    state: row.state,
    createTime: new Date(row.create_time_ms),
  };
}

function mapDevice(row: DeviceRow): DeviceRecord {
  return {
    uid: row.uid,
    displayName: row.display_name,
    status: row.status,
    certificateSerial: row.certificate_serial,
    createTime: new Date(row.create_time_ms),
    expireTime: new Date(row.expire_time_ms),
  };
}

function normalizePageSize(pageSize: number): number {
  return pageSize === 0 ? 50 : Math.min(pageSize, 100);
}

function decodePageToken(token: string): number {
  if (token.length === 0) {
    return 0;
  }
  const decoded = Buffer.from(token, "base64url").toString("utf8");
  if (Buffer.from(decoded).toString("base64url") !== token || !/^(0|[1-9][0-9]*)$/.test(decoded)) {
    throw new Error("Invalid page token");
  }
  const value = Number(decoded);
  if (!Number.isSafeInteger(value) || value < 0) {
    throw new Error("Invalid page token");
  }
  return value;
}

function page<T>(items: readonly T[], limit: number, offset: number): Page<T> {
  const hasNext = items.length > limit;
  return {
    items: items.slice(0, limit),
    nextPageToken: hasNext ? Buffer.from(String(offset + limit)).toString("base64url") : "",
  };
}

const SAFE_DIAGNOSTIC_VALUES = new Set([
  "type",
  "subtype",
  "method",
  "status",
  "code",
  "name",
  "kind",
  "role",
]);
const SENSITIVE_DIAGNOSTIC_FIELD =
  /token|secret|password|authorization|cookie|credential|content|text|input|output|result|message|prompt|argument|answer|path|url|email|account|encrypted/i;

function sanitizeDiagnostic(payload: unknown): unknown {
  const budget = { remaining: 512 };
  return sanitizeDiagnosticValue(payload, "", 0, budget);
}

function sanitizeDiagnosticValue(
  value: unknown,
  field: string,
  depth: number,
  budget: { remaining: number },
): unknown {
  if (budget.remaining <= 0 || depth > 6) {
    return "[truncated]";
  }
  budget.remaining -= 1;
  if (SENSITIVE_DIAGNOSTIC_FIELD.test(field)) {
    return redactedDiagnosticValue(value);
  }
  if (value === null || typeof value === "boolean" || typeof value === "number") {
    return value;
  }
  if (typeof value === "string") {
    return SAFE_DIAGNOSTIC_VALUES.has(field) ? value.slice(0, 128) : redactedDiagnosticValue(value);
  }
  if (Array.isArray(value)) {
    return value
      .slice(0, 16)
      .map((item) => sanitizeDiagnosticValue(item, field, depth + 1, budget));
  }
  if (typeof value === "object") {
    const result: Record<string, unknown> = {};
    for (const [index, [key, item]] of Object.entries(value).slice(0, 32).entries()) {
      const safeKey = /^[A-Za-z_][A-Za-z0-9_.-]{0,63}$/.test(key) ? key : `field_${index}`;
      result[safeKey] = sanitizeDiagnosticValue(item, key, depth + 1, budget);
    }
    return result;
  }
  return `[${typeof value}]`;
}

function redactedDiagnosticValue(value: unknown): Record<string, unknown> {
  return {
    redacted: true,
    type: Array.isArray(value) ? "array" : value === null ? "null" : typeof value,
    ...(typeof value === "string" ? { bytes: Buffer.byteLength(value) } : {}),
  };
}
