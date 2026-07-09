# realtime-me

A personal realtime status system: Android phone and Wear OS watch apps, a Go
gateway, a React status page on Cloudflare Workers, and Python probes for
Linux/macOS hosts. Protobuf is the single contract across all four languages.

## Layout

| Path | Contents |
| --- | --- |
| `proto/realtime/me/v1` | Canonical `.proto` contracts. Single source of truth. |
| `apps/status-gateway` | Go ConnectRPC gateway; queries Prometheus, serves the API. |
| `apps/status-page` | React SPA + Cloudflare Worker (`src/worker.ts`) that proxies the API. |
| `apps/mobile` | Android phone app (`me.realtime.mobile`). |
| `apps/watch` | Wear OS app (`me.realtime.watch`). |
| `libs/protocol` | Kotlin/Android library sharing the protos and the Wear data-layer contract. |
| `infra/status-stack` | Docker Compose: Prometheus, node-exporter, cAdvisor, gateway, cloudflared. |
| `scripts/probe` | Exporter payload downloaded onto probe hosts, plus the shared `status_common`. |
| `scripts/operator` | Tools you run from a clone against the gateway. |
| `scripts/*.sh` | Host installers, curled directly by URL. |
| `docs/projects` | Generated per-repo docs. Gitignored — may contain private repo internals. |

## Commands

```sh
npm run proto:lint            # buf lint
npm run proto:gen             # regenerate Go + TypeScript from every proto
npm run check:status          # go vet + go test, then tsc -b --noEmit
npm run build:status          # go build, then vite build
./gradlew :apps:watch:assembleDebug :apps:mobile:assembleDebug
```

`proto:gen` needs `buf`, `protoc-gen-go`, and `protoc-gen-connect-go` on `PATH`;
`protoc-gen-es` resolves from `node_modules`. None of them are vendored.

## Invariants

**`proto/` is the only place a contract is defined.** `buf` generates the Go
(`apps/status-gateway/internal/genproto`) and TypeScript
(`apps/status-page/src/gen`) trees; both are committed, so run `npm run proto:gen`
and commit the output alongside any `.proto` change. Kotlin is generated at build
time by Gradle and is not committed. Never hand-edit generated code.

**Kotlin reads the protos through a symlink**, not a copy:
`libs/protocol/src/main/proto/realtime` → `proto/realtime`. The Gradle protobuf
plugin compiles them as javalite, so Kotlin protos are absent from `buf.gen.yaml`
by design. Don't replace the symlink with copied files.

**Probes are pull-based, and probe hosts are unaware of the gateway.** They run
`status-device-reporter.py --serve` and `agent-status-reporter.py --serve` as
read-only HTTP exporters. Prometheus finds them through the gateway's HTTP
service discovery (`/api/prometheus/http-sd/*`) and stamps the gateway-minted
device uid on every series via target labels. A probe host therefore holds no
gateway URL, no ingest token, and no identity of its own. Only the phone pushes
(`IngestService`), because it cannot be scraped. Do not add a push path to a
probe script — that architecture was removed deliberately.

**`scripts/` has a published URL contract, and the probe payload is flat at
runtime.** The installers fetch each file from
`https://raw.githubusercontent.com/pood1e/realtime-me/main/scripts` (jsdelivr
first, overridable via `REALTIME_ME_RAW_BASE_URL`) and drop the payload into one
flat `INSTALL_DIR`. Keep that base URL stable and pass the subpath as the
filename (`download_file probe/status_common.py`), so the mirror list and the
env-var override keep working. `status_common.py` must land beside the reporters
in `INSTALL_DIR`, because that sibling layout is the only reason
`import status_common` resolves there. Renaming or moving anything under
`scripts/` breaks installers already deployed on hosts.

**`scripts/operator/register-device.py` reaches into `probe/` for
`status_common`** via an explicit `sys.path` insert. This inverts the intended
layer direction and is a known, accepted wart: the module is shared, but it
cannot live in a package because the probe payload must ship flat. Don't
"fix" it by duplicating the module.

**Device identity is backend-owned.** The gateway mints every device uid via
`EnrollmentService/EnrollDevice`; clients cache it and never construct one.

## Gotchas

- `gradle.properties` pins `org.gradle.java.home` to a macOS Homebrew JDK 17
  path. On Linux, override it rather than committing a new path.
- `infra/status-stack/prometheus/file_sd/cadvisor.yml` is intentionally an empty
  `[]`. cAdvisor sits behind the `containers` Compose profile, so an empty
  file_sd list is the opt-in switch; `cadvisor.yml.example` shows the contents.
  A `static_configs` entry would leave a permanently-down target.
- The status page's Worker proxies `/realtime.me.v1.*` and `/api/*` to
  `STATUS_API_BASE_URL`, so the browser always sees one origin.
