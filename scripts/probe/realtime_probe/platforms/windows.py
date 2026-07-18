"""Windows device adapter.

System and coding-agent collection are fully portable through psutil. Windows
does not expose global media sessions or Bluetooth battery state through a
stable command-line contract, so those optional signals remain empty instead of
requiring an unmaintained native bridge.
"""

from __future__ import annotations

from .base import AccessorySnapshot, MediaSnapshot, run, sanitize_text


class WindowsDeviceAdapter:
    def current_media(self) -> MediaSnapshot | None:
        return None

    def accessories(self) -> list[AccessorySnapshot]:
        return []

    def hardware_model(self) -> str:
        command = [
            "powershell.exe",
            "-NoLogo",
            "-NoProfile",
            "-NonInteractive",
            "-Command",
            "(Get-CimInstance -ClassName Win32_ComputerSystem).Model",
        ]
        return sanitize_text(run(command, timeout_seconds=8))
