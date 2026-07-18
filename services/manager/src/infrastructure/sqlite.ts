import { mkdirSync } from "node:fs";
import { dirname } from "node:path";
import Database from "better-sqlite3";

const DATABASE_VERSION = 1;

const INITIAL_SCHEMA = `
CREATE TABLE schema_migrations (
  version INTEGER PRIMARY KEY,
  apply_time_ms INTEGER NOT NULL
);

CREATE TABLE workspaces (
  uid TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  create_time_ms INTEGER NOT NULL
);

CREATE TABLE threads (
  uid TEXT PRIMARY KEY,
  workspace_uid TEXT NOT NULL REFERENCES workspaces(uid) ON DELETE CASCADE,
  runtime_uid TEXT NOT NULL,
  display_name TEXT NOT NULL,
  state TEXT NOT NULL,
  provider_session_id TEXT,
  next_sequence INTEGER NOT NULL DEFAULT 1,
  create_time_ms INTEGER NOT NULL,
  update_time_ms INTEGER NOT NULL
);

CREATE INDEX threads_workspace_idx ON threads(workspace_uid, update_time_ms DESC);

CREATE TABLE executions (
  uid TEXT PRIMARY KEY,
  thread_uid TEXT NOT NULL REFERENCES threads(uid) ON DELETE CASCADE,
  workspace_uid TEXT NOT NULL REFERENCES workspaces(uid) ON DELETE CASCADE,
  run_id TEXT NOT NULL UNIQUE,
  state TEXT NOT NULL,
  native_turn_id TEXT,
  start_time_ms INTEGER NOT NULL,
  end_time_ms INTEGER
);

CREATE INDEX executions_thread_idx ON executions(thread_uid, start_time_ms DESC);
CREATE UNIQUE INDEX executions_workspace_writer_idx
  ON executions(workspace_uid)
  WHERE state IN ('RUNNING', 'INPUT_REQUIRED');

CREATE TABLE runs (
  run_id TEXT PRIMARY KEY,
  execution_uid TEXT NOT NULL REFERENCES executions(uid) ON DELETE CASCADE,
  thread_uid TEXT NOT NULL REFERENCES threads(uid) ON DELETE CASCADE,
  parent_run_id TEXT,
  state TEXT NOT NULL,
  create_time_ms INTEGER NOT NULL,
  end_time_ms INTEGER
);

CREATE INDEX runs_execution_idx ON runs(execution_uid, create_time_ms);

CREATE TABLE agui_events (
  thread_uid TEXT NOT NULL REFERENCES threads(uid) ON DELETE CASCADE,
  sequence INTEGER NOT NULL,
  run_id TEXT NOT NULL,
  event_json TEXT NOT NULL,
  byte_size INTEGER NOT NULL,
  create_time_ms INTEGER NOT NULL,
  PRIMARY KEY(thread_uid, sequence)
);

CREATE INDEX agui_events_run_idx ON agui_events(run_id, sequence);

CREATE TABLE pending_interrupts (
  uid TEXT PRIMARY KEY,
  execution_uid TEXT NOT NULL REFERENCES executions(uid) ON DELETE CASCADE,
  thread_uid TEXT NOT NULL REFERENCES threads(uid) ON DELETE CASCADE,
  run_id TEXT NOT NULL,
  provider_request_id TEXT NOT NULL,
  tool_call_id TEXT NOT NULL,
  input_json TEXT NOT NULL,
  input_hash TEXT NOT NULL,
  state TEXT NOT NULL,
  create_time_ms INTEGER NOT NULL,
  UNIQUE(execution_uid, provider_request_id)
);

CREATE INDEX pending_interrupts_execution_idx
  ON pending_interrupts(execution_uid, state, create_time_ms);

CREATE TABLE terminal_sessions (
  uid TEXT PRIMARY KEY,
  workspace_uid TEXT NOT NULL REFERENCES workspaces(uid) ON DELETE CASCADE,
  display_name TEXT NOT NULL,
  cwd TEXT NOT NULL,
  tmux_name TEXT NOT NULL UNIQUE,
  columns INTEGER NOT NULL,
  rows INTEGER NOT NULL,
  state TEXT NOT NULL,
  create_time_ms INTEGER NOT NULL
);

CREATE INDEX terminal_sessions_workspace_idx
  ON terminal_sessions(workspace_uid, create_time_ms DESC);

CREATE TABLE devices (
  uid TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  status TEXT NOT NULL,
  certificate_serial TEXT NOT NULL UNIQUE,
  token_hash TEXT NOT NULL UNIQUE,
  create_time_ms INTEGER NOT NULL,
  expire_time_ms INTEGER NOT NULL
);

CREATE TABLE pairing_secrets (
  secret_hash TEXT PRIMARY KEY,
  expire_time_ms INTEGER NOT NULL,
  use_time_ms INTEGER
);

CREATE TABLE quota_snapshots (
  runtime_uid TEXT PRIMARY KEY,
  freshness TEXT NOT NULL,
  used_ratio REAL,
  reset_time_ms INTEGER,
  observe_time_ms INTEGER NOT NULL,
  source TEXT NOT NULL
);

CREATE TABLE raw_diagnostics (
  uid TEXT PRIMARY KEY,
  source TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  byte_size INTEGER NOT NULL,
  create_time_ms INTEGER NOT NULL
);

CREATE INDEX raw_diagnostics_time_idx ON raw_diagnostics(create_time_ms);
`;

export type SqliteDatabase = Database.Database;

export function openDatabase(path: string): SqliteDatabase {
  mkdirSync(dirname(path), { recursive: true, mode: 0o700 });
  const database = new Database(path);
  database.pragma("journal_mode = WAL");
  database.pragma("foreign_keys = ON");
  database.pragma("busy_timeout = 5000");
  database.pragma("synchronous = NORMAL");
  migrate(database);
  return database;
}

function migrate(database: SqliteDatabase): void {
  const table = database
    .prepare("SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'")
    .get();
  if (!table) {
    database.transaction(() => {
      database.exec(INITIAL_SCHEMA);
      database
        .prepare("INSERT INTO schema_migrations(version, apply_time_ms) VALUES (?, ?)")
        .run(DATABASE_VERSION, Date.now());
    })();
    return;
  }

  const current = database
    .prepare("SELECT COALESCE(MAX(version), 0) AS version FROM schema_migrations")
    .get() as { version: number };
  if (current.version !== DATABASE_VERSION) {
    throw new Error(`Unsupported database schema ${current.version}; expected ${DATABASE_VERSION}`);
  }
}
