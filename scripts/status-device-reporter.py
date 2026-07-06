#!/usr/bin/env python3
from __future__ import annotations

import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
import platform
import re
import shutil
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path

CPU_CORES = "system.cpu.logical.count"
CPU_USAGE = "system.cpu.utilization"
MEMORY_USAGE = "system.memory.usage"
MEMORY_LIMIT = "system.memory.limit"
FILESYSTEM_USAGE = "system.filesystem.usage"
FILESYSTEM_LIMIT = "system.filesystem.limit"
FILESYSTEM_UTILIZATION = "system.filesystem.utilization"


@dataclass(frozen=True)
class MemoryUsage:
    used: int
    total: int


@dataclass(frozen=True)
class MediaSnapshot:
    title: str
    artist: str = ""
    player: str = ""


def main() -> int:
    args = parse_args()
    if args.serve:
        return serve(args)

    payload = build_payload(args)
    if args.print:
        print(json.dumps(payload, indent=2))
        return 0

    token = args.token or os.getenv("STATUS_INGEST_TOKEN")
    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2

    return post_status(args.url.rstrip("/") + "/api/ingest/host", token, payload, args.timeout_seconds)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Publish this device status to realtime-me gateway.")
    parser.add_argument("--url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--token", default="")
    parser.add_argument("--device-id", default=os.getenv("STATUS_DEVICE_ID", socket.gethostname()))
    parser.add_argument("--device-name", default=os.getenv("STATUS_DEVICE_NAME", socket.gethostname()))
    parser.add_argument("--device-model", default=os.getenv("STATUS_DEVICE_MODEL", device_model()))
    parser.add_argument("--kind", default=os.getenv("STATUS_DEVICE_KIND", "host"))
    parser.add_argument("--role", default=os.getenv("STATUS_DEVICE_ROLE", "desktop"))
    parser.add_argument("--timeout-seconds", type=float, default=5)
    parser.add_argument("--serve", action="store_true")
    parser.add_argument("--bind", default=os.getenv("STATUS_DEVICE_EXPORTER_BIND", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("STATUS_DEVICE_EXPORTER_PORT", "18083")))
    parser.add_argument("--print", action="store_true")
    return parser.parse_args()


def build_payload(args: argparse.Namespace) -> dict:
    payload = {
        "device_id": args.device_id,
        "device_name": args.device_name,
        "device_model": args.device_model,
        "kind": args.kind,
        "role": args.role,
        "state": "online",
        "updated_at": utc_now(),
        "metrics": metrics(),
    }
    media = current_media()
    if media:
        payload["media"] = compact_dict({"title": media.title, "artist": media.artist})
    return payload


def serve(args: argparse.Namespace) -> int:
    class Handler(BaseHTTPRequestHandler):
        def do_GET(self) -> None:
            if self.path == "/healthz":
                self.write_json({"ok": True})
                return
            if self.path == "/api/device-status":
                self.write_json(build_payload(args))
                return
            if self.path == "/metrics":
                self.write_text(render_prometheus_metrics())
                return
            self.write_json({"error": "not_found"}, 404)

        def log_message(self, _format: str, *_args: object) -> None:
            return

        def write_json(self, payload: object, status: int = 200) -> None:
            data = json.dumps(payload).encode()
            self.send_response(status)
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

        def write_text(self, payload: str, status: int = 200) -> None:
            data = payload.encode()
            self.send_response(status)
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

    server = ThreadingHTTPServer((args.bind, args.port), Handler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 130
    return 0


def metrics() -> list[dict]:
    disk = shutil.disk_usage("/")
    memory = memory_usage()
    result = [
        sample(CPU_CORES, "{cpu}", float(os.cpu_count() or 0)),
        sample(CPU_USAGE, "1", cpu_usage()),
        sample(MEMORY_USAGE, "By", float(memory.used), {"system.memory.state": "used"}),
        sample(MEMORY_LIMIT, "By", float(memory.total)),
        sample(FILESYSTEM_USAGE, "By", float(disk.used), {"mountpoint": "/"}),
        sample(FILESYSTEM_LIMIT, "By", float(disk.total), {"mountpoint": "/"}),
        sample(FILESYSTEM_UTILIZATION, "1", disk.used / disk.total if disk.total else 0, {"mountpoint": "/"}),
    ]
    return [item for item in result if item["value"] >= 0]


def sample(name: str, unit: str, value: float, attributes: dict[str, str] | None = None) -> dict:
    payload = {"name": name, "unit": unit, "value": round(value, 6)}
    if attributes:
        payload["attributes"] = attributes
    return payload


def render_prometheus_metrics() -> str:
    lines = [
        "# HELP realtime_device_media_playing Device media playback state. OpenTelemetry name: realtime.device.media.playing.",
        "# TYPE realtime_device_media_playing gauge",
        "# UNIT realtime_device_media_playing 1",
    ]
    media = current_media()
    if media:
        labels = {"title": media.title, "artist": media.artist, "player": media.player}
        lines.append(f"realtime_device_media_playing{label_set(labels)} 1")
    lines.append("")
    return "\n".join(lines)


def current_media() -> MediaSnapshot | None:
    system = platform.system().lower()
    if system == "darwin":
        return darwin_media()
    if system == "linux":
        return linux_media()
    return None


def darwin_media() -> MediaSnapshot | None:
    for provider in (darwin_nowplaying_cli, darwin_music_media, darwin_spotify_media):
        media = provider()
        if media:
            return media
    return None


def darwin_nowplaying_cli() -> MediaSnapshot | None:
    if not command_exists("nowplaying-cli"):
        return None
    title = run(["nowplaying-cli", "get", "title"]).strip()
    if not title:
        return None
    artist = run(["nowplaying-cli", "get", "artist"]).strip()
    player = run(["nowplaying-cli", "get", "bundleIdentifier"]).strip()
    return MediaSnapshot(
        title=sanitize_media_text(title),
        artist=sanitize_media_text(artist),
        player=sanitize_media_text(player),
    )


def darwin_music_media() -> MediaSnapshot | None:
    script = """
tell application "System Events" to set musicRunning to exists (processes where name is "Music")
if musicRunning then
  tell application "Music"
    if player state is playing then
      return (name of current track) & linefeed & (artist of current track)
    end if
  end tell
end if
return ""
"""
    title, artist = media_title_artist(run_osascript(script))
    if not title:
        return None
    return MediaSnapshot(title=title, artist=artist, player="Music")


def darwin_spotify_media() -> MediaSnapshot | None:
    script = """
tell application "System Events" to set spotifyRunning to exists (processes where name is "Spotify")
if spotifyRunning then
  tell application "Spotify"
    if player state is playing then
      return (name of current track) & linefeed & (artist of current track)
    end if
  end tell
end if
return ""
"""
    title, artist = media_title_artist(run_osascript(script))
    if not title:
        return None
    return MediaSnapshot(title=title, artist=artist, player="Spotify")


def linux_media() -> MediaSnapshot | None:
    if not command_exists("playerctl"):
        return None
    players = [line.strip() for line in run(["playerctl", "--list-all"]).splitlines() if line.strip()]
    for player in players:
        if run(["playerctl", "--player", player, "status"]).strip().lower() != "playing":
            continue
        title = run(["playerctl", "--player", player, "metadata", "title"]).strip()
        if title:
            artist = first_non_empty(
                run(["playerctl", "--player", player, "metadata", "artist"]).strip(),
                run(["playerctl", "--player", player, "metadata", "xesam:artist"]).strip(),
            )
            return MediaSnapshot(
                title=sanitize_media_text(title),
                artist=sanitize_media_text(artist),
                player=sanitize_media_text(player),
            )
    return None


def media_title_artist(output: str) -> tuple[str, str]:
    parts = [sanitize_media_text(line) for line in output.splitlines()]
    parts = [part for part in parts if part]
    if not parts:
        return "", ""
    return parts[0], parts[1] if len(parts) > 1 else ""


def first_non_empty(*values: str) -> str:
    for value in values:
        if value:
            return value
    return ""


def compact_dict(values: dict[str, str]) -> dict[str, str]:
    return {key: value for key, value in values.items() if value}


def run_osascript(script: str) -> str:
    return run(["osascript", "-e", script]) if command_exists("osascript") else ""


def command_exists(name: str) -> bool:
    return shutil.which(name) is not None


def sanitize_media_text(value: str) -> str:
    text = re.sub(r"[\x00-\x1f\x7f]", " ", value)
    text = re.sub(r"\s+", " ", text).strip()
    if len(text) <= 120:
        return text
    return text[:119].rstrip() + "…"


def label_set(labels: dict[str, str]) -> str:
    pairs = []
    for key in sorted(labels):
        value = labels[key]
        if value:
            pairs.append(f'{prometheus_label_name(key)}="{escape_label_value(value)}"')
    return "{" + ",".join(pairs) + "}" if pairs else ""


def prometheus_label_name(value: str) -> str:
    name = re.sub(r"[^A-Za-z0-9_]", "_", value)
    if not name:
        return "label"
    if name[0].isdigit():
        return f"label_{name}"
    return name


def escape_label_value(value: str) -> str:
    return value.replace("\\", "\\\\").replace("\n", "\\n").replace('"', '\\"')


def cpu_usage() -> float:
    system = platform.system().lower()
    if system == "linux":
        first = linux_cpu_times()
        time.sleep(0.25)
        second = linux_cpu_times()
        total = sum(second) - sum(first)
        idle = second[3] - first[3]
        return clamp_ratio(1 - idle / total) if total > 0 else 0
    if system == "darwin":
        output = run(["top", "-l", "2", "-n", "0", "-s", "1"])
        matches = re.findall(r"CPU usage: .*?, .*?, ([0-9.]+)% idle", output)
        if matches:
            return clamp_ratio(1 - float(matches[-1]) / 100)
    return 0


def linux_cpu_times() -> list[int]:
    try:
        values = Path("/proc/stat").read_text().splitlines()[0].split()[1:]
    except OSError:
        return [0, 0, 0, 0]
    return [int(value) for value in values]


def memory_usage() -> MemoryUsage:
    system = platform.system().lower()
    if system == "linux":
        return linux_memory_usage()
    if system == "darwin":
        return darwin_memory_usage()
    return MemoryUsage(0, 0)


def linux_memory_usage() -> MemoryUsage:
    values = {}
    try:
        lines = Path("/proc/meminfo").read_text().splitlines()
    except OSError:
        return MemoryUsage(0, 0)
    for line in lines:
        key, value = line.split(":", 1)
        values[key] = int(value.strip().split()[0]) * 1024
    total = values.get("MemTotal", 0)
    available = values.get("MemAvailable", 0)
    used = max(0, total - available)
    return MemoryUsage(used, total)


def darwin_memory_usage() -> MemoryUsage:
    total = int(run(["sysctl", "-n", "hw.memsize"]).strip() or "0")
    stats = run(["vm_stat"])
    page_size_match = re.search(r"page size of (\d+) bytes", stats)
    page_size = int(page_size_match.group(1)) if page_size_match else 4096
    pages = {}
    for line in stats.splitlines():
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        pages[key.strip()] = int(re.sub(r"[^0-9]", "", value) or "0")
    free = (pages.get("Pages free", 0) + pages.get("Pages speculative", 0)) * page_size
    used = max(0, total - free)
    return MemoryUsage(used, total)


def device_model() -> str:
    system = platform.system().lower()
    if system == "darwin":
        os_name = " ".join(part for part in [
            run(["sw_vers", "-productName"]).strip(),
            run(["sw_vers", "-productVersion"]).strip(),
        ] if part)
        model = run(["sysctl", "-n", "hw.model"]).strip()
        cpu = run(["sysctl", "-n", "machdep.cpu.brand_string"]).strip()
        return " · ".join(part for part in [os_name, model, cpu] if part)[:120]
    if system == "linux":
        return linux_os_name() or platform.machine()
    return platform.platform()


def linux_os_name() -> str:
    try:
        for line in Path("/etc/os-release").read_text().splitlines():
            if line.startswith("PRETTY_NAME="):
                return line.split("=", 1)[1].strip().strip('"')
    except OSError:
        return ""
    return ""


def post_status(endpoint: str, token: str, payload: dict, timeout_seconds: float) -> int:
    request = urllib.request.Request(
        endpoint,
        data=json.dumps(payload).encode(),
        method="POST",
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json; charset=utf-8",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            if response.status < 200 or response.status > 299:
                print(f"gateway rejected device status: HTTP {response.status}", file=sys.stderr)
                return 1
    except urllib.error.HTTPError as error:
        print(f"gateway rejected device status: HTTP {error.code}", file=sys.stderr)
        return 1
    except OSError as error:
        print(f"gateway device status push failed: {error.__class__.__name__}", file=sys.stderr)
        return 1
    return 0


def run(command: list[str]) -> str:
    try:
        return subprocess.check_output(command, text=True, stderr=subprocess.DEVNULL, timeout=5)
    except (OSError, subprocess.SubprocessError):
        return ""


def utc_now() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def clamp_ratio(value: float) -> float:
    return max(0, min(1, value))


if __name__ == "__main__":
    raise SystemExit(main())
