import type { ServerResponse } from "node:http";
import { AgentCapabilitiesSchema, EventType, RunAgentInputSchema } from "@ag-ui/core";
import type { FastifyInstance, FastifyReply } from "fastify";
import type { EventStream } from "../application/event-stream.js";
import type { ExecutionCoordinator, StartRunResult } from "../application/execution-coordinator.js";
import type { RuntimeRegistry } from "../application/runtime-registry.js";
import type { StoredEvent } from "../domain/records.js";
import type { ResourceStore } from "../infrastructure/resource-store.js";

const MAX_SUBSCRIBER_BYTES = 4 * 1024 * 1024;
const HEARTBEAT_INTERVAL_MS = 15_000;
const REPLAY_PAGE_SIZE = 64;

interface AguiDependencies {
  readonly store: ResourceStore;
  readonly events: EventStream;
  readonly executions: ExecutionCoordinator;
  readonly runtimes: RuntimeRegistry;
}

export function registerAguiHttp(server: FastifyInstance, dependencies: AguiDependencies): void {
  const { store, events, executions, runtimes } = dependencies;

  server.post<{ Querystring: { after?: string } }>("/v1/ag-ui/runs", async (request, reply) => {
    const input = RunAgentInputSchema.parse(request.body);
    const thread = store.getThread(input.threadId);
    if (!thread) {
      return reply.code(404).send({ error: "Thread not found" });
    }
    const after = parseSequence(request.query.after);
    const buffered: StoredEvent[] = [];
    let connection: SseConnection | null = null;
    const unsubscribe = events.subscribe(thread.uid, (event) => {
      if (connection) {
        connection.write(event);
      } else {
        buffered.push(event);
      }
    });
    let start: StartRunResult;
    try {
      start = await executions.start(input);
    } catch (error) {
      unsubscribe();
      return sendRunStartError(reply, error);
    }

    connection = new SseConnection(reply, unsubscribe, after, input.runId);
    replay(events, thread.uid, after, (event) => connection?.write(event) ?? false);
    buffered.sort((left, right) => left.sequence - right.sequence);
    for (const event of buffered) {
      if (!connection.write(event)) {
        break;
      }
    }
    if (start.existing && store.getRun(input.runId)?.state !== "RUNNING") {
      connection.end();
    }
  });

  server.get<{
    Params: { threadUid: string };
    Querystring: { after?: string };
  }>("/v1/ag-ui/threads/:threadUid/events", async (request, reply) => {
    const thread = store.getThread(request.params.threadUid);
    if (!thread) {
      return reply.code(404).send({ error: "Thread not found" });
    }
    const after = parseSequence(request.query.after);
    const buffered: StoredEvent[] = [];
    let connection: SseConnection | null = null;
    const unsubscribe = events.subscribe(thread.uid, (event) => {
      if (connection) {
        connection.write(event);
      } else {
        buffered.push(event);
      }
    });
    connection = new SseConnection(reply, unsubscribe, after);
    replay(events, thread.uid, after, (event) => connection?.write(event) ?? false);
    buffered.sort((left, right) => left.sequence - right.sequence);
    for (const event of buffered) {
      if (!connection.write(event)) {
        break;
      }
    }
  });

  server.get<{ Params: { threadUid: string } }>(
    "/v1/ag-ui/threads/:threadUid/capabilities",
    async (request, reply) => {
      const thread = store.getThread(request.params.threadUid);
      if (!thread) {
        return reply.code(404).send({ error: "Thread not found" });
      }
      const runtime = runtimes.get(thread.runtimeUid);
      if (!runtime) {
        return reply.code(404).send({ error: "Runtime not found" });
      }
      return AgentCapabilitiesSchema.parse(runtime.capabilities);
    },
  );
}

class SseConnection {
  private readonly response: ServerResponse;
  private readonly heartbeat: NodeJS.Timeout;
  private lastSequence = 0;
  private closed = false;

  constructor(
    reply: FastifyReply,
    private readonly unsubscribe: () => void,
    initialSequence: number,
    private readonly closeOnRunId?: string,
  ) {
    this.lastSequence = initialSequence;
    reply.hijack();
    this.response = reply.raw;
    this.response.writeHead(200, {
      "Content-Type": "text/event-stream; charset=utf-8",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
      "X-Accel-Buffering": "no",
    });
    this.response.write(": connected\n\n");
    this.heartbeat = setInterval(() => this.writeHeartbeat(), HEARTBEAT_INTERVAL_MS);
    this.response.once("close", () => this.close(false));
  }

  write(stored: StoredEvent): boolean {
    if (this.closed) {
      return false;
    }
    if (stored.sequence <= this.lastSequence) {
      return true;
    }
    const data = `id: ${stored.sequence}\ndata: ${JSON.stringify(stored.event)}\n\n`;
    if (this.response.writableLength + Buffer.byteLength(data) > MAX_SUBSCRIBER_BYTES) {
      this.close(true);
      return false;
    }
    this.lastSequence = stored.sequence;
    this.response.write(data);
    if (this.closeOnRunId === stored.runId && isTerminalRunEvent(stored.event)) {
      this.close(true);
    }
    return !this.closed;
  }

  end(): void {
    this.close(true);
  }

  private writeHeartbeat(): void {
    if (this.closed) {
      return;
    }
    const heartbeat = `: heartbeat ${Date.now()}\n\n`;
    if (this.response.writableLength + Buffer.byteLength(heartbeat) > MAX_SUBSCRIBER_BYTES) {
      this.close(true);
      return;
    }
    this.response.write(heartbeat);
  }

  private close(endResponse: boolean): void {
    if (this.closed) {
      return;
    }
    this.closed = true;
    clearInterval(this.heartbeat);
    this.unsubscribe();
    if (endResponse && !this.response.writableEnded) {
      this.response.end();
    }
  }
}

function replay(
  events: EventStream,
  threadUid: string,
  after: number,
  write: (event: StoredEvent) => boolean,
): void {
  let cursor = after;
  while (true) {
    const page = events.list(threadUid, cursor, REPLAY_PAGE_SIZE);
    for (const event of page) {
      if (!write(event)) {
        return;
      }
      cursor = event.sequence;
    }
    if (page.length < REPLAY_PAGE_SIZE) {
      return;
    }
  }
}

function parseSequence(value: string | undefined): number {
  if (!value) {
    return 0;
  }
  const sequence = Number(value);
  if (!Number.isSafeInteger(sequence) || sequence < 0) {
    throw Object.assign(new Error("after must be a non-negative integer"), { statusCode: 400 });
  }
  return sequence;
}

function isTerminalRunEvent(event: unknown): boolean {
  if (!event || typeof event !== "object" || !("type" in event)) {
    return false;
  }
  const type = (event as { type: unknown }).type;
  return type === EventType.RUN_FINISHED || type === EventType.RUN_ERROR;
}

function sendRunStartError(reply: FastifyReply, error: unknown): FastifyReply {
  const message = (error instanceof Error ? error.message : "Unable to start run").slice(0, 512);
  const statusCode = /limit|disk space/i.test(message)
    ? 429
    : /not found/i.test(message)
      ? 404
      : /invalid|must|exceeds|text-only|empty/i.test(message)
        ? 400
        : /active|available|waiting|live/i.test(message)
          ? 409
          : 500;
  return reply.code(statusCode).send({ error: message });
}
