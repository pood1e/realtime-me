type AssetBinding = {
  fetch(request: Request): Promise<Response>;
};

type Env = {
  ASSETS: AssetBinding;
  STATUS_API_BASE_URL?: string;
};

// The browser speaks only ConnectRPC, POSTed to /<proto package>.<Service>/<Method>.
// Nothing under /api/ is proxied: those are the gateway's control-plane routes,
// such as Prometheus scrape discovery, and the browser must never reach them.
const PROXY_PREFIXES = ['/realtime.me.v1.'];

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
  const upstreamBaseUrl = env.STATUS_API_BASE_URL?.trim().replace(/\/+$/, '');
  if (!upstreamBaseUrl) {
    return Promise.resolve(json({ error: 'status_api_not_configured' }, 503));
  }

  const upstreamUrl = new URL(`${upstreamBaseUrl}${url.pathname}`);
  upstreamUrl.search = url.search;
  return fetch(new Request(upstreamUrl, request)).then((response) => withNoStore(response));
}

function withNoStore(response: Response): Response {
  const headers = new Headers(response.headers);
  headers.set('Cache-Control', 'no-store');
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
      'Cache-Control': 'no-store',
      'Content-Type': 'application/json; charset=utf-8',
    },
  });
}
