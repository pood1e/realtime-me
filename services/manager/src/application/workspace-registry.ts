import { realpath, stat } from "node:fs/promises";
import { isAbsolute, relative, resolve } from "node:path";
import type { WorkspaceRecord } from "../domain/records.js";
import type { ResourceStore } from "../infrastructure/resource-store.js";

export class WorkspaceRegistry {
  private constructor(
    private readonly store: ResourceStore,
    private readonly allowedRoots: readonly string[],
  ) {}

  static async create(
    store: ResourceStore,
    configuredRoots: readonly string[],
  ): Promise<WorkspaceRegistry> {
    const allowedRoots = await Promise.all(configuredRoots.map((root) => realpath(resolve(root))));
    return new WorkspaceRegistry(store, allowedRoots);
  }

  async register(displayName: string, requestedPath: string): Promise<WorkspaceRecord> {
    if (!isAbsolute(requestedPath)) {
      throw new Error("Workspace path must be absolute");
    }
    const path = await realpath(requestedPath);
    const metadata = await stat(path);
    if (!metadata.isDirectory()) {
      throw new Error("Workspace path must be a directory");
    }
    if (!this.allowedRoots.some((root) => contains(root, path))) {
      throw new Error("Workspace path is outside the configured allowed roots");
    }
    return this.store.createWorkspace(displayName, path);
  }
}

function contains(root: string, candidate: string): boolean {
  const path = relative(root, candidate);
  return path === "" || (!path.startsWith("..") && !isAbsolute(path));
}
