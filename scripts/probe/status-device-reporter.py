#!/usr/bin/env python3
"""Read-only device exporter for Prometheus.

This host is unaware of the gateway. It exposes media and Bluetooth-accessory
signals on /metrics; Prometheus discovers it through the gateway's HTTP service
discovery and stamps the gateway-minted device uid onto every series as a target
label. The exporter therefore holds no identity, no token, and no gateway
address. CPU, memory, and filesystem metrics come from node_exporter.
"""
from __future__ import annotations

import argparse
import json
import os
import platform
import re
import shutil
from dataclasses import dataclass
from pathlib import Path

from status_common import (
    decode_json,
    json_response,
    label_set,
    run,
    cached,
    run_server,
    text_response,
)


# What a track is, who owns it, and whether it is moving. Named rather than dumped
# with get-raw, whose answer also carries the cover art as tens of kilobytes of
# base64 -- an image this exporter has no use for and must never publish.
NOWPLAYING_FIELDS = ("title", "artist", "playbackRate", "clientBundleIdentifier")
# coreaudiod takes this assertion out while an audio device is running, and names
# it after the device the sound is going to, whichever one that turns out to be.
DARWIN_AUDIO_ASSERTION_PATTERN = re.compile(r"coreaudiod.*com\.apple\.audio\.", re.IGNORECASE)


@dataclass(frozen=True)
class MediaSnapshot:
    title: str
    artist: str = ""


@dataclass(frozen=True)
class AccessorySnapshot:
    kind: str
    name: str
    model: str = ""
    battery_percent: int | None = None


def main() -> int:
    args = parse_args()
    routes = {
        "/healthz": lambda: json_response({"ok": True}),
        "/metrics": cached(lambda: text_response(render_prometheus_metrics())),
    }
    return run_server(args.bind, args.port, routes)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Serve this device's media and accessory metrics for Prometheus.")
    parser.add_argument("--bind", default=os.getenv("STATUS_DEVICE_EXPORTER_BIND", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("STATUS_DEVICE_EXPORTER_PORT", "18083")))
    return parser.parse_args()


def render_prometheus_metrics() -> str:
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
        lines.append(f"realtime_device_media_playing{label_set({'title': media.title, 'artist': media.artist})} 1")
    for accessory in bluetooth_audio_accessories():
        labels = accessory_labels(accessory)
        lines.append(f"realtime_device_accessory_connected{label_set(labels)} 1")
        if accessory.battery_percent is not None:
            lines.append(f"realtime_device_accessory_battery_level_ratio{label_set(labels)} {accessory.battery_percent / 100:.6g}")
    lines.append("")
    return "\n".join(lines)


def accessory_labels(accessory: AccessorySnapshot) -> dict[str, str]:
    labels = {
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
    # Spotify/Music are tried first because they check play state themselves;
    # nowplaying-cli is the fallback for every other player.
    for provider in (darwin_spotify_media, darwin_music_media, darwin_nowplaying_cli):
        media = provider()
        if media:
            return media
    return None


def darwin_nowplaying_cli() -> MediaSnapshot | None:
    """The track MediaRemote holds, if this Mac is in fact playing it.

    MediaRemote is told what a player queued, never what it went on to do. An
    application that does not announce its own pause -- and plenty do not -- leaves
    its last track sitting there at a playback rate of 1 for as long as it stays
    open, so the rate is a claim rather than an observation, and on its own it
    pins a song to the page that stopped hours ago. What is actually coming out of
    the speakers is the thing to ask about, and coreaudiod answers it.

    Every field is read in one call, so the snapshot is of a single moment: asked
    a field at a time, a track that changes mid-scrape hands back one song's title
    over the next song's artist.
    """
    command = command_path("nowplaying-cli")
    if not command:
        return None
    if not darwin_audio_is_running():
        return None
    playing = decode_json(run([command, "get", "--json", *NOWPLAYING_FIELDS]))
    if read_float(playing.get("playbackRate")) <= 0:
        return None
    # MediaRemote keeps that item long after the application that queued it has
    # quit. Nothing owns a track like that, so the application still holding it is
    # what separates a live one from a ghost. It names the owner and never reaches
    # a label: the page is told what is playing, not which program plays it.
    title = sanitize_media_text(str(playing.get("title") or ""))
    owner = sanitize_media_text(str(playing.get("clientBundleIdentifier") or ""))
    if not title or not owner:
        return None
    return MediaSnapshot(title=title, artist=sanitize_media_text(str(playing.get("artist") or "")))


def darwin_audio_is_running() -> bool:
    """Whether sound is coming out of this Mac at all.

    coreaudiod holds a power assertion for exactly as long as an audio device is
    running, and drops it the moment the device goes quiet. It is the one signal
    here that no application can leave stale, because the system takes it rather
    than being told it.
    """
    assertions = run(["pmset", "-g", "assertions"], timeout_seconds=5)
    return any(DARWIN_AUDIO_ASSERTION_PATTERN.search(line) for line in assertions.splitlines())


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
    return MediaSnapshot(title=title, artist=artist)


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
    return MediaSnapshot(title=title, artist=artist)


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
            return MediaSnapshot(title=sanitize_media_text(title), artist=sanitize_media_text(artist))
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


def read_float(value: object) -> float:
    """A number from a field that was only ever meant to hold one."""
    try:
        return float(value)  # type: ignore[arg-type]
    except (TypeError, ValueError):
        return 0.0


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


if __name__ == "__main__":
    raise SystemExit(main())
