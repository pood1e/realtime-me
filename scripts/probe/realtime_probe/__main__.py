"""Realtime Me probe command-line entry point."""

from __future__ import annotations

import argparse
import logging
import os
from pathlib import Path

from . import VERSION
from .agents import AgentCollector
from .device import DeviceCollector
from .platforms import device_adapter
from .runtime import ProbeRuntime
from .server import cached, json_response, metrics_response, run_server
from .system import SystemCollector


def main() -> int:
    args = parse_args()
    logging.basicConfig(
        level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s"
    )
    adapter = device_adapter()
    runtime = ProbeRuntime(
        (
            SystemCollector(adapter),
            DeviceCollector(adapter),
            AgentCollector(
                codex_homes=parse_paths(args.codex_homes),
                claude_home=expand_path(args.claude_home),
                active_window_seconds=args.active_window_seconds,
            ),
        )
    )
    routes = {
        "/healthz": lambda: json_response(runtime.health()),
        "/metrics": cached(lambda: metrics_response(runtime.render_metrics())),
    }
    return run_server(args.bind, args.port, routes)


def parse_args() -> argparse.Namespace:
    default_codex_homes = os.pathsep.join(("~/.codex-api", "~/.codex"))
    parser = argparse.ArgumentParser(
        description="Serve cross-platform host, device, and coding-agent metrics."
    )
    parser.add_argument("--version", action="version", version=VERSION)
    parser.add_argument("--bind", default=os.getenv("REALTIME_PROBE_BIND", "127.0.0.1"))
    parser.add_argument(
        "--port",
        type=valid_port,
        default=os.getenv("REALTIME_PROBE_PORT", "18082"),
    )
    parser.add_argument(
        "--active-window-seconds",
        type=positive_integer,
        default=os.getenv("REALTIME_PROBE_ACTIVE_WINDOW_SECONDS", "300"),
    )
    parser.add_argument(
        "--codex-homes", default=os.getenv("REALTIME_CODEX_HOMES", default_codex_homes)
    )
    parser.add_argument(
        "--claude-home", default=os.getenv("REALTIME_CLAUDE_HOME", "~/.claude")
    )
    return parser.parse_args()


def valid_port(value: str) -> int:
    port = int(value)
    if not 1 <= port <= 65535:
        raise argparse.ArgumentTypeError("port must be between 1 and 65535")
    return port


def positive_integer(value: str) -> int:
    number = int(value)
    if number <= 0:
        raise argparse.ArgumentTypeError("value must be positive")
    return number


def parse_paths(value: str) -> list[Path]:
    values = value.split(os.pathsep)
    return [expand_path(item) for item in values if item.strip()]


def expand_path(value: str) -> Path:
    return Path(value.strip()).expanduser().resolve()


if __name__ == "__main__":
    raise SystemExit(main())
