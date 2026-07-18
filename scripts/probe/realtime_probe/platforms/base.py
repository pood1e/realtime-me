"""Shared device-adapter models and process helpers."""

from __future__ import annotations

import os
import re
import shutil
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Protocol


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


class DeviceAdapter(Protocol):
    def current_media(self) -> MediaSnapshot | None: ...

    def accessories(self) -> list[AccessorySnapshot]: ...

    def hardware_model(self) -> str: ...


def run(command: list[str], timeout_seconds: float = 5) -> str:
    try:
        return subprocess.check_output(
            command,
            text=True,
            stderr=subprocess.DEVNULL,
            timeout=timeout_seconds,
            encoding="utf-8",
            errors="replace",
        )
    except (OSError, subprocess.SubprocessError):
        return ""


def command_path(name: str) -> str | None:
    path = shutil.which(name)
    if path:
        return path
    for directory in ("/opt/homebrew/bin", "/usr/local/bin"):
        candidate = Path(directory) / name
        if candidate.exists() and os.access(candidate, os.X_OK):
            return str(candidate)
    return None


def sanitize_text(value: str) -> str:
    text = re.sub(r"[\x00-\x1f\x7f]", " ", value)
    text = re.sub(r"\s+", " ", text).strip()
    if text.lower() == "null":
        return ""
    return text if len(text) <= 120 else text[:119].rstrip() + "…"


def parse_percent(value: str) -> int | None:
    match = re.search(r"(\d{1,3})\s*%", value)
    return clamp_percent(int(match.group(1))) if match else None


def clamp_percent(value: int) -> int:
    return max(0, min(100, value))


def is_audio_accessory(*values: str) -> bool:
    text = " ".join(values).lower()
    return any(
        token in text
        for token in ("headphone", "headset", "earbud", "a2dp", "hfp", "avrcp", "audio")
    )


def unique_accessories(accessories: list[AccessorySnapshot]) -> list[AccessorySnapshot]:
    by_key: dict[tuple[str, str, str], AccessorySnapshot] = {}
    for accessory in accessories:
        key = (accessory.kind, accessory.name, accessory.model)
        existing = by_key.get(key)
        if existing is None or existing.battery_percent is None:
            by_key[key] = accessory
    return sorted(by_key.values(), key=lambda item: (item.kind, item.name, item.model))
