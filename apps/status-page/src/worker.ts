type AssetBinding = {
  fetch(request: Request): Promise<Response>;
};

type Env = {
  ASSETS: AssetBinding;
  STATUS_API_BASE_URL?: string;
};

const PUBLIC_STATUS_PATH = '/api/public-status';

export default {
  fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    if (url.pathname === PUBLIC_STATUS_PATH) return proxyStatusApi(url, env);
    return env.ASSETS.fetch(request);
  },
};

function proxyStatusApi(url: URL, env: Env): Promise<Response> {
  const upstreamBaseUrl = env.STATUS_API_BASE_URL?.trim().replace(/\/+$/, '');
  if (!upstreamBaseUrl) {
    return Promise.resolve(json({ error: 'status_api_not_configured' }, 503));
  }

  const upstreamUrl = new URL(`${upstreamBaseUrl}${PUBLIC_STATUS_PATH}`);
  upstreamUrl.search = url.search;
  return fetch(upstreamUrl, {
    headers: { Accept: 'application/json' },
  }).then((response) => withNoStore(response));
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
