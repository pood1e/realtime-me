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
  if (
    url.protocol !== "https:" &&
    url.hostname !== "localhost" &&
    url.hostname !== "127.0.0.1"
  ) {
    throw new Error(`${variableName} must use HTTPS outside local development`);
  }
  return url.origin;
}

export interface ContentSecurityPolicyExtensions {
  script?: readonly string[];
  image?: readonly string[];
  media?: readonly string[];
  connect?: readonly string[];
}

function contentSecurityPolicy(
  origin: string,
  extensions: ContentSecurityPolicyExtensions,
): string {
  return [
    "default-src 'self'",
    "base-uri 'none'",
    "object-src 'none'",
    "frame-ancestors 'none'",
    "form-action 'self'",
    `script-src 'self' ${extensions.script?.join(" ") ?? ""}`,
    "style-src 'self' 'unsafe-inline'",
    `img-src 'self' data: blob: ${origin} ${extensions.image?.join(" ") ?? ""}`,
    `media-src 'self' blob: ${origin} ${extensions.media?.join(" ") ?? ""}`,
    `font-src 'self' data: blob: ${origin}`,
    `connect-src 'self' ${origin} ${extensions.connect?.join(" ") ?? ""}`,
    `frame-src 'self' blob: ${origin}`,
    "worker-src 'self' blob:",
    "upgrade-insecure-requests",
  ].join("; ");
}

export function cloudflarePagesHeaders(
  variableName: string,
  fallbackApiBase: string,
  extensions: ContentSecurityPolicyExtensions = {},
): Plugin {
  let origin = apiOrigin(fallbackApiBase, variableName);

  return {
    name: "cloud-drive-cloudflare-pages-headers",
    apply: "build",
    configResolved(config) {
      origin = apiOrigin(
        config.env[variableName] ?? fallbackApiBase,
        variableName,
      );
    },
    generateBundle() {
      const lines = [
        "/*",
        `  Content-Security-Policy: ${contentSecurityPolicy(origin, extensions)}`,
      ];
      for (const header of securityHeaders) {
        lines.push(`  ${header}`);
      }
      this.emitFile({
        type: "asset",
        fileName: "_headers",
        source: `${lines.join("\n")}\n`,
      });
    },
  };
}
