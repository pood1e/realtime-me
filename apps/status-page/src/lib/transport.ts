import { Code, ConnectError, createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ProfileService } from '@/gen/realtime/me/v1/profile_pb';
import { StatusService } from '@/gen/realtime/me/v1/status_pb';

// POLL_INTERVAL_MS is the cadence for the public and internal status loops.
export const POLL_INTERVAL_MS = 10_000;

// apiBaseUrl is the same-origin gateway base the browser talks to. The Worker
// proxies the ConnectRPC calls to the upstream gateway.
export const apiBaseUrl = resolveApiBaseUrl();

const transport = createConnectTransport({ baseUrl: apiBaseUrl });

export const statusClient = createClient(StatusService, transport);
export const profileClient = createClient(ProfileService, transport);

export function authHeaders(token: string): HeadersInit {
  return { Authorization: `Bearer ${token}` };
}

export function isUnauthorized(error: unknown): boolean {
  return error instanceof ConnectError && error.code === Code.Unauthenticated;
}

function resolveApiBaseUrl(): string {
  const configured = import.meta.env.VITE_STATUS_API_BASE_URL as string | undefined;
  if (configured?.trim()) return configured.replace(/\/+$/, '');

  const { protocol, hostname } = window.location;
  if (hostname === 'localhost' || hostname === '127.0.0.1') return 'http://localhost:18080';
  if (hostname.startsWith('status.')) return `${protocol}//api-${hostname}`;
  return window.location.origin;
}
