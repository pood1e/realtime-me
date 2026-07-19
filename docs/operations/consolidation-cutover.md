# Unified Web and OIDC cutover

This is a one-time replacement of the legacy Status Worker, seven Library Pages
applications, browser query/password credentials, and their deployment scripts.
There is no dual route, legacy configuration fallback, or parallel writer.

## Preserve

- Status volumes `realtime-me-prometheus` and `realtime-me-status-gateway`;
- Library Compose project `cloud-drive`, PostgreSQL/database identity, objects,
  backups, and `/opt/cloud-drive` paths;
- Manager Unix user/home, SQLite, PKI, paired devices, provider logins, tmux
  socket, DDNS hostname, and workspace roots;
- Android `me.realtime` signing identity and Wear Data Layer paths.

## Convert configuration

| Remove | Add |
| --- | --- |
| Status `tokens.query`, browser localStorage token | `tokens.discovery` for Prometheus plus `OIDC_ISSUER` / `STATUS_AUTH_AUDIENCE` |
| `prometheus/query_token` | `prometheus/discovery_token` |
| Library password hash, session secret, private/public app origin lists | `PUBLIC_SITE_ORIGIN`, `CONSOLE_ORIGIN`, `OIDC_ISSUER`, `LIBRARY_AUTH_AUDIENCE` |
| Manager device-only application authorization | `OIDC_ISSUER` / `MANAGER_AUTH_AUDIENCE`; keep device mTLS + bearer unchanged |
| seven Pages projects and Status web app | one Site Worker and one Console BFF/SPA |

Register one confidential OIDC client with callback
`https://console.example.com/auth/callback`. ID and access tokens must carry the
common owner audience and canonical `permissions` array. Status, Library, and
Manager all require access-token `typ: at+jwt`; ID tokens stop at Console.

## Maintenance-window order

1. Freeze Library writes and Manager executions; take final Status, Library, and
   Manager backups with external checksums.
2. Install the reviewed monorepo and root-controlled generated Auth contracts,
   `libs/go/authn`, vendor tree, Dockerfiles, systemd units, and policy validators.
3. Create `prometheus/discovery_token`, install new Status OIDC configuration,
   then start Status and confirm public RPC plus scrape discovery.
4. Install new Library runtime configuration and restricted Compose policy; run
   backup + forward-only migrate + deploy.
5. Install Manager OIDC configuration and restart Manager. Verify the preserved
   Flutter mTLS device path before continuing.
6. Build Console SPA and binary, install `/etc/realtime-me/console.env`, start
   `realtime-me-console.service`, and add the Console Caddy hostname.
7. Verify OIDC login/logout and each permission independently against Status,
   Library, and Manager. Confirm cross-Origin mutations fail.
8. Publish `apps/web/site` with both public upstream variables. Verify its Worker
   cannot reach any internal/private procedure.
9. Delete legacy Pages projects, old Worker routes, old password/session secrets,
   and browser query-token instructions. Do not restart them.

## Acceptance

- anonymous Status, wallpapers, and token-scoped shares work through Site;
- Console browser storage contains no owner access token;
- downstream services reject missing audience/permission even if a route is
  manually invoked;
- Library upload/range/provider callback and Manager AG-UI/WebSocket/PTY streaming
  work through Console;
- Prometheus discovery and gateway process metrics accept only its workload token;
- Manager public device hostname still requires mTLS.

Library rollback is allowed only before its append-only migration boundary. Once
new writes or migration occur, fix forward or restore one matching PostgreSQL +
objects snapshot. Never restore an older binary onto a migrated database and never
regenerate the preserved Manager CA during rollback.
