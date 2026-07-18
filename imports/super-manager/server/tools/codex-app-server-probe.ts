import {
  JsonRpcProcess,
  runCommand,
  scrubApiCredentials,
} from "../src/infrastructure/processes.js";

const executable = process.env.SM_CODEX_PATH ?? "codex";
const expectedVersion = process.env.SM_CODEX_VERSION ?? "0.144.5";
const version = await runCommand(executable, ["--version"], {
  env: scrubApiCredentials(process.env),
});
if (version.exitCode !== 0) {
  throw new Error(version.stderr.trim() || "Codex version probe failed");
}
const detectedVersion = version.stdout.trim().split(/\s+/).at(-1) ?? "";
if (detectedVersion !== expectedVersion) {
  throw new Error(`Expected Codex ${expectedVersion}; found ${detectedVersion}`);
}

const client = new JsonRpcProcess(executable, ["app-server", "--listen", "stdio://"], {
  env: scrubApiCredentials(process.env),
});
try {
  await client.request("initialize", {
    clientInfo: { name: "super_manager_probe", version: "0.1.0" },
    capabilities: { experimentalApi: true, requestAttestation: false },
  });
  client.notify("initialized", {});
  const account = await client.request<{ account: { type: string } | null }>("account/read", {
    refreshToken: false,
  });
  await client.request("account/rateLimits/read", undefined);
  process.stdout.write(
    `${JSON.stringify({ version: detectedVersion, accountType: account.account?.type ?? null })}\n`,
  );
} finally {
  await client.close();
}
