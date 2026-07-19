# Realtime Me status stack

Self-hosted metrics and status gateway for a private metrics host.

## Services

- Prometheus stores raw metrics for 30 days by default.
- node-exporter exposes host metrics.
- cAdvisor exposes Docker container metrics when the `containers` profile is enabled.
- status-gateway receives phone/watch status, serves Prometheus HTTP service discovery, updates GitHub status, and exposes public and owner ConnectRPC APIs.
- status-public is the only container attached to the shared edge network and forwards three exact public RPC procedures.

Public ingress is owned by [`deploy/edge`](../edge/README.md), not this release unit. Only the
`status-public` Caddy sidecar joins `realtime-me-edge` as `status-public`; it strips credentials and
cannot route owner or workload procedures. The gateway binds its host port to the Status host's
existing LAN address and its dedicated OpenVPN address `10.66.0.10`; it never claims another LAN IP.
It retains a private backend connection and uses a separate egress network for OIDC and
GitHub. Prometheus alone also joins `probe-egress`; discovered targets must be inside the CIDRs and
single port declared under `probe` in `gateway.yaml`.

## Setup

```sh
cd deploy/status
cp .env.example .env
cp gateway.example.yaml gateway.yaml
cp projects.example.json projects.json
sudo install -d -m 0700 /etc/realtime-me
# Copy the management-plane key generated according to deploy/vpn/README.md.
sudo install -m 0400 -o root -g root /secure/internal-api-key \
  /etc/realtime-me/internal-api-key
openssl rand -base64 32 # paste into tokens.ingest
openssl rand -base64 32 # paste into tokens.discovery

# Prometheus authenticates to the gateway's scrape-discovery and raw process-metrics
# endpoints with the workload discovery token. Write it with no trailing newline
# before the first `up`.
printf %s "<tokens.discovery>" > prometheus/discovery_token
```

The gateway refuses to start unless both workload tokens are set and distinct.
It also refuses missing or invalid `probe.allowed_cidrs`/`probe.port` settings. Use
literal device IP addresses, keep the CIDRs narrow, and allow the Probe port through
each device firewall only from the Prometheus host.
Human internal Status and Metrics calls do not use either workload token: they
require both the management-plane key and an OIDC access token with
`PERMISSION_STATUS_INTERNAL_READ`. Only the authenticated Console BFF injects both.

Set `STATUS_GATEWAY_LAN_BIND` to this host's existing static LAN address and install the `status`
OpenVPN client before Compose so `STATUS_GATEWAY_VPN_BIND=10.66.0.10` exists. LAN agents connect
directly; remote agents use only the overlay address. Never bind the gateway to a wildcard or public address.

Set `github.status_token` in `gateway.yaml` to a classic GitHub token with only the `user` scope so the gateway can call `changeUserStatus`.

Configure the shared Cloudflare Tunnel public Status hostname to route only to
`http://status-public:8080`. Start `deploy/edge` once before this stack so the external network
exists. No Tunnel credential or private Status hostname belongs in the Status configuration.

## Run

```sh
docker compose up -d --build --remove-orphans
docker compose --profile containers up -d --build --remove-orphans
```

## Verify

```sh
GATEWAY=http://<STATUS_HOST_LAN_IP>:18080 # use 10.66.0.10 when remote

curl "$GATEWAY/healthz"
curl -X POST -H 'Content-Type: application/json' -d '{}' \
  "https://<PUBLIC_STATUS_HOST>/realtime.me.status.v1.StatusService/GetPublicStatus"
curl http://127.0.0.1:19090/-/ready

# Scrape discovery and /metrics require the workload discovery token; without it both return 401.
curl -o /dev/null -w '%{http_code}\n' "$GATEWAY/api/prometheus/http-sd/probe-agent"
curl -o /dev/null -w '%{http_code}\n' "$GATEWAY/metrics"
```

Install the unified probe on a Linux, macOS, or Windows device; register its one scrape target centrally through the gateway API:

```sh
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-probe.py \
  | sudo env REALTIME_PROBE_HOST=<device-lan-ip> python3 -
```

Open the Console `/status` route for detailed device, metric, GitHub sync, and
active-agent status. The browser receives only the Console session cookie; it never
stores the owner access token.

Do not commit `.env`; host-specific Compose settings stay local even though application secrets
live in `gateway.yaml`.
