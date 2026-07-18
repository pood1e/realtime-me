import type { DeviceRecord } from "../domain/records.js";
import type { ResourceStore } from "../infrastructure/resource-store.js";
import type { SecretStore } from "../infrastructure/secret-store.js";

export class DeviceAuthService {
  constructor(
    private readonly store: ResourceStore,
    private readonly secrets: SecretStore,
  ) {}

  authenticate(authorization: string | undefined): DeviceRecord | null {
    if (!authorization) {
      return null;
    }
    const match = /^Bearer ([A-Za-z0-9_-]{43})$/.exec(authorization);
    if (!match?.[1]) {
      return null;
    }
    return this.store.findActiveDeviceByTokenHash(this.secrets.hashDeviceToken(match[1]));
  }
}
