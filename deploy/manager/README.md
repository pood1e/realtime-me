# Linux deployment

The Manager host serves two deliberately different principals on private networks. Flutter clients
use the device mTLS path; human operators use the OIDC Console BFF on the same host:

```text
LAN Flutter  -- mTLS + device token --> manager.realtime.internal :443/:8443
LAN browser  -- OIDC ----------------> console.realtime.internal :9443
Remote clients -- vpn.example.com:1194/UDP --> OpenVPN 10.66.0.1
                                                   |-- Manager :443/:8443
                                                   `-- Console :9443
                                                          |-- loopback --> Manager :3080
                                                          |-- LAN ------> Status :18080
                                                          `-- LAN ------> Library :18081
```

Manager and Console remain loopback-only. Caddy binds both private hostnames to the host's existing
LAN address and OpenVPN `10.66.0.1`; LAN clients do not use VPN. There is no public Manager/Console
hostname, Tunnel route, or TCP port forward. The only permitted public ingress is OpenVPN UDP 1194.
The Manager proxy requires an application client certificate and strips the internal management
key, while Console reaches Manager over host loopback.

Complete [`deploy/vpn`](../vpn/README.md) before this guide. Status, Library, Console, and Manager
must share the root-only internal API key; only Console is allowed to inject it into owner calls.

## 1. Establish the private network boundary

1. Give the Manager host's existing `192.168.0.x` address a DHCP reservation. Do not allocate a
   second LAN address for this deployment.
2. In LAN DNS, resolve `manager.realtime.internal` and `console.realtime.internal` to that existing
   address. In VPN DNS, resolve both names to `10.66.0.1`.
3. Do not create public DNS or Cloudflare Tunnel routes for either internal name. Do not forward
   TCP 443, 8443, 9443, 3080, 8090, or 9876 on the router. Apply the same deny policy to public IPv6.
4. Forward only UDP 1194 to OpenVPN. Its client profile carries no default route and no
   `192.168.0.0/24` route, avoiding collisions when a remote site uses the same LAN prefix.
5. Use the committed nftables rules so only the owner/device VPN pool reaches Manager and Console;
   application mTLS, bearer, internal-key, and OIDC checks remain mandatory above that boundary.

## 2. Optional DDNS for the OpenVPN endpoint

Prefer the router's maintained native DDNS client when it supports the chosen DNS provider.
Otherwise install exact `ddns-go 6.17.2` and the provided `ddns-go.service`.

Create the service account and install the unit:

```bash
sudo useradd --system --home /var/lib/ddns-go --shell /usr/sbin/nologin ddns-go
sudo install -m 0755 ddns-go /usr/local/bin/ddns-go
sudo install -m 0644 deploy/manager/systemd/ddns-go.service /etc/systemd/system/ddns-go.service
sudo systemctl daemon-reload
sudo systemctl enable --now ddns-go.service
```

Its setup UI listens only on `127.0.0.1:9876`. Reach it through an SSH local forward, configure a
dedicated `vpn.example.com` A record, AAAA record, or both, then keep the DNS credential in
`/var/lib/ddns-go/config.yaml` with mode `0600`:

```bash
ssh -L 9876:127.0.0.1:9876 linux-host
chmod 600 /var/lib/ddns-go/config.yaml
```

The record identifies only the UDP OpenVPN endpoint; it must never be reused as a Manager or
Console origin. A normal HTTP CDN proxy does not carry this OpenVPN UDP service. DDNS cannot
traverse CGNAT: the endpoint still needs reachable UDP 1194 or an explicitly designed VPN relay.
`10 s` address checks with a provider comparison every `30` checks limits provider traffic while
reacting promptly to an observed address change.

## 3. Agent and provider CLIs

1. Install Node `24.18.0`, pnpm `11.10.0`, tmux, OpenSSL and qrencode.
2. Install the reviewed monorepo at `/opt/realtime-me`, then build Manager and Console there:

   ```bash
   pnpm install --frozen-lockfile
   pnpm --filter @realtime-me/manager build
   pnpm --filter @realtime-me/console build
   sudo install -d -o root -g root -m 0755 /opt/realtime-me/bin
   GOTOOLCHAIN=go1.26.5 go build -mod=vendor -trimpath \
     -o /opt/realtime-me/bin/realtime-me-console ./services/console/cmd/server
   ```

   Keep the installed worktree and binary root-owned and non-writable by service accounts.
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
6. Install the environment file and CLI wrapper, keep the internal service origins, and set the
   only allowed workspace roots. Keep `agent.env` as plain `KEY=value` lines without shell expansion:

   ```bash
   sudo install -d -m 0750 -o root -g super-manager /etc/super-manager
   sudo install -m 0640 -o root -g super-manager deploy/manager/agent/agent.env.example \
     /etc/super-manager/agent.env
   sudo install -m 0755 -o root -g root deploy/manager/agent/smctl /usr/local/bin/smctl
   ```

7. Install and enable `super-manager-terminal.service` and `super-manager.service`:

   ```bash
   sudo install -m 0644 deploy/manager/systemd/super-manager-terminal.service \
     deploy/manager/systemd/super-manager.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now super-manager-terminal.service super-manager.service
   ```
8. Run management commands through the wrapper so they use the same data directory, service hostname,
   provider paths, `HOME`, and pinned Node runtime as the service:

   ```bash
   sudo -u super-manager -H /usr/local/bin/smctl doctor
   sudo -u super-manager -H /usr/local/bin/smctl pki init
   ```

The tmux server is in a separate systemd cgroup. Restarting only `super-manager.service` detaches
PTY bridges but leaves shells running.

## 4. Owner identity and Console

Register one confidential OIDC client with the exact callback
`https://console.realtime.internal:9443/auth/callback`. Configure the provider to put the canonical
`permissions` string array in both ID and access tokens and to issue the common `realtime-me`
audience. The available values are:

- `PERMISSION_STATUS_INTERNAL_READ`
- `PERMISSION_LIBRARY_MANAGE`
- `PERMISSION_MANAGER_CONTROL`

Status, Library, and Manager require RFC 9068 access tokens (`typ: at+jwt`). Use the same issuer
and audience in all downstream and Console configuration; permissions retain the service boundary.

Install the shared key and root-only Console environment, replace every example origin and secret,
then install the hardened dynamic-user unit. The systemd units expose the key through
`LoadCredential`, never through either environment file:

```bash
sudo install -d -o root -g root -m 0700 /etc/realtime-me
sudo install -m 0400 -o root -g root /secure/internal-api-key \
  /etc/realtime-me/internal-api-key
sudo install -m 0600 -o root -g root deploy/manager/console.env.example \
  /etc/realtime-me/console.env
sudoedit /etc/realtime-me/console.env
sudo install -m 0644 deploy/manager/systemd/realtime-me-console.service \
  /etc/systemd/system/realtime-me-console.service
sudo systemctl daemon-reload
sudo systemctl enable --now realtime-me-console.service
```

Console binds authorization state to a short-lived host-only cookie, keeps OAuth tokens in a
bounded server-side session, checks same-origin mutations, validates private cleartext upstreams
against `MANAGEMENT_CIDRS`, and injects access tokens plus the internal key only into its three
upstreams.

## 5. Caddy and router

Install exact Caddy `2.11.4`, then set `MANAGER_LAN_ADDRESS` to the Manager host's existing LAN
address; no new `192.168.0.x` address is allocated. Keep the second bind for both private sites at
OpenVPN `10.66.0.1`.
Install the Caddyfile, its OpenVPN ordering drop-in, and the application TLS material:

```bash
sudo install -m 0644 deploy/manager/host/Caddyfile.example /etc/caddy/Caddyfile
sudo install -m 0644 deploy/manager/caddy.env.example /etc/realtime-me/caddy.env
sudoedit /etc/realtime-me/caddy.env
sudo install -d -m 0755 /etc/systemd/system/caddy.service.d
sudo install -m 0644 deploy/manager/systemd/caddy.service.d/realtime-me.conf \
  /etc/systemd/system/caddy.service.d/realtime-me.conf
sudo install -d -m 0750 -o root -g caddy /etc/caddy/super-manager
sudo install -m 0644 -o root -g caddy /var/lib/super-manager/data/pki/ca.cert.pem \
  /var/lib/super-manager/data/pki/server.cert.pem /etc/caddy/super-manager/
sudo install -m 0640 -o root -g caddy /var/lib/super-manager/data/pki/server.key.pem \
  /etc/caddy/super-manager/server.key.pem
sudo caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
sudo systemctl daemon-reload
sudo systemctl reload caddy
```

Keep `ca.key.pem` readable only by `super-manager`; Caddy never needs it. The Manager server
certificate is issued for `manager.realtime.internal`.
Export Caddy's local root CA to owner devices and trust it only there so
`https://console.realtime.internal:9443` validates on LAN and OpenVPN.

Do not forward any Caddy TCP port. Use split DNS for both private names: LAN clients receive the
existing LAN address and VPN clients receive `10.66.0.1`. Verify that Caddy is not listening on a
WAN or public IPv6 address. Manager port `443` requires an application client certificate and a
bearer token. Port `8443` routes only `PairDevice`; network reachability plus a 32-byte, ten-minute,
one-time secret and rate limit protect initial enrollment.

## 6. Pairing

Run locally on the Linux host as the service user:

```bash
sudo -u super-manager -H /usr/local/bin/smctl pair create
```

Scan the QR code in the unified Flutter phone app. The payload bootstraps the private CA, service URL, pairing
URL, expiry, and one-time secret. After redemption, all normal traffic uses both mTLS and a
revocable inner bearer token.

Device certificates expire after 365 days. Re-pair before expiry; the client never falls back to
an expired certificate or weaker TLS. Server certificates expire after 825 days and must be
reissued with `smctl pki renew-server`, recopied to Caddy, and reloaded before that date.
The same renewal is required after changing the internal Manager hostname.

If an expired/revoked client can only erase its local credentials, remove its still-active server
record locally:

```bash
sudo -u super-manager -H /usr/local/bin/smctl device list
sudo -u super-manager -H /usr/local/bin/smctl device revoke <device-uid>
```

## Operational limits

- A residential address change affects only new remote OpenVPN connections until the optional VPN
  DDNS record and recursive caches converge; it never makes Manager a public service.
- Prefer wired Ethernet and a DHCP reservation for the Linux host.
- Back up `/var/lib/super-manager/data` with mode-preserving encrypted storage. It contains the CA
  key, device-token hashing secret, SQLite history and provider mapping data.
- Monitor SQLite growth. The MVP replays all retained semantic events and does not yet compact old
  thread history automatically.
- Expose only OpenVPN UDP 1194 publicly. Keep every Caddy TCP port private and run both CLIs and the
  API under the dedicated non-root account. The workspace allowlist restricts registered starting
  directories; Unix ownership and permissions are the actual filesystem boundary for bypass-mode agents.
