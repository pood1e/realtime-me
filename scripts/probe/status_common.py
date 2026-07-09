#!/usr/bin/env python3
"""Shared helpers for the realtime-me probe scripts.

The gateway speaks ConnectRPC (unary JSON over HTTP/1.1 POST) and owns every
device identity. This module centralises the Connect client, the backend-owned
device enrollment plus uid cache, the read-only exporter HTTP server, and the
Prometheus label rendering so the reporters stay small and consistent.
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import time
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from typing import Callable, NamedTuple

JSON_CONTENT_TYPE = "application/json; charset=utf-8"
METRICS_CONTENT_TYPE = "text/plain; version=0.0.4; charset=utf-8"

DEVICE_KIND_ENUMS = {
    "host": "DEVICE_KIND_HOST",
    "virtual_machine": "DEVICE_KIND_VIRTUAL_MACHINE",
    "phone": "DEVICE_KIND_PHONE",
    "watch": "DEVICE_KIND_WATCH",
}
DEVICE_ROLE_ENUMS = {
    "server": "DEVICE_ROLE_SERVER",
    "desktop": "DEVICE_ROLE_DESKTOP",
    "vm": "DEVICE_ROLE_VM",
}


def device_kind_enum(value: str) -> str:
    return DEVICE_KIND_ENUMS.get(value, "DEVICE_KIND_UNSPECIFIED")


def device_role_enum(value: str) -> str:
    return DEVICE_ROLE_ENUMS.get(value, "DEVICE_ROLE_UNSPECIFIED")


class ConnectError(Exception):
    """A ConnectRPC error carrying the machine-readable Connect code."""

    def __init__(self, code: str) -> None:
        super().__init__(code)
        self.code = code


def connect_post(base: str, service: str, method: str, token: str, message: dict, timeout_seconds: float = 5) -> dict:
    """Call a ConnectRPC unary method and return the decoded response message.

    Raises ConnectError with the Connect error code on a 4xx/5xx response and
    OSError on a transport failure, so callers can surface a clean status
    without leaking gateway internals.
    """
    endpoint = f"{base.rstrip('/')}/realtime.me.v1.{service}/{method}"
    request = urllib.request.Request(
        endpoint,
        data=json.dumps(message).encode(),
        method="POST",
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {token}",
            "Connect-Protocol-Version": "1",
            "Content-Type": JSON_CONTENT_TYPE,
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            body = response.read().decode()
    except urllib.error.HTTPError as error:
        raise ConnectError(connect_error_code(error)) from error
    return json.loads(body) if body.strip() else {}


def connect_error_code(error: urllib.error.HTTPError) -> str:
    try:
        payload = json.loads(error.read().decode())
    except (OSError, ValueError):
        payload = {}
    code = payload.get("code") if isinstance(payload, dict) else None
    return str(code) if code else f"http_{error.code}"


def ensure_device_uid(
    base: str,
    token: str,
    identity_file: Path,
    kind: str,
    role: str,
    display_name: str,
    model: str,
    timeout_seconds: float = 5,
) -> str:
    """Return this device's gateway-owned uid, enrolling once when needed.

    The uid is minted by the gateway and cached locally; only the first run
    without a cached uid enrolls. Serving exporters that hold no ingest token
    simply reuse the cached uid.
    """
    cached = read_cached_uid(identity_file)
    if cached:
        return cached
    if not token:
        return ""
    response = connect_post(
        base,
        "EnrollmentService",
        "EnrollDevice",
        token,
        {"kind": kind, "role": role, "displayName": display_name, "model": model},
        timeout_seconds,
    )
    uid = str(response.get("deviceUid") or "")
    if uid:
        write_cached_uid(identity_file, uid)
    return uid


def read_cached_uid(identity_file: Path) -> str:
    try:
        data = json.loads(identity_file.read_text())
    except (OSError, ValueError):
        return ""
    return str(data.get("deviceUid") or "") if isinstance(data, dict) else ""


def write_cached_uid(identity_file: Path, uid: str) -> None:
    identity_file.parent.mkdir(parents=True, exist_ok=True)
    temporary = identity_file.with_name(f".{identity_file.name}.tmp")
    temporary.write_text(json.dumps({"deviceUid": uid}))
    os.chmod(temporary, 0o600)
    os.replace(temporary, identity_file)


class Response(NamedTuple):
    body: bytes
    content_type: str
    status: int = 200


def json_response(payload: object, status: int = 200) -> Response:
    return Response(json.dumps(payload, ensure_ascii=False).encode(), JSON_CONTENT_TYPE, status)


def text_response(text: str, status: int = 200) -> Response:
    return Response(text.encode(), METRICS_CONTENT_TYPE, status)


def run_server(bind: str, port: int, routes: dict[str, Callable[[], Response]]) -> int:
    """Serve the read-only exporter routes until interrupted."""

    class Handler(BaseHTTPRequestHandler):
        def do_GET(self) -> None:
            route = routes.get(self.path)
            self.write_response(route() if route else json_response({"error": "not_found"}, 404))

        def write_response(self, response: Response) -> None:
            self.send_response(response.status)
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Type", response.content_type)
            self.send_header("Content-Length", str(len(response.body)))
            self.end_headers()
            self.wfile.write(response.body)

        def log_message(self, _format: str, *_args: object) -> None:
            return

    server = ThreadingHTTPServer((bind, port), Handler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 130
    return 0


def label_set(labels: dict[str, str]) -> str:
    pairs = []
    for key in sorted(labels):
        value = labels[key]
        if value:
            pairs.append(f'{prometheus_label_name(key)}="{escape_label_value(value)}"')
    return "{" + ",".join(pairs) + "}" if pairs else ""


def prometheus_label_name(value: str) -> str:
    name = re.sub(r"[^A-Za-z0-9_]", "_", value)
    if not name:
        return "label"
    if name[0].isdigit():
        return f"label_{name}"
    return name


def escape_label_value(value: str) -> str:
    return value.replace("\\", "\\\\").replace("\n", "\\n").replace('"', '\\"')


def run(command: list[str], timeout_seconds: float = 5) -> str:
    try:
        return subprocess.check_output(command, text=True, stderr=subprocess.DEVNULL, timeout=timeout_seconds)
    except (OSError, subprocess.SubprocessError):
        return ""


def utc_now() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
