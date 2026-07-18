import { type AGUIEvent, EventSchemas } from "@ag-ui/core";
import type { ExecutionState, StoredEvent, ThreadState } from "../domain/records.js";
import type { SqliteDatabase } from "../infrastructure/sqlite.js";

interface EventRow {
  thread_uid: string;
  sequence: number;
  run_id: string;
  event_json: string;
  create_time_ms: number;
}

export interface EventProjection {
  readonly threadState?: ThreadState;
  readonly executionUid?: string;
  readonly executionState?: ExecutionState;
}

type Subscriber = (event: StoredEvent) => void;
const MAX_EVENT_BYTES = 1024 * 1024;

export class EventStream {
  private readonly subscribers = new Map<string, Set<Subscriber>>();

  constructor(private readonly database: SqliteDatabase) {}

  append(
    threadUid: string,
    runId: string,
    event: AGUIEvent,
    projection: EventProjection = {},
  ): StoredEvent {
    const normalized = EventSchemas.parse({ timestamp: Date.now(), ...event });
    const eventJson = JSON.stringify(normalized);
    if (Buffer.byteLength(eventJson) > MAX_EVENT_BYTES) {
      throw new Error("AG-UI event exceeds 1 MiB");
    }
    const createTime = new Date();
    const stored = this.database.transaction(() => {
      const thread = this.database
        .prepare("SELECT next_sequence FROM threads WHERE uid = ?")
        .get(threadUid) as { next_sequence: number } | undefined;
      if (!thread) {
        throw new Error("Thread not found while appending an event");
      }
      const sequence = thread.next_sequence;
      this.database
        .prepare(
          `INSERT INTO agui_events(
            thread_uid, sequence, run_id, event_json, byte_size, create_time_ms
          ) VALUES (?, ?, ?, ?, ?, ?)`,
        )
        .run(
          threadUid,
          sequence,
          runId,
          eventJson,
          Buffer.byteLength(eventJson),
          createTime.getTime(),
        );
      this.database
        .prepare(
          `UPDATE threads SET
            next_sequence = next_sequence + 1,
            state = COALESCE(?, state),
            update_time_ms = ?
           WHERE uid = ?`,
        )
        .run(projection.threadState ?? null, createTime.getTime(), threadUid);
      if (projection.executionUid && projection.executionState) {
        const terminal = ["SUCCEEDED", "FAILED", "CANCELED", "LOST"].includes(
          projection.executionState,
        );
        this.database
          .prepare("UPDATE executions SET state = ?, end_time_ms = ? WHERE uid = ?")
          .run(
            projection.executionState,
            terminal ? createTime.getTime() : null,
            projection.executionUid,
          );
      }
      return {
        threadUid,
        sequence,
        runId,
        event: normalized,
        createTime,
      } satisfies StoredEvent;
    })();

    const subscribers = this.subscribers.get(threadUid);
    for (const subscriber of subscribers ?? []) {
      try {
        subscriber(stored);
      } catch {
        subscribers?.delete(subscriber);
      }
    }
    if (subscribers?.size === 0) {
      this.subscribers.delete(threadUid);
    }
    return stored;
  }

  list(threadUid: string, afterSequence: number, limit = 500): readonly StoredEvent[] {
    const rows = this.database
      .prepare(
        `SELECT thread_uid, sequence, run_id, event_json, create_time_ms
         FROM agui_events
         WHERE thread_uid = ? AND sequence > ?
         ORDER BY sequence LIMIT ?`,
      )
      .all(threadUid, afterSequence, limit) as EventRow[];
    return rows.map((row) => ({
      threadUid: row.thread_uid,
      sequence: row.sequence,
      runId: row.run_id,
      event: JSON.parse(row.event_json) as unknown,
      createTime: new Date(row.create_time_ms),
    }));
  }

  subscribe(threadUid: string, subscriber: Subscriber): () => void {
    const subscribers = this.subscribers.get(threadUid) ?? new Set<Subscriber>();
    subscribers.add(subscriber);
    this.subscribers.set(threadUid, subscribers);
    return () => {
      subscribers.delete(subscriber);
      if (subscribers.size === 0) {
        this.subscribers.delete(threadUid);
      }
    };
  }
}
