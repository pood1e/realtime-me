# Realtime Me probe

`realtime_probe` is the single pull-based collector for Linux, macOS, and Windows hosts. It exposes one `/metrics` endpoint and never stores a gateway URL, device identity, or token. Prometheus applies the enrolled device labels through HTTP service discovery.

The collector registry has three independent capabilities:

- `system`: CPU, memory, root-filesystem, OS, architecture, and hardware model through `psutil`;
- `device`: current media and Bluetooth audio accessories through an OS adapter;
- `agents`: privacy-minimized Codex and Claude Code state from local process metadata and their own state files.

Prompts, objectives, task titles, tool arguments, responses, and transcript content are never exported. Agent transcripts are read only far enough to derive lifecycle brackets and model identifiers.

## Operating-system adapters

| Capability | Linux | macOS | Windows |
| --- | --- | --- | --- |
| System and process collection | `psutil` | `psutil` | `psutil` |
| Codex and Claude Code | yes | yes | yes |
| Current media | `playerctl` | Music, Spotify, `nowplaying-cli` | unavailable |
| Bluetooth audio | BlueZ `bluetoothctl` | `system_profiler` | unavailable |
| Service manager | systemd | LaunchAgent | Task Scheduler |

Windows omits the two optional device signals rather than depending on an unmaintained native bridge. The other metric families and the scrape contract are identical on every operating system.

## Runtime settings

The installer persists these options in the service command:

- `REALTIME_PROBE_BIND` and `REALTIME_PROBE_PORT` (installer defaults: the detected private address on port `18082`);
- `REALTIME_PROBE_ACTIVE_WINDOW_SECONDS` (default: `300`);
- `REALTIME_CODEX_HOMES`, separated with the operating system's path separator;
- `REALTIME_CLAUDE_HOME` (default: `~/.claude`).

`/healthz` reports liveness and the latest result of each collector. A collector failure is isolated: the scrape remains available, `realtime_probe_collector_success` identifies the failed capability, and the service log retains the traceback.

## Installation integrity

Install only from a reviewed 40-character Git commit and pass it as
`REALTIME_PROBE_RELEASE`. The installer rejects branch refs, mixed CDN content,
unlisted runtime files, oversized downloads, and SHA-256 mismatches. Dependency
wheels are version- and hash-pinned and source builds are disabled.
The locked wheel set covers x86-64 and ARM64 on all three operating systems,
including 64-bit Raspberry Pi OS; 32-bit installations are not supported.

Linux installs under `/opt/realtime-me-probe`; macOS uses the root-owned
`/Library/Application Support/RealtimeMeProbe` plus a per-user LaunchAgent;
Windows uses `%ProgramFiles%\RealtimeMeProbe` and a limited scheduled-task principal.
The runtime is replace-on-success; an existing runtime is restored if replacement
activation fails. The service identity has read/execute permission but cannot
change its code. Linux and macOS select `SUDO_USER`; unattended root installs
must set `REALTIME_PROBE_USER` to the non-root account whose agent state is
collected.

Use `REALTIME_PROBE_SOURCE_URLS` only for immutable HTTPS mirrors. After changing
the collector or requirements, run `make generate-probe-integrity`; never edit
`integrity.json` or the installer digest by hand.
