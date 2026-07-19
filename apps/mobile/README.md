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

Or run the Flutter checks directly:

```sh
(cd apps/mobile && flutter pub get)
(cd apps/mobile && flutter analyze)
(cd apps/mobile && flutter build apk --debug)
```

The Pigeon source is `pigeons/status_bridge.dart`. Commit both generated bridge files
with every contract change.

## Status gateway configuration

Gateway addresses are Android build properties rather than committed private values:

```sh
cd apps/mobile/android
./gradlew app:assembleDebug \
  -PstatusGatewayLanUrl=http://<lan-host>:18080 \
  -PstatusGatewayPublicUrl=https://api-status.example.com
```

The ingest token is entered in the app and stored through Android Keystore-backed
encrypted preferences. OpenAI, Anthropic, and GitHub credentials are never stored in
the phone app.
