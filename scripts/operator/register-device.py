#!/usr/bin/env python3
"""Register a device's probe with the gateway from the operator side.

Probe hosts stay unaware of the gateway: they only run a read-only endpoint. This tool, run
wherever you can reach the gateway, mints the device's backend-owned uid and
registers its scrape target so Prometheus discovers and pulls it. Prometheus
stamps the uid onto every series through the service-discovery target labels, so
the probe needs no identity, token, or gateway address.

Example:
  STATUS_INGEST_TOKEN=... python3 scripts/operator/register-device.py \
    --url http://<gateway-host>:18080 --host <probe-host> \
    --name "Studio Mac" --kind host
"""

from __future__ import annotations

import argparse
import math
import os
import re
import sys
from pathlib import Path

from status_client import ConnectError, connect_post, ensure_device_uid


def main() -> int:
    args = parse_args()
    token = (os.getenv("STATUS_INGEST_TOKEN", "") or args.token).strip()
    if not token:
        print("Set STATUS_INGEST_TOKEN or pass --token.", file=sys.stderr)
        return 2

    identity_file = resolve_identity_file(args)
    try:
        device_uid = enroll(args, token, identity_file)
        try:
            register_target(args, token, device_uid)
        except ConnectError as error:
            if error.code != "not_found":
                raise
            identity_file.unlink(missing_ok=True)
            device_uid = enroll(args, token, identity_file)
            register_target(args, token, device_uid)
    except (ConnectError, OSError, RuntimeError) as error:
        print(f"registration failed: {error}", file=sys.stderr)
        return 1

    print(f"registered {args.name} ({device_uid}) at {args.host}", file=sys.stderr)
    print(device_uid)
    return 0


def enroll(args: argparse.Namespace, token: str, identity_file: Path) -> str:
    device_uid = ensure_device_uid(
        args.url,
        token,
        identity_file,
        args.kind,
        args.role,
        args.name,
        args.model,
        args.timeout_seconds,
    )
    if not device_uid:
        raise RuntimeError("enrollment did not return a device uid")
    return device_uid


def register_target(args: argparse.Namespace, token: str, device_uid: str) -> None:
    # One endpoint exposes the canonical system, device, and coding-agent
    # metrics on every desktop OS. The complete-set RPC removes every retired
    # multi-exporter target owned by this device.
    target = {
        "job": "SCRAPE_JOB_PROBE",
        "target": host_port(args.host, args.probe_port),
    }
    connect_post(
        args.url,
        "IngestService",
        "RegisterScrapeTargets",
        token,
        {"deviceUid": device_uid, "targets": [target]},
        args.timeout_seconds,
    )


def resolve_identity_file(args: argparse.Namespace) -> Path:
    if args.identity_file:
        return Path(args.identity_file).expanduser()
    slug = re.sub(r"[^A-Za-z0-9_.-]+", "-", args.name).strip("-").lower() or "device"
    return Path.home() / ".realtime-me" / "devices" / f"{slug}.json"


def host_port(host: str, port: int) -> str:
    value = host.strip()
    if ":" in value and not value.startswith("["):
        value = f"[{value}]"
    return f"{value}:{port}"


def valid_port(value: str) -> int:
    number = int(value)
    if not 1 <= number <= 65535:
        raise argparse.ArgumentTypeError("port must be between 1 and 65535")
    return number


def positive_float(value: str) -> float:
    number = float(value)
    if not math.isfinite(number) or number <= 0:
        raise argparse.ArgumentTypeError("value must be a positive finite number")
    return number


def non_empty(value: str) -> str:
    normalized = value.strip()
    if not normalized:
        raise argparse.ArgumentTypeError("value must not be empty")
    return normalized


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Register a device's unified probe with the realtime-me gateway."
    )
    parser.add_argument(
        "--url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080")
    )
    parser.add_argument(
        "--token", default="", help="Ingest token (or set STATUS_INGEST_TOKEN)."
    )
    parser.add_argument(
        "--host",
        type=non_empty,
        required=True,
        help="Device LAN address Prometheus should scrape.",
    )
    parser.add_argument(
        "--name", type=non_empty, required=True, help="Human-readable device name."
    )
    parser.add_argument("--model", default="")
    parser.add_argument("--kind", choices=("host", "virtual_machine"), default="host")
    parser.add_argument(
        "--role", choices=("server", "desktop", "vm"), default="desktop"
    )
    parser.add_argument(
        "--probe-port",
        type=valid_port,
        default=os.getenv("REALTIME_PROBE_PORT", "18082"),
    )
    parser.add_argument(
        "--identity-file",
        default=os.getenv("STATUS_IDENTITY_FILE", ""),
        help="Where the device's minted uid is cached so re-registration is stable "
        "(default ~/.realtime-me/devices/<name>.json).",
    )
    parser.add_argument("--timeout-seconds", type=positive_float, default="5")
    return parser.parse_args()


if __name__ == "__main__":
    raise SystemExit(main())
