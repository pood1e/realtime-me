#!/usr/bin/env node
import { ClaudeAdapter } from "./adapters/claude/claude-adapter.js";
import { CodexAdapter } from "./adapters/codex/codex-adapter.js";
import { PairingAuthority } from "./application/pairing-authority.js";
import { RuntimeRegistry } from "./application/runtime-registry.js";
import { assertSafeProcess, loadConfig, type ServerConfig } from "./infrastructure/config.js";
import { type CommandResult, runCommand } from "./infrastructure/processes.js";
import { ResourceStore } from "./infrastructure/resource-store.js";
import { SecretStore } from "./infrastructure/secret-store.js";
import { openDatabase } from "./infrastructure/sqlite.js";

async function main(): Promise<void> {
  const [command, subcommand, ...rest] = process.argv.slice(2);
  const config = loadConfig();
  assertSafeProcess(config);
  if (command === "pair" && subcommand === "create" && rest.length === 0) {
    await withPairingAuthority(config, async (pairing) => {
      const offer = await pairing.createOffer();
      const qr = await safeRunCommand("qrencode", ["-t", "UTF8", offer.payload]);
      process.stdout.write(`Pair before ${offer.expireTime.toISOString()}\n\n`);
      if (qr.exitCode === 0) {
        process.stdout.write(`${qr.stdout}\n`);
      } else {
        process.stdout.write(`${offer.payload}\n`);
        process.stderr.write("qrencode is unavailable; printed the pairing payload instead.\n");
      }
    });
    return;
  }
  if (command === "pki" && subcommand === "init" && rest.length === 0) {
    await withPairingAuthority(config, (pairing) => {
      process.stdout.write(
        `${JSON.stringify(
          {
            caCertificate: pairing.pki.caCertificate,
            serverCertificate: pairing.pki.serverCertificate,
            serverKey: pairing.pki.serverKey,
          },
          null,
          2,
        )}\n`,
      );
    });
    return;
  }
  if (command === "pki" && subcommand === "renew-server" && rest.length === 0) {
    await withPairingAuthority(config, async (pairing) => {
      await pairing.renewServerCertificate();
      process.stdout.write(
        `${JSON.stringify(
          {
            serverCertificate: pairing.pki.serverCertificate,
            serverKey: pairing.pki.serverKey,
          },
          null,
          2,
        )}\n`,
      );
    });
    return;
  }
  if (command === "device" && subcommand === "list" && rest.length === 0) {
    await withStore(config, (store) => {
      process.stdout.write(
        `${JSON.stringify(
          store.listDevices().map((device) => ({
            uid: device.uid,
            displayName: device.displayName,
            status: device.status,
            expireTime: device.expireTime.toISOString(),
          })),
          null,
          2,
        )}\n`,
      );
    });
    return;
  }
  if (command === "device" && subcommand === "revoke" && rest.length === 1) {
    await withStore(config, (store) => {
      if (!store.revokeDevice(rest[0] ?? "")) {
        throw new Error("Active device not found");
      }
    });
    process.stdout.write("Device revoked\n");
    return;
  }
  if (command === "doctor" && subcommand === undefined) {
    const database = openDatabase(config.databasePath);
    const store = new ResourceStore(database);
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
    try {
      await runtimes.probeAll();
      const [tmux, openssl] = await Promise.all([
        safeRunCommand(config.tmuxPath, ["-V"]),
        safeRunCommand(config.opensslPath, ["version"]),
      ]);
      process.stdout.write(
        `${JSON.stringify(
          {
            runtimes: runtimes.list().map((runtime) => ({
              name: runtime.displayName,
              version: runtime.version,
              availability: runtime.availability,
              diagnostic: runtime.diagnostic,
            })),
            tmux: commandSummary(tmux),
            openssl: commandSummary(openssl),
            allowedWorkspaceRoots: config.allowedWorkspaceRoots,
          },
          null,
          2,
        )}\n`,
      );
    } finally {
      await runtimes.close();
      database.close();
    }
    return;
  }
  process.stderr.write(
    [
      "Usage:",
      "  smctl doctor",
      "  smctl pki init",
      "  smctl pki renew-server",
      "  smctl pair create",
      "  smctl device list",
      "  smctl device revoke <device-uid>",
      "",
    ].join("\n"),
  );
  process.exitCode = 2;
}

async function withPairingAuthority<T>(
  config: ServerConfig,
  operation: (pairing: PairingAuthority) => T | Promise<T>,
): Promise<T> {
  return withStore(config, async (store) => {
    const pairing = await PairingAuthority.open({
      store,
      secrets: await SecretStore.open(config.dataDirectory),
      opensslPath: config.opensslPath,
      dataDirectory: config.dataDirectory,
      publicUrl: config.publicUrl,
      pairingUrl: config.pairingUrl,
    });
    return operation(pairing);
  });
}

async function withStore<T>(
  config: ServerConfig,
  operation: (store: ResourceStore) => T | Promise<T>,
): Promise<T> {
  const database = openDatabase(config.databasePath);
  try {
    return await operation(new ResourceStore(database));
  } finally {
    database.close();
  }
}

async function safeRunCommand(command: string, args: readonly string[]): Promise<CommandResult> {
  try {
    return await runCommand(command, args, { timeoutMs: 5_000 });
  } catch (error) {
    return {
      exitCode: -1,
      stdout: "",
      stderr: error instanceof Error ? error.message : String(error),
    };
  }
}

function commandSummary(result: CommandResult): {
  readonly available: boolean;
  readonly output: string;
} {
  return {
    available: result.exitCode === 0,
    output: (result.stdout || result.stderr).trim().slice(0, 256),
  };
}

main().catch((error: unknown) => {
  process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
  process.exitCode = 1;
});
