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
`status-device-reporter.py` and `agent-status-reporter.py` as read-only HTTP
exporters serving `/healthz` and `/metrics`, and nothing else. Prometheus finds
them through the gateway's HTTP service discovery (`/api/prometheus/http-sd/*`)
and stamps the gateway-minted device uid on every series via target labels. A
probe host therefore holds no gateway URL, no ingest token, and no identity of
its own — the exporters neither read nor emit a device uid. Only the phone
pushes (`IngestService/ReportMobileStatus`), because it cannot be scraped.
Do not add a push path to a probe script, and do not give one an identity.

**`IngestService` has exactly two methods.** `ReportMobileStatus` and
`RegisterScrapeTargets`. Hosts, VMs, and coding agents are never pushed; the
gateway derives their `DeviceState` and `Agent` entirely from Prometheus queries
(`internal/gateway/prometheus.go`). The gateway's own `/metrics` therefore
exports only what it owns: the phone's pushed status and its GitHub sync state.
Re-exporting host or agent series here would duplicate what the exporters
already publish.

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
`EnrollmentService/EnrollDevice`; clients cache it and never construct one. A
`ScrapeTarget` therefore carries only a job and a `host:port` — every label
Prometheus stamps on it (`device_name`, `device_model`, `device_kind`,
`device_role`) is read from the enrollment, so a caller cannot assert an identity
the gateway did not mint. `RegisterScrapeTargets` declares a device's *complete*
target set: an empty set deregisters the device. When the gateway does not
recognise a uid it answers `not_found`, which is the signal a client uses to drop
its cached uid and enroll again.

**The agent exporter only reads what the agents already wrote, and never their
titles.** It issues no request to any agent and runs no agent command. Claude
Code is read from `~/.claude/sessions/<pid>.json` (which names the live pid, its
transcript, and its own `busy` flag) plus the per-sub-agent transcripts beside
it; Codex is read from the rollout `.jsonl` its process holds open and from
`state_5.sqlite` / `goals_1.sqlite` in `$CODEX_HOME`. Neither agent's file layout
is symmetric, so neither detection can be folded into the other: Codex holds its
rollout open and brackets each turn with `task_started` and `task_complete`
(Codex serialises those v1 names and reads `turn_*` only as aliases; nothing
between the brackets is worth reading), and a thread Codex spawned names its
parent in the `session_meta` record heading its own rollout — never the
`thread_spawn_edges` table, which keeps an edge `open` for a finished child that
is merely resumable and records none at all for an ephemeral one — while Claude Code appends
and closes, answers a background sub-agent the instant it launches, and splits
one reply across `thinking`/`text`/`tool_use` records. A Claude sub-agent is
therefore live until its session announces its `<task-id>`, and only a later
write means it was resumed. Prompts, objectives, task titles, sub-agent
descriptions and `threads.title` are never read into a metric — a `model` and a
*count*, of agents and of sub-agents per model, are the only things added to
`realtime_agent_*`. A host runs as many agents of one kind as it likes: three
codex sessions in three terminals are three agents, counted by
`realtime_agent_running_count` and expanded by the gateway into one `Agent` each,
whose uid it mints from the host, the kind, the model and the ordinal so that no
thread id ever reaches the page. A sub-agent may run a different model from the
agent that spawned it, so that count is labelled by model too and expands into
one `Subagent` per worker. Neither count carries an identity, so the budget and
the sub-agents — which the exporter can only report for a kind — are carried by
the first agent of that kind.

**Probe exporters cache `/metrics` and serve it from a fixed thread pool.** A
scrape shells out to `ps`, `lsof` and `bluetoothctl`, so `status_common.cached()`
keeps the scrape rate from becoming the process-spawn rate, and `PooledHTTPServer`
keeps a burst of connections from becoming a burst of threads. The cache TTL must
stay under Prometheus's scrape interval. Prometheus label values carry no
free text beyond the track title and artist — the artwork URL is gone, because a
per-play CDN URL is an unbounded series and rendering it made every visitor's
browser fetch from a third party.

**PromQL lives only in the gateway.** `internal/gateway/metrics.go` is the one
place a query expression is written. Clients name a `MetricSeries` and pass
domain selectors (device uid, agent kind, accessory); the gateway resolves the
metric name, labels, and job, escapes every selector, and bounds the window to
`maxMetricRangePoints`. There is no PromQL passthrough — the read token cannot
run an arbitrary query. Don't reintroduce a `query=` parameter.

## Gotchas

- `gradle.properties` pins `org.gradle.java.home` to a macOS Homebrew JDK 17
  path. On Linux, override it rather than committing a new path.
- `infra/status-stack/prometheus/file_sd/cadvisor.yml` is intentionally an empty
  `[]`. cAdvisor sits behind the `containers` Compose profile, so an empty
  file_sd list is the opt-in switch; `cadvisor.yml.example` shows the contents.
  A `static_configs` entry would leave a permanently-down target.
- The installers try jsdelivr before GitHub raw, and jsdelivr caches a branch ref
  for hours, so `@main` keeps serving the old file for the rest of the day you
  push a probe change. Reinstall against the commit instead, keeping raw as the
  fallback, because raw alone is slow enough on some networks that only the
  largest probe file times out:
  `REALTIME_ME_RAW_BASE_URLS="https://cdn.jsdelivr.net/gh/pood1e/realtime-me@<sha>/scripts https://raw.githubusercontent.com/pood1e/realtime-me/main/scripts"`.
- `GetPublicStatus` is unauthenticated, so its Prometheus fan-out sits behind a
  2-second single-flight cache in `internal/gateway/status.go`. Don't add a query
  to the assembly without checking it runs inside `parallel(...)`.
- Prometheus runs without `--web.enable-lifecycle`: nothing here reloads it, and
  the flag serves unauthenticated `/-/reload` and `/-/quit`. Config changes need a
  container restart.
- The status page's Worker proxies only `/realtime.me.v1.*` to
  `STATUS_API_BASE_URL`, so the browser always sees one origin. It deliberately
  proxies nothing under `/api/`: those are the gateway's control-plane routes,
  such as scrape discovery, and a browser must never reach them.
- `STATUS_INGEST_TOKEN` (write) and `STATUS_QUERY_TOKEN` (read) are separate
  secrets and the gateway refuses to start without both. The read token reaches
  the internal dashboard, `MetricsService`, and scrape discovery; Prometheus
  presents it from `infra/status-stack/prometheus/query_token`, which is
  gitignored and must exist before the first `docker compose up`.
