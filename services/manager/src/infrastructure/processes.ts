import { type ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import { createInterface } from "node:readline";

const MAX_CAPTURE_BYTES = 1024 * 1024;

export interface CommandResult {
  readonly exitCode: number;
  readonly stdout: string;
  readonly stderr: string;
}

export async function runCommand(
  command: string,
  args: readonly string[],
  options: {
    readonly cwd?: string;
    readonly env?: NodeJS.ProcessEnv;
    readonly timeoutMs?: number;
  } = {},
): Promise<CommandResult> {
  const child = spawn(command, args, {
    cwd: options.cwd,
    env: options.env,
    shell: false,
    stdio: ["ignore", "pipe", "pipe"],
  });
  let stdout: Buffer<ArrayBufferLike> = Buffer.alloc(0);
  let stderr: Buffer<ArrayBufferLike> = Buffer.alloc(0);
  child.stdout.on("data", (chunk: Buffer) => {
    stdout = appendBounded(stdout, chunk);
  });
  child.stderr.on("data", (chunk: Buffer) => {
    stderr = appendBounded(stderr, chunk);
  });

  const timeoutMs = options.timeoutMs ?? 10_000;
  const timeout = setTimeout(() => child.kill("SIGKILL"), timeoutMs);
  try {
    const exitCode = await new Promise<number>((resolve, reject) => {
      child.once("error", reject);
      child.once("close", (code) => resolve(code ?? -1));
    });
    return {
      exitCode,
      stdout: stdout.toString("utf8"),
      stderr: stderr.toString("utf8"),
    };
  } finally {
    clearTimeout(timeout);
  }
}

export interface JsonRpcError {
  readonly code: number;
  readonly message: string;
  readonly data?: unknown;
}

type RequestId = string | number;
type JsonObject = Record<string, unknown>;
type RequestHandler = (method: string, id: RequestId, params: unknown) => Promise<void>;
type NotificationHandler = (method: string, params: unknown) => void;
type ExitHandler = (error: Error) => void;

interface PendingRequest {
  readonly resolve: (result: unknown) => void;
  readonly reject: (error: Error) => void;
  readonly timeout: NodeJS.Timeout;
}

export class JsonRpcProcess {
  private readonly child: ChildProcessWithoutNullStreams;
  private readonly pending = new Map<RequestId, PendingRequest>();
  private readonly notificationHandlers = new Set<NotificationHandler>();
  private readonly requestHandlers = new Set<RequestHandler>();
  private readonly exitHandlers = new Set<ExitHandler>();
  private nextRequestId = 1;
  private closed = false;
  private stderrBytes = 0;

  constructor(
    command: string,
    args: readonly string[],
    options: { readonly cwd?: string; readonly env?: NodeJS.ProcessEnv } = {},
  ) {
    this.child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env,
      shell: false,
      stdio: ["pipe", "pipe", "pipe"],
    });
    this.child.stderr.on("data", (chunk: Buffer) => {
      this.stderrBytes = Math.min(Number.MAX_SAFE_INTEGER, this.stderrBytes + chunk.length);
    });
    const lines = createInterface({
      input: this.child.stdout,
      crlfDelay: Number.POSITIVE_INFINITY,
    });
    lines.on("line", (line) => this.handleLine(line));
    this.child.once("error", (error) => this.handleExit(error));
    this.child.once("close", (code, signal) => {
      this.handleExit(
        new Error(
          `JSON-RPC process exited with code ${code ?? "unknown"} and signal ${signal ?? "none"}${this.stderrBytes > 0 ? ` (${this.stderrBytes} stderr bytes)` : ""}`,
        ),
      );
    });
  }

  onNotification(handler: NotificationHandler): () => void {
    this.notificationHandlers.add(handler);
    return () => this.notificationHandlers.delete(handler);
  }

  onRequest(handler: RequestHandler): () => void {
    this.requestHandlers.add(handler);
    return () => this.requestHandlers.delete(handler);
  }

  onExit(handler: ExitHandler): () => void {
    this.exitHandlers.add(handler);
    return () => this.exitHandlers.delete(handler);
  }

  async request<T>(method: string, params: unknown, timeoutMs = 30_000): Promise<T> {
    if (this.closed) {
      throw new Error("JSON-RPC process is closed");
    }
    const id = this.nextRequestId;
    this.nextRequestId += 1;
    const result = new Promise<T>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pending.delete(id);
        reject(new Error(`JSON-RPC request ${method} timed out`));
      }, timeoutMs);
      this.pending.set(id, {
        resolve: (value) => resolve(value as T),
        reject,
        timeout,
      });
    });
    try {
      this.write({ method, id, params });
    } catch (error) {
      const pending = this.pending.get(id);
      if (pending) {
        clearTimeout(pending.timeout);
        this.pending.delete(id);
        pending.reject(asError(error));
      }
    }
    return result;
  }

  notify(method: string, params: unknown): void {
    this.write({ method, params });
  }

  respond(id: RequestId, result: unknown): void {
    this.write({ id, result });
  }

  respondError(id: RequestId, error: JsonRpcError): void {
    this.write({ id, error });
  }

  async close(): Promise<void> {
    if (this.closed) {
      return;
    }
    this.closed = true;
    await new Promise<void>((resolve) => {
      let settled = false;
      const finish = () => {
        if (settled) {
          return;
        }
        settled = true;
        clearTimeout(timeout);
        this.child.off("close", finish);
        resolve();
      };
      const timeout = setTimeout(() => {
        this.child.kill("SIGKILL");
        finish();
      }, 5_000);
      this.child.once("close", finish);
      this.child.stdin.end();
      this.child.kill("SIGTERM");
      if (this.child.exitCode !== null || this.child.signalCode !== null) {
        finish();
      }
    });
  }

  private write(message: JsonObject): void {
    if (this.closed || !this.child.stdin.writable) {
      throw new Error("JSON-RPC process stdin is not writable");
    }
    this.child.stdin.write(`${JSON.stringify(message)}\n`);
  }

  private handleLine(line: string): void {
    if (Buffer.byteLength(line) > 8 * 1024 * 1024) {
      this.handleExit(new Error("JSON-RPC frame exceeded 8 MiB"));
      this.child.kill("SIGKILL");
      return;
    }
    let message: JsonObject;
    try {
      message = JSON.parse(line) as JsonObject;
    } catch {
      return;
    }
    if ((typeof message.id === "number" || typeof message.id === "string") && !message.method) {
      const pending = this.pending.get(message.id);
      if (!pending) {
        return;
      }
      clearTimeout(pending.timeout);
      this.pending.delete(message.id);
      if (message.error && typeof message.error === "object") {
        const error = message.error as { message?: unknown; code?: unknown };
        pending.reject(
          new Error(
            `JSON-RPC error ${String(error.code ?? "unknown")}: ${String(error.message ?? "unknown")}`,
          ),
        );
      } else {
        pending.resolve(message.result);
      }
      return;
    }
    if (typeof message.method !== "string") {
      return;
    }
    if (typeof message.id === "number" || typeof message.id === "string") {
      for (const handler of this.requestHandlers) {
        void handler(message.method, message.id, message.params);
      }
      return;
    }
    for (const handler of this.notificationHandlers) {
      handler(message.method, message.params);
    }
  }

  private handleExit(error: Error): void {
    if (this.closed && this.pending.size === 0) {
      return;
    }
    this.closed = true;
    for (const pending of this.pending.values()) {
      clearTimeout(pending.timeout);
      pending.reject(error);
    }
    this.pending.clear();
    for (const handler of this.exitHandlers) {
      handler(error);
    }
  }
}

function asError(error: unknown): Error {
  return error instanceof Error ? error : new Error(String(error));
}

export function scrubApiCredentials(environment: NodeJS.ProcessEnv): NodeJS.ProcessEnv {
  const scrubbed = { ...environment };
  for (const name of [
    "ANTHROPIC_API_KEY",
    "ANTHROPIC_AUTH_TOKEN",
    "ANTHROPIC_BASE_URL",
    "OPENAI_API_KEY",
    "OPENAI_BASE_URL",
    "CODEX_API_KEY",
  ]) {
    delete scrubbed[name];
  }
  return scrubbed;
}

function appendBounded(
  existing: Buffer<ArrayBufferLike>,
  chunk: Buffer<ArrayBufferLike>,
): Buffer<ArrayBufferLike> {
  const combined = Buffer.concat([existing, chunk]);
  return combined.length <= MAX_CAPTURE_BYTES ? combined : combined.subarray(-MAX_CAPTURE_BYTES);
}
