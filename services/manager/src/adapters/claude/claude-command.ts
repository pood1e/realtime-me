import { scrubApiCredentials } from "../../infrastructure/processes.js";

export function claudeEnvironment(environment: NodeJS.ProcessEnv = process.env): NodeJS.ProcessEnv {
  return {
    ...scrubApiCredentials(environment),
    CLAUDE_CODE_ENTRYPOINT: "sdk-ts",
  };
}

export function claudeArguments(providerSessionId?: string | null): string[] {
  return [
    "--print",
    "--input-format",
    "stream-json",
    "--output-format",
    "stream-json",
    "--verbose",
    "--include-partial-messages",
    "--replay-user-messages",
    "--permission-mode",
    "bypassPermissions",
    "--permission-prompt-tool",
    "stdio",
    ...(providerSessionId ? ["--resume", providerSessionId] : []),
  ];
}
