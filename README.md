# realtime-me

Realtime Pixel Watch status publisher for GitHub.

The watch does not need internet access. It collects local health/device data and sends the latest snapshot to the paired Android phone through the Wear OS Data Layer. The phone companion stores a GitHub personal access token locally and updates the GitHub user status through the GitHub GraphQL API.

## Architecture

- `apps/watch`: headless Wear OS app for Pixel Watch.
  - Requests sensor/background permission once, then exits.
  - Runs a health foreground service for heart rate, daily steps, battery, charging, and wrist state.
  - Restarts collection after boot/package replacement when permissions are still granted.
  - Publishes watch snapshots to the paired phone over Data Layer, with a one-minute normal refresh loop.
- `apps/mobile`: Android phone companion.
  - Receives watch snapshots from Data Layer.
  - Opens GitHub's token request page; no token is typed into the app.
  - Stores the GitHub token with Android Keystore-backed AES/GCM encrypted shared preferences.
  - Reads the authenticated GitHub username/current status and calls `changeUserStatus` directly; no Cloudflare or other relay is used.
  - Keeps a foreground sync service active after token setup; WorkManager is only a 15-minute fallback because Android periodic work cannot run every minute.
- `libs/protocol` and `proto/realtime/me/v1/watch.proto`: shared protobuf contract for the Data Layer payload.

Wear OS Data Layer requires the phone and watch APKs to use the same package name and signing certificate. Both APKs therefore use application ID `me.realtime`, while keeping separate source namespaces for mobile and watch code.

## GitHub token

Use a classic personal access token with the smallest GitHub scope that can update user status through GraphQL:

- Scope: `user`.
- Repository permissions: none requested by this app.
- Used GraphQL operations: `viewer.login`, `viewer.status`, and `changeUserStatus`.

In the phone app:

1. Tap **Request minimal GitHub token**. The browser opens GitHub's classic token page prefilled with `user`.
2. Create the token in GitHub and copy it once when GitHub displays it.
3. Return to the phone app and tap **Save copied token**.
4. The app shows the connected GitHub account after the token is accepted.

The app does not use OAuth, a GitHub OAuth client ID, manual token text input, or any repository scope. Tokens are not stored on the watch, committed to this repository, or printed to logs.

## Sync frequency

- Watch → phone:
  - First snapshot is published immediately after the watch service starts.
  - In normal mode, heart-rate/step/wrist/battery changes can publish after a 2-second minimum interval.
  - A foreground refresh loop republishes the current watch state about once per minute, even if sensor values do not change.
  - Low-battery or off-wrist mode slows unchanged refreshes to about every 5 minutes; changed values still have a 30-second minimum interval.
- Phone → GitHub:
  - Incoming watch snapshots are processed immediately.
  - The phone foreground service runs a one-minute sync loop while a token is configured.
  - The app keeps a 15-minute WorkManager fallback for OS-managed background recovery.
  - GitHub writes are throttled to one per minute.

## Build

Requirements:

- Android SDK with API 37 installed.
- JDK 17. This repo pins Gradle to `/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home` in `gradle.properties`; change that path if needed.

Commands:

```sh
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :libs:protocol:generateDebugProto
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :apps:watch:assembleDebug :apps:mobile:assembleDebug
buf lint
```

## Debug token fallback

For debug builds, a copied token can be injected over ADB instead of using the phone clipboard:

```sh
GITHUB_TOKEN=github_pat_placeholder \
  $ANDROID_HOME/platform-tools/adb -s <phone-serial> shell am broadcast \
  -a me.realtime.mobile.debug.SET_GITHUB_TOKEN \
  -n me.realtime/me.realtime.mobile.debug.DebugTokenReceiver \
  --es token "$GITHUB_TOKEN"
```

The debug receiver exists only in debug builds and stores the token through the same encrypted store used by the app UI.

## Runtime setup

1. Install `apps/mobile` on the paired Android phone and configure the GitHub token.
2. Install `apps/watch` on the Pixel Watch.
3. Open the watch app once and grant the requested sensor/background permissions. The watch activity exits after starting the foreground collection service.
4. After reboot or app update, collection restarts automatically if permissions remain granted.

## Status format

The phone writes a short public GitHub status such as:

```text
❤️72 · 👣8.4k · 🔋64%
```

The emoji changes for charging, low battery, and off-wrist states. Off-wrist status does not include heart rate. Statuses expire after 20 minutes so a stale status clears naturally if the phone/watch stop syncing.

## Verification

```sh
buf lint
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :apps:watch:assembleDebug :apps:mobile:assembleDebug --no-daemon
ANDROID_HOME=$HOME/Library/Android/sdk ANDROID_SDK_ROOT=$HOME/Library/Android/sdk ./gradlew :apps:watch:lintDebug :apps:mobile:lintDebug --no-daemon
```
