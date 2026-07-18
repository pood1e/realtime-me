# Linux deployment

The MVP uses the residential public address directly. There is no VPS, tunnel, CDN, or relay in
the data path:

```text
Android -- DNS lookup --> residential public IPv4/IPv6
                              |
                         router/firewall
                              |
                    Caddy :443 / :8443
                              |
                       127.0.0.1:3080
                              |
                    Super Manager + CLIs

ddns-go -- provider API --> A/AAAA record
```

The API process remains loopback-only. Caddy exposes the authenticated API on `443` and the
single-use pairing RPC on `8443`.

## 1. Check that direct inbound access is possible

DDNS only updates DNS. It cannot traverse carrier-grade NAT.

1. Compare the router WAN IPv4 address with an independently observed public IPv4 address. If
   they differ, or the WAN address is private/shared (`10/8`, `172.16/12`, `192.168/16`, or
   `100.64/10`), IPv4 port forwarding will not work.
2. If globally routed IPv6 is available, AAAA-only access is valid, but both the Linux firewall
   and router IPv6 firewall must allow inbound TCP, and the Android network must support IPv6.
   Publish only address families that have passed an external inbound check; a broken AAAA record
   can delay or prevent otherwise valid IPv4 connections.
3. If neither family is globally reachable, direct DDNS is not viable. Use an overlay VPN or
   restore a domestic relay as a separate deployment profile; do not silently run both paths.
4. Check whether the ISP filters inbound `443` or `8443`. If needed, map two external high ports
   to the same internal ports and include those external ports in `SM_PUBLIC_URL` and
   `SM_PAIRING_URL`.
5. Check router NAT loopback from the home LAN. If it is unavailable, configure router split DNS
   so the same DDNS hostname resolves to the Linux host's private address at home. Keep using the
   hostname covered by the private certificate; do not add an insecure IP-based client profile.

## 2. DNS updater

Prefer the router's maintained native DDNS client when it supports the chosen DNS provider.
Otherwise install exact `ddns-go 6.17.2` and the provided `ddns-go.service`.

Create the service account and install the unit:

```bash
sudo useradd --system --home /var/lib/ddns-go --shell /usr/sbin/nologin ddns-go
sudo install -m 0755 ddns-go /usr/local/bin/ddns-go
sudo install -m 0644 deploy/systemd/ddns-go.service /etc/systemd/system/ddns-go.service
sudo systemctl daemon-reload
sudo systemctl enable --now ddns-go.service
```

Its setup UI listens only on `127.0.0.1:9876`. Reach it through an SSH local forward, configure
an A record, an AAAA record, or both, then keep the DNS credential in
`/var/lib/ddns-go/config.yaml` with mode `0600`:

```bash
ssh -L 9876:127.0.0.1:9876 linux-host
chmod 600 /var/lib/ddns-go/config.yaml
```

Use a DNS-only record. A CDN/proxy would reintroduce an extra hop and may interfere with
application mTLS. `10 s` address checks with a provider comparison every `30` checks limits
provider traffic while reacting promptly to an observed address change.

## 3. Agent and provider CLIs

1. Install Node `24.18.0`, pnpm `11.10.0`, tmux, OpenSSL and qrencode.
2. Build this repository with `pnpm install --frozen-lockfile && pnpm build`.
3. Install exact provider versions `@openai/codex@0.144.5` and
   `@anthropic-ai/claude-code@2.1.195` under `/opt/super-manager/providers`.
4. Create a locked service account with a real shell for tmux, then create its state and workspace
   directories:

   ```bash
   sudo useradd --system --create-home --home-dir /var/lib/super-manager \
     --shell /bin/bash super-manager
   sudo install -d -m 0700 -o super-manager -g super-manager /var/lib/super-manager/data
   sudo install -d -m 0750 -o super-manager -g super-manager /srv/workspaces
   ```

5. Authenticate both CLIs as `super-manager`. Do not configure API keys; `smctl doctor` rejects
   unsupported subscription state:

   ```bash
   provider_path=/opt/super-manager/node/bin:/opt/super-manager/providers/node_modules/.bin:/usr/local/bin:/usr/bin:/bin
   sudo -u super-manager -H env PATH="$provider_path" codex login
   sudo -u super-manager -H env PATH="$provider_path" claude auth login
   ```
6. Install the environment file and CLI wrapper, replace the example domain, and set the only
   allowed workspace roots. Keep `agent.env` as plain `KEY=value` lines without shell expansion:

   ```bash
   sudo install -d -m 0750 -o root -g super-manager /etc/super-manager
   sudo install -m 0640 -o root -g super-manager deploy/agent/agent.env.example \
     /etc/super-manager/agent.env
   sudo install -m 0755 -o root -g root deploy/agent/smctl /usr/local/bin/smctl
   ```

7. Install and enable `super-manager-terminal.service` and `super-manager.service`.
8. Run management commands through the wrapper so they use the same data directory, hostname,
   provider paths, `HOME`, and pinned Node runtime as the service:

   ```bash
   sudo -u super-manager -H /usr/local/bin/smctl doctor
   sudo -u super-manager -H /usr/local/bin/smctl pki init
   ```

The tmux server is in a separate systemd cgroup. Restarting only `super-manager.service` detaches
PTY bridges but leaves shells running.

## 4. Caddy and router

Install exact Caddy `2.11.4`. Replace `manager.example.com` in
`deploy/host/Caddyfile.example`, then install the application TLS material:

```bash
sudo install -d -m 0750 -o root -g caddy /etc/caddy/super-manager
sudo install -m 0644 -o root -g caddy /var/lib/super-manager/data/pki/ca.cert.pem \
  /var/lib/super-manager/data/pki/server.cert.pem /etc/caddy/super-manager/
sudo install -m 0640 -o root -g caddy /var/lib/super-manager/data/pki/server.key.pem \
  /etc/caddy/super-manager/server.key.pem
sudo caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
sudo systemctl reload caddy
```

Keep `ca.key.pem` readable only by `super-manager`; Caddy never needs it. The private server
certificate is issued for the DDNS hostname, so a public IP change does not require reissuance.

Forward router TCP `443 -> host:443` and `8443 -> host:8443`. For IPv6, add firewall rules rather
than NAT. Do not forward `3080` or `9876`. Port `443` requires an application client certificate
and a bearer token. Port `8443` routes only `PairDevice`; the RPC also requires a 32-byte,
ten-minute, one-time secret and is rate-limited. The `8443` rule may remain disabled except while
adding a device.

## 5. Pairing

Run locally on the Linux host as the service user:

```bash
sudo -u super-manager -H /usr/local/bin/smctl pair create
```

Scan the QR code in the Android app. The payload bootstraps the private CA, service URL, pairing
URL, expiry, and one-time secret. After redemption, all normal traffic uses both mTLS and a
revocable inner bearer token.

Device certificates expire after 365 days. Re-pair before expiry; the client never falls back to
an expired certificate or weaker TLS. Server certificates expire after 825 days and must be
reissued with `smctl pki renew-server`, recopied to Caddy, and reloaded before that date.
The same renewal is required after changing the DDNS hostname.

If an expired/revoked client can only erase its local credentials, remove its still-active server
record locally:

```bash
sudo -u super-manager -H /usr/local/bin/smctl device list
sudo -u super-manager -H /usr/local/bin/smctl device revoke <device-uid>
```

## Operational limits

- A residential address change causes a short outage until the updater and recursive DNS caches
  converge; use the lowest TTL allowed by the provider without assuming zero downtime.
- Prefer wired Ethernet and a DHCP reservation for the Linux host.
- Back up `/var/lib/super-manager/data` with mode-preserving encrypted storage. It contains the CA
  key, device-token hashing secret, SQLite history and provider mapping data.
- Monitor SQLite growth. The MVP replays all retained semantic events and does not yet compact old
  thread history automatically.
- Expose only the two Caddy ports and keep both CLIs and the API under the dedicated non-root
  account. The workspace allowlist restricts registered starting directories; Unix ownership and
  permissions are the actual filesystem boundary for bypass-mode agents.
