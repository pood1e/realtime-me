"""Operating-system adapter selection."""

from __future__ import annotations

import platform

from .base import DeviceAdapter
from .darwin import DarwinDeviceAdapter
from .linux import LinuxDeviceAdapter
from .windows import WindowsDeviceAdapter


def device_adapter() -> DeviceAdapter:
    adapters: dict[str, type[DeviceAdapter]] = {
        "darwin": DarwinDeviceAdapter,
        "linux": LinuxDeviceAdapter,
        "windows": WindowsDeviceAdapter,
    }
    adapter = adapters.get(platform.system().lower())
    if adapter is None:
        raise RuntimeError(f"unsupported operating system: {platform.system()}")
    return adapter()
