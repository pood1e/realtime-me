import { resolve } from "node:path";
import { z } from "zod";

const EnvironmentSchema = z.object({
  NODE_ENV: z.enum(["development", "production"]).default("development"),
  SM_HOST: z.string().default("127.0.0.1"),
  SM_PORT: z.coerce.number().int().min(1).max(65535).default(3080),
  SM_LOG_LEVEL: z
    .enum(["fatal", "error", "warn", "info", "debug", "trace", "silent"])
    .default("info"),
  SM_DATA_DIR: z.string().default("./data"),
  SM_ALLOWED_WORKSPACE_ROOTS: z.string().default(""),
  SM_CODEX_PATH: z.string().default("codex"),
  SM_CODEX_VERSION: z.string().default("0.144.5"),
  SM_CLAUDE_PATH: z.string().default("claude"),
  SM_CLAUDE_VERSION: z.string().default("2.1.195"),
  SM_TMUX_PATH: z.string().default("tmux"),
  SM_TMUX_SOCKET_NAME: z
    .string()
    .regex(/^[A-Za-z0-9_-]{1,64}$/)
    .default("super-manager"),
  SM_OPENSSL_PATH: z.string().default("openssl"),
  SM_SERVICE_URL: z.string().url().default("https://127.0.0.1"),
  SM_PAIRING_URL: z.string().url().optional(),
  OIDC_ISSUER: z.string().url(),
  MANAGER_AUTH_AUDIENCE: z.string().trim().regex(/^\S+$/),
  INTERNAL_API_KEY_FILE: z.string().trim().default(""),
});

export interface ServerConfig {
  readonly production: boolean;
  readonly host: string;
  readonly port: number;
  readonly logLevel: string;
  readonly dataDirectory: string;
  readonly databasePath: string;
  readonly allowedWorkspaceRoots: readonly string[];
  readonly codexPath: string;
  readonly codexVersion: string;
  readonly claudePath: string;
  readonly claudeVersion: string;
  readonly tmuxPath: string;
  readonly tmuxSocketName: string;
  readonly opensslPath: string;
  readonly serviceUrl: string;
  readonly pairingUrl: string;
  readonly oidcIssuer: string;
  readonly oidcAudience: string;
  readonly internalApiKeyFile: string;
}

export function loadConfig(environment: NodeJS.ProcessEnv = process.env): ServerConfig {
  const parsed = EnvironmentSchema.parse(environment);
  const dataDirectory = resolve(parsed.SM_DATA_DIR);
  const configuredRoots = parsed.SM_ALLOWED_WORKSPACE_ROOTS.split(":")
    .map((root) => root.trim())
    .filter((root) => root.length > 0)
    .map((root) => resolve(root));
  const allowedWorkspaceRoots =
    configuredRoots.length > 0
      ? configuredRoots
      : parsed.NODE_ENV === "production"
        ? []
        : [resolve(process.cwd(), "..")];
  const serviceEndpoint = normalizeEndpoint(parsed.SM_SERVICE_URL, "SM_SERVICE_URL");
  const defaultPairingUrl = new URL(serviceEndpoint);
  defaultPairingUrl.port = "8443";
  const pairingEndpoint = normalizeEndpoint(
    parsed.SM_PAIRING_URL ?? defaultPairingUrl.toString(),
    "SM_PAIRING_URL",
  );
  if (parsed.NODE_ENV === "production") {
    if (serviceEndpoint.protocol !== "https:" || pairingEndpoint.protocol !== "https:") {
      throw new Error("Service and pairing endpoints must use HTTPS in production");
    }
    if (new URL(parsed.OIDC_ISSUER).protocol !== "https:") {
      throw new Error("OIDC_ISSUER must use HTTPS in production");
    }
    if (serviceEndpoint.hostname !== pairingEndpoint.hostname) {
      throw new Error("Service and pairing endpoints must use the same hostname");
    }
  }

  return {
    production: parsed.NODE_ENV === "production",
    host: parsed.SM_HOST,
    port: parsed.SM_PORT,
    logLevel: parsed.SM_LOG_LEVEL,
    dataDirectory,
    databasePath: resolve(dataDirectory, "super-manager.sqlite3"),
    allowedWorkspaceRoots,
    codexPath: parsed.SM_CODEX_PATH,
    codexVersion: parsed.SM_CODEX_VERSION,
    claudePath: parsed.SM_CLAUDE_PATH,
    claudeVersion: parsed.SM_CLAUDE_VERSION,
    tmuxPath: parsed.SM_TMUX_PATH,
    tmuxSocketName: parsed.SM_TMUX_SOCKET_NAME,
    opensslPath: parsed.SM_OPENSSL_PATH,
    serviceUrl: serviceEndpoint.toString().replace(/\/$/, ""),
    pairingUrl: pairingEndpoint.toString().replace(/\/$/, ""),
    oidcIssuer: normalizeIssuer(parsed.OIDC_ISSUER),
    oidcAudience: parsed.MANAGER_AUTH_AUDIENCE,
    internalApiKeyFile: parsed.INTERNAL_API_KEY_FILE,
  };
}

export function assertSafeProcess(config: ServerConfig): void {
  if (config.production && process.getuid?.() === 0) {
    throw new Error("Realtime Me Manager must not run as root");
  }
  if (config.production && config.allowedWorkspaceRoots.length === 0) {
    throw new Error("At least one allowed workspace root is required");
  }
  if (config.production && config.host !== "127.0.0.1" && config.host !== "::1") {
    throw new Error("Realtime Me Manager must listen on loopback in production");
  }
}

function normalizeEndpoint(value: string, name: string): URL {
  const endpoint = new URL(value);
  if (
    endpoint.username.length > 0 ||
    endpoint.password.length > 0 ||
    endpoint.pathname !== "/" ||
    endpoint.search.length > 0 ||
    endpoint.hash.length > 0
  ) {
    throw new Error(`${name} must be an origin without credentials, path, query, or fragment`);
  }
  return endpoint;
}

function normalizeIssuer(value: string): string {
  const issuer = new URL(value);
  if (
    issuer.username.length > 0 ||
    issuer.password.length > 0 ||
    issuer.search.length > 0 ||
    issuer.hash.length > 0
  ) {
    throw new Error("OIDC_ISSUER must not contain credentials, query, or fragment");
  }
  if (issuer.protocol !== "https:" && !(issuer.protocol === "http:" && loopback(issuer.hostname))) {
    throw new Error("OIDC_ISSUER must use HTTPS outside loopback development");
  }
  return issuer.toString().replace(/\/$/, "");
}

function loopback(hostname: string): boolean {
  return hostname === "localhost" || hostname === "[::1]" || /^127(?:\.\d{1,3}){3}$/.test(hostname);
}
