# Consolidated deployment cutover

This is a one-time conversion from the standalone `realtime-me`, `cloud-driver`,
and `super-manager` deployments to the consolidated repository. It is not a
rolling compatibility procedure: old writers remain stopped after the migration.

## Preserved external identities

Do not rename, recreate, or empty these resources during source installation:

- Status volumes `realtime-me-prometheus` and `realtime-me-status-gateway`;
- Library Compose project `cloud-drive`, bind mounts under `/srv/cloud-drive/data`,
  database name/user, backup snapshots, and `/opt/cloud-drive` installation root;
- Manager Unix user/home, `/var/lib/super-manager/data`, PKI, device records,
  provider login homes, tmux socket, DDNS hostname, and systemd unit names;
- Android application ID `me.realtime`, signing key, Keystore alias, and Wear Data
  Layer paths.

## New configuration ownership

| Old owner | New owner | Required conversion |
| --- | --- | --- |
| Status `CLOUDFLARE_TUNNEL_TOKEN` | `deploy/edge/.env` references one root-only token file | Delete the token key from Status `.env`. |
| Library `CLOUDFLARE_TUNNEL_TOKEN_FILE` and `cloudflared` service | `deploy/edge` | Delete both Library keys; never copy the token into `/etc/cloud-drive/runtime.env`. |
| Library local `edge` network | external `realtime-me-edge` | Create it through `deploy/edge/compose.yaml` before either origin starts. |
| Library `docker-compose.yml` | `deploy/library/compose.yaml` | Stage releases as `incoming-compose/compose.yaml`; there is no old filename fallback. |
| Library `incoming-api` staging | `incoming-source` containing `services/library` contents | Root separately owns the monorepo module, vendor tree, generated Go contracts, and Dockerfile. |
| Library Pages config/scripts | `deploy/web` | Move the local values to `deploy/web/pages.env`; do not retain a second file. |
| Manager source under `/opt/super-manager/source` | monorepo under `/opt/realtime-me` | Keep `/opt/super-manager` only for Node, provider CLIs, home, and persistent credentials. |
| Standalone Manager Android application | Agent/Terminal areas in `apps/mobile` | Re-pair after installing the unified Flutter APK, then revoke the old device. |

The shared remotely managed Tunnel must route the Status host to
`http://status-api:8080`, both Library API hosts to `http://library-api:8080`, and
all unmatched hostnames to an HTTP 404 service.

## Maintenance-window order

1. Stop old Library API/worker writes, Status ingest, Manager execution, and both
   old `cloudflared` containers.
2. Take final consistent Library PostgreSQL/object, Status volume, and Manager
   SQLite/PKI/provider-home backups. Record checksums outside the repository.
3. Install the reviewed monorepo at `/opt/realtime-me`; install the reviewed
   Library release at the existing `/opt/cloud-drive` path.
4. Copy `deploy/edge/.env.example` to `deploy/edge/.env`, point it at the root-only
   Tunnel token file, then run `docker compose ... create` for the edge unit.
5. Install the new Status configuration and start Status. Install the new Library
   `compose.yaml` and restricted operator policy, then run the Library deploy
   script. Confirm both containers are attached to `realtime-me-edge` under their
   stable aliases.
6. Start the single edge connector and verify every configured hostname plus the
   unmatched-host 404 rule.
7. Install the Manager units/wrapper from `deploy/manager`, reload Caddy and
   systemd, then verify pairing, mTLS/bearer control calls, SSE replay, and terminal
   attach against the preserved state.
8. Deploy the Status Worker and all seven Library Pages applications from
   `deploy/web`.
9. Publish the unified Flutter phone APK with the existing `me.realtime` signature,
   publish the Wear APK, pair Manager again, revoke the standalone app's old device,
   and uninstall the standalone APK.
10. Remove old service definitions and local configuration copies. Do not restart
    an old writer or connector after the new baseline is accepted.

## Acceptance and rollback boundary

Before opening writes, verify Status ingest/query separation and Prometheus
scrapes; Library login, upload, worker, anonymous share/wallpaper, backup and
restore; Manager pairing, AG-UI, terminal, restart semantics; and edge streaming,
range requests, WebSocket, CORS, and large uploads.

Rollback is allowed only before an append-only Library migration crosses its
schema boundary. After that point, retain the candidate and fix forward. Restoring
an older Library binary over a migrated database is prohibited. Status and Manager
can be restored only from their matching final snapshots; do not mix backup times
or regenerate the Manager CA.
