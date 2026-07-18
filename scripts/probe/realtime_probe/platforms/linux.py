"""Linux media, Bluetooth, and hardware adapters."""

from __future__ import annotations

import re
from pathlib import Path

from .base import (
    AccessorySnapshot,
    MediaSnapshot,
    clamp_percent,
    command_path,
    is_audio_accessory,
    parse_percent,
    run,
    sanitize_text,
    unique_accessories,
)


class LinuxDeviceAdapter:
    def current_media(self) -> MediaSnapshot | None:
        playerctl = command_path("playerctl")
        if not playerctl:
            return None
        players = [
            line.strip()
            for line in run([playerctl, "--list-all"]).splitlines()
            if line.strip()
        ]
        for player in players:
            if (
                run([playerctl, "--player", player, "status"]).strip().lower()
                != "playing"
            ):
                continue
            title = run([playerctl, "--player", player, "metadata", "title"]).strip()
            if not title:
                continue
            artists = (
                run([playerctl, "--player", player, "metadata", "artist"]).strip(),
                run(
                    [playerctl, "--player", player, "metadata", "xesam:artist"]
                ).strip(),
            )
            return MediaSnapshot(
                sanitize_text(title),
                sanitize_text(next((item for item in artists if item), "")),
            )
        return None

    def accessories(self) -> list[AccessorySnapshot]:
        bluetoothctl = command_path("bluetoothctl")
        if not bluetoothctl:
            return []
        devices = _parse_devices(run([bluetoothctl, "devices", "Connected"]))
        if not devices:
            devices = _parse_devices(run([bluetoothctl, "devices"]))
        accessories: list[AccessorySnapshot] = []
        for address, fallback_name in devices:
            info = run([bluetoothctl, "info", address])
            if not info or _field(info, "Connected").lower() not in {"yes", "true"}:
                continue
            name = sanitize_text(_field(info, "Name") or fallback_name)
            icon = _field(info, "Icon")
            uuids = " ".join(re.findall(r"UUID: (.+)", info))
            if not name or not is_audio_accessory(icon, uuids):
                continue
            alias = sanitize_text(_field(info, "Alias"))
            accessories.append(
                AccessorySnapshot(
                    kind="bluetooth_audio",
                    name=name,
                    model=alias if alias != name else "",
                    battery_percent=_battery(info),
                )
            )
        return unique_accessories(accessories)

    def hardware_model(self) -> str:
        for path in (
            Path("/sys/devices/virtual/dmi/id/product_name"),
            Path("/sys/firmware/devicetree/base/model"),
        ):
            try:
                model = sanitize_text(path.read_text(errors="replace"))
            except OSError:
                continue
            if model:
                return model
        return ""


def _parse_devices(output: str) -> list[tuple[str, str]]:
    devices: list[tuple[str, str]] = []
    for line in output.splitlines():
        match = re.match(r"Device\s+([0-9A-Fa-f:]{17})\s+(.+)", line.strip())
        if match:
            devices.append((match.group(1), sanitize_text(match.group(2))))
    return devices


def _field(info: str, field: str) -> str:
    match = re.search(rf"^\s*{re.escape(field)}:\s*(.+?)\s*$", info, re.MULTILINE)
    return sanitize_text(match.group(1)) if match else ""


def _battery(info: str) -> int | None:
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
