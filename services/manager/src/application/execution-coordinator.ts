import { createHash } from "node:crypto";
import { statfs } from "node:fs/promises";
import {
  type AGUIEvent,
  EventSchemas,
  EventType,
  type InputContent,
  type RunAgentInput,
  RunAgentInputSchema,
} from "@ag-ui/core";
import type { ExecutionRecord, PendingInterruptRecord } from "../domain/records.js";
import { redactSensitiveText } from "../domain/redaction.js";
import type {
  InputAnswer,
  InputRequest,
  ProviderSink,
  RuntimeAdapter,
  StructuredQuestion,
} from "../domain/runtime.js";
import type { ResourceStore } from "../infrastructure/resource-store.js";
import type { EventStream } from "./event-stream.js";
import type { RuntimeRegistry } from "./runtime-registry.js";

const MAX_PROMPT_BYTES = 128 * 1024;
const MAX_PENDING_INPUT_BYTES = 256 * 1024;
const MAX_PAUSED_EVENT_BYTES = 4 * 1024 * 1024;
const MAX_PROVIDER_EVENT_BYTES = 1024 * 1024;
const MAX_EXECUTION_EVENT_BYTES = 64 * 1024 * 1024;
const MINIMUM_FREE_BYTES = 2 * 1024 * 1024 * 1024;

interface PendingInput {
  readonly record: PendingInterruptRecord;
  readonly questions: readonly StructuredQuestion[];
  readonly resolve: (answer: InputAnswer) => void;
  readonly reject: (error: Error) => void;
  announced: boolean;
  announcedRunId: string | null;
}

interface ActiveExecution {
  readonly execution: ExecutionRecord;
  readonly adapter: RuntimeAdapter;
  readonly pending: Map<string, PendingInput>;
  readonly pausedEvents: AGUIEvent[];
  currentRunId: string;
  pausedEventBytes: number;
  providerEventBytes: number;
  providerOutputTruncated: boolean;
  paused: boolean;
  canceling: boolean;
  deferredCompletion: { readonly error?: unknown } | null;
  terminal: boolean;
  interruptTimer: NodeJS.Timeout | null;
}

export interface StartRunResult {
  readonly threadUid: string;
  readonly runId: string;
  readonly existing: boolean;
}

export class ExecutionCoordinator {
  private readonly active = new Map<string, ActiveExecution>();
  private closing = false;

  constructor(
    private readonly store: ResourceStore,
    private readonly events: EventStream,
    private readonly runtimes: RuntimeRegistry,
    private readonly dataDirectory: string,
  ) {}

  async start(inputValue: unknown): Promise<StartRunResult> {
    if (this.closing) {
      throw new Error("Server is shutting down");
    }
    const input = RunAgentInputSchema.parse(inputValue);
    validateRunInput(input);
    const priorRun = this.store.getRun(input.runId);
    if (priorRun) {
      if (priorRun.threadUid !== input.threadId) {
        throw new Error("runId already belongs to another thread");
      }
      return { threadUid: priorRun.threadUid, runId: priorRun.runId, existing: true };
    }
    if (input.resume && input.resume.length > 0) {
      return this.resume(input);
    }
    return this.startExecution(input);
  }

  async cancel(executionUid: string): Promise<ExecutionRecord> {
    const context = this.active.get(executionUid);
    if (!context || context.terminal || context.canceling) {
      throw new Error("Execution is not active");
    }
    context.canceling = true;
    try {
      await context.adapter.cancel(executionUid);
    } catch (error) {
      context.canceling = false;
      const deferred = context.deferredCompletion;
      context.deferredCompletion = null;
      if (deferred) {
        if ("error" in deferred) {
          this.finishError(context, deferred.error);
        } else {
          this.finishSuccess(context);
        }
      }
      throw error;
    }
    context.canceling = false;
    context.deferredCompletion = null;
    this.finishCanceled(context);
    return this.store.getExecution(executionUid) ?? context.execution;
  }

  async steer(executionUid: string, instruction: string): Promise<ExecutionRecord> {
    const context = this.active.get(executionUid);
    if (!context || context.terminal || context.canceling) {
      throw new Error("Execution is not active");
    }
    await context.adapter.steer(executionUid, instruction);
    if (context.terminal || this.active.get(executionUid) !== context) {
      return this.store.getExecution(executionUid) ?? context.execution;
    }
    this.events.append(context.execution.threadUid, context.currentRunId, {
      type: EventType.CUSTOM,
      name: "realtime.me.manager.steer",
      value: { instruction },
    });
    return this.store.getExecution(executionUid) ?? context.execution;
  }

  shutdown(): void {
    if (this.closing) {
      return;
    }
    this.closing = true;
    for (const context of this.active.values()) {
      context.terminal = true;
      if (context.interruptTimer) {
        clearTimeout(context.interruptTimer);
        context.interruptTimer = null;
      }
      for (const pending of context.pending.values()) {
        pending.reject(new Error("Server is shutting down"));
      }
      context.pending.clear();
    }
    this.active.clear();
  }

  private async startExecution(input: RunAgentInput): Promise<StartRunResult> {
    await this.assertDiskCapacity();
    if (this.store.countActiveExecutions() >= 2) {
      throw new Error("The global structured execution limit has been reached");
    }
    const thread = this.store.getThread(input.threadId);
    if (!thread) {
      throw new Error("Thread not found");
    }
    const workspace = this.store.getWorkspace(thread.workspaceUid);
    if (!workspace) {
      throw new Error("Workspace not found");
    }
    if (workspace.activeExecutionUid) {
      throw new Error("Workspace already has an active structured execution");
    }
    const prompt = extractPrompt(input);
    if (Buffer.byteLength(prompt) > MAX_PROMPT_BYTES) {
      throw new Error("Prompt exceeds 128 KiB");
    }
    const adapter = this.runtimes.requireAvailable(thread.runtimeUid);
    const execution = this.store.createExecution(thread, input.runId);
    const context: ActiveExecution = {
      execution,
      adapter,
      pending: new Map(),
      pausedEvents: [],
      currentRunId: input.runId,
      pausedEventBytes: 0,
      providerEventBytes: 0,
      providerOutputTruncated: false,
      paused: false,
      canceling: false,
      deferredCompletion: null,
      terminal: false,
      interruptTimer: null,
    };
    try {
      this.store.createRun(input.runId, execution.uid, thread.uid, input.parentRunId ?? null);
      this.events.append(
        thread.uid,
        input.runId,
        {
          type: EventType.RUN_STARTED,
          threadId: thread.uid,
          runId: input.runId,
          ...(input.parentRunId ? { parentRunId: input.parentRunId } : {}),
          input,
        },
        { threadState: "RUNNING", executionUid: execution.uid, executionState: "RUNNING" },
      );
    } catch (error) {
      this.store.deleteExecution(execution.uid);
      this.store.setThreadState(thread.uid, "IDLE");
      throw error;
    }
    this.active.set(execution.uid, context);
    const sink = this.createSink(context);
    void adapter
      .execute(
        {
          executionUid: execution.uid,
          threadUid: thread.uid,
          workspacePath: workspace.path,
          providerSessionId: thread.providerSessionId,
          prompt,
        },
        sink,
      )
      .then(() => this.finishSuccess(context))
      .catch((error: unknown) => this.finishError(context, error));
    return { threadUid: thread.uid, runId: input.runId, existing: false };
  }

  private async resume(input: RunAgentInput): Promise<StartRunResult> {
    const execution = this.store.getActiveExecutionForThread(input.threadId);
    if (!execution) {
      throw new Error("No live execution is waiting for this resume");
    }
    const context = this.active.get(execution.uid);
    if (!context || context.terminal || !context.paused) {
      throw new Error("The provider callback is no longer live");
    }
    const pending = [...context.pending.values()].filter((item) => item.announced);
    if (pending.length === 0) {
      throw new Error("No structured input is pending");
    }
    const parentRunId = pending[0]?.announcedRunId;
    if (
      !parentRunId ||
      pending.some((item) => item.announcedRunId !== parentRunId) ||
      input.parentRunId !== parentRunId
    ) {
      throw new Error("parentRunId does not match the interrupted run");
    }
    const entries = new Map(input.resume?.map((entry) => [entry.interruptId, entry]) ?? []);
    if (
      input.resume?.length !== entries.size ||
      entries.size !== pending.length ||
      pending.some((item) => !entries.has(item.record.uid))
    ) {
      throw new Error("Resume must answer every pending interrupt exactly once");
    }
    const resolutions = pending.map((item) => {
      const entry = entries.get(item.record.uid);
      if (!entry) {
        throw new Error("Pending interrupt disappeared during resume");
      }
      verifyInputHash(item.record);
      return entry.status === "cancelled"
        ? ({ item, status: "cancelled" } as const)
        : ({
            item,
            status: "resolved",
            answer: parseInputAnswer(entry.payload, item.questions),
          } as const);
    });
    this.store.createRun(input.runId, execution.uid, input.threadId, parentRunId);
    try {
      this.events.append(
        input.threadId,
        input.runId,
        {
          type: EventType.RUN_STARTED,
          threadId: input.threadId,
          runId: input.runId,
          parentRunId,
        },
        { threadState: "RUNNING", executionUid: execution.uid, executionState: "RUNNING" },
      );
    } catch (error) {
      this.store.deleteRun(input.runId);
      throw error;
    }
    context.currentRunId = input.runId;
    context.paused = false;
    await this.flushPausedEvents(context);

    for (const resolution of resolutions) {
      const { item } = resolution;
      if (resolution.status === "cancelled") {
        this.store.setInterruptState(item.record.uid, "CANCELED");
        context.pending.delete(item.record.uid);
        item.reject(new Error("User canceled structured input"));
        continue;
      }
      this.store.setInterruptState(item.record.uid, "RESOLVED");
      context.pending.delete(item.record.uid);
      item.resolve(resolution.answer);
    }
    if (context.pending.size > 0) {
      this.scheduleInterrupt(context);
    }
    return { threadUid: input.threadId, runId: input.runId, existing: false };
  }

  private createSink(context: ActiveExecution): ProviderSink {
    return {
      emit: async (event) => {
        if (context.terminal) {
          return;
        }
        if (context.paused) {
          this.bufferPausedEvent(context, event);
          return;
        }
        this.appendProviderEvent(context, event);
      },
      setProviderSession: async (providerSessionId) => {
        this.store.setThreadProviderSession(context.execution.threadUid, providerSessionId);
      },
      setNativeTurn: async (nativeTurnId) => {
        this.store.setExecutionNativeTurn(context.execution.uid, nativeTurnId);
      },
      requestInput: (request) => this.requestInput(context, request),
    };
  }

  private requestInput(context: ActiveExecution, request: InputRequest): Promise<InputAnswer> {
    if (context.terminal) {
      return Promise.reject(new Error("Execution is no longer active"));
    }
    const questions = validateQuestions(request.questions);
    if (!request.providerRequestId || request.providerRequestId.length > 512) {
      return Promise.reject(new Error("Provider request id is invalid"));
    }
    if (!request.toolCallId || request.toolCallId.length > 512) {
      return Promise.reject(new Error("Tool call id is invalid"));
    }
    const providerInput = { providerInput: request.providerInput, questions };
    if (Buffer.byteLength(JSON.stringify(providerInput)) > MAX_PENDING_INPUT_BYTES) {
      return Promise.reject(new Error("Structured input exceeds 256 KiB"));
    }
    const record = this.store.createInterrupt({
      executionUid: context.execution.uid,
      threadUid: context.execution.threadUid,
      runId: context.currentRunId,
      providerRequestId: request.providerRequestId,
      toolCallId: request.toolCallId,
      providerInput,
      inputHash: hashInput(providerInput),
    });
    const result = new Promise<InputAnswer>((resolve, reject) => {
      context.pending.set(record.uid, {
        record,
        questions,
        resolve,
        reject,
        announced: false,
        announcedRunId: null,
      });
    });
    if (!context.paused) {
      this.scheduleInterrupt(context);
    }
    return result;
  }

  private scheduleInterrupt(context: ActiveExecution): void {
    if (context.interruptTimer) {
      clearTimeout(context.interruptTimer);
    }
    context.interruptTimer = setTimeout(() => {
      context.interruptTimer = null;
      this.finishInterruptedRun(context);
    }, 100);
  }

  private finishInterruptedRun(context: ActiveExecution): void {
    if (context.terminal || context.paused || context.pending.size === 0) {
      return;
    }
    context.paused = true;
    const pending = [...context.pending.values()].filter((item) => !item.announced);
    if (pending.length === 0) {
      return;
    }
    for (const item of pending) {
      item.announced = true;
      item.announcedRunId = context.currentRunId;
    }
    const interrupts = pending.map((pending) => ({
      id: pending.record.uid,
      reason: "input_required",
      message: pending.questions.map((question) => question.question).join("\n"),
      toolCallId: pending.record.toolCallId,
      responseSchema: questionResponseSchema(pending.questions),
      metadata: { questions: pending.questions },
    }));
    this.events.append(
      context.execution.threadUid,
      context.currentRunId,
      {
        type: EventType.RUN_FINISHED,
        threadId: context.execution.threadUid,
        runId: context.currentRunId,
        outcome: { type: "interrupt", interrupts },
      },
      {
        threadState: "INPUT_REQUIRED",
        executionUid: context.execution.uid,
        executionState: "INPUT_REQUIRED",
      },
    );
    this.store.setRunState(context.currentRunId, "INTERRUPTED");
  }

  private finishSuccess(context: ActiveExecution): void {
    if (context.terminal) {
      return;
    }
    if (context.canceling) {
      context.deferredCompletion = {};
      return;
    }
    if (context.pending.size > 0) {
      this.finishError(context, new Error("Provider completed while structured input was pending"));
      return;
    }
    context.terminal = true;
    this.events.append(
      context.execution.threadUid,
      context.currentRunId,
      {
        type: EventType.RUN_FINISHED,
        threadId: context.execution.threadUid,
        runId: context.currentRunId,
        outcome: { type: "success" },
      },
      {
        threadState: "IDLE",
        executionUid: context.execution.uid,
        executionState: "SUCCEEDED",
      },
    );
    this.store.setRunState(context.currentRunId, "SUCCEEDED");
    this.active.delete(context.execution.uid);
  }

  private finishError(context: ActiveExecution, error: unknown): void {
    if (context.terminal) {
      return;
    }
    if (context.canceling) {
      context.deferredCompletion = { error };
      return;
    }
    context.terminal = true;
    if (context.interruptTimer) {
      clearTimeout(context.interruptTimer);
    }
    const message = boundedError(error);
    const lost = /exited|closed|process|stdin|stdout/i.test(message);
    for (const pending of context.pending.values()) {
      this.store.setInterruptState(pending.record.uid, lost ? "LOST" : "CANCELED");
      pending.reject(new Error(message));
    }
    context.pending.clear();
    this.events.append(
      context.execution.threadUid,
      context.currentRunId,
      {
        type: EventType.RUN_ERROR,
        message,
        code: lost ? "PROVIDER_LOST" : "PROVIDER_FAILED",
      },
      {
        threadState: lost ? "LOST" : "IDLE",
        executionUid: context.execution.uid,
        executionState: lost ? "LOST" : "FAILED",
      },
    );
    this.store.setRunState(context.currentRunId, "FAILED");
    this.active.delete(context.execution.uid);
  }

  private finishCanceled(context: ActiveExecution): void {
    if (context.terminal) {
      return;
    }
    context.terminal = true;
    if (context.interruptTimer) {
      clearTimeout(context.interruptTimer);
      context.interruptTimer = null;
    }
    for (const pending of context.pending.values()) {
      this.store.setInterruptState(pending.record.uid, "CANCELED");
      pending.reject(new Error("Execution canceled"));
    }
    context.pending.clear();
    this.events.append(
      context.execution.threadUid,
      context.currentRunId,
      { type: EventType.RUN_ERROR, message: "Execution canceled", code: "CANCELED" },
      {
        threadState: "IDLE",
        executionUid: context.execution.uid,
        executionState: "CANCELED",
      },
    );
    this.store.setRunState(context.currentRunId, "CANCELED");
    this.active.delete(context.execution.uid);
  }

  private bufferPausedEvent(context: ActiveExecution, event: AGUIEvent): void {
    const normalized = EventSchemas.parse(event);
    const bytes = Buffer.byteLength(JSON.stringify(normalized));
    if (context.pausedEventBytes + bytes > MAX_PAUSED_EVENT_BYTES) {
      if (context.pausedEvents.at(-1)?.type !== EventType.ACTIVITY_SNAPSHOT) {
        context.pausedEvents.push({
          type: EventType.ACTIVITY_SNAPSHOT,
          messageId: `paused-output-${context.execution.uid}`,
          activityType: "output_truncated",
          content: { message: "Output while waiting for input exceeded 4 MiB" },
          replace: false,
        });
      }
      return;
    }
    context.pausedEvents.push(normalized);
    context.pausedEventBytes += bytes;
  }

  private async flushPausedEvents(context: ActiveExecution): Promise<void> {
    for (const event of context.pausedEvents) {
      this.appendProviderEvent(context, event);
    }
    context.pausedEvents.length = 0;
    context.pausedEventBytes = 0;
  }

  private appendProviderEvent(context: ActiveExecution, event: AGUIEvent): void {
    const bytes = Buffer.byteLength(JSON.stringify(event));
    if (
      bytes > MAX_PROVIDER_EVENT_BYTES ||
      context.providerEventBytes + bytes > MAX_EXECUTION_EVENT_BYTES
    ) {
      if (!context.providerOutputTruncated) {
        context.providerOutputTruncated = true;
        this.events.append(context.execution.threadUid, context.currentRunId, {
          type: EventType.ACTIVITY_SNAPSHOT,
          messageId: `provider-output-${context.execution.uid}`,
          activityType: "output_truncated",
          content: { message: "Provider output exceeded the execution storage limit" },
          replace: false,
        });
      }
      return;
    }
    context.providerEventBytes += bytes;
    this.events.append(context.execution.threadUid, context.currentRunId, event);
  }

  private async assertDiskCapacity(): Promise<void> {
    const filesystem = await statfs(this.dataDirectory);
    const available = filesystem.bavail * filesystem.bsize;
    if (available < MINIMUM_FREE_BYTES) {
      throw new Error("Less than 2 GiB of disk space remains; new executions are disabled");
    }
  }
}

function extractPrompt(input: RunAgentInput): string {
  const [message] = input.messages;
  if (message?.role !== "user") {
    throw new Error("A single user message is required");
  }
  if (typeof message.content === "string") {
    if (!message.content.trim()) {
      throw new Error("The user message is empty");
    }
    return message.content;
  }
  const text = message.content
    .filter((part): part is Extract<InputContent, { type: "text" }> => part.type === "text")
    .map((part) => part.text)
    .join("\n");
  if (!text.trim() || message.content.some((part) => part.type !== "text")) {
    throw new Error("MVP accepts text-only user input");
  }
  return text;
}

function validateRunInput(input: RunAgentInput): void {
  if (!input.runId || input.runId.length > 128) {
    throw new Error("runId must contain between 1 and 128 characters");
  }
  if (input.parentRunId !== undefined && (!input.parentRunId || input.parentRunId.length > 128)) {
    throw new Error("parentRunId must contain between 1 and 128 characters");
  }
  if (input.tools.length > 0 || input.context.length > 0) {
    throw new Error("MVP does not accept client tools or context");
  }
  if (!isEmptyRecord(input.state) || !isEmptyRecord(input.forwardedProps)) {
    throw new Error("MVP does not accept client state or forwarded properties");
  }
  if (input.resume && input.resume.length > 0) {
    if (input.messages.length > 0 || !input.parentRunId) {
      throw new Error("Resume runs require parentRunId and no messages");
    }
    return;
  }
  if (input.parentRunId || input.messages.length !== 1 || input.messages[0]?.role !== "user") {
    throw new Error("New runs require exactly one user message and no parentRunId");
  }
  const message = input.messages[0];
  if (message.encryptedValue || message.name) {
    throw new Error("MVP does not accept encrypted or named user messages");
  }
}

function isEmptyRecord(value: unknown): boolean {
  return (
    value === undefined ||
    (value !== null &&
      typeof value === "object" &&
      !Array.isArray(value) &&
      Object.keys(value).length === 0)
  );
}

function questionResponseSchema(questions: readonly StructuredQuestion[]): Record<string, unknown> {
  return {
    type: "object",
    properties: {
      answers: {
        type: "object",
        properties: Object.fromEntries(
          questions.map((question) => [
            question.id,
            {
              type: "array",
              minItems: 1,
              ...(question.multiple ? {} : { maxItems: 1 }),
              items: {
                type: "string",
                ...(question.options.length > 0 && !question.allowOther
                  ? { enum: question.options }
                  : {}),
              },
            },
          ]),
        ),
        required: questions.map((question) => question.id),
        additionalProperties: false,
      },
    },
    required: ["answers"],
    additionalProperties: false,
  };
}

function parseInputAnswer(payload: unknown, questions: readonly StructuredQuestion[]): InputAnswer {
  if (
    !payload ||
    typeof payload !== "object" ||
    Array.isArray(payload) ||
    !("answers" in payload) ||
    Object.keys(payload).length !== 1
  ) {
    throw new Error("Resume payload must contain answers");
  }
  const raw = (payload as { answers: unknown }).answers;
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw new Error("Resume answers must be an object");
  }
  const answerKeys = Object.keys(raw);
  if (
    answerKeys.length !== questions.length ||
    answerKeys.some((key) => !questions.some((question) => question.id === key))
  ) {
    throw new Error("Resume answers must match the pending questions exactly");
  }
  const answers: Record<string, readonly string[]> = {};
  for (const question of questions) {
    const value = (raw as Record<string, unknown>)[question.id];
    if (
      !Array.isArray(value) ||
      value.length === 0 ||
      value.some((answer) => typeof answer !== "string" || answer.length === 0) ||
      (!question.multiple && value.length !== 1)
    ) {
      throw new Error(`Invalid answer for question ${question.id}`);
    }
    const strings = value as string[];
    if (
      question.options.length > 0 &&
      !question.allowOther &&
      strings.some((answer) => !question.options.includes(answer))
    ) {
      throw new Error(`Answer for question ${question.id} is not an allowed option`);
    }
    answers[question.id] = strings;
  }
  return { answers };
}

function hashInput(input: unknown): string {
  return createHash("sha256").update(JSON.stringify(input)).digest("hex");
}

function verifyInputHash(record: PendingInterruptRecord): void {
  if (hashInput(record.input) !== record.inputHash) {
    throw new Error("Pending input integrity check failed");
  }
}

function boundedError(error: unknown): string {
  const message = error instanceof Error ? error.message : String(error);
  return redactSensitiveText(message).trim() || "Provider execution failed";
}

function validateQuestions(
  questions: readonly StructuredQuestion[],
): readonly StructuredQuestion[] {
  if (questions.length === 0 || questions.length > 16) {
    throw new Error("Structured input must contain between 1 and 16 questions");
  }
  const identifiers = new Set<string>();
  for (const question of questions) {
    if (!question.id.trim() || question.id.length > 128 || identifiers.has(question.id)) {
      throw new Error("Structured question ids must be unique and at most 128 characters");
    }
    identifiers.add(question.id);
    if (question.header.length > 128) {
      throw new Error("Structured question header exceeds 128 characters");
    }
    if (!question.question.trim() || question.question.length > 4096) {
      throw new Error("Structured question text is invalid");
    }
    if (
      question.options.length > 32 ||
      new Set(question.options).size !== question.options.length ||
      question.options.some((option) => !option.trim() || option.length > 512)
    ) {
      throw new Error("Structured question options are invalid");
    }
  }
  return questions;
}
