import { claudeArguments, claudeEnvironment } from "../src/adapters/claude/claude-command.js";
import { runCommand } from "../src/infrastructure/processes.js";

const executable = process.env.SM_CLAUDE_PATH ?? "claude";
const expectedVersion = process.env.SM_CLAUDE_VERSION ?? "2.1.195";
const environment = claudeEnvironment();
const [version, auth, help, control] = await Promise.all([
  runCommand(executable, ["--version"], { env: environment }),
  runCommand(executable, ["auth", "status", "--json"], { env: environment }),
  runCommand(executable, ["--help"], { env: environment }),
  runCommand(executable, claudeArguments(), { env: environment }),
]);
if (
  version.exitCode !== 0 ||
  auth.exitCode !== 0 ||
  help.exitCode !== 0 ||
  control.exitCode !== 0
) {
  throw new Error("Claude Code compatibility probe failed");
}
const detectedVersion = version.stdout.trim().split(/\s+/)[0] ?? "";
if (detectedVersion !== expectedVersion) {
  throw new Error(`Expected Claude Code ${expectedVersion}; found ${detectedVersion}`);
}
for (const flag of ["--input-format", "--output-format", "--permission-mode"]) {
  if (!help.stdout.includes(flag)) {
    throw new Error(`Claude Code no longer declares ${flag}`);
  }
}
const authStatus = JSON.parse(auth.stdout) as {
  loggedIn?: boolean;
  authMethod?: string;
  apiProvider?: string;
  subscriptionType?: string;
};
process.stdout.write(
  `${JSON.stringify({
    version: detectedVersion,
    loggedIn: authStatus.loggedIn === true,
    authMethod: authStatus.authMethod ?? null,
    apiProvider: authStatus.apiProvider ?? null,
    subscriptionType: authStatus.subscriptionType ?? null,
    stdioControlAccepted: true,
  })}\n`,
);
