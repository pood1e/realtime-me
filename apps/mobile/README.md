# Realtime Me Flutter client

`apps/mobile` is the only phone application. It combines Status, Agent workspace,
remote terminal, pairing, runtime, device, and connection settings in one Material 3
Flutter UI.

## Ownership boundary

- Flutter owns every phone screen and navigation flow using Riverpod and `go_router`.
- Generated Dart protobuf packages are the application contracts; the UI does not
  define parallel JSON DTOs.
- Pigeon exposes the Android Status store through a typed Host API and Event Channel.
  Snapshot payloads cross the bridge as protobuf bytes.
- Native Android owns Wear OS Data Layer listeners, Android Keystore token storage,
  WorkManager recovery, and foreground/background Status sync. These continue when
  the Flutter engine is not running.
- Manager pairing uses the private CA and internal service origin created by `smctl pair create`, a
  device PKCS#12 identity, and a bearer token. The app rejects plaintext HTTP and has
  no certificate-verification bypass.

## Generate and verify

Run from the repository root:

```sh
make generate-mobile
make verify-mobile
```

Or run the Flutter checks directly. Flutter drives the tracked Android Gradle
wrapper; Gradle still compiles and packages the native host, plugins, and Kotlin
Status service even though every phone screen is Flutter:

```sh
(cd apps/mobile && flutter pub get)
(cd apps/mobile && flutter analyze)
(cd apps/mobile && flutter build apk --debug)
```

The Pigeon source is `pigeons/status_bridge.dart`. Commit both generated bridge files
with every contract change.

## Status gateway configuration

The app uses one canonical private origin, `http://status.realtime.internal:18080`.
Split DNS resolves it to the Status host's existing LAN address for local clients
and to OpenVPN `10.66.0.10` for remote clients. There is no public ingest fallback
and no `192.168.0.0/24` route is pushed through the VPN.

## Signed release

Copy the signing template, point it at the existing upload keystore, and keep both
files out of version control:

```sh
cd apps/mobile/android
cp key.properties.example key.properties
# Edit key.properties, then return to apps/mobile.
cd ..
flutter build appbundle --release --build-name 0.1.0 --build-number 1
cd ../..
./gradlew :apps:watch:bundleRelease
```

Release builds fail closed when `key.properties` or its keystore is missing. The
same file signs the phone and Wear OS bundles because both form factors use the
`me.realtime` package. Preserve and back up that signing identity; updates must
use the same key. Wear artifacts use a separate version-code range so Play can
accept both form factors under one listing.

The ingest token is entered in the app and stored through Android Keystore-backed
encrypted preferences. OpenAI, Anthropic, and GitHub credentials are never stored in
the phone app.
