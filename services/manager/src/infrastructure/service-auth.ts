import { timingSafeEqual } from "node:crypto";
import { readFileSync } from "node:fs";

export const INTERNAL_API_KEY_HEADER = "x-realtime-internal-key";

const encodedKeyPattern = /^[0-9a-f]{64}$/i;

export function loadInternalApiKey(path: string): Buffer {
  if (path.length === 0) {
    throw new Error("INTERNAL_API_KEY_FILE is required");
  }
  const content = readFileSync(path, "utf8");
  const encoded = content.length === 65 && content.endsWith("\n") ? content.slice(0, -1) : content;
  if (!encodedKeyPattern.test(encoded)) {
    throw new Error("Internal API key must contain exactly 32 hexadecimal bytes");
  }
  return Buffer.from(encoded, "hex");
}

export function matchesInternalApiKey(
  presented: string | readonly string[] | undefined,
  expected: Uint8Array,
): boolean {
  if (typeof presented !== "string" || !encodedKeyPattern.test(presented)) {
    return false;
  }
  return timingSafeEqual(Buffer.from(presented, "hex"), expected);
}
