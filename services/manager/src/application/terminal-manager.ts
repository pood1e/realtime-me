import { Buffer } from "node:buffer";
import type { IPty } from "node-pty";
import * as pty from "node-pty";
import type { TerminalSessionRecord } from "../domain/records.js";
import { runCommand } from "../infrastructure/processes.js";
import type { ResourceStore } from "../infrastructure/resource-store.js";

const MAX_TERMINALS = 4;

export interface TerminalAttachment {
  write(data: Uint8Array): void;
  resize(columns: number, rows: number): void;
  detach(): void;
}

export interface TerminalAttachmentSink {
  output(data: Uint8Array): void;
  exited(exitCode: number | null, signal: number | null): void;
}

interface ActiveAttachment {
  readonly process: IPty;
  detached: boolean;
}

export class TerminalManager {
  private readonly attachments = new Map<string, ActiveAttachment>();

  constructor(
    private readonly store: ResourceStore,
    private readonly tmuxPath: string,
    private readonly tmuxSocketName: string,
  ) {}

  async create(input: {
    workspaceUid: string;
    displayName: string;
    cwd: string;
    columns: number;
    rows: number;
  }): Promise<TerminalSessionRecord> {
    await this.reconcile();
    if (this.store.countRunningTerminals() >= MAX_TERMINALS) {
      throw new Error("The terminal session limit has been reached");
    }
    const session = this.store.createTerminalSession(input);
    try {
      await this.tmux([
        "new-session",
        "-d",
        "-s",
        session.tmuxName,
        "-c",
        session.cwd,
        "-x",
        String(session.columns),
        "-y",
        String(session.rows),
      ]);
      await this.tmux(["set-option", "-t", session.tmuxName, "history-limit", "50000"]);
      return session;
    } catch (error) {
      try {
        await this.cleanupFailedCreation(session);
      } catch (cleanupError) {
        throw new AggregateError(
          [error, cleanupError],
          "Terminal creation failed and the tmux session could not be cleaned up",
        );
      }
      throw error;
    }
  }

  async get(uid: string): Promise<TerminalSessionRecord | null> {
    const session = this.store.getTerminalSession(uid);
    if (!session || session.state === "CLOSED") {
      return session;
    }
    if (!(await this.isAlive(session.tmuxName))) {
      this.store.setTerminalState(uid, "CLOSED");
      return this.store.getTerminalSession(uid);
    }
    return session;
  }

  async list(workspaceUid: string): Promise<readonly TerminalSessionRecord[]> {
    await this.reconcile();
    return this.store.listTerminalSessions(workspaceUid);
  }

  async attach(uid: string, sink: TerminalAttachmentSink): Promise<TerminalAttachment> {
    const session = await this.get(uid);
    if (session?.state !== "RUNNING") {
      throw new Error("Terminal session is not running");
    }
    if (this.attachments.has(uid)) {
      throw new Error("Terminal session already has a writable attachment");
    }
    const ptyProcess = pty.spawn(
      this.tmuxPath,
      ["-L", this.tmuxSocketName, "attach-session", "-t", session.tmuxName],
      {
        name: "xterm-256color",
        cols: session.columns,
        rows: session.rows,
        cwd: session.cwd,
        env: { ...processEnvironment(), TERM: "xterm-256color" },
      },
    );
    const active: ActiveAttachment = { process: ptyProcess, detached: false };
    this.attachments.set(uid, active);
    ptyProcess.onData((data) => sink.output(Buffer.from(data, "utf8")));
    ptyProcess.onExit(({ exitCode, signal }) => {
      this.release(uid, active);
      if (!active.detached) {
        sink.exited(exitCode, signal ?? null);
      }
      void this.markClosedIfGone(uid, session.tmuxName).catch(() => {
        this.store.setTerminalState(uid, "CLOSED");
      });
    });
    return {
      write: (data) => ptyProcess.write(Buffer.from(data).toString("utf8")),
      resize: (columns, rows) => {
        assertDimensions(columns, rows);
        ptyProcess.resize(columns, rows);
        this.store.setTerminalSize(uid, columns, rows);
      },
      detach: () => {
        if (!active.detached) {
          active.detached = true;
          this.release(uid, active);
          ptyProcess.kill();
        }
      },
    };
  }

  async delete(uid: string): Promise<boolean> {
    const session = this.store.getTerminalSession(uid);
    if (!session) {
      return false;
    }
    const attachment = this.attachments.get(uid);
    if (attachment) {
      attachment.detached = true;
      this.release(uid, attachment);
      attachment.process.kill();
    }
    if (session.state === "RUNNING") {
      const result = await runCommand(
        this.tmuxPath,
        ["-L", this.tmuxSocketName, "kill-session", "-t", session.tmuxName],
        { timeoutMs: 5_000 },
      );
      if (result.exitCode !== 0 && (await this.isAlive(session.tmuxName))) {
        throw new Error(result.stderr.trim() || "Failed to stop tmux session");
      }
    }
    return this.store.deleteTerminalSession(uid);
  }

  async reconcile(): Promise<void> {
    await Promise.all(
      this.store.listRunningTerminalSessions().map(async (session) => {
        if (!(await this.isAlive(session.tmuxName))) {
          this.store.setTerminalState(session.uid, "CLOSED");
        }
      }),
    );
  }

  close(): void {
    for (const [uid, attachment] of this.attachments) {
      attachment.detached = true;
      attachment.process.kill();
      this.attachments.delete(uid);
    }
  }

  private async tmux(args: readonly string[]): Promise<void> {
    const result = await runCommand(this.tmuxPath, ["-L", this.tmuxSocketName, ...args], {
      timeoutMs: 5_000,
    });
    if (result.exitCode !== 0) {
      throw new Error(result.stderr.trim() || `tmux ${args[0] ?? "command"} failed`);
    }
  }

  private async isAlive(tmuxName: string): Promise<boolean> {
    const result = await runCommand(
      this.tmuxPath,
      ["-L", this.tmuxSocketName, "has-session", "-t", tmuxName],
      { timeoutMs: 5_000 },
    );
    if (result.exitCode === 0) {
      return true;
    }
    if (result.exitCode === 1) {
      return false;
    }
    throw new Error(result.stderr.trim() || "Failed to inspect tmux session");
  }

  private async cleanupFailedCreation(session: TerminalSessionRecord): Promise<void> {
    const result = await runCommand(
      this.tmuxPath,
      ["-L", this.tmuxSocketName, "kill-session", "-t", session.tmuxName],
      { timeoutMs: 5_000 },
    );
    if (result.exitCode !== 0 && (await this.isAlive(session.tmuxName))) {
      throw new Error(result.stderr.trim() || "Failed to clean up tmux session");
    }
    this.store.deleteTerminalSession(session.uid);
  }

  private async markClosedIfGone(uid: string, tmuxName: string): Promise<void> {
    if (!(await this.isAlive(tmuxName))) {
      this.store.setTerminalState(uid, "CLOSED");
    }
  }

  private release(uid: string, attachment: ActiveAttachment): void {
    if (this.attachments.get(uid) === attachment) {
      this.attachments.delete(uid);
    }
  }
}

function assertDimensions(columns: number, rows: number): void {
  if (columns < 20 || columns > 1000 || rows < 5 || rows > 1000) {
    throw new Error("Terminal dimensions are out of range");
  }
}

function processEnvironment(): Record<string, string> {
  return Object.fromEntries(
    Object.entries(process.env).filter(
      (entry): entry is [string, string] => entry[1] !== undefined,
    ),
  );
}
