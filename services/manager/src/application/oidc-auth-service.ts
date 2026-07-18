import { enumToJson } from "@bufbuild/protobuf";
import { createRemoteJWKSet, errors as joseErrors, jwtVerify } from "jose";
import { z } from "zod";
import { type Permission, PermissionSchema } from "../gen/realtime/me/auth/v1/permission_pb.js";

const DiscoveryDocument = z.object({
  issuer: z.string().url(),
  jwks_uri: z.string().url(),
});

const ACCESS_TOKEN_ALGORITHMS = [
  "RS256",
  "RS384",
  "RS512",
  "PS256",
  "PS384",
  "PS512",
  "ES256",
  "ES384",
  "ES512",
  "EdDSA",
];

export class PermissionDeniedError extends Error {
  constructor() {
    super("Permission denied");
    this.name = "PermissionDeniedError";
  }
}

export interface OwnerPrincipal {
  readonly subject: string;
}

// OidcAuthService verifies short-lived owner access tokens independently of device credentials.
export class OidcAuthService {
  private keys: ReturnType<typeof createRemoteJWKSet> | undefined;
  private discovery: Promise<ReturnType<typeof createRemoteJWKSet>> | undefined;

  constructor(
    private readonly issuer: string,
    private readonly audience: string,
  ) {}

  private async discover(): Promise<ReturnType<typeof createRemoteJWKSet>> {
    const response = await fetch(`${this.issuer}/.well-known/openid-configuration`, {
      headers: { Accept: "application/json" },
      signal: AbortSignal.timeout(10_000),
    });
    if (!response.ok) {
      throw new Error(`OIDC discovery failed with HTTP ${response.status}`);
    }
    const discovery = DiscoveryDocument.parse(await response.json());
    if (discovery.issuer !== this.issuer) {
      throw new Error("OIDC discovery issuer does not match OIDC_ISSUER");
    }
    const jwksUrl = new URL(discovery.jwks_uri);
    if (
      jwksUrl.username.length > 0 ||
      jwksUrl.password.length > 0 ||
      jwksUrl.hash.length > 0 ||
      (jwksUrl.protocol !== "https:" && !loopback(jwksUrl.hostname))
    ) {
      throw new Error(
        "OIDC jwks_uri must be credential-free and use HTTPS outside loopback development",
      );
    }
    return createRemoteJWKSet(jwksUrl, { timeoutDuration: 5_000, cooldownDuration: 30_000 });
  }

  private async keySet(): Promise<ReturnType<typeof createRemoteJWKSet>> {
    if (this.keys) return this.keys;
    this.discovery ??= this.discover();
    try {
      this.keys = await this.discovery;
      return this.keys;
    } finally {
      this.discovery = undefined;
    }
  }

  async authenticate(
    authorization: string | undefined,
    required: Permission,
  ): Promise<OwnerPrincipal | null> {
    const token = bearerToken(authorization);
    if (token?.split(".").length !== 3) return null;
    try {
      const { payload } = await jwtVerify(token, await this.keySet(), {
        issuer: this.issuer,
        audience: this.audience,
        typ: "at+jwt",
        algorithms: ACCESS_TOKEN_ALGORITHMS,
        requiredClaims: ["sub", "permissions"],
      });
      if (typeof payload.sub !== "string" || payload.sub.trim().length === 0) return null;
      const permissionName = enumToJson(PermissionSchema, required);
      if (
        typeof permissionName !== "string" ||
        !Array.isArray(payload.permissions) ||
        !payload.permissions.includes(permissionName)
      ) {
        throw new PermissionDeniedError();
      }
      return { subject: payload.sub.trim() };
    } catch (error) {
      if (error instanceof PermissionDeniedError) throw error;
      if (identityUnavailable(error)) throw error;
      return null;
    }
  }
}

function identityUnavailable(error: unknown): boolean {
  return (
    !(error instanceof joseErrors.JOSEError) ||
    error.constructor === joseErrors.JOSEError ||
    error instanceof joseErrors.JWKSTimeout ||
    error instanceof joseErrors.JWKInvalid ||
    error instanceof joseErrors.JWKSInvalid
  );
}

function bearerToken(authorization: string | undefined): string | null {
  if (!authorization?.startsWith("Bearer ")) return null;
  const token = authorization.slice("Bearer ".length);
  return token.length > 0 && !/\s/.test(token) ? token : null;
}

function loopback(hostname: string): boolean {
  return hostname === "localhost" || hostname === "[::1]" || /^127(?:\.\d{1,3}){3}$/.test(hostname);
}
