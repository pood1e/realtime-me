import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import type { Client } from "@connectrpc/connect";
import {
  GetSessionRequestSchema,
  LoginRequestSchema,
  LogoutRequestSchema,
  SessionService,
} from "@cloud-drive/contracts";

import { privateTransport } from "./core";

export class SessionClient {
  private readonly client: Client<typeof SessionService>;

  constructor(baseUrl: string) {
    this.client = createClient(SessionService, privateTransport(baseUrl));
  }

  getSession(signal?: AbortSignal) {
    return this.client.getSession(create(GetSessionRequestSchema), { signal });
  }

  login(password: string, returnUrl: string, signal?: AbortSignal) {
    return this.client.login(
      create(LoginRequestSchema, { password, returnUrl }),
      { signal },
    );
  }

  async logout(signal?: AbortSignal): Promise<void> {
    await this.client.logout(create(LogoutRequestSchema), { signal });
  }
}

export function authenticationUrl(
  authOrigin: string,
  returnUrl = window.location.href,
): string {
  const url = new URL(authOrigin);
  url.searchParams.set("return_to", returnUrl);
  return url.toString();
}
