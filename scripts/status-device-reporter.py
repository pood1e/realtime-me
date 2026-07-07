#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import platform
import re
import shutil
import socket
import sys
import time
from dataclasses import dataclass
from pathlib import Path

import status_common
from status_common import (
    ONLINE_STATE_ONLINE,
    ConnectError,
    device_kind_enum,
    device_role_enum,
    ensure_device_uid,
    json_response,
    label_set,
    run,
    run_server,
    text_response,
)

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


@dataclass(frozen=True)
class AccessorySnapshot:
    kind: str
    name: str
    model: str = ""
    battery_percent: int | None = None


def main() -> int:
    args = parse_args()
    if args.serve:
        return serve(args)

    token = os.getenv("STATUS_INGEST_TOKEN", "").strip()
    try:
        device_uid = resolve_device_uid(args, token)
    except (ConnectError, OSError) as error:
        print(f"device enrollment failed: {enrollment_reason(error)}", file=sys.stderr)
        return 1

    if args.print_uid:
        if not device_uid:
            print("device is not enrolled; set STATUS_INGEST_TOKEN to enroll", file=sys.stderr)
            return 1
        print(device_uid)
        return 0

    report = build_report(args, device_uid)
    if args.print:
        print(json.dumps(report, indent=2, ensure_ascii=False))
        return 0

    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2
    if not device_uid:
        print("device is not enrolled", file=sys.stderr)
        return 1
    return push_device_status(args, token, report)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Publish this device status to realtime-me gateway.")
    parser.add_argument("--url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--device-name", default=os.getenv("STATUS_DEVICE_NAME", socket.gethostname()))
    parser.add_argument("--device-model", default=os.getenv("STATUS_DEVICE_MODEL", device_model()))
    parser.add_argument("--kind", default=os.getenv("STATUS_DEVICE_KIND", "host"))
    parser.add_argument("--role", default=os.getenv("STATUS_DEVICE_ROLE", "desktop"))
    parser.add_argument("--identity-file", default=os.getenv("STATUS_IDENTITY_FILE", ""))
    parser.add_argument("--timeout-seconds", type=float, default=5)
    parser.add_argument("--serve", action="store_true")
    parser.add_argument("--bind", default=os.getenv("STATUS_DEVICE_EXPORTER_BIND", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("STATUS_DEVICE_EXPORTER_PORT", "18083")))
    parser.add_argument("--print", action="store_true")
    parser.add_argument("--print-uid", action="store_true")
    return parser.parse_args()


def identity_file(args: argparse.Namespace) -> Path:
    return Path(args.identity_file) if args.identity_file else status_common.default_identity_file()


def resolve_device_uid(args: argparse.Namespace, token: str) -> str:
    return ensure_device_uid(
        args.url,
        token,
        identity_file(args),
        device_kind_enum(args.kind),
        device_role_enum(args.role),
        args.device_name,
        args.device_model,
        args.timeout_seconds,
    )


def build_report(args: argparse.Namespace, device_uid: str) -> dict:
    report = {
        "deviceUid": device_uid,
        "kind": device_kind_enum(args.kind),
        "role": device_role_enum(args.role),
        "displayName": args.device_name,
        "model": args.device_model,
        "state": ONLINE_STATE_ONLINE,
        "metrics": metrics(),
    }
    media = current_media()
    if media:
        report["media"] = compact_dict({"title": media.title, "artist": media.artist})
    accessories = bluetooth_audio_accessories()
    if accessories:
        report["accessories"] = [accessory_payload(accessory) for accessory in accessories]
    return report


def serve(args: argparse.Namespace) -> int:
    def device_uid() -> str:
        return read_cached_uid(args)

    routes = {
        "/healthz": lambda: json_response({"ok": True}),
        "/api/device-status": lambda: json_response(build_report(args, device_uid())),
        "/metrics": lambda: text_response(render_prometheus_metrics(device_uid())),
    }
    return run_server(args.bind, args.port, routes)


def read_cached_uid(args: argparse.Namespace) -> str:
    return status_common.read_cached_uid(identity_file(args))


def push_device_status(args: argparse.Namespace, token: str, report: dict) -> int:
    try:
        status_common.connect_post(
            args.url,
            "IngestService",
            "ReportDeviceStatus",
            token,
            {"device": report},
            args.timeout_seconds,
        )
    except ConnectError as error:
        print(f"gateway rejected device status: {error.code}", file=sys.stderr)
        return 1
    except OSError as error:
        print(f"gateway device status push failed: {error.__class__.__name__}", file=sys.stderr)
        return 1
    return 0


def enrollment_reason(error: Exception) -> str:
    return error.code if isinstance(error, ConnectError) else error.__class__.__name__


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


def render_prometheus_metrics(device_uid: str) -> str:
    lines = [
        "# HELP realtime_device_media_playing Device media playback state. OpenTelemetry name: realtime.device.media.playing.",
        "# TYPE realtime_device_media_playing gauge",
        "# UNIT realtime_device_media_playing 1",
        "# HELP realtime_device_accessory_connected Connected accessory state labelled by accessory name and kind. OpenTelemetry name: realtime.device.accessory.connected.",
        "# TYPE realtime_device_accessory_connected gauge",
        "# UNIT realtime_device_accessory_connected 1",
        "# HELP realtime_device_accessory_battery_level_ratio Accessory battery level as a fraction of total capacity. OpenTelemetry name: realtime.device.accessory.battery.level.",
        "# TYPE realtime_device_accessory_battery_level_ratio gauge",
        "# UNIT realtime_device_accessory_battery_level_ratio 1",
    ]
    media = current_media()
    if media:
        labels = {"device_uid": device_uid, "title": media.title, "artist": media.artist, "player": media.player}
        lines.append(f"realtime_device_media_playing{label_set(labels)} 1")
    for accessory in bluetooth_audio_accessories():
        labels = accessory_labels(device_uid, accessory)
        lines.append(f"realtime_device_accessory_connected{label_set(labels)} 1")
        if accessory.battery_percent is not None:
            lines.append(f"realtime_device_accessory_battery_level_ratio{label_set(labels)} {accessory.battery_percent / 100:.6g}")
    lines.append("")
    return "\n".join(lines)


def accessory_payload(accessory: AccessorySnapshot) -> dict:
    payload: dict[str, str | int] = {
        "kind": accessory.kind,
        "displayName": accessory.name,
    }
    if accessory.model:
        payload["model"] = accessory.model
    if accessory.battery_percent is not None:
        payload["batteryPercent"] = accessory.battery_percent
    return payload


def accessory_labels(device_uid: str, accessory: AccessorySnapshot) -> dict[str, str]:
    labels = {
        "device_uid": device_uid,
        "accessory_kind": accessory.kind,
        "accessory_name": accessory.name,
    }
    if accessory.model:
        labels["accessory_model"] = accessory.model
    return labels


def bluetooth_audio_accessories() -> list[AccessorySnapshot]:
    system = platform.system().lower()
    if system == "darwin":
        return darwin_bluetooth_audio_accessories()
    if system == "linux":
        return linux_bluetooth_audio_accessories()
    return []


def darwin_bluetooth_audio_accessories() -> list[AccessorySnapshot]:
    output = run(["system_profiler", "SPBluetoothDataType", "-json"], timeout_seconds=8)
    if not output:
        return []
    try:
        payload = json.loads(output)
    except json.JSONDecodeError:
        return []
    devices = []
    for controller in payload.get("SPBluetoothDataType", []):
        if not isinstance(controller, dict):
            continue
        connected = controller.get("device_connected", [])
        if isinstance(connected, list):
            devices.extend(darwin_connected_accessories(connected))
    return unique_accessories(devices)


def darwin_connected_accessories(devices: list) -> list[AccessorySnapshot]:
    accessories = []
    for item in devices:
        if not isinstance(item, dict):
            continue
        for raw_name, details in item.items():
            if not isinstance(details, dict):
                continue
            name = sanitize_media_text(str(raw_name))
            minor_type = sanitize_media_text(str(details.get("device_minorType", "")))
            services = sanitize_media_text(str(details.get("device_services", "")))
            if not name or not is_audio_accessory(minor_type, services):
                continue
            accessories.append(AccessorySnapshot(
                kind="bluetooth_audio",
                name=name,
                model=minor_type if minor_type and minor_type.lower() != "headphones" else "",
                battery_percent=parse_percent(str(details.get("device_batteryLevelMain", ""))),
            ))
    return accessories


def linux_bluetooth_audio_accessories() -> list[AccessorySnapshot]:
    if not command_exists("bluetoothctl"):
        return []
    output = run(["bluetoothctl", "devices", "Connected"])
    devices = parse_bluetoothctl_devices(output)
    if not devices:
        devices = parse_bluetoothctl_devices(run(["bluetoothctl", "devices"]))
    accessories = []
    for address, fallback_name in devices:
        info = run(["bluetoothctl", "info", address])
        if not info or not bluetoothctl_connected(info):
            continue
        name = sanitize_media_text(bluetoothctl_field(info, "Name") or fallback_name)
        icon = bluetoothctl_field(info, "Icon")
        uuids = " ".join(re.findall(r"UUID: (.+)", info))
        if not name or not is_audio_accessory(icon, uuids):
            continue
        accessories.append(AccessorySnapshot(
            kind="bluetooth_audio",
            name=name,
            model=sanitize_media_text(bluetoothctl_field(info, "Alias")) if bluetoothctl_field(info, "Alias") != name else "",
            battery_percent=parse_bluetoothctl_battery(info),
        ))
    return unique_accessories(accessories)


def parse_bluetoothctl_devices(output: str) -> list[tuple[str, str]]:
    devices = []
    for line in output.splitlines():
        match = re.match(r"Device\s+([0-9A-Fa-f:]{17})\s+(.+)", line.strip())
        if match:
            devices.append((match.group(1), sanitize_media_text(match.group(2))))
    return devices


def bluetoothctl_connected(info: str) -> bool:
    connected = bluetoothctl_field(info, "Connected").lower()
    return connected == "yes" or connected == "true"


def bluetoothctl_field(info: str, field: str) -> str:
    pattern = re.compile(rf"^\s*{re.escape(field)}:\s*(.+?)\s*$", re.MULTILINE)
    match = pattern.search(info)
    return sanitize_media_text(match.group(1)) if match else ""


def parse_bluetoothctl_battery(info: str) -> int | None:
    for line in info.splitlines():
        if "Battery Percentage:" not in line:
            continue
        parenthesized = re.search(r"\((\d{1,3})\)", line)
        if parenthesized:
            return clamp_percent(int(parenthesized.group(1)))
        percent = parse_percent(line)
        if percent is not None:
            return percent
    return None


def is_audio_accessory(*values: str) -> bool:
    text = " ".join(values).lower()
    tokens = ("headphone", "headset", "earbud", "a2dp", "hfp", "avrcp", "audio")
    return any(token in text for token in tokens)


def unique_accessories(accessories: list[AccessorySnapshot]) -> list[AccessorySnapshot]:
    by_key: dict[tuple[str, str, str], AccessorySnapshot] = {}
    for accessory in accessories:
        key = (accessory.kind, accessory.name, accessory.model)
        existing = by_key.get(key)
        if existing is None or existing.battery_percent is None:
            by_key[key] = accessory
    return sorted(by_key.values(), key=lambda item: (item.kind, item.name, item.model))


def parse_percent(value: str) -> int | None:
    match = re.search(r"(\d{1,3})\s*%", value)
    if not match:
        return None
    return clamp_percent(int(match.group(1)))


def clamp_percent(value: int) -> int:
    return max(0, min(100, value))


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
    command = command_path("nowplaying-cli")
    if not command:
        return None
    title = run([command, "get", "title"]).strip()
    if not title:
        return None
    artist = run([command, "get", "artist"]).strip()
    player = run([command, "get", "bundleIdentifier"]).strip()
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
    return command_path(name) is not None


def command_path(name: str) -> str | None:
    path = shutil.which(name)
    if path:
        return path
    for directory in ("/opt/homebrew/bin", "/usr/local/bin"):
        candidate = Path(directory) / name
        if candidate.exists() and os.access(candidate, os.X_OK):
            return str(candidate)
    return None


def sanitize_media_text(value: str) -> str:
    text = re.sub(r"[\x00-\x1f\x7f]", " ", value)
    text = re.sub(r"\s+", " ", text).strip()
    if text.lower() == "null":
        return ""
    if len(text) <= 120:
        return text
    return text[:119].rstrip() + "…"


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
    used_pages = sum(
        pages.get(key, 0)
        for key in (
            "Pages active",
            "Pages inactive",
            "Pages throttled",
            "Pages wired down",
            "Pages occupied by compressor",
        )
    )
    used = used_pages * page_size
    if used <= 0:
        free = (pages.get("Pages free", 0) + pages.get("Pages speculative", 0)) * page_size
        used = total - free
    used = max(0, min(used, total))
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


def clamp_ratio(value: float) -> float:
    return max(0, min(1, value))


if __name__ == "__main__":
    raise SystemExit(main())
