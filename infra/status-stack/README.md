# Realtime Me status stack

Self-hosted metrics and status gateway for a private metrics host.

## Services

- Prometheus stores raw metrics for 30 days by default.
- node-exporter exposes host metrics.
- cAdvisor exposes Docker container metrics when the `containers` profile is enabled.
- status-gateway receives phone/watch/agent status, updates GitHub status, exposes public JSON and Prometheus metrics.
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
curl -H "Authorization: Bearer $STATUS_INGEST_TOKEN" http://<STATUS_GATEWAY_BIND>:18080/api/internal-status
curl http://127.0.0.1:19090/-/ready
```

Do not commit `.env`; it contains reusable credentials.
