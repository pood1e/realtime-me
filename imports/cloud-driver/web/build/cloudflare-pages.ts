import type { Plugin } from "vite";

const securityHeaders = [
  "X-Content-Type-Options: nosniff",
  "Referrer-Policy: no-referrer",
  "Permissions-Policy: camera=(), microphone=(), geolocation=(), payment=()",
  "Cross-Origin-Opener-Policy: same-origin",
] as const;

function apiOrigin(value: unknown, variableName: string): string {
  if (typeof value !== "string" || !value.trim()) {
    throw new Error(`${variableName} must be a non-empty absolute URL`);
  }

  const url = new URL(value);
  if (url.protocol !== "https:" && url.hostname !== "localhost" && url.hostname !== "127.0.0.1") {
    throw new Error(`${variableName} must use HTTPS outside local development`);
  }
  return url.origin;
}

function contentSecurityPolicy(origin: string): string {
  return [
    "default-src 'self'",
    "base-uri 'none'",
    "object-src 'none'",
    "frame-ancestors 'none'",
    "form-action 'self'",
    "script-src 'self'",
    "style-src 'self'",
    `img-src 'self' data: blob: ${origin}`,
    `media-src 'self' blob: ${origin}`,
    `connect-src 'self' ${origin}`,
    `frame-src ${origin}`,
    "upgrade-insecure-requests",
  ].join("; ");
}

export function cloudflarePagesHeaders(variableName: string, fallbackApiBase: string): Plugin {
  let origin = apiOrigin(fallbackApiBase, variableName);

  return {
    name: "cloud-drive-cloudflare-pages-headers",
    apply: "build",
    configResolved(config) {
      origin = apiOrigin(config.env[variableName] ?? fallbackApiBase, variableName);
    },
    generateBundle() {
      const lines = ["/*", `  Content-Security-Policy: ${contentSecurityPolicy(origin)}`];
      for (const header of securityHeaders) {
        lines.push(`  ${header}`);
      }
      this.emitFile({ type: "asset", fileName: "_headers", source: `${lines.join("\n")}\n` });
    },
  };
}
