"""Canonical system metrics shared by Linux, macOS, and Windows."""

from __future__ import annotations

import platform
from pathlib import Path

import psutil

from .metrics import MetricFamily, Sample
from .platforms.base import DeviceAdapter


class SystemCollector:
    name = "system"

    def __init__(self, adapter: DeviceAdapter) -> None:
        self._adapter = adapter
        # psutil calculates a non-blocking percentage from the previous call.
        # Prime its baseline at startup so the first scrape covers real time.
        psutil.cpu_percent(interval=None)

    def collect(self) -> tuple[MetricFamily, ...]:
        memory = psutil.virtual_memory()
        root = Path.home().anchor if platform.system().lower() == "windows" else "/"
        disk = psutil.disk_usage(root)
        labels = {
            "os": platform.system().lower(),
            "os_version": platform.release(),
            "architecture": platform.machine().lower(),
            "hostname": platform.node(),
            "model": self._adapter.hardware_model(),
        }
        return (
            MetricFamily(
                "realtime_system_info",
                "Host operating-system and hardware information.",
                (Sample(1, labels),),
            ),
            MetricFamily(
                "realtime_system_cpu_logical_count",
                "Logical CPU count. OpenTelemetry name: system.cpu.logical.count.",
                (Sample(psutil.cpu_count(logical=True) or 0),),
                unit="{cpu}",
            ),
            MetricFamily(
                "realtime_system_cpu_utilization_ratio",
                "Whole-system CPU utilization ratio. OpenTelemetry name: system.cpu.utilization.",
                (Sample(psutil.cpu_percent(interval=None) / 100),),
                unit="1",
            ),
            MetricFamily(
                "realtime_system_memory_total_bytes",
                "Physical memory limit. OpenTelemetry name: system.memory.limit.",
                (Sample(memory.total),),
                unit="bytes",
            ),
            MetricFamily(
                "realtime_system_memory_available_bytes",
                "Physical memory immediately available to applications.",
                (Sample(memory.available),),
                unit="bytes",
            ),
            MetricFamily(
                "realtime_system_filesystem_total_bytes",
                "Root filesystem limit. OpenTelemetry name: system.filesystem.limit.",
                (Sample(disk.total, {"mountpoint": root}),),
                unit="bytes",
            ),
            MetricFamily(
                "realtime_system_filesystem_available_bytes",
                "Root filesystem space available to the current user.",
                (Sample(disk.free, {"mountpoint": root}),),
                unit="bytes",
            ),
        )
