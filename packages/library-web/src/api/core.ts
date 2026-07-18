import type { Transport } from "@connectrpc/connect";
import { Code, ConnectError } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

export const DEFAULT_TIMEOUT_MS = 30_000;

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status = 0) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

export function normalizeBaseUrl(value: string): string {
  return value.replace(/\/+$/, "");
}

export function resolveApiUrl(baseUrl: string, path: string): string {
  return new URL(path.replace(/^\/+/, ""), `${normalizeBaseUrl(baseUrl)}/`).toString();
}

export function required<T>(value: T | undefined, field: string): T {
  if (value === undefined) throw new ApiError(`The API response is missing ${field}.`);
  return value;
}

export function privateTransport(baseUrl: string): Transport {
  return createConnectTransport({
    baseUrl: normalizeBaseUrl(baseUrl),
    fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    useBinaryFormat: false,
    defaultTimeoutMs: DEFAULT_TIMEOUT_MS,
  });
}

export function publicTransport(baseUrl: string): Transport {
  return createConnectTransport({
    baseUrl: normalizeBaseUrl(baseUrl),
    fetch: (input, init) =>
      fetch(input, {
        ...init,
        credentials: "omit",
        referrerPolicy: "no-referrer",
      }),
    useBinaryFormat: false,
    defaultTimeoutMs: DEFAULT_TIMEOUT_MS,
  });
}

export function apiBaseUrl(value: string | undefined, fallback: string): string {
  return normalizeBaseUrl(value?.trim() || fallback);
}

export function isUnauthenticatedError(error: unknown): boolean {
  return (
    (error instanceof ApiError && error.status === 401) ||
    (error instanceof ConnectError && error.code === Code.Unauthenticated)
  );
}

export function isUnavailableShareError(error: unknown): boolean {
  return (
    (error instanceof ApiError && [401, 403, 404, 410].includes(error.status)) ||
    (error instanceof ConnectError &&
      [Code.Unauthenticated, Code.PermissionDenied, Code.NotFound].includes(error.code))
  );
}
