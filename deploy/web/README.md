# Public Site release

`apps/web/site` is the single anonymous web surface. It contains the public
Status/Profile/Projects pages plus Library wallpapers and token-scoped shares.
The Cloudflare Worker serves the SPA and proxies only the explicit public
procedures allowlisted in `src/worker.ts`; it strips cookies and authorization.

```sh
VITE_CONSOLE_URL=https://console.realtime.internal:9443 pnpm --filter @realtime-me/site build
pnpm --dir apps/web/site deploy -- \
  --var STATUS_API_BASE_URL:https://api-status.example.com \
  --var LIBRARY_PUBLIC_API_BASE_URL:https://api-library-public.example.com
```

Set the custom domain in `apps/web/site/wrangler.jsonc` before deployment. There
are no separate Library Pages projects.

The authenticated application is built from `apps/web/console` and served by
the Console BFF. Its host deployment and OIDC configuration live under
[`deploy/manager`](../manager/README.md).
