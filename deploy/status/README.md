# Realtime Me status stack

Self-hosted metrics and status gateway for a private metrics host.

## Services

- Prometheus stores raw metrics for 30 days by default.
- node-exporter exposes host metrics.
- cAdvisor exposes Docker container metrics when the `containers` profile is enabled.
- status-gateway receives phone/watch status, serves Prometheus HTTP service discovery, updates GitHub status, exposes public JSON, Prometheus metrics, and the LAN-only internal status page.

Public ingress is owned by [`deploy/edge`](../edge/README.md), not this release unit. The gateway
joins the external `realtime-me-edge` network as `status-api` while Prometheus and exporters remain
on the internal Status backend network.

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

Configure the shared Cloudflare Tunnel to route `api-status.<BASE_DOMAIN>` to
`http://status-api:8080`. Start `deploy/edge` once before bringing up this stack so the external
network exists. No Tunnel credential belongs in the Status `.env` file.

## Run

```sh
docker compose up -d --build --remove-orphans
docker compose --profile containers up -d --build --remove-orphans
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
curl -o /dev/null -w '%{http_code}\n' "$GATEWAY/api/prometheus/http-sd/probe-agent"
```

Install the unified probe on a Linux, macOS, or Windows device; register its one scrape target centrally through the gateway API:

```sh
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-probe.py \
  | sudo env REALTIME_PROBE_HOST=<device-lan-ip> python3 -
```

Open `http://<STATUS_GATEWAY_BIND>:18080/internal` on the LAN for detailed device, metric, GitHub sync, and active-agent status. The page stores the internal access token only in browser local storage.

Do not commit `.env`; host-specific Compose settings stay local even though application secrets
live in `gateway.yaml`.
