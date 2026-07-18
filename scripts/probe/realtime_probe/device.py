"""Media and connected-accessory metrics."""

from __future__ import annotations

from .metrics import MetricFamily, Sample
from .platforms.base import AccessorySnapshot, DeviceAdapter


class DeviceCollector:
    name = "device"

    def __init__(self, adapter: DeviceAdapter) -> None:
        self._adapter = adapter

    def collect(self) -> tuple[MetricFamily, ...]:
        media = self._adapter.current_media()
        accessories = self._adapter.accessories()
        media_samples = (
            ()
            if media is None
            else (Sample(1, {"title": media.title, "artist": media.artist}),)
        )
        connected = tuple(Sample(1, _labels(item)) for item in accessories)
        batteries = tuple(
            Sample(item.battery_percent / 100, _labels(item))
            for item in accessories
            if item.battery_percent is not None
        )
        return (
            MetricFamily(
                "realtime_device_media_playing",
                "Device media playback state. OpenTelemetry name: realtime.device.media.playing.",
                media_samples,
                unit="1",
            ),
            MetricFamily(
                "realtime_device_accessory_connected",
                "Connected accessory state labelled by name and kind. "
                "OpenTelemetry name: realtime.device.accessory.connected.",
                connected,
                unit="1",
            ),
            MetricFamily(
                "realtime_device_accessory_battery_level_ratio",
                "Accessory battery fraction. OpenTelemetry name: realtime.device.accessory.battery.level.",
                batteries,
                unit="1",
            ),
        )


def _labels(accessory: AccessorySnapshot) -> dict[str, str]:
    return {
        "accessory_kind": accessory.kind,
        "accessory_name": accessory.name,
        "accessory_model": accessory.model,
    }
