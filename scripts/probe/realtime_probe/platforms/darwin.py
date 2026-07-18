"""macOS media, Bluetooth, and hardware adapters."""

from __future__ import annotations

import json
import re
from pathlib import Path

from .base import (
    AccessorySnapshot,
    MediaSnapshot,
    command_path,
    is_audio_accessory,
    parse_percent,
    run,
    sanitize_text,
    unique_accessories,
)

NOW_PLAYING_FIELDS = ("title", "artist", "playbackRate", "clientBundleIdentifier")
ASSERTION_HEADER_PATTERN = re.compile(r"^\s*pid \d+\(")
AUDIO_ASSERTION_PATTERN = re.compile(r"coreaudiod.*com\.apple\.audio\.", re.IGNORECASE)
ASSERTION_OWNER_PATTERN = re.compile(r"Created for PID:\s*(\d+)")


class DarwinDeviceAdapter:
    def current_media(self) -> MediaSnapshot | None:
        for provider in (
            self._spotify_media,
            self._music_media,
            self._now_playing_media,
        ):
            media = provider()
            if media:
                return media
        return None

    def accessories(self) -> list[AccessorySnapshot]:
        output = run(
            ["system_profiler", "SPBluetoothDataType", "-json"], timeout_seconds=8
        )
        if not output:
            return []
        try:
            payload = json.loads(output)
        except json.JSONDecodeError:
            return []
        accessories: list[AccessorySnapshot] = []
        controllers = (
            payload.get("SPBluetoothDataType", []) if isinstance(payload, dict) else []
        )
        for controller in controllers:
            connected = (
                controller.get("device_connected", [])
                if isinstance(controller, dict)
                else []
            )
            if isinstance(connected, list):
                accessories.extend(_connected_accessories(connected))
        return unique_accessories(accessories)

    def hardware_model(self) -> str:
        return sanitize_text(run(["sysctl", "-n", "hw.model"]))

    def _now_playing_media(self) -> MediaSnapshot | None:
        command = command_path("nowplaying-cli")
        if not command:
            return None
        playing = _decode_object(run([command, "get", "--json", *NOW_PLAYING_FIELDS]))
        if _read_float(playing.get("playbackRate")) <= 0:
            return None
        title = sanitize_text(str(playing.get("title") or ""))
        owner = sanitize_text(str(playing.get("clientBundleIdentifier") or ""))
        if not title or not owner or owner not in _playing_applications():
            return None
        return MediaSnapshot(title, sanitize_text(str(playing.get("artist") or "")))

    def _music_media(self) -> MediaSnapshot | None:
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
        return _media_from_apple_script(script)

    def _spotify_media(self) -> MediaSnapshot | None:
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
        return _media_from_apple_script(script)


def _connected_accessories(devices: list[object]) -> list[AccessorySnapshot]:
    accessories: list[AccessorySnapshot] = []
    for item in devices:
        if not isinstance(item, dict):
            continue
        for raw_name, details in item.items():
            if not isinstance(details, dict):
                continue
            name = sanitize_text(str(raw_name))
            minor_type = sanitize_text(str(details.get("device_minorType", "")))
            services = sanitize_text(str(details.get("device_services", "")))
            if not name or not is_audio_accessory(minor_type, services):
                continue
            accessories.append(
                AccessorySnapshot(
                    kind="bluetooth_audio",
                    name=name,
                    model=minor_type
                    if minor_type and minor_type.lower() != "headphones"
                    else "",
                    battery_percent=parse_percent(
                        str(details.get("device_batteryLevelMain", ""))
                    ),
                )
            )
    return accessories


def _playing_applications() -> set[str]:
    executables = _audio_out_executables(_audio_out_processes())
    return {bundle for bundle in map(_bundle_identifier, executables) if bundle}


def _audio_out_processes() -> set[int]:
    processes: set[int] = set()
    audio = False
    for line in run(["pmset", "-g", "assertions"]).splitlines():
        if ASSERTION_HEADER_PATTERN.match(line):
            audio = bool(AUDIO_ASSERTION_PATTERN.search(line))
            continue
        owner = ASSERTION_OWNER_PATTERN.search(line) if audio else None
        if owner:
            processes.add(int(owner.group(1)))
    return processes


def _audio_out_executables(processes: set[int]) -> list[Path]:
    if not processes:
        return []
    output = run(
        ["ps", "-p", ",".join(str(pid) for pid in sorted(processes)), "-o", "comm="]
    )
    return [Path(line.strip()) for line in output.splitlines() if line.strip()]


def _bundle_identifier(executable: Path) -> str:
    for directory in reversed(executable.parents):
        if directory.suffix != ".app":
            continue
        info = directory / "Contents" / "Info.plist"
        return run(
            ["plutil", "-extract", "CFBundleIdentifier", "raw", "-o", "-", str(info)]
        ).strip()
    return ""


def _media_from_apple_script(script: str) -> MediaSnapshot | None:
    osascript = command_path("osascript")
    if not osascript:
        return None
    parts = [
        sanitize_text(line) for line in run([osascript, "-e", script]).splitlines()
    ]
    parts = [part for part in parts if part]
    return (
        MediaSnapshot(parts[0], parts[1] if len(parts) > 1 else "") if parts else None
    )


def _decode_object(text: str) -> dict[str, object]:
    try:
        value = json.loads(text)
    except json.JSONDecodeError:
        return {}
    return value if isinstance(value, dict) else {}


def _read_float(value: object) -> float:
    if not isinstance(value, (int, float, str)):
        return 0
    try:
        return float(value)
    except (TypeError, ValueError):
        return 0
