import type { RuntimeAdapter, RuntimeStatus } from "../domain/runtime.js";
import { CLAUDE_RUNTIME_UID, CODEX_RUNTIME_UID } from "../domain/runtime.js";

export class RuntimeRegistry {
  private readonly adaptersByUid = new Map<string, RuntimeAdapter>();
  private readonly statusesByUid = new Map<string, RuntimeStatus>();

  constructor(adapters: readonly RuntimeAdapter[]) {
    for (const adapter of adapters) {
      const uid = adapter.kind === "codex" ? CODEX_RUNTIME_UID : CLAUDE_RUNTIME_UID;
      this.adaptersByUid.set(uid, adapter);
    }
  }

  async probeAll(): Promise<void> {
    await Promise.all(
      [...this.adaptersByUid.entries()].map(async ([uid, adapter]) => {
        const status = await adapter.probe();
        if (status.uid !== uid) {
          throw new Error(`Runtime adapter ${adapter.kind} returned an unexpected uid`);
        }
        this.statusesByUid.set(uid, status);
      }),
    );
  }

  list(): readonly RuntimeStatus[] {
    return [...this.adaptersByUid.keys()].flatMap((uid) => {
      const status = this.statusesByUid.get(uid);
      return status ? [status] : [];
    });
  }

  get(uid: string): RuntimeStatus | null {
    return this.statusesByUid.get(uid) ?? null;
  }

  requireAvailable(uid: string): RuntimeAdapter {
    const status = this.statusesByUid.get(uid);
    const adapter = this.adaptersByUid.get(uid);
    if (!status || !adapter || status.availability !== "AVAILABLE") {
      throw new Error("Runtime is not available");
    }
    return adapter;
  }

  async close(): Promise<void> {
    await Promise.all([...this.adaptersByUid.values()].map((adapter) => adapter.close()));
  }
}
