# realtime-me

A personal Status, Library, and coding-agent Manager monorepo. It contains a
Flutter phone app, Kotlin Wear app, Go/TypeScript services, two React web apps,
and one Python probe for Linux, macOS, and Windows. Protobuf is the canonical
cross-language contract.

## Layout

| Path | Contents |
| --- | --- |
| `proto/realtime/me/*` | Canonical versioned contracts for auth, status, site, library, and manager. |
| `services/status` | Go ConnectRPC gateway; queries Prometheus, serves the API. |
| `services/library` | Go Library API, worker, and migrate commands. |
| `services/manager` | TypeScript/Fastify Manager service and CLI. |
| `services/console` | Go OIDC BFF, authenticated proxy, and static Console host. |
| `apps/web/site` | Anonymous React SPA + explicit public allowlist Worker. |
| `apps/web/console` | Unified owner Status/Library/Manager SPA. |
| `apps/mobile` | Flutter phone app with the Android platform bridge (`me.realtime.mobile`). |
| `apps/watch` | Wear OS app (`me.realtime.watch`). |
| `packages/status-protocol-android` | Kotlin/Android library sharing the protos and the Wear data-layer contract. |
| `deploy/status` | Docker Compose: Prometheus, node-exporter, cAdvisor, and gateway. |
| `deploy/edge` | The only cloudflared connector and shared edge network. |
| `scripts/probe` | Cross-platform probe package and its pinned runtime manifest. |
| `scripts/operator` | Tools you run from a clone against the gateway. |
| `scripts/install-probe.py` | Cross-platform host installer, executed directly from its published URL. |
| `docs/projects` | Generated per-repo docs. Gitignored — may contain private repo internals. |

## Commands

```sh
pnpm check:proto                         # buf lint
pnpm generate                            # regenerate Go, TypeScript, and Dart contracts
pnpm check                               # all service/web static and build checks
pnpm build                               # Status, Library, Manager, Site, and Console
./gradlew :apps:watch:assembleDebug      # Wear OS application
(cd apps/mobile && flutter build apk --debug) # Flutter phone application
```

Generation needs `buf`, `protoc-gen-go`, and `protoc-gen-connect-go` on `PATH`;
the repository wrappers resolve the TypeScript and Dart plugins from the
workspace dependencies. None of the Go generators are vendored.

## Invariants

**`proto/` is the only place a contract is defined.** `buf` generates committed
Go under `gen/go`, TypeScript under `packages/*-contracts-web/src/gen`, and Dart
under `packages/*-contracts-dart/lib/gen`; run `pnpm generate` and commit the
output alongside every `.proto` change. Kotlin is generated at build time by
Gradle and is not committed. Never hand-edit generated code.

**Kotlin reads the status protos through a symlink**, not a copy:
`packages/status-protocol-android/src/main/proto/realtime/me/status` →
`proto/realtime/me/status`. The Gradle protobuf plugin compiles them as javalite,
so Kotlin output is absent from the Buf generation templates by design. Don't
replace the symlink with copied files.

**Probes are pull-based, and probe hosts are unaware of the gateway.** Linux,
macOS, and Windows all run `python -m realtime_probe`, a read-only HTTP endpoint
serving `/healthz` and `/metrics`, and nothing else. Prometheus finds it through
the gateway's single `/api/prometheus/http-sd/probe-agent` endpoint and stamps
the gateway-minted device uid on every series via target labels. A probe host
therefore holds no gateway URL, no ingest token, and no identity of its own — the
probe neither reads nor emits a device uid. Only the phone pushes
(`IngestService/ReportMobileStatus`), because it cannot be scraped. Do not add a
push path to the probe, and do not give it an identity.

**`IngestService` has exactly two methods.** `ReportMobileStatus` and
`RegisterScrapeTargets`. Hosts, VMs, and coding agents are never pushed; the
gateway derives their `DeviceState` and `Agent` entirely from Prometheus queries
(`internal/gateway/prometheus.go`). The gateway's own `/metrics` therefore
exports only what it owns: the phone's pushed status and its GitHub sync state.
Re-exporting host or agent series here would duplicate what the probes
already publish.

**Who the owner is and what the owner built are two documents.** `ProfileService`
serves the name, avatar, and contact links the topbar carries on *every* page;
`ProjectsService` serves `/projects`, and nothing else. Neither is a "page" — the
contract does not model screens.

**Status settings live in one YAML; data lives in JSON beside it.** `gateway.yaml` is
the two workload tokens, both kinds of GitHub credential, and the profile; an unknown
key is a startup error. Human authentication is OIDC wiring in Compose rather than a
hand-written gateway setting. `projects.json` is data: the curated repositories and their summaries,
which are prose and read badly in YAML. What the container *is* — port, state paths,
the address of Prometheus — stays in `compose.yaml`, because Compose decides it and
nobody should keep it in step by hand. Nothing the gateway reads goes in `.env`: `.env`
is rewritten whenever a token rotates, and the profile once vanished for four days
because its line left with one. Both hand-written files are gitignored (five secrets;
the names of private repos), so an `.example` sits beside each.

**The login is the only identity written down.** `gateway.yaml` carries
`profile.github_login` and the links GitHub cannot supply — Telegram, Discord, an email. The
display name, the avatar, and the GitHub link are *derived* from the login, never
configured: `github.com/<login>.png` already resolves to the current avatar, and
GitHub's own `name` field is the login here, so asking GitHub for any of it would
fetch a string the login already spells and hand the topbar a way to go nameless
whenever GitHub is down. Writing the login into four fields is four chances to change
three of them.

**`projects.json` curates; it does not describe.** It names the repositories the page
may show — as `owner/name`, never a bare name: the projects span four owners, and a
bare name identifies nothing. It carries `summary`, the one field GitHub cannot give
back. Everything else on a card — description, languages, stars, topics, archived,
created, the commit sparkline — the gateway reads from GitHub once a day and serves
from memory. Do not fetch on demand: a refresh is three calls per project, and against
5,000 requests an hour a per-visitor fetch is spent inside a couple of dozen page
loads. There is deliberately **no listing call**: the gateway reads only what
`projects.json` names, which is what keeps a private repository created *from now on*
from walking onto a public page by itself.

**`github.projects_tokens` is plural, read-only, and not `github.status_token`.** A
fine-grained token reaches a single user or organization and no further, so the
curated projects need one per owner — each Metadata: read-only, none able to write a
byte. Do not collapse them into one classic token: classic has no read-only grade for
private repositories, only `repo`, which is read *and write* over every repository the
owner has, on a host that needs to write to none of them. `github.status_token` is a
different secret entirely — it *writes* the owner's GitHub status and needs the `user`
scope. Never widen it to read repositories.

**A missing config file is a fault, not an empty document.** `loadJSONConfig` returns
an error when a configured path cannot be read, and the service it feeds answers
`unavailable` rather than serving an empty one. It does not exit: this process also
carries phone ingest and Prometheus scrape discovery, and a cosmetic file must not
take the metrics pipeline down. Swallowing the missing file is what let a lost
profile sit for days behind a healthy 200 — and why the topbar no longer hardcodes a
name and avatar to fall back on.

**`scripts/` has one published installer and one integrity-governed runtime.**
`install-probe.py` requires a 40-character `REALTIME_PROBE_RELEASE`, verifies the
embedded digest of `probe/integrity.json`, verifies every runtime file, and lets
pip accept only hash-pinned wheels. `scripts/probe/generate-integrity.py` owns the
manifest and installer digest. A custom immutable HTTPS mirror may be selected
only through `REALTIME_PROBE_SOURCE_URLS`. Installation is atomic and the service
identity cannot modify its runtime. Operator-side ConnectRPC code belongs in
`scripts/operator/status_client.py` and must not be imported by the
credential-free probe.

**Device identity is backend-owned.** The gateway mints every device uid via
`EnrollmentService/EnrollDevice`; clients cache it and never construct one. A
`ScrapeTarget` therefore carries only a job and a `host:port` — every label
Prometheus stamps on it (`device_name`, `device_model`, `device_kind`,
`device_role`) is read from the enrollment, so a caller cannot assert an identity
the gateway did not mint. `RegisterScrapeTargets` declares a device's *complete*
target set: an empty set deregisters the device. When the gateway does not
recognise a uid it answers `not_found`, which is the signal a client uses to drop
its cached uid and enroll again.

**The agent collector only reads what the agents already wrote, and never their
titles.** It issues no request to any agent and runs no agent command. Claude
Code is read from `~/.claude/sessions/<pid>.json` (which names the live pid, the
tick the kernel started it on, its transcript, and its own `busy` flag) plus the
per-sub-agent transcripts beside it; a pid is only on loan, and on a host that
runs Claude Code the program that inherits one is often another `claude`, so a
session counts only while `/proc/<pid>/stat` still reports the start tick it
recorded — where a kernel does not publish that, the pid has to stand alone; Codex is read from the rollout `.jsonl` its process holds open and from
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
write means it was resumed. A Codex thread is alive for as long as its rollout is
open, but Codex holds a spawned thread's rollout open well after that thread's
last turn ended, so being alive is not being at work: an agent and a sub-agent
alike count only while the newest bracket in the rollout is an open one. Every
sub-agent the page draws is one more mascot, so a finished one that still counts
is visible.

**A Claude session puts its sub-agents out two ways, and files them apart.** The
ones it drives from its own turn write `subagents/agent-<id>.jsonl`; the ones a
*workflow* script drives write `subagents/workflows/<run>/agent-<id>.jsonl`, and
a workflow is where the sub-agents actually get numerous. Neither test that
retires a task sub-agent retires one of those: the session never announces them,
and their transcripts end on the tool result that fed the last reply rather than
on the reply, so they close on a `user` record and never on `end_turn`. The
`journal.jsonl` beside them is what brackets each — `started` when the script
spawns it, `result` when it returns — and only a record's type and the `agentId`
it names are ever read, never the `result` payload, which is the workflow's own
work. Count both populations, or a nine-agent workflow draws one lone mascot.

Prompts, objectives, task titles, sub-agent
descriptions and `threads.title` are never read into a metric — a `model` and a
*count*, of agents and of sub-agents per model, are the only things added to
`realtime_agent_*`. A host runs as many agents of one kind as it likes: three
codex sessions in three terminals are three agents, counted by
`realtime_agent_running_count` and expanded by the gateway into one `Agent` each,
whose uid it mints from the host, the kind, the model and the ordinal so that no
thread id ever reaches the page. A sub-agent may run a different model from the
agent that spawned it, so that count is labelled by model too and expands into
one `Subagent` per worker. Neither count carries an identity, so the budget and
the sub-agents — which the collector can only report for a kind — are carried by
the first agent of that kind.

**The probe caches `/metrics` and serves it from a fixed thread pool.** Process
and open-file inspection is shared through `psutil`; media and accessory details
live behind Linux, macOS, and Windows device adapters. `server.cached()` keeps
the scrape rate from becoming the adapter-command rate, and `PooledHTTPServer`
keeps a burst of connections from becoming a burst of threads. The cache TTL
must stay under Prometheus's scrape interval. Prometheus label values carry no
free text beyond the track title and artist — the artwork URL is gone, because a
per-play CDN URL is an unbounded series and rendering it made every visitor's
browser fetch from a third party.

**PromQL lives only in the gateway.** `internal/gateway/metrics.go` is the one
place a query expression is written. Clients name a `MetricSeries` and pass
domain selectors (device uid, agent kind, accessory); the gateway resolves the
metric name, labels, and job, escapes every selector, and bounds the window to
`maxMetricRangePoints`. There is no PromQL passthrough — an authenticated owner
cannot run an arbitrary query. Don't reintroduce a `query=` parameter.

## Gotchas

- Export `JAVA_HOME` to a JDK supported by the current Gradle and Android Gradle
  Plugin versions; never commit a host-specific `org.gradle.java.home` path
  property. Configure it per host instead.
- `deploy/status/prometheus/file_sd/cadvisor.yml` is intentionally an empty
  `[]`. cAdvisor sits behind the `containers` Compose profile, so an empty
  file_sd list is the opt-in switch; `cadvisor.yml.example` shows the contents.
  A `static_configs` entry would leave a permanently-down target.
- Probe installation never accepts a branch ref. Pass the reviewed commit through
  `REALTIME_PROBE_RELEASE`; the installer tries commit-pinned jsdelivr and GitHub
  raw mirrors and rejects stale or mixed content by SHA-256.
- `GetPublicStatus` is unauthenticated, so its Prometheus fan-out sits behind a
  2-second single-flight cache in `internal/gateway/status.go`. Don't add a query
  to the assembly without checking it runs inside `parallel(...)`.
- Prometheus runs without `--web.enable-lifecycle`: nothing here reloads it, and
  the flag serves unauthenticated `/-/reload` and `/-/quit`. Config changes need a
  container restart.
- The Site Worker's Status allowlist contains only public `GetPublicStatus`,
  `GetProfile`, and `ListProjects`; it must never proxy Metrics, internal Status,
  enrollment, ingest, or `/api/` control routes. Console reaches internal APIs
  through same-origin BFF prefixes and downstream OIDC permission checks.
- `tokens.ingest` and `tokens.discovery` are separate workload secrets. The
  discovery token is used only by Prometheus for HTTP service discovery and the
  gateway process scrape, and is loaded from
  `deploy/status/prometheus/discovery_token`; humans use OIDC and the browser
  stores only the Console session cookie.
