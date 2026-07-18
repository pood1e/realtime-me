const CREDENTIAL_ASSIGNMENT =
  /\b(api[_-]?key|authorization|bearer|token|secret|password|cookie)\b\s*[:=]\s*["']?(?:Bearer\s+)?[^\s,"'}]+/gi;
const TOKEN_PREFIX = /\b(?:sk|sess|key|token)-[A-Za-z0-9._-]{8,}\b/gi;
const BEARER_VALUE = /\bBearer\s+[A-Za-z0-9._~-]{8,}/gi;
const JWT = /\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b/g;
const URL_CREDENTIALS = /(https?:\/\/)[^/\s:@]+:[^@\s/]+@/gi;

export function redactSensitiveText(value: string, maximumLength = 1024): string {
  return value
    .replace(CREDENTIAL_ASSIGNMENT, "$1=[REDACTED]")
    .replace(TOKEN_PREFIX, "[REDACTED_TOKEN]")
    .replace(BEARER_VALUE, "Bearer [REDACTED]")
    .replace(JWT, "[REDACTED_JWT]")
    .replace(URL_CREDENTIALS, "$1[REDACTED]@")
    .slice(0, maximumLength);
}
