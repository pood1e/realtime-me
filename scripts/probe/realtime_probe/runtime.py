"""Collector orchestration and probe self-observability."""

from __future__ import annotations

import logging
import platform
import threading
import time
from collections.abc import Sequence
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from typing import Protocol

from . import VERSION
from .metrics import MetricFamily, Sample, render


class Collector(Protocol):
    name: str

    def collect(self) -> Sequence[MetricFamily]: ...


@dataclass(frozen=True)
class CollectionResult:
    families: tuple[MetricFamily, ...]
    success: bool
    duration_seconds: float


class ProbeRuntime:
    def __init__(self, collectors: Sequence[Collector]) -> None:
        self._collectors = tuple(collectors)
        self._pool = ThreadPoolExecutor(
            max_workers=max(1, len(collectors)), thread_name_prefix="collector"
        )
        self._lock = threading.Lock()
        self._last_success: dict[str, bool | None] = {
            collector.name: None for collector in collectors
        }

    def render_metrics(self) -> str:
        started_at = time.monotonic()
        families: list[MetricFamily] = []
        successes: list[Sample] = []
        durations: list[Sample] = []
        results = self._pool.map(self._collect, self._collectors)
        latest: dict[str, bool] = {}
        for collector, result in zip(self._collectors, results, strict=True):
            families.extend(result.families)
            successes.append(Sample(int(result.success), {"collector": collector.name}))
            durations.append(
                Sample(
                    result.duration_seconds,
                    {"collector": collector.name},
                )
            )
            latest[collector.name] = result.success
        with self._lock:
            self._last_success.update(latest)

        families.extend(
            (
                MetricFamily(
                    "realtime_probe_info",
                    "Static probe build and platform information.",
                    (
                        Sample(
                            1,
                            {
                                "version": VERSION,
                                "os": platform.system().lower(),
                                "architecture": platform.machine().lower(),
                            },
                        ),
                    ),
                ),
                MetricFamily(
                    "realtime_probe_collector_success",
                    "Whether the collector completed its latest scrape.",
                    successes,
                ),
                MetricFamily(
                    "realtime_probe_collector_duration_seconds",
                    "Collector execution time for the latest scrape.",
                    durations,
                    unit="seconds",
                ),
                MetricFamily(
                    "realtime_probe_scrape_duration_seconds",
                    "Total probe collection time for the latest scrape.",
                    (Sample(time.monotonic() - started_at),),
                    unit="seconds",
                ),
            )
        )
        return render(families)

    @staticmethod
    def _collect(collector: Collector) -> CollectionResult:
        started_at = time.monotonic()
        try:
            families = tuple(collector.collect())
            success = True
        except Exception:  # One unavailable source must not fail the scrape.
            logging.exception("collector %s failed", collector.name)
            families = ()
            success = False
        return CollectionResult(families, success, time.monotonic() - started_at)

    def health(self) -> dict[str, object]:
        with self._lock:
            results = dict(self._last_success)
        collectors = {
            name: "pending" if success is None else "ok" if success else "error"
            for name, success in results.items()
        }
        return {
            "ok": all(success is not False for success in results.values()),
            "version": VERSION,
            "platform": platform.system().lower(),
            "collectors": collectors,
        }
