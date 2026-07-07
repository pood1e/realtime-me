#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from typing import Any

from status_common import ConnectError, connect_post


def main() -> int:
    args = parse_args()
    token = os.getenv("STATUS_INGEST_TOKEN", "").strip()
    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2

    ok = True
    if args.device_url:
        device = fetch_json(args.device_url, args.timeout_seconds)
        ok = forward_device(args.gateway_url, token, device, args.timeout_seconds) and ok
    if args.agent_url:
        agents = fetch_json(args.agent_url, args.timeout_seconds)
        if not isinstance(agents, list):
            print("agent exporter returned invalid payload", file=sys.stderr)
            ok = False
        else:
            for agent in agents:
                ok = forward_agent(args.gateway_url, token, agent, args.timeout_seconds) and ok
    return 0 if ok else 1


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Forward local status exporter JSON to realtime-me gateway.")
    parser.add_argument("--gateway-url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--device-url", default=os.getenv("STATUS_DEVICE_SOURCE_URL", ""))
    parser.add_argument("--agent-url", default=os.getenv("STATUS_AGENT_SOURCE_URL", ""))
    parser.add_argument("--timeout-seconds", type=float, default=float(os.getenv("STATUS_FORWARDER_TIMEOUT_SECONDS", "5")))
    return parser.parse_args()


def fetch_json(url: str, timeout_seconds: float) -> Any:
    request = urllib.request.Request(url, headers={"Accept": "application/json"})
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            return json.loads(response.read().decode())
    except (OSError, json.JSONDecodeError, urllib.error.HTTPError) as error:
        print(f"status exporter fetch failed: {error.__class__.__name__}", file=sys.stderr)
        return None


def forward_device(gateway_url: str, token: str, device: Any, timeout_seconds: float) -> bool:
    if not isinstance(device, dict):
        print("device exporter returned invalid payload", file=sys.stderr)
        return False
    return forward(gateway_url, "ReportDeviceStatus", token, {"device": device}, timeout_seconds)


def forward_agent(gateway_url: str, token: str, agent: Any, timeout_seconds: float) -> bool:
    if not isinstance(agent, dict):
        print("agent exporter returned invalid payload", file=sys.stderr)
        return False
    return forward(gateway_url, "ReportAgentStatus", token, agent, timeout_seconds)


def forward(gateway_url: str, method: str, token: str, message: dict, timeout_seconds: float) -> bool:
    try:
        connect_post(gateway_url, "IngestService", method, token, message, timeout_seconds)
    except ConnectError as error:
        print(f"gateway rejected forwarded status: {error.code}", file=sys.stderr)
        return False
    except OSError as error:
        print(f"gateway forward failed: {error.__class__.__name__}", file=sys.stderr)
        return False
    return True


if __name__ == "__main__":
    raise SystemExit(main())
