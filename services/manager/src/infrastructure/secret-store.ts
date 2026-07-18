import { createHmac, randomBytes } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { join } from "node:path";

const PEPPER_FILE = "device-token-pepper";

export class SecretStore {
  private constructor(private readonly tokenPepper: Buffer) {}

  static async open(directory: string): Promise<SecretStore> {
    await mkdir(directory, { recursive: true, mode: 0o700 });
    const path = join(directory, PEPPER_FILE);
    let tokenPepper: Buffer;
    try {
      tokenPepper = await readFile(path);
    } catch (error) {
      if (!isMissingFile(error)) {
        throw error;
      }
      tokenPepper = randomBytes(32);
      try {
        await writeFile(path, tokenPepper, { mode: 0o600, flag: "wx" });
      } catch (writeError) {
        if (!isExistingFile(writeError)) {
          throw writeError;
        }
        tokenPepper = await readFile(path);
      }
    }
    if (tokenPepper.length !== 32) {
      throw new Error("Invalid device token pepper");
    }
    return new SecretStore(tokenPepper);
  }

  hashDeviceToken(token: string): string {
    return this.hash("device-token", Buffer.from(token, "utf8"));
  }

  hashPairingSecret(secret: Uint8Array): string {
    return this.hash("pairing-secret", secret);
  }

  private hash(purpose: string, value: Uint8Array): string {
    return createHmac("sha256", this.tokenPepper)
      .update(purpose, "utf8")
      .update(Buffer.from([0]))
      .update(value)
      .digest("hex");
  }
}

function isMissingFile(error: unknown): error is NodeJS.ErrnoException {
  return error instanceof Error && "code" in error && error.code === "ENOENT";
}

function isExistingFile(error: unknown): error is NodeJS.ErrnoException {
  return error instanceof Error && "code" in error && error.code === "EEXIST";
}
