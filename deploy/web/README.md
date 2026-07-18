# Web release unit

Status is deployed as a Worker with static assets. The seven Library frontends
are deployed to independent Cloudflare Pages projects. This directory owns their
deployment configuration; application source remains under `apps/web`.

## Status Worker

```sh
pnpm --filter @realtime-me/status-web build
pnpm --dir apps/web/status deploy -- \
  --var STATUS_API_BASE_URL:https://api-status.example.com
```

The Worker proxies only the public Status and Site ConnectRPC procedures listed
in `apps/web/status/src/worker.ts`.

## Library Pages

```sh
cp deploy/web/pages.env.example deploy/web/pages.env
$EDITOR deploy/web/pages.env
deploy/web/deploy-library-pages.sh
```

The local environment file supplies exact HTTPS origins and existing Pages
project names. It contains no API, account, or Tunnel credentials.
