import { randomUUID } from "node:crypto";
import { EventType } from "@ag-ui/core";
import { fastifyConnectPlugin } from "@connectrpc/connect-fastify";
import { createValidateInterceptor } from "@connectrpc/validate";
import fastifyRateLimit from "@fastify/rate-limit";
import fastifyWebsocket from "@fastify/websocket";
import Fastify, { type FastifyInstance } from "fastify";
import { ClaudeAdapter } from "./adapters/claude/claude-adapter.js";
import { CodexAdapter } from "./adapters/codex/codex-adapter.js";
import { DeviceAuthService } from "./application/device-auth-service.js";
import { EventStream } from "./application/event-stream.js";
import { ExecutionCoordinator } from "./application/execution-coordinator.js";
import { OidcAuthService, PermissionDeniedError } from "./application/oidc-auth-service.js";
import { PairingAuthority } from "./application/pairing-authority.js";
import { RuntimeRegistry } from "./application/runtime-registry.js";
import { TerminalManager } from "./application/terminal-manager.js";
import { WorkspaceRegistry } from "./application/workspace-registry.js";
import { Permission } from "./gen/realtime/me/auth/v1/permission_pb.js";
import type { ServerConfig } from "./infrastructure/config.js";
import { ResourceStore } from "./infrastructure/resource-store.js";
import { SecretStore } from "./infrastructure/secret-store.js";
import { openDatabase, type SqliteDatabase } from "./infrastructure/sqlite.js";
import { registerAguiHttp } from "./transport/agui-http.js";
import { createControlRoutes } from "./transport/control/routes.js";
import { registerTerminalWebSocket } from "./transport/terminal-websocket.js";

const PAIR_DEVICE_PATH = "/realtime.me.manager.control.v1.PairingService/PairDevice";

export interface ServerApplication {
  readonly server: FastifyInstance;
  readonly pairing: PairingAuthority;
  close(): Promise<void>;
}

export async function createServer(config: ServerConfig): Promise<ServerApplication> {
  const database = openDatabase(config.databasePath);
  try {
    return await buildServer(config, database);
  } catch (error) {
    database.close();
    throw error;
  }
}

async function buildServer(
  config: ServerConfig,
  database: SqliteDatabase,
): Promise<ServerApplication> {
  const store = new ResourceStore(database);
  const secrets = await SecretStore.open(config.dataDirectory);
  const workspaces = await WorkspaceRegistry.create(store, config.allowedWorkspaceRoots);
  const events = new EventStream(database);
  const runtimes = new RuntimeRegistry([
    new CodexAdapter({
      executable: config.codexPath,
      expectedVersion: config.codexVersion,
      store,
    }),
    new ClaudeAdapter({
      executable: config.claudePath,
      expectedVersion: config.claudeVersion,
      store,
    }),
  ]);
  await runtimes.probeAll();
  const pairing = await PairingAuthority.open({
    store,
    secrets,
    opensslPath: config.opensslPath,
    dataDirectory: config.dataDirectory,
    publicUrl: config.publicUrl,
    pairingUrl: config.pairingUrl,
  });
  const terminals = new TerminalManager(store, config.tmuxPath, config.tmuxSocketName);
  await terminals.reconcile();
  const executions = new ExecutionCoordinator(store, events, runtimes, config.dataDirectory);
  projectRecoveredExecutions(store, events);
  const ownerAuth = new OidcAuthService(config.oidcIssuer, config.oidcAudience);

  const server = Fastify({
    bodyLimit: 256 * 1024,
    forceCloseConnections: true,
    trustProxy: "127.0.0.1",
    logger: {
      level: config.logLevel,
      redact: {
        paths: ["req.headers.authorization", "request.headers.authorization"],
        censor: "[REDACTED]",
      },
    },
  });
  await server.register(fastifyRateLimit, {
    global: true,
    max: 300,
    timeWindow: "1 minute",
  });
  await server.register(fastifyWebsocket, {
    options: { maxPayload: 1024 * 1024 + 1024 },
  });

  const deviceAuth = new DeviceAuthService(store, secrets);
  const pairingLimiter = new PairingLimiter();
  server.addHook("onRequest", async (request, reply) => {
    const path = request.url.split("?", 1)[0] ?? "";
    if (path === "/healthz" || path === "/readyz") {
      return;
    }
    if (path === PAIR_DEVICE_PATH) {
      if (!pairingLimiter.accept(request.ip)) {
        return reply.code(429).send({ error: "Pairing rate limit exceeded" });
      }
      return;
    }
    if (deviceAuth.authenticate(request.headers.authorization)) {
      return;
    }
    try {
      if (await ownerAuth.authenticate(request.headers.authorization, Permission.MANAGER_CONTROL)) {
        return;
      }
    } catch (error) {
      if (error instanceof PermissionDeniedError) {
        return reply.code(403).send({ error: "Permission denied" });
      }
      request.log.error({ err: error }, "OIDC authentication failed");
      return reply.code(503).send({ error: "Identity service unavailable" });
    }
    return reply.code(401).header("WWW-Authenticate", "Bearer").send({ error: "Unauthorized" });
  });

  server.get("/healthz", async () => ({ status: "ok" }));
  server.get("/readyz", async () => ({ status: "ready" }));
  const dependencies = { store, workspaces, runtimes, executions, terminals, pairing };
  await server.register(fastifyConnectPlugin, {
    routes: createControlRoutes(dependencies),
    interceptors: [createValidateInterceptor()],
    grpc: false,
    grpcWeb: false,
    connect: true,
  });
  registerAguiHttp(server, { store, events, executions, runtimes });
  registerTerminalWebSocket(server, terminals);
  server.setErrorHandler((error, _request, reply) => {
    const normalized = error instanceof Error ? error : new Error(String(error));
    const statusCode = statusForError(normalized);
    const message = statusCode >= 500 ? "Internal server error" : normalized.message.slice(0, 512);
    if (statusCode >= 500) {
      server.log.error({ err: normalized }, "request failed");
    }
    void reply.code(statusCode).send({ error: message });
  });

  let closed = false;
  return {
    server,
    pairing,
    close: async () => {
      if (closed) {
        return;
      }
      closed = true;
      executions.shutdown();
      await server.close();
      terminals.close();
      await runtimes.close();
      database.close();
    },
  };
}

function projectRecoveredExecutions(store: ResourceStore, events: EventStream): void {
  for (const execution of store.recoverLostExecutions()) {
    const parent = store.getLatestRunForExecution(execution.uid);
    const runId = `recovery-${randomUUID()}`;
    store.createRun(runId, execution.uid, execution.threadUid, parent?.runId ?? null);
    events.append(
      execution.threadUid,
      runId,
      {
        type: EventType.RUN_ERROR,
        message: "The server restarted while the provider execution was active",
        code: "PROVIDER_LOST",
      },
      {
        threadState: "LOST",
        executionUid: execution.uid,
        executionState: "LOST",
      },
    );
    store.setRunState(runId, "FAILED");
  }
}

class PairingLimiter {
  private readonly attempts = new Map<string, number[]>();
  private nextSweep = 0;

  accept(address: string): boolean {
    const now = Date.now();
    const cutoff = now - 60_000;
    if (now >= this.nextSweep || this.attempts.size >= 4096) {
      for (const [key, values] of this.attempts) {
        const active = values.filter((time) => time > cutoff);
        if (active.length === 0) {
          this.attempts.delete(key);
        } else {
          this.attempts.set(key, active);
        }
      }
      this.nextSweep = now + 60_000;
    }
    if (!this.attempts.has(address) && this.attempts.size >= 4096) {
      return false;
    }
    const attempts = (this.attempts.get(address) ?? []).filter((time) => time > cutoff);
    if (attempts.length >= 5) {
      this.attempts.set(address, attempts);
      return false;
    }
    attempts.push(now);
    this.attempts.set(address, attempts);
    return true;
  }
}

function statusForError(error: Error & { statusCode?: number; name?: string }): number {
  if (typeof error.statusCode === "number" && error.statusCode >= 400 && error.statusCode < 600) {
    return error.statusCode;
  }
  if (error.name === "ZodError" || /validation/i.test(error.name ?? "")) {
    return 400;
  }
  return 500;
}
