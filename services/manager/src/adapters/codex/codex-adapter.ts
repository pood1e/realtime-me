import type { AGUIEvent, AgentCapabilities } from "@ag-ui/core";
import {
  activityDelta,
  assistantMessageDelta,
  assistantMessageEnd,
  assistantMessageStart,
  reasoningDelta,
  reasoningEnd,
  reasoningStart,
  toolCallResult,
  toolCallStart,
} from "../../application/agui-events.js";
import { redactSensitiveText } from "../../domain/redaction.js";
import {
  CODEX_RUNTIME_UID,
  type InputAnswer,
  type ProviderRun,
  type ProviderSink,
  type RuntimeAdapter,
  type RuntimeStatus,
} from "../../domain/runtime.js";
import { JsonRpcProcess, runCommand, scrubApiCredentials } from "../../infrastructure/processes.js";
import type { ResourceStore } from "../../infrastructure/resource-store.js";
import type { ServerNotification } from "./gen/ServerNotification.js";
import type { ThreadItem } from "./gen/v2/ThreadItem.js";
import type { ThreadResumeResponse } from "./gen/v2/ThreadResumeResponse.js";
import type { ThreadStartResponse } from "./gen/v2/ThreadStartResponse.js";
import type { ToolRequestUserInputParams } from "./gen/v2/ToolRequestUserInputParams.js";
import type { ToolRequestUserInputResponse } from "./gen/v2/ToolRequestUserInputResponse.js";
import type { TurnCompletedNotification } from "./gen/v2/TurnCompletedNotification.js";
import type { TurnStartResponse } from "./gen/v2/TurnStartResponse.js";

const MAX_TOOL_OUTPUT_BYTES = 4 * 1024 * 1024;

interface ActiveTurn {
  readonly run: ProviderRun;
  readonly sink: ProviderSink;
  readonly providerThreadId: string;
  readonly completion: Promise<void>;
  readonly resolve: () => void;
  readonly reject: (error: Error) => void;
  readonly messageIds: Set<string>;
  readonly reasoningIds: Set<string>;
  readonly outputs: Map<string, BoundedOutput>;
  nativeTurnId: string | null;
}

interface CodexAdapterOptions {
  readonly executable: string;
  readonly expectedVersion: string;
  readonly store: ResourceStore;
}

export class CodexAdapter implements RuntimeAdapter {
  readonly kind = "codex" as const;
  private client: JsonRpcProcess | null = null;
  private clientOpening: Promise<JsonRpcProcess> | null = null;
  private closing = false;
  private readonly turnsByExecution = new Map<string, ActiveTurn>();
  private readonly turnsByProviderThread = new Map<string, ActiveTurn>();

  constructor(private readonly options: CodexAdapterOptions) {}

  async probe(): Promise<RuntimeStatus> {
    const updateTime = new Date();
    try {
      const versionResult = await runCommand(this.options.executable, ["--version"]);
      if (versionResult.exitCode !== 0) {
        return this.status(
          "UNHEALTHY",
          "",
          `Codex version check failed with exit code ${versionResult.exitCode}`,
          updateTime,
        );
      }
      const version = versionResult.stdout.trim().split(/\s+/).at(-1) ?? "";
      if (version !== this.options.expectedVersion) {
        return this.status(
          "INCOMPATIBLE",
          version,
          `Expected Codex ${this.options.expectedVersion}`,
          updateTime,
        );
      }
      const client = await this.ensureClient();
      const account = await client.request<{ account: { type: string } | null }>("account/read", {
        refreshToken: false,
      });
      if (account.account?.type !== "chatgpt") {
        return this.status(
          "NOT_AUTHENTICATED",
          version,
          "Codex must be logged in with ChatGPT",
          updateTime,
        );
      }
      await this.refreshQuota(client);
      return this.status("AVAILABLE", version, "", updateTime);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      const missing = isMissingExecutable(error);
      return this.status(
        missing ? "NOT_INSTALLED" : "UNHEALTHY",
        "",
        redactSensitiveText(message, 512),
        updateTime,
      );
    }
  }

  async execute(run: ProviderRun, sink: ProviderSink): Promise<void> {
    const client = await this.ensureClient();
    const providerThreadId = await this.prepareThread(client, run, sink);
    const active = createActiveTurn(run, sink, providerThreadId);
    this.turnsByExecution.set(run.executionUid, active);
    this.turnsByProviderThread.set(providerThreadId, active);
    try {
      const response = await client.request<TurnStartResponse>("turn/start", {
        threadId: providerThreadId,
        input: [{ type: "text", text: run.prompt, text_elements: [] }],
        cwd: run.workspacePath,
        approvalPolicy: "never",
        sandboxPolicy: { type: "dangerFullAccess" },
      });
      active.nativeTurnId = response.turn.id;
      await sink.setNativeTurn(response.turn.id);
      await active.completion;
    } finally {
      for (const output of active.outputs.values()) {
        await output.flush();
      }
      this.turnsByExecution.delete(run.executionUid);
      this.turnsByProviderThread.delete(providerThreadId);
    }
  }

  async cancel(executionUid: string): Promise<void> {
    const active = this.requireActive(executionUid);
    if (!active.nativeTurnId) {
      throw new Error("Codex turn has not started");
    }
    const client = await this.ensureClient();
    await client.request("turn/interrupt", {
      threadId: active.providerThreadId,
      turnId: active.nativeTurnId,
    });
  }

  async steer(executionUid: string, instruction: string): Promise<void> {
    const active = this.requireActive(executionUid);
    if (!active.nativeTurnId) {
      throw new Error("Codex turn has not started");
    }
    const client = await this.ensureClient();
    await client.request("turn/steer", {
      threadId: active.providerThreadId,
      expectedTurnId: active.nativeTurnId,
      input: [{ type: "text", text: instruction, text_elements: [] }],
    });
  }

  async close(): Promise<void> {
    this.closing = true;
    const client = this.client ?? (await this.clientOpening?.catch(() => null)) ?? null;
    this.client = null;
    await client?.close();
  }

  private async ensureClient(): Promise<JsonRpcProcess> {
    if (this.closing) {
      throw new Error("Codex adapter is shutting down");
    }
    if (this.client) {
      return this.client;
    }
    if (this.clientOpening) {
      return this.clientOpening;
    }
    const opening = this.openClient();
    this.clientOpening = opening;
    try {
      return await opening;
    } finally {
      if (this.clientOpening === opening) {
        this.clientOpening = null;
      }
    }
  }

  private async openClient(): Promise<JsonRpcProcess> {
    const client = new JsonRpcProcess(
      this.options.executable,
      ["app-server", "--listen", "stdio://"],
      { env: scrubApiCredentials(process.env) },
    );
    client.onNotification((method, params) => {
      void this.handleNotification({ method, params } as ServerNotification).catch((error) => {
        const active = findActive(this.turnsByProviderThread, params);
        if (active) {
          active.reject(asError(error));
        } else {
          this.options.store.appendRawDiagnostic("codex-handler-error", {
            method,
            error: asError(error).message,
          });
        }
      });
    });
    client.onRequest(async (method, id, params) => {
      try {
        await this.handleServerRequest(client, method, id, params);
      } catch (error) {
        findActive(this.turnsByProviderThread, params)?.reject(asError(error));
        try {
          client.respondError(id, { code: -32_603, message: "Request handling failed" });
        } catch {
          // The process exit callback will fail any active turn.
        }
      }
    });
    client.onExit((error) => {
      if (this.client === client) {
        this.client = null;
      }
      for (const active of this.turnsByExecution.values()) {
        active.reject(error);
      }
    });
    try {
      await client.request("initialize", {
        clientInfo: {
          name: "realtime_me_manager",
          title: "Realtime Me Manager",
          version: "0.1.0",
        },
        capabilities: {
          experimentalApi: true,
          requestAttestation: false,
        },
      });
      client.notify("initialized", {});
      if (this.closing) {
        await client.close();
        throw new Error("Codex adapter is shutting down");
      }
      this.client = client;
      return client;
    } catch (error) {
      await client.close();
      throw error;
    }
  }

  private async prepareThread(
    client: JsonRpcProcess,
    run: ProviderRun,
    sink: ProviderSink,
  ): Promise<string> {
    if (run.providerSessionId) {
      const response = await client.request<ThreadResumeResponse>("thread/resume", {
        threadId: run.providerSessionId,
        cwd: run.workspacePath,
        approvalPolicy: "never",
        sandbox: "danger-full-access",
        excludeTurns: true,
      });
      return response.thread.id;
    }
    const response = await client.request<ThreadStartResponse>("thread/start", {
      cwd: run.workspacePath,
      modelProvider: "openai",
      approvalPolicy: "never",
      sandbox: "danger-full-access",
      ephemeral: false,
    });
    await sink.setProviderSession(response.thread.id);
    return response.thread.id;
  }

  private async handleNotification(notification: ServerNotification): Promise<void> {
    if (notification.method === "account/rateLimits/updated") {
      this.writeQuota(notification.params.rateLimits);
      return;
    }
    const active = findActive(this.turnsByProviderThread, notification.params);
    if (!active) {
      this.options.store.appendRawDiagnostic("codex", notification);
      return;
    }
    switch (notification.method) {
      case "turn/started":
        active.nativeTurnId = notification.params.turn.id;
        await active.sink.setNativeTurn(notification.params.turn.id);
        return;
      case "item/started":
        await this.itemStarted(active, notification.params.item);
        return;
      case "item/completed":
        await this.itemCompleted(active, notification.params.item);
        return;
      case "item/agentMessage/delta":
        if (!active.messageIds.has(notification.params.itemId)) {
          active.messageIds.add(notification.params.itemId);
          await active.sink.emit(assistantMessageStart(notification.params.itemId));
        }
        await active.sink.emit(
          assistantMessageDelta(notification.params.itemId, notification.params.delta),
        );
        return;
      case "item/reasoning/summaryTextDelta": {
        const messageId = `reasoning-${notification.params.itemId}`;
        if (!active.reasoningIds.has(messageId)) {
          active.reasoningIds.add(messageId);
          await active.sink.emit(reasoningStart(messageId));
        }
        await active.sink.emit(reasoningDelta(messageId, notification.params.delta));
        return;
      }
      case "item/commandExecution/outputDelta":
        await this.output(active, notification.params.itemId, notification.params.delta);
        return;
      case "item/plan/delta":
        await active.sink.emit(
          activityDelta(`plan-${notification.params.itemId}`, "plan", {
            delta: notification.params.delta,
          }),
        );
        return;
      case "turn/completed":
        this.completeTurn(active, notification.params);
        return;
      default:
        return;
    }
  }

  private async handleServerRequest(
    client: JsonRpcProcess,
    method: string,
    id: string | number,
    params: unknown,
  ): Promise<void> {
    if (method === "item/tool/requestUserInput") {
      const request = params as ToolRequestUserInputParams;
      const active = this.turnsByProviderThread.get(request.threadId);
      if (!active) {
        client.respondError(id, { code: -32_001, message: "Execution is no longer active" });
        return;
      }
      try {
        const answer = await active.sink.requestInput({
          providerRequestId: String(id),
          toolCallId: request.itemId,
          questions: request.questions.map((question) => ({
            id: question.id,
            header: question.header,
            question: question.question,
            options: question.options?.map((option) => option.label) ?? [],
            multiple: false,
            secret: question.isSecret,
            allowOther: question.isOther,
          })),
          providerInput: request,
        });
        client.respond(id, toCodexAnswers(answer));
      } catch (error) {
        client.respondError(id, {
          code: -32_000,
          message: error instanceof Error ? error.message : String(error),
        });
      }
      return;
    }
    if (
      method === "item/commandExecution/requestApproval" ||
      method === "item/fileChange/requestApproval"
    ) {
      client.respond(id, { decision: "accept" });
      return;
    }
    client.respondError(id, { code: -32_601, message: `Unsupported server request ${method}` });
  }

  private async itemStarted(active: ActiveTurn, item: ThreadItem): Promise<void> {
    switch (item.type) {
      case "agentMessage":
        active.messageIds.add(item.id);
        await active.sink.emit(assistantMessageStart(item.id));
        return;
      case "reasoning": {
        const messageId = `reasoning-${item.id}`;
        active.reasoningIds.add(messageId);
        await active.sink.emit(reasoningStart(messageId));
        return;
      }
      case "commandExecution":
        await emitAll(
          active.sink,
          toolCallStart(item.id, "shell", { command: item.command, cwd: item.cwd }),
        );
        active.outputs.set(
          item.id,
          new BoundedOutput(active.sink, item.id, (error) => active.reject(error)),
        );
        return;
      case "fileChange":
        await emitAll(
          active.sink,
          toolCallStart(item.id, "file_change", { changes: item.changes }),
        );
        return;
      case "mcpToolCall":
        await emitAll(
          active.sink,
          toolCallStart(item.id, `${item.server}/${item.tool}`, item.arguments),
        );
        return;
      case "dynamicToolCall":
        await emitAll(active.sink, toolCallStart(item.id, item.tool, item.arguments));
        return;
      case "plan":
        await active.sink.emit(activityDelta(`plan-${item.id}`, "plan", { text: item.text }));
        return;
      default:
        return;
    }
  }

  private async itemCompleted(active: ActiveTurn, item: ThreadItem): Promise<void> {
    switch (item.type) {
      case "agentMessage":
        if (!active.messageIds.has(item.id)) {
          await active.sink.emit(assistantMessageStart(item.id));
          if (item.text) {
            await active.sink.emit(assistantMessageDelta(item.id, item.text));
          }
        }
        await active.sink.emit(assistantMessageEnd(item.id));
        return;
      case "reasoning": {
        const messageId = `reasoning-${item.id}`;
        if (!active.reasoningIds.has(messageId) && item.summary.length > 0) {
          await active.sink.emit(reasoningStart(messageId));
          await active.sink.emit(reasoningDelta(messageId, item.summary.join("\n")));
        }
        await active.sink.emit(reasoningEnd(messageId));
        return;
      }
      case "commandExecution": {
        const output = active.outputs.get(item.id);
        await output?.flush();
        active.outputs.delete(item.id);
        await active.sink.emit(
          toolCallResult(item.id, {
            status: item.status,
            exitCode: item.exitCode,
            output: output?.content ?? item.aggregatedOutput ?? "",
          }),
        );
        return;
      }
      case "fileChange":
        await active.sink.emit(
          toolCallResult(item.id, { status: item.status, changes: item.changes }),
        );
        return;
      case "mcpToolCall":
        await active.sink.emit(
          toolCallResult(item.id, {
            status: item.status,
            result: item.result,
            error: item.error,
          }),
        );
        return;
      case "dynamicToolCall":
        await active.sink.emit(
          toolCallResult(item.id, {
            status: item.status,
            success: item.success,
            content: item.contentItems,
          }),
        );
        return;
      default:
        return;
    }
  }

  private async output(active: ActiveTurn, itemId: string, delta: string): Promise<void> {
    const output =
      active.outputs.get(itemId) ??
      new BoundedOutput(active.sink, itemId, (error) => active.reject(error));
    active.outputs.set(itemId, output);
    await output.append(delta);
  }

  private completeTurn(active: ActiveTurn, notification: TurnCompletedNotification): void {
    if (notification.turn.status === "failed") {
      active.reject(new Error(notification.turn.error?.message ?? "Codex turn failed"));
      return;
    }
    active.resolve();
  }

  private requireActive(executionUid: string): ActiveTurn {
    const active = this.turnsByExecution.get(executionUid);
    if (!active) {
      throw new Error("Codex execution is not active");
    }
    return active;
  }

  private async refreshQuota(client: JsonRpcProcess): Promise<void> {
    const result = await client.request<{
      rateLimits: {
        primary: { usedPercent: number; resetsAt: number | null } | null;
      };
    }>("account/rateLimits/read", undefined);
    this.writeQuota(result.rateLimits);
  }

  private writeQuota(rateLimits: {
    primary: { usedPercent: number; resetsAt: number | null } | null;
  }): void {
    const usedRatio = rateLimits.primary
      ? Math.min(Math.max(rateLimits.primary.usedPercent / 100, 0), 1)
      : null;
    this.options.store.putQuota({
      runtimeUid: CODEX_RUNTIME_UID,
      freshness: "FRESH",
      usedRatio,
      resetTime: rateLimits.primary?.resetsAt ? new Date(rateLimits.primary.resetsAt * 1000) : null,
      observeTime: new Date(),
      source: "codex-app-server",
    });
  }

  private status(
    availability: RuntimeStatus["availability"],
    version: string,
    diagnostic: string,
    updateTime: Date,
  ): RuntimeStatus {
    return {
      uid: CODEX_RUNTIME_UID,
      kind: "codex",
      displayName: "Codex",
      version,
      availability,
      diagnostic,
      capabilities: codexCapabilities(version),
      updateTime,
    };
  }
}

class BoundedOutput {
  private chunks: string[] = [];
  private bytes = 0;
  private pending = "";
  private timer: NodeJS.Timeout | null = null;
  private truncated = false;

  constructor(
    private readonly sink: ProviderSink,
    private readonly toolCallId: string,
    private readonly onError: (error: Error) => void,
  ) {}

  get content(): string {
    return this.chunks.join("");
  }

  async append(delta: string): Promise<void> {
    const remaining = MAX_TOOL_OUTPUT_BYTES - this.bytes;
    if (remaining > 0) {
      const exceeded = Buffer.byteLength(delta) > remaining;
      const bounded = truncateUtf8(delta, remaining);
      this.chunks.push(bounded);
      this.bytes += Buffer.byteLength(bounded);
      this.pending += bounded;
      if (exceeded) {
        this.appendTruncationMarker();
      }
    } else if (!this.truncated) {
      this.appendTruncationMarker();
    }
    if (Buffer.byteLength(this.pending) >= 64 * 1024) {
      await this.flush();
      return;
    }
    if (!this.timer) {
      this.timer = setTimeout(() => {
        this.timer = null;
        void this.flush().catch((error: unknown) => this.onError(asError(error)));
      }, 50);
    }
  }

  async flush(): Promise<void> {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    const delta = this.pending;
    this.pending = "";
    if (!delta) {
      return;
    }
    await this.sink.emit(
      activityDelta(`output-${this.toolCallId}`, "command_output", {
        toolCallId: this.toolCallId,
        delta,
      }),
    );
  }

  private appendTruncationMarker(): void {
    if (this.truncated) {
      return;
    }
    const marker = "\n[output truncated after 4 MiB]\n";
    this.chunks.push(marker);
    this.pending += marker;
    this.truncated = true;
  }
}

function truncateUtf8(value: string, maximumBytes: number): string {
  const bytes = Buffer.from(value);
  if (bytes.length <= maximumBytes) {
    return value;
  }
  let end = maximumBytes;
  while (end > 0 && (bytes[end] ?? 0) >> 6 === 2) {
    end -= 1;
  }
  return bytes.subarray(0, end).toString("utf8");
}

function createActiveTurn(
  run: ProviderRun,
  sink: ProviderSink,
  providerThreadId: string,
): ActiveTurn {
  let resolve: (() => void) | undefined;
  let reject: ((error: Error) => void) | undefined;
  const completion = new Promise<void>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  if (!resolve || !reject) {
    throw new Error("Failed to initialize Codex completion");
  }
  return {
    run,
    sink,
    providerThreadId,
    completion,
    resolve,
    reject,
    messageIds: new Set(),
    reasoningIds: new Set(),
    outputs: new Map(),
    nativeTurnId: null,
  };
}

function findActive(turns: ReadonlyMap<string, ActiveTurn>, params: unknown): ActiveTurn | null {
  if (!params || typeof params !== "object" || !("threadId" in params)) {
    return null;
  }
  return turns.get(String(params.threadId)) ?? null;
}

function toCodexAnswers(answer: InputAnswer): ToolRequestUserInputResponse {
  return {
    answers: Object.fromEntries(
      Object.entries(answer.answers).map(([id, answers]) => [id, { answers: [...answers] }]),
    ),
  };
}

async function emitAll(sink: ProviderSink, events: readonly AGUIEvent[]): Promise<void> {
  for (const event of events) {
    await sink.emit(event);
  }
}

function codexCapabilities(version: string): AgentCapabilities {
  return {
    identity: { name: "Codex", type: "codex-cli", version, provider: "OpenAI" },
    transport: { streaming: true, resumable: true },
    tools: { supported: true, parallelCalls: true, clientProvided: false },
    state: { persistentState: true },
    reasoning: { supported: true, streaming: true, encrypted: false },
    execution: { codeExecution: true, sandboxed: false },
    humanInTheLoop: {
      supported: true,
      approvals: false,
      interventions: true,
      interrupts: true,
      approveWithEdits: false,
    },
  };
}

function isMissingExecutable(error: unknown): boolean {
  return error instanceof Error && "code" in error && error.code === "ENOENT";
}

function asError(error: unknown): Error {
  return error instanceof Error ? error : new Error(String(error));
}
