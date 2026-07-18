type AssetBinding = {
  fetch(request: Request): Promise<Response>;
};

type Env = {
  ASSETS: AssetBinding;
  STATUS_API_BASE_URL?: string;
  LIBRARY_PUBLIC_API_BASE_URL?: string;
};

const STATUS_PROCEDURES = new Set([
  "/realtime.me.status.v1.StatusService/GetPublicStatus",
  "/realtime.me.site.v1.ProfileService/GetProfile",
  "/realtime.me.site.v1.ProjectsService/ListProjects",
]);

const LIBRARY_PROCEDURES = new Set([
  "/realtime.me.library.drive.v1.ShareService/ResolveShare",
  "/realtime.me.library.drive.v1.ShareService/ListSharedItems",
  "/realtime.me.library.drive.v1.ShareService/GetSharedDownload",
]);

const LIBRARY_PREFIXES = [
  "/realtime.me.library.wallpapers.v1.WallpaperPublicService/",
  "/v1/wallpapers/",
  "/v1/shares/",
  "/i/",
];

export default {
  fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    if (STATUS_PROCEDURES.has(url.pathname)) {
      return proxyPublic(request, url, env.STATUS_API_BASE_URL, false);
    }
    if (
      LIBRARY_PROCEDURES.has(url.pathname) ||
      LIBRARY_PREFIXES.some((prefix) => url.pathname.startsWith(prefix))
    ) {
      return proxyPublic(request, url, env.LIBRARY_PUBLIC_API_BASE_URL, true);
    }
    return env.ASSETS.fetch(request).then(withSiteHeaders);
  },
};

function proxyPublic(
  request: Request,
  url: URL,
  configuredBaseUrl: string | undefined,
  preserveCache: boolean,
): Promise<Response> {
  const baseUrl = configuredOrigin(configuredBaseUrl);
  if (!baseUrl) return Promise.resolve(json({ error: "upstream_not_configured" }, 503));
  if (!allowedMethod(request.method, url.pathname)) {
    return Promise.resolve(json({ error: "method_not_allowed" }, 405));
  }
  const upstreamUrl = new URL(url.pathname, baseUrl);
  upstreamUrl.search = url.search;
  const headers = new Headers(request.headers);
  headers.delete("Authorization");
  headers.delete("Cookie");
  const upstreamRequest = new Request(new Request(upstreamUrl, request), { headers });
  return fetch(upstreamRequest).then((response) => withPublicHeaders(response, preserveCache));
}

function configuredOrigin(value: string | undefined): URL | undefined {
  if (!value?.trim()) return undefined;
  try {
    const url = new URL(value.trim());
    if (
      url.protocol !== "https:" ||
      url.username ||
      url.password ||
      (url.pathname !== "/" && url.pathname !== "") ||
      url.search ||
      url.hash
    ) {
      return undefined;
    }
    return url;
  } catch {
    return undefined;
  }
}

function allowedMethod(method: string, path: string): boolean {
  if (path.startsWith("/realtime.me.")) return method === "POST";
  return method === "GET" || method === "HEAD";
}

function withPublicHeaders(response: Response, preserveCache: boolean): Response {
  const headers = new Headers(response.headers);
  headers.delete("Set-Cookie");
  headers.set("Cross-Origin-Resource-Policy", "same-origin");
  if (!preserveCache) headers.set("Cache-Control", "no-store");
  applySecurityHeaders(headers);
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

function withSiteHeaders(response: Response): Response {
  const headers = new Headers(response.headers);
  applySecurityHeaders(headers);
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

function applySecurityHeaders(headers: Headers): void {
  headers.set("Permissions-Policy", "camera=(), microphone=(), geolocation=()");
  headers.set("Referrer-Policy", "no-referrer");
  headers.set("X-Content-Type-Options", "nosniff");
  headers.set("X-Frame-Options", "DENY");
}

function json(body: unknown, status: number): Response {
  const headers = new Headers({
    "Cache-Control": "no-store",
    "Content-Type": "application/json; charset=utf-8",
  });
  applySecurityHeaders(headers);
  return new Response(JSON.stringify(body), {
    status,
    headers,
  });
}
