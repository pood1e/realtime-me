#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from typing import Any


def main() -> int:
    args = parse_args()
    token = args.token or os.getenv("STATUS_INGEST_TOKEN")
    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2

    ok = True
    if args.device_url:
        device = fetch_json(args.device_url, args.timeout_seconds)
        ok = post(args.gateway_url, "/api/ingest/host", token, device, args.timeout_seconds) and ok
    if args.agent_url:
        agents = fetch_json(args.agent_url, args.timeout_seconds)
        if not isinstance(agents, list):
            print("agent exporter returned invalid payload", file=sys.stderr)
            ok = False
        else:
            for agent in agents:
                ok = post(args.gateway_url, "/api/ingest/agent", token, agent, args.timeout_seconds) and ok
    return 0 if ok else 1


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Forward local status exporter JSON to realtime-me gateway.")
    parser.add_argument("--gateway-url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--token", default="")
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


def post(gateway_url: str, path: str, token: str, payload: Any, timeout_seconds: float) -> bool:
    if not isinstance(payload, dict):
        print("status exporter returned invalid payload", file=sys.stderr)
        return False
    request = urllib.request.Request(
        gateway_url.rstrip("/") + path,
        data=json.dumps(payload, ensure_ascii=False).encode(),
        method="POST",
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json; charset=utf-8",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            if response.status < 200 or response.status > 299:
                print(f"gateway rejected forwarded status: HTTP {response.status}", file=sys.stderr)
                return False
    except urllib.error.HTTPError as error:
        print(f"gateway rejected forwarded status: HTTP {error.code}", file=sys.stderr)
        return False
    except OSError as error:
        print(f"gateway forward failed: {error.__class__.__name__}", file=sys.stderr)
        return False
    return True


if __name__ == "__main__":
    raise SystemExit(main())
