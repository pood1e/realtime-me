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
cd infra/status-stack
cp .env.example .env
openssl rand -base64 32 # paste into STATUS_INGEST_TOKEN
```

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
curl http://<STATUS_GATEWAY_BIND>:18080/healthz
curl http://<STATUS_GATEWAY_BIND>:18080/api/public-status
curl -H "Authorization: Bearer $STATUS_INGEST_TOKEN" http://<STATUS_GATEWAY_BIND>:18080/api/internal/status
curl http://127.0.0.1:19090/-/ready
```

Register an additional Linux device by running the installer on that device and passing its reachable LAN address as `STATUS_EXPORTER_HOST`:

```sh
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-linux-probe.sh \
  | sudo env STATUS_GATEWAY_URL=http://<gateway-host>:18080 \
      STATUS_EXPORTER_HOST=<device-lan-ip> \
      bash
```

Open `http://<STATUS_GATEWAY_BIND>:18080/internal` on the LAN for detailed device, metric, GitHub sync, and active-agent status. The page stores the internal access token only in browser local storage.

Do not commit `.env`; it contains reusable credentials.
