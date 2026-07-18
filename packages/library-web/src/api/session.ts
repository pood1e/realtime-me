import { create } from "@bufbuild/protobuf";
import type { Client } from "@connectrpc/connect";
import { createClient } from "@connectrpc/connect";
import {
  GetSessionRequestSchema,
  LoginRequestSchema,
  LogoutRequestSchema,
  SessionService,
} from "@realtime-me/library-contracts";

import { privateTransport } from "./core";

const SESSION_VALIDATION_TTL_MS = 5 * 60_000;

export class SessionClient {
  private readonly client: Client<typeof SessionService>;

  constructor(baseUrl: string) {
    this.client = createClient(SessionService, privateTransport(baseUrl));
  }

  getSession(signal?: AbortSignal) {
    return this.client.getSession(create(GetSessionRequestSchema), signal ? { signal } : undefined);
  }

  login(password: string, returnUrl: string, signal?: AbortSignal) {
    return this.client.login(
      create(LoginRequestSchema, { password, returnUrl }),
      signal ? { signal } : undefined,
    );
  }

  async logout(signal?: AbortSignal): Promise<void> {
    try {
      await this.client.logout(create(LogoutRequestSchema), signal ? { signal } : undefined);
    } finally {
      clearRecentSessionValidation();
    }
  }
}

export function hasRecentSessionValidation(apiBase: string): boolean {
  try {
    return Number(window.sessionStorage.getItem(validationKey(apiBase))) > Date.now();
  } catch {
    return false;
  }
}

export function markSessionValidated(apiBase: string): void {
  try {
    window.sessionStorage.setItem(
      validationKey(apiBase),
      String(Date.now() + SESSION_VALIDATION_TTL_MS),
    );
  } catch {
    // Session storage is only a presentation optimization; API auth remains authoritative.
  }
}

export function clearRecentSessionValidation(): void {
  try {
    for (let index = window.sessionStorage.length - 1; index >= 0; index -= 1) {
      const key = window.sessionStorage.key(index);
      if (key?.startsWith("realtime-me.library.session-valid-until:")) {
        window.sessionStorage.removeItem(key);
      }
    }
  } catch {
    // Nothing to clear when storage is unavailable.
  }
}

function validationKey(apiBase: string): string {
  return `realtime-me.library.session-valid-until:${apiBase}`;
}

export function authenticationUrl(authOrigin: string, returnUrl = window.location.href): string {
  const url = new URL(authOrigin);
  url.searchParams.set("return_to", returnUrl);
  return url.toString();
}
