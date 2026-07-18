type AssetBinding = {
  fetch(request: Request): Promise<Response>;
};

type Env = {
  ASSETS: AssetBinding;
  STATUS_API_BASE_URL?: string;
};

// The browser speaks only ConnectRPC, POSTed to /<proto package>.<Service>/<Method>,
// and it reads. The four read services are named one by one rather than by their
// shared package prefix, which would also carry EnrollmentService and IngestService
// -- the write half of the API, which no browser has any business reaching.
// Nothing under /api/ is proxied either: those are the gateway's control-plane
// routes, such as Prometheus scrape discovery.
const PROXY_PREFIXES = [
  "/realtime.me.status.v1.StatusService/",
  "/realtime.me.site.v1.ProfileService/",
  "/realtime.me.site.v1.ProjectsService/",
  "/realtime.me.status.v1.MetricsService/",
];

export default {
  fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    if (PROXY_PREFIXES.some((prefix) => url.pathname.startsWith(prefix))) {
      return proxyStatusApi(request, url, env);
    }
    return env.ASSETS.fetch(request);
  },
};

function proxyStatusApi(request: Request, url: URL, env: Env): Promise<Response> {
  const upstreamBaseUrl = env.STATUS_API_BASE_URL?.trim().replace(/\/+$/, "");
  if (!upstreamBaseUrl) {
    return Promise.resolve(json({ error: "status_api_not_configured" }, 503));
  }

  const upstreamUrl = new URL(`${upstreamBaseUrl}${url.pathname}`);
  upstreamUrl.search = url.search;
  return fetch(new Request(upstreamUrl, request)).then((response) => withNoStore(response));
}

function withNoStore(response: Response): Response {
  const headers = new Headers(response.headers);
  headers.set("Cache-Control", "no-store");
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

function json(body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Cache-Control": "no-store",
      "Content-Type": "application/json; charset=utf-8",
    },
  });
}
