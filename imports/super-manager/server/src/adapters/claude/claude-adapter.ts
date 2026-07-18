import { type ChildProcessWithoutNullStreams, spawn } from "node:child_process";
import { randomUUID } from "node:crypto";
import { createInterface } from "node:readline";
import { type AgentCapabilities, EventType } from "@ag-ui/core";
import {
  activityDelta,
  assistantMessageDelta,
  assistantMessageEnd,
  assistantMessageStart,
  toolCallResult,
} from "../../application/agui-events.js";
import { redactSensitiveText } from "../../domain/redaction.js";
import {
  CLAUDE_RUNTIME_UID,
  type InputAnswer,
  type ProviderRun,
  type ProviderSink,
  type RuntimeAdapter,
  type RuntimeStatus,
  type StructuredQuestion,
} from "../../domain/runtime.js";
import { runCommand } from "../../infrastructure/processes.js";
import type { ResourceStore } from "../../infrastructure/resource-store.js";
import { claudeArguments, claudeEnvironment } from "./claude-command.js";

interface ClaudeAdapterOptions {
  readonly executable: string;
  readonly expectedVersion: string;
  readonly store: ResourceStore;
}

interface ClaudeMessage {
  readonly type: string;
  readonly [key: string]: unknown;
}

interface ClaudeContentBlock {
  readonly type: string;
  readonly id?: string;
  readonly name?: string;
  readonly text?: string;
  readonly input?: unknown;
  readonly content?: unknown;
  readonly tool_use_id?: string;
}

interface StreamBlock {
  readonly kind: "text" | "tool";
  readonly messageId: string;
  readonly toolCallId?: string;
}

interface ActiveClaudeExecution {
  readonly run: ProviderRun;
  readonly sink: ProviderSink;
  readonly child: ChildProcessWithoutNullStreams;
  readonly completion: Promise<void>;
  readonly resolve: () => void;
  readonly reject: (error: Error) => void;
  readonly blocks: Map<number, StreamBlock>;
  readonly emittedMessages: Set<string>;
  readonly emittedToolCalls: Set<string>;
  readonly knownToolCalls: Array<{ id: string; name: string; input: unknown; used: boolean }>;
  readonly pendingControls: Map<
    string,
    { resolve: () => void; reject: (error: Error) => void; timeout: NodeJS.Timeout }
  >;
  resultReceived: boolean;
  sessionId: string | null;
}

export class ClaudeAdapter implements RuntimeAdapter {
  readonly kind = "claude" as const;
  private readonly active = new Map<string, ActiveClaudeExecution>();

  constructor(private readonly options: ClaudeAdapterOptions) {}

  async probe(): Promise<RuntimeStatus> {
    const updateTime = new Date();
    const environment = claudeEnvironment();
    try {
      const versionResult = await runCommand(this.options.executable, ["--version"], {
        env: environment,
      });
      if (versionResult.exitCode !== 0) {
        return this.status(
          "UNHEALTHY",
          "",
          `Claude Code version check failed with exit code ${versionResult.exitCode}`,
          updateTime,
        );
      }
      const version = versionResult.stdout.trim().split(/\s+/)[0] ?? "";
      if (version !== this.options.expectedVersion) {
        return this.status(
          "INCOMPATIBLE",
          version,
          `Expected Claude Code ${this.options.expectedVersion}`,
          updateTime,
        );
      }
      const authResult = await runCommand(this.options.executable, ["auth", "status", "--json"], {
        env: environment,
      });
      if (authResult.exitCode !== 0) {
        return this.status(
          "NOT_AUTHENTICATED",
          version,
          "Claude Code subscription authentication check failed",
          updateTime,
        );
      }
      const auth = JSON.parse(authResult.stdout) as {
        loggedIn?: boolean;
        authMethod?: string;
        apiProvider?: string;
        subscriptionType?: string;
      };
      if (
        auth.loggedIn !== true ||
        auth.authMethod !== "claude.ai" ||
        auth.apiProvider !== "firstParty" ||
        !auth.subscriptionType
      ) {
        return this.status(
          "NOT_AUTHENTICATED",
          version,
          "Claude Code must use a first-party claude.ai subscription login",
          updateTime,
        );
      }
      const controlResult = await runCommand(this.options.executable, claudeArguments(), {
        env: environment,
        timeoutMs: 10_000,
      });
      if (controlResult.exitCode !== 0) {
        return this.status(
          "INCOMPATIBLE",
          version,
          "Claude Code hidden stdio control is unavailable",
          updateTime,
        );
      }
      return this.status("AVAILABLE", version, "", updateTime);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      return this.status(
        isMissingExecutable(error) ? "NOT_INSTALLED" : "UNHEALTHY",
        "",
        redactSensitiveText(message, 512),
        updateTime,
      );
    }
  }

  async execute(run: ProviderRun, sink: ProviderSink): Promise<void> {
    const child = spawn(this.options.executable, claudeArguments(run.providerSessionId), {
      cwd: run.workspacePath,
      env: claudeEnvironment(),
      shell: false,
      stdio: ["pipe", "pipe", "pipe"],
    });
    const active = createActiveExecution(run, sink, child);
    this.active.set(run.executionUid, active);
    let stderrBytes = 0;
    child.stderr.on("data", (chunk: Buffer) => {
      stderrBytes = Math.min(Number.MAX_SAFE_INTEGER, stderrBytes + chunk.length);
    });
    const lines = createInterface({ input: child.stdout, crlfDelay: Number.POSITIVE_INFINITY });
    let processing = Promise.resolve();
    lines.on("line", (line) => {
      processing = processing
        .then(() => this.handleLine(active, line))
        .catch((error) => {
          active.reject(asError(error));
          if (!child.killed) {
            child.kill("SIGTERM");
          }
        });
    });
    child.once("error", (error) => active.reject(error));
    child.once("close", (code, signal) => {
      void processing.then(() => {
        if (active.resultReceived) {
          return;
        }
        active.reject(
          new Error(
            `Claude Code exited with code ${code ?? "unknown"} and signal ${signal ?? "none"}${stderrBytes > 0 ? ` (${stderrBytes} stderr bytes)` : ""}`,
          ),
        );
      });
    });

    child.stdin.write(
      `${JSON.stringify({
        type: "user",
        message: { role: "user", content: run.prompt },
        parent_tool_use_id: null,
        ...(run.providerSessionId ? { session_id: run.providerSessionId } : {}),
      })}\n`,
    );

    try {
      await active.completion;
    } finally {
      this.active.delete(run.executionUid);
      for (const pending of active.pendingControls.values()) {
        clearTimeout(pending.timeout);
        pending.reject(new Error("Claude Code process ended"));
      }
      active.pendingControls.clear();
      if (child.stdin.writable) {
        child.stdin.end();
      }
      if (!child.killed) {
        child.kill("SIGTERM");
      }
    }
  }

  async cancel(executionUid: string): Promise<void> {
    const active = this.requireActive(executionUid);
    await sendControlRequest(active, { subtype: "interrupt" });
  }

  async steer(_executionUid: string, _instruction: string): Promise<void> {
    throw new Error("Claude Code stdio control does not declare steering support");
  }

  async close(): Promise<void> {
    const executions = [...this.active.values()];
    for (const active of executions) {
      active.child.kill("SIGTERM");
    }
    if (executions.length === 0) {
      return;
    }
    const completions = Promise.allSettled(executions.map((active) => active.completion));
    let timeout: NodeJS.Timeout | undefined;
    const stopped = await Promise.race([
      completions.then(() => true),
      new Promise<false>((resolve) => {
        timeout = setTimeout(() => resolve(false), 5_000);
      }),
    ]);
    if (timeout) {
      clearTimeout(timeout);
    }
    if (!stopped) {
      for (const active of executions) {
        active.child.kill("SIGKILL");
      }
      await completions;
    }
  }

  private async handleLine(active: ActiveClaudeExecution, line: string): Promise<void> {
    if (!line.trim()) {
      return;
    }
    if (Buffer.byteLength(line) > 8 * 1024 * 1024) {
      active.reject(new Error("Claude Code frame exceeded 8 MiB"));
      active.child.kill("SIGKILL");
      return;
    }
    let message: ClaudeMessage;
    try {
      message = JSON.parse(line) as ClaudeMessage;
    } catch {
      this.options.store.appendRawDiagnostic("claude-invalid-json", { line: line.slice(0, 4096) });
      return;
    }
    switch (message.type) {
      case "control_response":
        handleControlResponse(active, message);
        return;
      case "control_request":
        void this.handlePermissionRequest(active, message).catch((error) => {
          active.reject(asError(error));
        });
        return;
      case "control_cancel_request":
        this.options.store.appendRawDiagnostic("claude-control-cancel", message);
        return;
      case "system":
        await this.handleSystem(active, message);
        return;
      case "stream_event":
        await this.handleStreamEvent(active, message);
        return;
      case "assistant":
        await this.handleAssistant(active, message);
        return;
      case "user":
        await this.handleUser(active, message);
        return;
      case "tool_progress":
        await active.sink.emit(
          activityDelta(
            `progress-${String(message.tool_use_id ?? randomUUID())}`,
            "tool_progress",
            {
              toolCallId: message.tool_use_id,
              toolName: message.tool_name,
              elapsedSeconds: message.elapsed_time_seconds,
            },
          ),
        );
        return;
      case "rate_limit_event":
        this.writeQuota(message);
        return;
      case "result":
        await this.handleResult(active, message);
        return;
      default:
        this.options.store.appendRawDiagnostic("claude", message);
    }
  }

  private async handleSystem(active: ActiveClaudeExecution, message: ClaudeMessage): Promise<void> {
    if (message.subtype !== "init" || typeof message.session_id !== "string") {
      return;
    }
    active.sessionId = message.session_id;
    if (active.run.providerSessionId !== message.session_id) {
      await active.sink.setProviderSession(message.session_id);
    }
  }

  private async handleStreamEvent(
    active: ActiveClaudeExecution,
    message: ClaudeMessage,
  ): Promise<void> {
    const event = asObject(message.event);
    if (!event || typeof event.type !== "string") {
      return;
    }
    const index = typeof event.index === "number" ? event.index : 0;
    const streamId = String(message.uuid ?? message.session_id ?? active.run.executionUid);
    if (event.type === "content_block_start") {
      const content = asObject(event.content_block);
      if (!content || typeof content.type !== "string") {
        return;
      }
      if (content.type === "text") {
        const messageId = `${streamId}-text-${index}`;
        active.blocks.set(index, { kind: "text", messageId });
        active.emittedMessages.add(messageId);
        await active.sink.emit(assistantMessageStart(messageId));
        if (typeof content.text === "string" && content.text) {
          await active.sink.emit(assistantMessageDelta(messageId, content.text));
        }
        return;
      }
      if (content.type === "tool_use" && typeof content.id === "string") {
        const name = typeof content.name === "string" ? content.name : "tool";
        active.blocks.set(index, {
          kind: "tool",
          messageId: `${streamId}-tool-${index}`,
          toolCallId: content.id,
        });
        active.emittedToolCalls.add(content.id);
        active.knownToolCalls.push({
          id: content.id,
          name,
          input: content.input ?? {},
          used: false,
        });
        await active.sink.emit({
          type: EventType.TOOL_CALL_START,
          toolCallId: content.id,
          toolCallName: name,
        });
      }
      return;
    }
    const block = active.blocks.get(index);
    if (!block) {
      return;
    }
    if (event.type === "content_block_delta") {
      const delta = asObject(event.delta);
      if (!delta) {
        return;
      }
      if (block.kind === "text" && typeof delta.text === "string") {
        await active.sink.emit(assistantMessageDelta(block.messageId, delta.text));
      } else if (
        block.kind === "tool" &&
        block.toolCallId &&
        typeof delta.partial_json === "string"
      ) {
        await active.sink.emit({
          type: EventType.TOOL_CALL_ARGS,
          toolCallId: block.toolCallId,
          delta: delta.partial_json,
        });
      }
      return;
    }
    if (event.type === "content_block_stop") {
      if (block.kind === "text") {
        await active.sink.emit(assistantMessageEnd(block.messageId));
      } else if (block.toolCallId) {
        await active.sink.emit({ type: EventType.TOOL_CALL_END, toolCallId: block.toolCallId });
      }
      active.blocks.delete(index);
    }
  }

  private async handleAssistant(
    active: ActiveClaudeExecution,
    envelope: ClaudeMessage,
  ): Promise<void> {
    const message = asObject(envelope.message);
    const content = Array.isArray(message?.content)
      ? (message.content as ClaudeContentBlock[])
      : [];
    const envelopeId = String(envelope.uuid ?? message?.id ?? randomUUID());
    for (const [index, block] of content.entries()) {
      if (block.type === "tool_use" && block.id && block.name) {
        if (!active.knownToolCalls.some((tool) => tool.id === block.id)) {
          active.knownToolCalls.push({
            id: block.id,
            name: block.name,
            input: block.input ?? {},
            used: false,
          });
        }
        if (!active.emittedToolCalls.has(block.id)) {
          active.emittedToolCalls.add(block.id);
          await active.sink.emit({
            type: EventType.TOOL_CALL_START,
            toolCallId: block.id,
            toolCallName: block.name,
          });
          await active.sink.emit({
            type: EventType.TOOL_CALL_ARGS,
            toolCallId: block.id,
            delta: JSON.stringify(block.input ?? {}),
          });
          await active.sink.emit({ type: EventType.TOOL_CALL_END, toolCallId: block.id });
        }
      } else if (block.type === "text" && block.text) {
        const messageId = `${envelopeId}-text-${index}`;
        if (!active.emittedMessages.has(messageId)) {
          active.emittedMessages.add(messageId);
          await active.sink.emit(assistantMessageStart(messageId));
          await active.sink.emit(assistantMessageDelta(messageId, block.text));
          await active.sink.emit(assistantMessageEnd(messageId));
        }
      }
    }
  }

  private async handleUser(active: ActiveClaudeExecution, envelope: ClaudeMessage): Promise<void> {
    const message = asObject(envelope.message);
    const content = Array.isArray(message?.content)
      ? (message.content as ClaudeContentBlock[])
      : [];
    for (const block of content) {
      if (block.type === "tool_result" && block.tool_use_id) {
        const tool = active.knownToolCalls.find((candidate) => candidate.id === block.tool_use_id);
        const content =
          tool && isQuestionTool(tool.name) ? "Structured answer submitted" : block.content;
        await active.sink.emit(toolCallResult(block.tool_use_id, content ?? ""));
      }
    }
  }

  private async handlePermissionRequest(
    active: ActiveClaudeExecution,
    envelope: ClaudeMessage,
  ): Promise<void> {
    const requestId = String(envelope.request_id ?? "");
    const request = asObject(envelope.request);
    if (!requestId || request?.subtype !== "can_use_tool") {
      sendControlError(active, requestId, "Unsupported control request");
      return;
    }
    const toolName = String(request.tool_name ?? "");
    const input = request.input ?? {};
    if (!isQuestionTool(toolName)) {
      sendControlSuccess(active, requestId, { behavior: "allow", updatedInput: input });
      return;
    }
    const toolCallId =
      (typeof request.tool_use_id === "string" ? request.tool_use_id : null) ??
      consumeMatchingToolCall(active, toolName, input) ??
      `question-${requestId}`;
    try {
      const questions = parseQuestions(input);
      const answer = await active.sink.requestInput({
        providerRequestId: requestId,
        toolCallId,
        questions,
        providerInput: { toolName, input },
      });
      if (Object.keys(answer.answers).length === 0) {
        sendControlSuccess(active, requestId, {
          behavior: "deny",
          message: "No answers were provided.",
        });
        return;
      }
      sendControlSuccess(active, requestId, {
        behavior: "allow",
        updatedInput: buildUpdatedInput(toolName, input, questions, answer),
      });
    } catch (error) {
      sendControlError(active, requestId, error instanceof Error ? error.message : String(error));
    }
  }

  private async handleResult(active: ActiveClaudeExecution, message: ClaudeMessage): Promise<void> {
    active.resultReceived = true;
    if (typeof message.session_id === "string" && active.sessionId !== message.session_id) {
      active.sessionId = message.session_id;
      await active.sink.setProviderSession(message.session_id);
    }
    if (message.is_error === true || String(message.subtype ?? "").startsWith("error")) {
      active.reject(new Error(String(message.result ?? message.subtype ?? "Claude Code failed")));
      return;
    }
    active.resolve();
  }

  private writeQuota(message: ClaudeMessage): void {
    const info = asObject(message.rate_limit_info ?? message.rateLimitInfo ?? message);
    const rawRatio = numberValue(info?.utilization ?? info?.used_ratio ?? info?.usedPercent);
    const normalizedRatio = rawRatio === null ? null : rawRatio > 1 ? rawRatio / 100 : rawRatio;
    const usedRatio = normalizedRatio === null ? null : Math.min(Math.max(normalizedRatio, 0), 1);
    const reset = numberValue(info?.resets_at ?? info?.reset_time ?? info?.resetsAt);
    this.options.store.putQuota({
      runtimeUid: CLAUDE_RUNTIME_UID,
      freshness: "FRESH",
      usedRatio,
      resetTime: reset === null ? null : new Date(reset > 10_000_000_000 ? reset : reset * 1000),
      observeTime: new Date(),
      source: "claude-stream-json",
    });
  }

  private requireActive(executionUid: string): ActiveClaudeExecution {
    const active = this.active.get(executionUid);
    if (!active) {
      throw new Error("Claude Code execution is not active");
    }
    return active;
  }

  private status(
    availability: RuntimeStatus["availability"],
    version: string,
    diagnostic: string,
    updateTime: Date,
  ): RuntimeStatus {
    return {
      uid: CLAUDE_RUNTIME_UID,
      kind: "claude",
      displayName: "Claude Code",
      version,
      availability,
      diagnostic,
      capabilities: claudeCapabilities(version),
      updateTime,
    };
  }
}

function createActiveExecution(
  run: ProviderRun,
  sink: ProviderSink,
  child: ChildProcessWithoutNullStreams,
): ActiveClaudeExecution {
  let resolve: (() => void) | undefined;
  let reject: ((error: Error) => void) | undefined;
  const completion = new Promise<void>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  if (!resolve || !reject) {
    throw new Error("Failed to initialize Claude completion");
  }
  return {
    run,
    sink,
    child,
    completion,
    resolve,
    reject,
    blocks: new Map(),
    emittedMessages: new Set(),
    emittedToolCalls: new Set(),
    knownToolCalls: [],
    pendingControls: new Map(),
    resultReceived: false,
    sessionId: run.providerSessionId,
  };
}

function parseQuestions(input: unknown): readonly StructuredQuestion[] {
  const object = asObject(input);
  if (!Array.isArray(object?.questions) || object.questions.length === 0) {
    throw new Error("Question tool input did not contain questions");
  }
  return object.questions.map((value, index) => {
    const question = asObject(value);
    if (!question || typeof question.question !== "string") {
      throw new Error("Question tool input was malformed");
    }
    const options = Array.isArray(question.options)
      ? question.options
          .map((option) => asObject(option)?.label)
          .filter((label): label is string => typeof label === "string")
      : [];
    return {
      id: typeof question.id === "string" ? question.id : String(index),
      header:
        typeof question.header === "string" && question.header
          ? question.header
          : `Question ${index + 1}`,
      question: question.question,
      options,
      multiple: question.multiSelect === true || question.multiple === true,
      secret: question.isSecret === true,
      allowOther: question.isOther !== false,
    };
  });
}

function buildUpdatedInput(
  toolName: string,
  input: unknown,
  questions: readonly StructuredQuestion[],
  answer: InputAnswer,
): Record<string, unknown> {
  const original = asObject(input) ?? {};
  if (toolName === "request_user_input") {
    return {
      ...original,
      answers: Object.fromEntries(
        Object.entries(answer.answers).map(([id, answers]) => [id, { answers: [...answers] }]),
      ),
    };
  }
  return {
    ...original,
    answers: Object.fromEntries(
      questions.flatMap((question) => {
        const values = answer.answers[question.id];
        return values && values.length > 0 ? [[question.question, values.join(",")]] : [];
      }),
    ),
  };
}

function isQuestionTool(name: string): boolean {
  return ["AskUserQuestion", "ask_user_question", "request_user_input"].includes(name);
}

function consumeMatchingToolCall(
  active: ActiveClaudeExecution,
  name: string,
  input: unknown,
): string | null {
  const serialized = JSON.stringify(input);
  for (let index = active.knownToolCalls.length - 1; index >= 0; index -= 1) {
    const tool = active.knownToolCalls[index];
    if (tool && !tool.used && tool.name === name && JSON.stringify(tool.input) === serialized) {
      tool.used = true;
      return tool.id;
    }
  }
  return null;
}

function sendControlSuccess(
  active: ActiveClaudeExecution,
  requestId: string,
  response: unknown,
): void {
  writeClaude(active, {
    type: "control_response",
    response: { subtype: "success", request_id: requestId, response },
  });
}

function sendControlError(active: ActiveClaudeExecution, requestId: string, error: string): void {
  writeClaude(active, {
    type: "control_response",
    response: { subtype: "error", request_id: requestId, error },
  });
}

function sendControlRequest(
  active: ActiveClaudeExecution,
  request: Record<string, unknown>,
): Promise<void> {
  const requestId = randomUUID();
  const { promise, resolve, reject } = Promise.withResolvers<void>();
  const timeout = setTimeout(() => {
    active.pendingControls.delete(requestId);
    reject(new Error("Claude Code control request timed out"));
  }, 10_000);
  const pending = { resolve, reject, timeout };
  active.pendingControls.set(requestId, pending);
  try {
    writeClaude(active, { type: "control_request", request_id: requestId, request });
  } catch (error) {
    clearTimeout(pending.timeout);
    active.pendingControls.delete(requestId);
    pending.reject(asError(error));
  }
  return promise;
}

function handleControlResponse(active: ActiveClaudeExecution, message: ClaudeMessage): void {
  const response = asObject(message.response);
  const requestId = typeof response?.request_id === "string" ? response.request_id : "";
  const pending = active.pendingControls.get(requestId);
  if (!pending) {
    return;
  }
  clearTimeout(pending.timeout);
  active.pendingControls.delete(requestId);
  if (response?.subtype === "success") {
    pending.resolve();
  } else {
    pending.reject(new Error(String(response?.error ?? "Claude control request failed")));
  }
}

function writeClaude(active: ActiveClaudeExecution, message: unknown): void {
  if (!active.child.stdin.writable) {
    throw new Error("Claude Code stdin is not writable");
  }
  active.child.stdin.write(`${JSON.stringify(message)}\n`);
}

function asObject(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function numberValue(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function claudeCapabilities(version: string): AgentCapabilities {
  return {
    identity: {
      name: "Claude Code",
      type: "claude-code-cli",
      version,
      provider: "Anthropic",
    },
    transport: { streaming: true, resumable: true },
    tools: { supported: true, parallelCalls: true, clientProvided: false },
    state: { persistentState: true },
    reasoning: { supported: false, streaming: false, encrypted: false },
    execution: { codeExecution: true, sandboxed: false },
    humanInTheLoop: {
      supported: true,
      approvals: false,
      interventions: false,
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
