# realtime-me

Personal realtime status, local content library, and remote coding-agent/terminal
management in one monorepo.

## Architecture

Three bounded contexts share contracts and tooling, not storage or implementation:

- **Status** (`services/status`): phone/Watch ingestion, Prometheus discovery,
  GitHub status, public profile/projects, internal metrics;
- **Library** (`services/library`): Drive, Books, Music, Images, Wallpapers,
  Shares, PostgreSQL, objects, worker, and migrations;
- **Manager** (`services/manager`): Codex/Claude runtimes, AG-UI, workspaces,
  persistent terminals, SQLite, and paired-device control.

Two web applications present them:

- `apps/web/site`: anonymous Status/Profile/Projects plus Library wallpapers and
  token-scoped shares; a Worker proxies only an explicit public allowlist;
- `apps/web/console`: the authenticated owner console for Status, Library, and
  Manager, served by the Go OIDC BFF in `services/console`.

The Console is reachable directly on the trusted LAN or through OpenVPN's separate overlay. It uses browser-bound authorization-code state +
PKCE, bounded server-side sessions, a host-only `HttpOnly` cookie, same-origin CSRF checks, an
internal service key, and downstream JWT permissions. The browser stores neither downstream
credential.

Workload/device credentials remain deliberately separate:

- Status producers use an ingest token;
- Prometheus uses a workload token limited to target discovery and gateway process scraping;
- Manager Flutter clients use device mTLS plus a revocable bearer;
- anonymous users reach only public Status/wallpaper/share handlers.

Manager and Console have no public hostname or TCP port forward. Local clients use
the existing LAN addresses; remote clients first join the non-conflicting
`10.66.0.0/24` OpenVPN overlay. Only the OpenVPN UDP endpoint may be public.

See [the unified architecture](docs/architecture/project-consolidation.md),
[Library architecture](docs/library/architecture.md), and
[the one-time cutover](docs/operations/consolidation-cutover.md).

## Clients and collection

- `apps/mobile`: the only phone APK, Flutter `me.realtime`, with Status, Agent,
  Terminal, Pairing, and Settings;
- `apps/mobile/android`: native Wear listener, foreground sync, WorkManager,
  Keystore, and snapshot bridge that continue without a Flutter engine;
- `apps/watch`: Kotlin Wear OS collection and Data Layer publishing;
- `scripts/probe`: one credential-free Python `/metrics` endpoint for Linux,
  macOS, and Windows.

The probe supports Linux `aarch64`, including 64-bit Raspberry Pi OS. OS-specific
installers only adapt systemd, launchd, and Task Scheduler; collection and metrics
remain one implementation.

## Repository layout

```text
apps/mobile                 Flutter phone app + Android native core
apps/watch                  Kotlin Wear OS app
apps/web/site               public SPA + Cloudflare Worker
apps/web/console            owner SPA
services/status             Go Status service
services/library            Go API/worker/migrate
services/manager            TypeScript/Fastify Manager
services/console            Go OIDC BFF/static host
libs/go/authn               shared Go OIDC verifier
libs/go/serviceauth         shared Go internal-key verifier
proto/realtime/me           canonical language-neutral contracts
gen/go                      generated Go contracts
packages/*-contracts-*      generated TypeScript/Dart contracts
packages/web-ui             shared React primitives
packages/web-shell          shared theme and Console shell
packages/status-web         reusable Status feature
packages/library-web        reusable Library feature/API layer
deploy                      independent runtime release units
```

`proto/` is the only contract source. Generated code is committed and must be
updated through the root generation command; do not hand-edit it.

## Requirements

- Go 1.26.x;
- Node 24.18+ and pnpm 11.10;
- Flutter 3.44.6 / Dart 3.12.2;
- JDK 17 and Android SDK API 37;
- Buf plus `protoc-gen-go` and `protoc-gen-connect-go`.

## Build and verify

```sh
make generate
pnpm check
pnpm build
go vet ./...
go build ./...
./gradlew :apps:watch:assembleDebug
(cd apps/mobile && flutter analyze && flutter build apk --debug)
```

The complete cross-language gate is:

```sh
make verify
```

## Owner OIDC contract

Register one confidential client with callback:

```text
https://console.realtime.internal:9443/auth/callback
```

Issue the common owner access-token audience (recommended `realtime-me`) and put
the following canonical strings in the `permissions` array of ID and access
tokens:

```text
PERMISSION_STATUS_INTERNAL_READ
PERMISSION_LIBRARY_MANAGE
PERMISSION_MANAGER_CONTROL
```

Each owner downstream first verifies the internal management key, then independently verifies
issuer, audience, expiry, subject, its permission, and the bounded RFC 9068 `typ: at+jwt`
access-token profile. ID tokens are accepted only by the Console login boundary. VPN and key
setup lives in [`deploy/vpn`](deploy/vpn/README.md).

## Status stack quick start

```sh
cd deploy/status
cp gateway.example.yaml gateway.yaml
cp projects.example.json projects.json
cp .env.example .env

# First complete deploy/vpn and install the shared key on this host.
sudo install -d -m 0700 /etc/realtime-me
sudo install -m 0400 /secure/internal-api-key /etc/realtime-me/internal-api-key

openssl rand -base64 32 # tokens.ingest
openssl rand -base64 32 # tokens.discovery
printf %s '<tokens.discovery>' > prometheus/discovery_token
chmod 700 . && chmod 644 gateway.yaml projects.json

docker compose --env-file ../edge/.env -f ../edge/compose.yaml up -d
docker compose up -d --build --remove-orphans
```

`gateway.yaml` owns workload tokens, GitHub credentials, and the public profile.
`.env` owns Compose wiring plus the Status OIDC issuer/audience. The internal
dashboard is available only through Console `/status`; there is no browser query
token.

Install the same probe on each desktop/server and then register its single target
with the Status ingest API:

```sh
# Linux
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-probe.py \
  | sudo python3 -

# macOS (login user, not sudo)
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-probe.py \
  | python3 -

# Windows PowerShell
irm https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-probe.py | py -3 -
```

## Web release

Public Site:

```sh
VITE_CONSOLE_URL=https://console.realtime.internal:9443 pnpm --filter @realtime-me/site build
pnpm --dir apps/web/site deploy -- \
  --var STATUS_API_BASE_URL:https://api-status.example.com \
  --var LIBRARY_PUBLIC_API_BASE_URL:https://api-library-public.example.com
```

Console is built with `pnpm --filter @realtime-me/console build` and served by
`services/console`. Host deployment, Caddy, systemd, and environment examples are
documented in [`deploy/manager`](deploy/manager/README.md). Library and edge
operations remain in [`deploy/library`](deploy/library/README.md) and
[`deploy/edge`](deploy/edge/README.md).

## Mobile development

Private LAN/OpenVPN and public Status endpoints are Android build properties, never committed
addresses:

```sh
cd apps/mobile/android
./gradlew app:assembleDebug \
  -PstatusGatewayLanUrl=http://status.realtime.internal:18080 \
  -PstatusGatewayPublicUrl=https://api-status.example.com
```

For debug builds only, inject an ingest token without using the clipboard:

```sh
STATUS_INGEST_TOKEN=replace-with-generated-token \
  $ANDROID_HOME/platform-tools/adb -s <phone-serial> shell am broadcast \
  -a me.realtime.mobile.debug.SET_STATUS_GATEWAY_TOKEN \
  -n me.realtime/me.realtime.mobile.debug.DebugTokenReceiver \
  --es token "$STATUS_INGEST_TOKEN"
```
