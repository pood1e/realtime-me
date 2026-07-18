"""Portable process inspection backed by psutil."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

import psutil


@dataclass(frozen=True)
class ProcessSnapshot:
    pid: int
    command_line: str
    create_time: float


def running_processes() -> dict[int, ProcessSnapshot]:
    snapshots: dict[int, ProcessSnapshot] = {}
    attributes = ("pid", "name", "cmdline", "create_time")
    for process in psutil.process_iter(attributes):
        try:
            command = " ".join(process.info.get("cmdline") or ()) or str(
                process.info.get("name") or ""
            )
            lowered = command.lower()
            if "realtime_probe" in lowered:
                continue
            snapshot = ProcessSnapshot(
                pid=int(process.info["pid"]),
                command_line=lowered,
                create_time=float(process.info.get("create_time") or 0),
            )
            snapshots[snapshot.pid] = snapshot
        except (psutil.Error, TypeError, ValueError):
            continue
    return snapshots


def open_files(process_id: int) -> list[Path]:
    try:
        return [Path(item.path) for item in psutil.Process(process_id).open_files()]
    except (psutil.Error, OSError):
        return []
