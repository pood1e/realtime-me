# Unified Web and OIDC cutover

This is a one-time replacement of the legacy Status Worker, seven Library Pages
applications, browser query/password credentials, and their deployment scripts.
There is no dual route, legacy configuration fallback, or parallel writer.

## Preserve

- Status volumes `realtime-me-prometheus` and `realtime-me-status-gateway`;
- Library Compose project `cloud-drive`, PostgreSQL/database identity, objects,
  backups, and `/opt/cloud-drive` paths;
- Manager Unix user/home, SQLite, PKI, paired devices, provider logins, tmux
  socket, internal service hostname, and workspace roots;
- Android `me.realtime` signing identity and Wear Data Layer paths.

## Convert configuration

| Remove | Add |
| --- | --- |
| Status `tokens.query`, browser localStorage token | `tokens.discovery` for Prometheus plus `OIDC_ISSUER` / `STATUS_AUTH_AUDIENCE` |
| `prometheus/query_token` | `prometheus/discovery_token` |
| Library password hash, session secret, private/public app origin lists | `PUBLIC_SITE_ORIGIN`, `CONSOLE_ORIGIN`, `OIDC_ISSUER`, `LIBRARY_AUTH_AUDIENCE` |
| Manager device-only application authorization | `OIDC_ISSUER` / `MANAGER_AUTH_AUDIENCE`; keep device mTLS + bearer unchanged |
| seven Pages projects and Status web app | one Site Worker and one Console BFF/SPA |
| private Status/Library Tunnel hostnames | existing LAN binds + non-conflicting OpenVPN overlay binds + one internal API key |

Register one confidential OIDC client with callback
`https://console.realtime.internal:9443/auth/callback`. ID and access tokens must carry the
common owner audience and canonical `permissions` array. Status, Library, and
Manager all require access-token `typ: at+jwt`; ID tokens stop at Console.

## Maintenance-window order

1. Freeze Library writes and Manager executions; take final Status, Library, and
   Manager backups with external checksums.
2. Install OpenVPN according to `deploy/vpn`, issue unique client certificates,
   verify that no `192.168.0.0/24` route is pushed, and generate/distribute the root-only internal API key.
3. Install the reviewed monorepo and root-controlled generated Auth contracts,
   `libs/go/authn`, `libs/go/serviceauth`, vendor tree, Dockerfiles, systemd units,
   public allowlists, and policy validators.
4. Create `prometheus/discovery_token`, install new Status OIDC configuration,
   then start Status and confirm public RPC plus scrape discovery.
5. Install new Library runtime configuration and restricted Compose policy; run
   backup + forward-only migrate + deploy.
6. Install Manager OIDC configuration and restart Manager. Verify the preserved
   Flutter mTLS device path before continuing.
7. Build Console SPA and binary, install `/etc/realtime-me/console.env`, start
   `realtime-me-console.service`, and add the VPN-bound Console Caddy hostname.
8. Verify OpenVPN, internal-key rejection, OIDC login/logout, and each permission
   independently against Status, Library, and Manager. Confirm cross-Origin mutations fail.
9. Publish `apps/web/site` with both public upstream variables. Verify its Worker
   cannot reach any internal/private procedure.
10. Delete private Tunnel hostnames, legacy Pages projects, old Worker routes, old password/session secrets,
   and browser query-token instructions. Do not restart them.

## Acceptance

- anonymous Status, wallpapers, and token-scoped shares work through Site;
- Console browser storage contains no owner access token;
- downstream services reject missing audience/permission even if a route is
  manually invoked;
- private routes are unreachable outside the trusted LAN/OpenVPN boundary and reject
  a missing or wrong internal API key before OIDC;
- Library upload/range/provider callback and Manager AG-UI/WebSocket/PTY streaming
  work through Console;
- Prometheus discovery and gateway process metrics accept only its workload token;
- Manager device hostname is reachable only through LAN/OpenVPN and still requires mTLS.

Library rollback is allowed only before its append-only migration boundary. Once
new writes or migration occur, fix forward or restore one matching PostgreSQL +
objects snapshot. Never restore an older binary onto a migrated database and never
regenerate the preserved Manager CA during rollback.
