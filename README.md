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

The status stack stores raw time-series data in Prometheus on your own host. Cloudflare only needs to expose the public API/page through a Tunnel or Worker custom domain. Linux host and VM metrics are scraped by Prometheus through node-exporter and HTTP service discovery. Extra device signals, such as the currently playing media title and connected Bluetooth audio accessory battery on macOS/Linux, are scraped from `status-device-reporter.py --serve`.

```sh
cd infra/status-stack
cp .env.example .env
openssl rand -base64 32 # paste into STATUS_INGEST_TOKEN
```

Set these values in `.env`:

- `STATUS_GATEWAY_BIND`: LAN address that should accept phone updates, or `127.0.0.1` when only Cloudflare Tunnel should reach it.
- `GITHUB_TOKEN`: GitHub token with the `user` scope.
- `GITHUB_STATUS_MIN_INTERVAL_SECONDS`: default `10`.
- `GITHUB_STATUS_TTL_MINUTES`: default `20`.

```sh
docker compose up -d --build
```

The public page consumes:

```text
GET /api/public-status
```

The phone app publishes:

```text
POST /api/ingest/mobile
Authorization: Bearer <STATUS_INGEST_TOKEN>
```

Linux probes only install local exporters. Prometheus discovers and pulls them from the status gateway's HTTP service discovery endpoint, so the probe host does not need the gateway URL or ingest token.

```sh
curl -fsSL https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts/install-linux-probe.sh \
  | sudo env STATUS_EXPORTER_HOST=<device-lan-ip> bash
```

Register scrape targets centrally on the gateway side. Use `INSTALL_AGENT=1` when the device should also expose Codex/Claude active-agent state. Use `STATUS_DEVICE_ROLE=vm STATUS_DEVICE_KIND=virtual_machine` for VMs. The Linux installer does not hardcode LAN addresses; pass `STATUS_EXPORTER_HOST` when automatic route detection is not suitable. Media title collection on Linux uses `playerctl` when available, and Bluetooth audio accessory discovery uses BlueZ `bluetoothctl`. On macOS, run `scripts/status-device-reporter.py --serve` under the logged-in user so media session and Bluetooth accessory metadata are visible.


## Public status page

`apps/status-page/wrangler.jsonc` deploys the page directly to the Worker custom domain `me.pood1e.space`; no Pages project or `pages.dev` deployment is required.

Build the static assets, then deploy with the public gateway URL as a runtime variable:

```sh
npm run build --workspace apps/status-page
npx wrangler deploy --config apps/status-page/wrangler.jsonc \
  --var STATUS_API_BASE_URL:https://api-status.example.com
```

The Worker serves the SPA and proxies only `GET /api/public-status` to `STATUS_API_BASE_URL`, so the browser reads status from the same origin:

```text
https://me.pood1e.space/api/public-status
```

## Runtime setup

1. Deploy `infra/status-stack` and configure `STATUS_INGEST_TOKEN` plus `GITHUB_TOKEN` on the host.
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
