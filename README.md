# realtime-me

Personal realtime status, content library, and remote Agent/terminal management in one monorepo.

The watch does not need internet access. It collects local health/device data and sends the latest snapshot to the paired Android phone through the Wear OS Data Layer. The phone only forwards phone/watch status to a self-hosted gateway; the gateway owns GitHub status updates and Prometheus metrics.

## Architecture

- `apps/watch`: headless Wear OS app for Pixel Watch.
  - Requests sensor/background permission once, then exits.
  - Runs a health foreground service for heart rate, daily steps, battery, charging, and wrist state.
  - Restarts collection after boot/package replacement when permissions are still granted.
  - Publishes watch snapshots to the paired phone over Data Layer.
- `apps/mobile`: the single Flutter phone application (`me.realtime`).
  - Provides Status, Agent, terminal, pairing, and settings surfaces with Material 3, Riverpod, and go_router.
  - Receives watch snapshots from Data Layer.
  - Stores the Status ingest token with Android Keystore-backed AES/GCM preferences and keeps the
    separately scoped Manager PKCS#12/bearer credentials in Flutter secure storage; Android backup
    is disabled for both.
  - Pushes phone/watch status to configured gateway endpoints; private LAN/public URLs are build properties, not committed values.
  - Keeps a foreground sync service active after token setup; WorkManager is only a 15-minute OS-managed recovery fallback.
  - Uses generated Pigeon Host APIs and an Event Channel to read the native snapshot store; Flutter is not required for background sync to stay alive.
- `services/status`: Go gateway for mobile ingestion, Prometheus HTTP service discovery, GitHub `changeUserStatus`, public status JSON, and internal metrics charts.
- `apps/web/status`: Vite/React status page using shadcn/ui, Radix, Lucide, Tailwind, and a Cloudflare Worker custom domain.
- `services/library` and `apps/web/library`: local-first content API/worker plus seven independent web applications.
- `services/manager`: Fastify/ConnectRPC control plane for subscription-backed Codex/Claude agents and persistent terminals.
- `packages/web-ui`: the single shared React primitive and theme-token layer for Status and Library.
- `deploy/edge`: the only `cloudflared` connector and owner of the shared `realtime-me-edge` network.
- `deploy/status`, `deploy/library`, `deploy/manager`, and `deploy/web`: independent release units for each runtime boundary.
- `packages/status-protocol-android` and `proto/realtime/me/status/v1/watch.proto`: shared protobuf contract for the Data Layer payload.

Wear OS Data Layer requires the phone and watch APKs to use the same package name and signing certificate. Both APKs therefore use application ID `me.realtime`, while keeping separate source namespaces for mobile and watch code.

Detailed boundaries and operations:

- [Consolidation architecture and migration plan](docs/architecture/project-consolidation.md)
- [Library architecture](docs/library/architecture.md) and [production operations](deploy/library/README.md)
- [Manager architecture](docs/manager/architecture.md) and [Linux deployment](deploy/manager/README.md)
- [One-time consolidated cutover](docs/operations/consolidation-cutover.md)

## GitHub status

GitHub is updated by `services/status`, not by the phone. Put a classic personal access token in `github.status_token` in the gateway's `gateway.yaml`:

- Scope: `user`.
- Repository permissions: none.
- GraphQL operation: `changeUserStatus`.

No GitHub token is stored on the watch or phone.

## Metrics

`/metrics` is Prometheus-compatible and uses OpenTelemetry-style metric names/units in the metadata. Examples:

- `realtime.device.battery.level` → `realtime_device_battery_level_ratio`, unit `1`.
- `realtime.device.last_update` → `realtime_device_last_update_time_seconds`, unit `s`.
- `realtime.watch.heart_rate` → `realtime_watch_heart_rate_beats_per_minute`, unit `{beat}/min`.
- `realtime.github.status.sync.state` → `realtime_github_status_sync_state`, unit `1`.

Attributes are exposed as Prometheus labels such as `device_id`, `device_type`, `state`, and `wrist_state`.

## Sync frequency

- Watch → phone:
  - First snapshot is published immediately after the watch service starts.
  - In normal mode, heart-rate/step/wrist/battery changes can publish after a 2-second minimum interval.
  - A foreground refresh loop republishes the current watch state about once per minute.
  - Low-battery or off-wrist mode slows unchanged refreshes to about every 5 minutes; changed values still have a 30-second minimum interval.
- Phone → status gateway:
  - Incoming watch snapshots are processed immediately.
  - The phone foreground service pushes on watch, connectivity, and charging changes, with a 5-minute heartbeat while a gateway token is configured.
  - LAN HTTP and public HTTPS endpoints are build-time configuration. Use LAN first to avoid Cloudflare traffic, and public HTTPS as fallback when configured.
- Gateway → GitHub:
  - The gateway formats the latest watch data and updates GitHub status at most once every 10 seconds by default.
  - Statuses expire after 20 minutes by default so stale status clears naturally.

## Build

Requirements:

- Android SDK with API 37 installed.
- JDK 17. This repo pins Gradle to `/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home` in `gradle.properties`; change that path if needed.
- Flutter 3.44.6 / Dart 3.12.2.
- Go 1.26.4.
- Node 24.18+ and pnpm 11.10.

Commands:

```sh
make generate
pnpm check
./gradlew :apps:watch:assembleDebug
(cd apps/mobile && flutter analyze && flutter build apk --debug)
```

To enable the phone app's LAN status gateway without committing a private address:

```sh
cd apps/mobile/android
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk \
  ./gradlew app:assembleDebug -PstatusGatewayLanUrl=http://<lan-host>:18080
```

For a public HTTPS fallback:

```sh
cd apps/mobile/android
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk \
  ./gradlew app:assembleDebug \
  -PstatusGatewayPublicUrl=https://api-status.example.com
```

## Debug gateway token injection

For debug builds, a copied gateway token can be injected over ADB instead of using the phone clipboard:

```sh
STATUS_INGEST_TOKEN=replace-with-generated-token \
  $ANDROID_HOME/platform-tools/adb -s <phone-serial> shell am broadcast \
  -a me.realtime.mobile.debug.SET_STATUS_GATEWAY_TOKEN \
  -n me.realtime/me.realtime.mobile.debug.DebugTokenReceiver \
  --es token "$STATUS_INGEST_TOKEN"
```

The debug receiver exists only in debug builds and stores the token through the same encrypted store used by the app UI.

## Self-hosted status stack

The status stack stores raw time-series data in Prometheus on your own host. Cloudflare only needs to expose the public API/page through a Tunnel or Worker custom domain. Linux host and VM metrics are scraped by Prometheus through node-exporter and HTTP service discovery. Extra device signals, such as the currently playing media title and connected Bluetooth audio accessory battery on macOS/Linux, are scraped from `status-device-reporter.py`.

Every setting the gateway reads lives in one file, `deploy/status/gateway.yaml`:
the two bearer tokens, the two kinds of GitHub credential, and the owner's profile.
An unknown key in it is a startup error, never a setting that quietly does nothing.

```sh
cd deploy/status
cp gateway.example.yaml gateway.yaml   # tokens, GitHub credentials, profile
cp projects.example.json projects.json # which repositories the page may show
cp .env.example .env                   # only what Compose itself interpolates

openssl rand -base64 32 # paste into tokens.ingest  (write)
openssl rand -base64 32 # paste into tokens.query   (read)
printf %s "<tokens.query>" > prometheus/query_token  # Prometheus presents the same one

# The gateway runs as uid 10001, not as you, so it cannot open a 0600 file you own.
# Put the read bit on the files and take it off the directory: the daemon mounts them
# as root, so the directory's mode never reaches the container — it only keeps other
# accounts on the host away from five secrets.
chmod 700 . && chmod 644 gateway.yaml projects.json
```

`gateway.example.yaml` explains each setting where it stands. The two that catch
people out:

- `tokens.query` **must differ** from `tokens.ingest`. It is pasted into a browser and
  kept in localStorage; if they were the same value, every visitor with the dashboard
  open could enroll a device. The gateway refuses to start when they match.
- `github.projects_tokens` is **plural and read-only**, and is not
  `github.status_token`. See [Profile and projects](#profile-and-projects) below.

`.env` now holds only what Docker Compose has to interpolate for itself —
`STATUS_GATEWAY_BIND` and `PROMETHEUS_RETENTION` — because
Compose cannot read the YAML. Everything the *gateway* reads is in the YAML, and the
container's own paths and ports are set in `compose.yaml`, where you never have to
think about them.

Start [`deploy/edge`](deploy/edge/README.md) once before the Status or Library stack. The shared
connector routes Status through `status-api:8080` and both Library API hostnames through
`library-api:8080`; neither application unit stores the Tunnel token.

```sh
docker compose up -d --build --remove-orphans
```

The gateway speaks ConnectRPC (`POST /realtime.me.status.v1.<Service>/<Method>` or
`POST /realtime.me.site.v1.<Service>/<Method>`, JSON or binary protobuf). The main procedures:

```text
StatusService/GetPublicStatus       # public, unauthenticated — what the page reads
StatusService/GetInternalStatus     # Bearer <STATUS_QUERY_TOKEN>
MetricsService/GetMetricRange       # Bearer <STATUS_QUERY_TOKEN> — chart time series
ProfileService/GetProfile           # public — the owner's name, avatar, contact links
ProjectsService/ListProjects        # public — the /projects page
EnrollmentService/EnrollDevice      # Bearer <STATUS_INGEST_TOKEN> — mints the device uid
IngestService/ReportMobileStatus    # Bearer <STATUS_INGEST_TOKEN> — phone push
IngestService/RegisterScrapeTargets # Bearer <STATUS_INGEST_TOKEN> — central device registration
```

Probe hosts run only Prometheus exporters and stay unaware of the gateway. Prometheus discovers and pulls them through the gateway's HTTP service discovery, so the probe host needs no gateway URL or ingest token. Install the exporters on the host:

```sh
# Linux (systemd)
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-linux-probe.sh | sudo bash

# macOS (run as your login user, not sudo — LaunchAgents and media access are per-user)
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-macos-probe.sh | bash
```

Then register the host once, from anywhere that can reach the gateway (the installer prints this line with the host and ports filled in):

```sh
STATUS_INGEST_TOKEN=... python3 scripts/operator/register-device.py \
  --url http://<gateway-host>:18080 --host <device-lan-ip> --name "<name>" --kind host
```

The gateway mints the device's uid and serves it to Prometheus as a service-discovery target label, so the exporters carry no identity of their own. Use `INSTALL_AGENT=1` (installer) plus `--install-agent` (register) when the device should also expose Codex/Claude active-agent state, and `--kind virtual_machine --role vm` (register) for VMs. Media title collection on Linux uses `playerctl` when available, Bluetooth audio accessory discovery uses BlueZ `bluetoothctl`, and on macOS both come from the logged-in user's session. Exporters bind `0.0.0.0` so the gateway can scrape them across the LAN; set `STATUS_EXPORTER_BIND`/`STATUS_EXPORTER_HOST` to override. Only the phone pushes, since it cannot be scraped.


## Public status page

`apps/web/status/wrangler.jsonc` deploys the page directly to the Worker custom domain `me.pood1e.space`; no Pages project or `pages.dev` deployment is required.

Build the static assets, then deploy with the public gateway URL as a runtime variable:

```sh
pnpm --filter @realtime-me/status-web build
pnpm --dir apps/web/status deploy -- \
  --var STATUS_API_BASE_URL:https://api-status.example.com
```

The Worker serves the SPA and proxies an explicit allowlist of public/read-only
ConnectRPC calls to `STATUS_API_BASE_URL`, so the browser reads status and profile
from the same origin without exposing ingest or enrollment routes:

```text
https://me.pood1e.space/realtime.me.status.v1.StatusService/GetPublicStatus
https://me.pood1e.space/realtime.me.site.v1.ProjectsService/ListProjects
```

## Profile and projects

The owner's identity and the owner's work are two documents, served by two services,
because they answer to different pages. `ProfileService/GetProfile` is the name,
avatar, and contact links the topbar carries on *every* screen.
`ProjectsService/ListProjects` is the `/projects` page, and nothing else.

The profile is a *setting*, so it sits in `gateway.yaml` with the rest of them. It
holds the login, and the ways to reach the owner that GitHub cannot supply. The name,
the avatar, and the GitHub link are *not* written down: all three are read from the
login, and `github.com/<login>.png` always resolves to the current avatar, so the page
follows a change on GitHub without anyone editing a file.

```yaml
profile:
  github_login: your-login
  links:
    - platform: telegram
      label: Telegram
      uri: https://t.me/your-handle
    - platform: email
      label: Email
      uri: mailto:me@example.com
```

`projects.json` is *data*, not settings, so it stays JSON and stays in its own file:
long prose reads badly in YAML, and these summaries are prose. It *curates*, and does
not describe — it names the repositories the page may show, as `owner/name`, because
the projects reach across organizations — and carries the one field GitHub cannot give
back:

```json
{
  "projects": [
    { "repo": "pood1e/realtime-me", "summary": "Optional. Stands in for GitHub's own description." },
    { "repo": "some-org/some-other-repo" }
  ]
}
```

Everything else a card draws — the description, languages, stars, topics, the
archived flag, the created month, the commit sparkline — the gateway reads from
GitHub once a day (`github.projects_refresh_hours`, default 24) and serves from
memory. A snapshot of those fields ages the moment it is written; a live one does
not. The page cannot fetch on demand: a refresh costs three calls per project, and
against GitHub's 5,000-request hourly budget a per-visitor fetch would be spent
inside a couple of dozen page loads — and would make every visitor wait on all of it.

Curation is explicit on purpose. There is no listing call: the gateway reads only the
repositories `projects.json` names, so a private repository the owner creates *from
now on* cannot walk onto a public page by itself. A private project that *is* curated
appears with a badge and no link — the response withholds `repository_url` for it,
always.

This needs `github.projects_tokens`: **read-only** tokens, separate from the
`github.status_token` that *writes* the owner's GitHub status. It is a list because a
fine-grained token reaches the repositories of a single user or organization and no
further, while the curated projects span several owners — so create one per owner at
<https://github.com/settings/personal-access-tokens/new>, each with *Repository
access: All repositories* and *Metadata: read-only*, and nothing else. An organization
may hold its token pending until an owner approves it, under *Settings → Personal
access tokens → Pending requests*.

Read-only is the reason for the shape. A classic token has no read-only grade for
private repositories: the only scope that reaches them is `repo`, which is read *and
write* over every repository you have, on a host that needs to write to none of them.

If either file is configured but unreadable, the gateway logs the path, keeps serving
status and ingest, and answers that one service with `unavailable`. The page then says
it cannot load, rather than rendering an empty life as though nobody had written one.

## Runtime setup

1. Deploy `deploy/status`. Fill in `gateway.yaml` (tokens, GitHub credentials, profile) and `projects.json` (the curated repositories), each copied from the `.example` beside it.
2. Build/install `apps/mobile` with gateway endpoint build properties and save the gateway token once.
3. Install `apps/watch` on the Pixel Watch.
4. Open the watch app once and grant the requested sensor/background permissions. The watch activity exits after starting the foreground collection service.
5. After reboot or app update, collection restarts automatically if permissions remain granted.

## Status format

The gateway writes a short public GitHub status such as:

```text
❤️72 · 👣8.4k · 🔋64%
```

The emoji changes for charging, low battery, and off-wrist states. Off-wrist status does not include heart rate.

## Verification

```sh
make verify
```
