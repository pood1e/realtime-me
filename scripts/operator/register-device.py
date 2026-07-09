#!/usr/bin/env python3
"""Register a device's exporters with the gateway from the operator side.

Probe hosts stay unaware of the gateway: they only run exporters. This tool, run
wherever you can reach the gateway, mints the device's backend-owned uid and
registers its scrape targets so Prometheus discovers and pulls them. Prometheus
stamps the uid onto every series through the service-discovery target labels, so
the exporters themselves need no identity, token, or gateway address.

Example:
  STATUS_INGEST_TOKEN=... python3 scripts/operator/register-device.py \
    --url http://<gateway-host>:18080 --host <probe-host> --name "Studio Mac" --kind host
"""
from __future__ import annotations

import argparse
import os
import re
import sys
from pathlib import Path

# The probe payload ships as flat siblings into a single install dir, so
# status_common cannot live in a package shared with this operator-side tool.
sys.path.insert(0, str(Path(__file__).resolve().parent.parent / "probe"))

import status_common
from status_common import ConnectError, device_kind_enum, device_role_enum, ensure_device_uid


def main() -> int:
    args = parse_args()
    token = (os.getenv("STATUS_INGEST_TOKEN", "") or args.token).strip()
    if not token:
        print("Set STATUS_INGEST_TOKEN or pass --token.", file=sys.stderr)
        return 2

    kind = device_kind_enum(args.kind)
    role = device_role_enum(args.role)
    identity_file = resolve_identity_file(args)
    try:
        device_uid = ensure_device_uid(
            args.url, token, identity_file, kind, role, args.name, args.model, args.timeout_seconds
        )
    except (ConnectError, OSError) as error:
        print(f"enrollment failed: {error}", file=sys.stderr)
        return 1
    if not device_uid:
        print("enrollment did not return a device uid", file=sys.stderr)
        return 1

    common = {
        "deviceUid": device_uid,
        "displayName": args.name,
        "model": args.model,
        "kind": kind,
        "role": role,
    }
    node_job = (
        "SCRAPE_JOB_VM_NODE_EXPORTER"
        if args.kind == "virtual_machine" or args.role == "vm"
        else "SCRAPE_JOB_NODE_EXPORTER"
    )
    targets = [dict(common, job=node_job, target=f"{args.host}:{args.node_port}")]
    if not args.no_device:
        targets.append(dict(common, job="SCRAPE_JOB_DEVICE_EXPORTER", target=f"{args.host}:{args.device_port}"))
    if args.install_agent:
        targets.append(dict(common, job="SCRAPE_JOB_AGENT_EXPORTER", target=f"{args.host}:{args.agent_port}"))

    try:
        status_common.connect_post(
            args.url, "IngestService", "RegisterScrapeTargets", token, {"targets": targets}, args.timeout_seconds
        )
    except ConnectError as error:
        print(f"target registration failed: {error.code}", file=sys.stderr)
        return 1

    print(f"registered {args.name} ({device_uid}) at {args.host}", file=sys.stderr)
    print(device_uid)
    return 0


def resolve_identity_file(args: argparse.Namespace) -> Path:
    if args.identity_file:
        return Path(args.identity_file).expanduser()
    slug = re.sub(r"[^A-Za-z0-9_.-]+", "-", args.name).strip("-").lower() or "device"
    return Path.home() / ".realtime-me" / "devices" / f"{slug}.json"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Register a device's exporters with the realtime-me gateway.")
    parser.add_argument("--url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--token", default="", help="Ingest token (or set STATUS_INGEST_TOKEN).")
    parser.add_argument("--host", required=True, help="Device LAN address Prometheus should scrape.")
    parser.add_argument("--name", required=True, help="Human-readable device name.")
    parser.add_argument("--model", default="")
    parser.add_argument("--kind", default="host")
    parser.add_argument("--role", default="desktop")
    parser.add_argument("--node-port", type=int, default=int(os.getenv("STATUS_NODE_EXPORTER_PORT", "9100")))
    parser.add_argument("--device-port", type=int, default=int(os.getenv("STATUS_DEVICE_EXPORTER_PORT", "18083")))
    parser.add_argument("--agent-port", type=int, default=int(os.getenv("STATUS_AGENT_EXPORTER_PORT", "18082")))
    parser.add_argument("--install-agent", action="store_true", help="Also register the agent exporter target.")
    parser.add_argument("--no-device", action="store_true", help="Skip the device exporter target (headless hosts with no media/Bluetooth).")
    parser.add_argument(
        "--identity-file",
        default=os.getenv("STATUS_IDENTITY_FILE", ""),
        help="Where the device's minted uid is cached so re-registration is stable "
        "(default ~/.realtime-me/devices/<name>.json).",
    )
    parser.add_argument("--timeout-seconds", type=float, default=5)
    return parser.parse_args()


if __name__ == "__main__":
    raise SystemExit(main())
