# Realtime Me status stack

Self-hosted metrics and status gateway for a private metrics host.

## Services

- Prometheus stores raw metrics for 30 days by default.
- node-exporter exposes host metrics.
- cAdvisor exposes Docker container metrics when the `containers` profile is enabled.
- status-gateway receives phone/watch status, serves Prometheus HTTP service discovery, updates GitHub status, exposes public JSON, Prometheus metrics, and the LAN-only internal status page.
- cloudflared is optional and runs only with the `tunnel` compose profile.

## Setup

```sh
cd deploy/status
cp .env.example .env
openssl rand -base64 32 # paste into STATUS_INGEST_TOKEN  (write access)
openssl rand -base64 32 # paste into STATUS_QUERY_TOKEN   (read access)

# Prometheus authenticates to the gateway's scrape-discovery endpoint with the
# read token. Write it with no trailing newline, before the first `up`.
printf %s "<STATUS_QUERY_TOKEN>" > prometheus/query_token
```

The gateway refuses to start unless both tokens are set, and they must differ:
the read token is pasted into a browser and kept in local storage, so it must
never carry the authority to enroll devices or register scrape targets.

Set `STATUS_GATEWAY_BIND` to the LAN address that should accept local device updates, or keep `127.0.0.1` when only Cloudflare Tunnel should reach the gateway.

Set `GITHUB_TOKEN` to a classic GitHub token with only the `user` scope so the gateway can call `changeUserStatus`.

Configure Cloudflare Tunnel to route `api-status.<BASE_DOMAIN>` to `http://status-gateway:8080`, then put the tunnel token in `.env`.

## Run

```sh
docker compose up -d --build
docker compose --profile containers up -d --build
docker compose --profile tunnel up -d --build
```

## Verify

```sh
GATEWAY=http://<STATUS_GATEWAY_BIND>:18080

curl "$GATEWAY/healthz"
curl -X POST -H 'Content-Type: application/json' -d '{}' \
  "$GATEWAY/realtime.me.status.v1.StatusService/GetPublicStatus"
curl -X POST -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $STATUS_QUERY_TOKEN" -d '{}' \
  "$GATEWAY/realtime.me.status.v1.StatusService/GetInternalStatus"
curl http://127.0.0.1:19090/-/ready

# Scrape discovery requires the read token; without it this must return 401.
curl -o /dev/null -w '%{http_code}\n' "$GATEWAY/api/prometheus/http-sd/node-exporter"
```

Install an additional Linux probe on that device; register its scrape targets centrally through the gateway API:

```sh
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-linux-probe.sh \
  | sudo env STATUS_EXPORTER_HOST=<device-lan-ip> bash
```

Open `http://<STATUS_GATEWAY_BIND>:18080/internal` on the LAN for detailed device, metric, GitHub sync, and active-agent status. The page stores the internal access token only in browser local storage.

Do not commit `.env`; it contains reusable credentials.
