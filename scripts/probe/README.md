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

- `REALTIME_PROBE_BIND` and `REALTIME_PROBE_PORT` (installer defaults: `0.0.0.0:18082`);
- `REALTIME_PROBE_ACTIVE_WINDOW_SECONDS` (default: `300`);
- `REALTIME_CODEX_HOMES`, separated with the operating system's path separator;
- `REALTIME_CLAUDE_HOME` (default: `~/.claude`).

`/healthz` reports liveness and the latest result of each collector. A collector failure is isolated: the scrape remains available, `realtime_probe_collector_success` identifies the failed capability, and the service log retains the traceback.
