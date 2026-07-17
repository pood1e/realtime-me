import { mkdtemp, readdir, readFile, rename, rm, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { runCommand, scrubApiCredentials } from "../src/infrastructure/processes.js";

const executable = process.env.SM_CODEX_PATH ?? "codex";
const expectedVersion = process.env.SM_CODEX_VERSION ?? "0.144.5";
const adapterDirectory = fileURLToPath(new URL("../src/adapters/codex", import.meta.url));
const temporaryDirectory = await mkdtemp(join(dirname(adapterDirectory), ".codex-protocol-"));
const generatedTypes = join(temporaryDirectory, "gen");
const generatedSchemas = join(temporaryDirectory, "schema");

try {
  const environment = scrubApiCredentials(process.env);
  const version = await checkedCommand(["--version"], environment);
  const detectedVersion = version.trim().split(/\s+/).at(-1) ?? "";
  if (detectedVersion !== expectedVersion) {
    throw new Error(`Expected Codex ${expectedVersion}; found ${detectedVersion}`);
  }
  await checkedCommand(
    ["app-server", "generate-ts", "--experimental", "--out", generatedTypes],
    environment,
  );
  await checkedCommand(
    ["app-server", "generate-json-schema", "--experimental", "--out", generatedSchemas],
    environment,
  );
  await normalizeJsonSchemas(generatedSchemas);
  await replaceDirectory(generatedTypes, join(adapterDirectory, "gen"));
  await replaceDirectory(generatedSchemas, join(adapterDirectory, "schema"));
} finally {
  await rm(temporaryDirectory, { recursive: true, force: true });
}

async function checkedCommand(
  args: readonly string[],
  environment: NodeJS.ProcessEnv,
): Promise<string> {
  const result = await runCommand(executable, args, { env: environment, timeoutMs: 60_000 });
  if (result.exitCode !== 0) {
    throw new Error(`Codex protocol generation failed: ${result.stderr.trim().slice(-1024)}`);
  }
  return result.stdout;
}

async function normalizeJsonSchemas(directory: string): Promise<void> {
  const entries = await readdir(directory, { withFileTypes: true, recursive: true });
  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith(".json")) {
      continue;
    }
    const path = join(entry.parentPath, entry.name);
    const value = JSON.parse(await readFile(path, "utf8")) as unknown;
    await writeFile(path, `${JSON.stringify(sortObjectKeys(value), null, 2)}\n`);
  }
}

function sortObjectKeys(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(sortObjectKeys);
  }
  if (!value || typeof value !== "object") {
    return value;
  }
  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) => (left < right ? -1 : left > right ? 1 : 0))
      .map(([key, item]) => [key, sortObjectKeys(item)]),
  );
}

async function replaceDirectory(source: string, destination: string): Promise<void> {
  await rm(destination, { recursive: true, force: true });
  await rename(source, destination);
}
