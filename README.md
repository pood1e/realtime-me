# realtime-me

Realtime Pixel Watch and device status publisher.

The watch does not need internet access. It collects local health/device data and sends the latest snapshot to the paired Android phone through the Wear OS Data Layer. The phone only forwards phone/watch status to a self-hosted gateway; the gateway owns GitHub status updates and Prometheus metrics.

## Architecture

- `apps/watch`: headless Wear OS app for Pixel Watch.
  - Requests sensor/background permission once, then exits.
  - Runs a health foreground service for heart rate, daily steps, battery, charging, and wrist state.
  - Restarts collection after boot/package replacement when permissions are still granted.
  - Publishes watch snapshots to the paired phone over Data Layer.
- `apps/mobile`: Android phone companion.
  - Receives watch snapshots from Data Layer.
  - Stores only the self-hosted gateway ingest token with Android Keystore-backed AES/GCM encrypted shared preferences.
  - Pushes phone/watch status to configured gateway endpoints; private LAN/public URLs are build properties, not committed values.
  - Keeps a foreground sync service active after token setup; WorkManager is only a 15-minute OS-managed recovery fallback.
- `apps/status-gateway`: Go gateway for mobile ingestion, Prometheus HTTP service discovery, GitHub `changeUserStatus`, public status JSON, and internal metrics charts.
- `apps/status-page`: Vite/React status page using shadcn/ui, Radix, Lucide, Tailwind, and a Cloudflare Worker custom domain.
- `infra/status-stack`: Prometheus, node-exporter, cAdvisor, status-gateway, and optional cloudflared service definitions.
- `libs/protocol` and `proto/realtime/me/v1/watch.proto`: shared protobuf contract for the Data Layer payload.

Wear OS Data Layer requires the phone and watch APKs to use the same package name and signing certificate. Both APKs therefore use application ID `me.realtime`, while keeping separate source namespaces for mobile and watch code.

## GitHub status

GitHub is updated by `apps/status-gateway`, not by the phone. Put a classic personal access token in the gateway host `.env`:

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
  - The phone foreground service pushes the latest phone/watch status every 10 seconds while a gateway token is configured.
  - LAN HTTP and public HTTPS endpoints are build-time configuration. Use LAN first to avoid Cloudflare traffic, and public HTTPS as fallback when configured.
- Gateway → GitHub:
  - The gateway formats the latest watch data and updates GitHub status at most once every 10 seconds by default.
  - Statuses expire after 20 minutes by default so stale status clears naturally.

## Build

Requirements:

- Android SDK with API 37 installed.
- JDK 17. This repo pins Gradle to `/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home` in `gradle.properties`; change that path if needed.
- Go 1.26.
- Node 22+.

Commands:

```sh
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :libs:protocol:generateDebugProto
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :apps:watch:assembleDebug :apps:mobile:assembleDebug
npm run build:status
buf lint
```

To enable the phone app's LAN status gateway without committing a private address:

```sh
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk \
  ./gradlew :apps:mobile:assembleDebug \
  -PstatusGatewayLanUrl=http://<lan-host>:18080 \
  -PstatusGatewayAllowCleartext=true
```

For a public HTTPS fallback:

```sh
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk \
  ./gradlew :apps:mobile:assembleDebug \
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

```sh
cd infra/status-stack
cp .env.example .env
openssl rand -base64 32 # paste into STATUS_INGEST_TOKEN  (write)
openssl rand -base64 32 # paste into STATUS_QUERY_TOKEN   (read)
printf %s "<STATUS_QUERY_TOKEN>" > prometheus/query_token
```

Set these values in `.env`:

- `STATUS_INGEST_TOKEN`: write access — enrollment, phone push, scrape-target registration.
- `STATUS_QUERY_TOKEN`: read access — the internal dashboard, charts, and scrape discovery. Must differ from the ingest token: it is pasted into a browser. Prometheus reads the same value from `prometheus/query_token`.
- `STATUS_GATEWAY_BIND`: LAN address that should accept phone updates, or `127.0.0.1` when only Cloudflare Tunnel should reach it.
- `GITHUB_TOKEN`: GitHub token with the `user` scope.
- `GITHUB_STATUS_MIN_INTERVAL_SECONDS`: default `10`.
- `GITHUB_STATUS_TTL_MINUTES`: default `20`.

```sh
docker compose up -d --build
```

The gateway speaks ConnectRPC (`POST /realtime.me.v1.<Service>/<Method>`, JSON or binary protobuf). The main procedures:

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

`apps/status-page/wrangler.jsonc` deploys the page directly to the Worker custom domain `me.pood1e.space`; no Pages project or `pages.dev` deployment is required.

Build the static assets, then deploy with the public gateway URL as a runtime variable:

```sh
npm run build --workspace apps/status-page
npx wrangler deploy --config apps/status-page/wrangler.jsonc \
  --var STATUS_API_BASE_URL:https://api-status.example.com
```

The Worker serves the SPA and proxies the ConnectRPC calls (`/realtime.me.v1.*`) and `/api/*` to `STATUS_API_BASE_URL`, so the browser reads status and profile from the same origin:

```text
https://me.pood1e.space/realtime.me.v1.StatusService/GetPublicStatus
https://me.pood1e.space/realtime.me.v1.ProjectsService/ListProjects
```

## Profile and projects

The owner's identity and the owner's work are two documents, served by two services,
because they answer to different pages. `ProfileService/GetProfile` is the name,
avatar, and contact links the topbar carries on *every* screen.
`ProjectsService/ListProjects` is the `/projects` page, and nothing else.

Each is backed by one small hand-written file, mounted into the gateway from
`infra/status-stack/`. Both are gitignored — the profile carries a real email, the
curated list names private repositories — so copy each from the `.example` beside it
before the first `docker compose up`. The paths are named in `compose.yaml`, not in
`.env`: `.env` is rewritten whenever a token rotates, and a config line that quietly
leaves with it is not a failure anyone notices.

`profile.json` is the whole of the owner's identity:

```json
{
  "profile": {
    "display_name": "pood1e",
    "avatar_url": "https://github.com/pood1e.png",
    "github_login": "pood1e",
    "links": [
      { "label": "GitHub", "uri": "https://github.com/pood1e", "platform": "github" },
      { "label": "Email", "uri": "mailto:me@example.com", "platform": "email" }
    ]
  }
}
```

`projects.json` *curates*, and does not describe. It names the repositories the page
may show — as `owner/name`, because the token reaches organizations too — and carries
the one field GitHub cannot give back:

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
GitHub on a timer (`GITHUB_PROJECTS_REFRESH_MINUTES`, default 30) and serves from
memory. A snapshot of those fields ages the moment it is written; a live one does
not. The page cannot fetch on demand: one refresh costs a call for the repository
list plus two per project, and against GitHub's 5,000-request hourly budget a
per-visitor fetch would be spent inside seventy page loads.

Curation is explicit on purpose. Publishing whatever the token can see would put
every private repository the owner creates *from now on* onto a public page. A
private project that *is* curated appears with a badge and no link — the response
withholds `repository_url` for it, always.

This needs `GITHUB_PROJECTS_TOKEN`: a **second, read-only** token, separate from the
`GITHUB_TOKEN` that writes the owner's GitHub status. A fine-grained token with
*Repository access: All repositories* and *Metadata: read-only* is enough for every
call the gateway makes with it — it reads no file in any repository, and the write
token must not be widened to cover this.

If either file is configured but unreadable, the gateway logs the path, keeps serving
status and ingest, and answers that one service with `unavailable`. The page then says
it cannot load, rather than rendering an empty life as though nobody had written one.

## Runtime setup

1. Deploy `infra/status-stack`. Configure `STATUS_INGEST_TOKEN`, `STATUS_QUERY_TOKEN`, `GITHUB_TOKEN`, and `GITHUB_PROJECTS_TOKEN` in `.env`, and copy `profile.example.json` and `projects.example.json` to `profile.json` and `projects.json` beside them.
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
buf lint
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :apps:watch:assembleDebug :apps:mobile:assembleDebug --no-daemon
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :apps:watch:lintDebug :apps:mobile:lintDebug --no-daemon
npm run check:status
npm run build:status
```
