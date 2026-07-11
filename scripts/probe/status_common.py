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
import threading
import time
import traceback
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from typing import Any, Callable, NamedTuple

JSON_CONTENT_TYPE = "application/json; charset=utf-8"
METRICS_CONTENT_TYPE = "text/plain; version=0.0.4; charset=utf-8"

# Rendering a scrape shells out to ps, lsof, bluetoothctl and friends. Serving it
# from a short-lived cache keeps the scrape rate from becoming the process spawn
# rate. The TTL stays well under Prometheus's 15s scrape interval, so a scrape
# never reads a sample belonging to the previous one.
METRICS_CACHE_TTL_SECONDS = 5.0

# A thread per connection lets any LAN client turn concurrent GETs into unbounded
# threads. A fixed pool turns them into a queue instead.
WORKER_THREADS = 8
REQUEST_TIMEOUT_SECONDS = 10

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


def cached(render: Callable[[], Response], ttl_seconds: float = METRICS_CACHE_TTL_SECONDS) -> Callable[[], Response]:
    """Serve one render to every caller within the TTL, and render once at a time."""
    lock = threading.Lock()
    state: dict[str, object] = {"expires_at": 0.0, "response": None}

    def serve() -> Response:
        with lock:
            response = state["response"]
            if response is not None and time.monotonic() < state["expires_at"]:
                return response  # type: ignore[return-value]
            fresh = render()
            state["response"] = fresh
            state["expires_at"] = time.monotonic() + ttl_seconds
            return fresh

    return serve


class PooledHTTPServer(HTTPServer):
    """Serves every request from a fixed thread pool, so a burst of connections
    becomes a queue rather than a thread apiece."""

    request_queue_size = 32

    def __init__(self, address: tuple[str, int], handler: type[BaseHTTPRequestHandler], workers: int) -> None:
        super().__init__(address, handler)
        self.pool = ThreadPoolExecutor(max_workers=workers, thread_name_prefix="probe")

    def process_request(self, request: object, client_address: object) -> None:
        self.pool.submit(self.handle_request_in_pool, request, client_address)

    def handle_request_in_pool(self, request: object, client_address: object) -> None:
        # Nobody waits on the future this runs in, so an exception it does not
        # catch is discarded: the connection would close unanswered and the host
        # would hold the only record of why.
        try:
            self.finish_request(request, client_address)
        except Exception:  # noqa: BLE001 - the pool would otherwise swallow it
            self.handle_error(request, client_address)
        finally:
            self.shutdown_request(request)

    def server_close(self) -> None:
        super().server_close()
        self.pool.shutdown(wait=False)


def render_route(route: Callable[[], Response]) -> Response:
    """Answer a route, turning a render that raised into a scrape that says so.

    A scrape reads files, databases and the output of other programs, any of
    which can surprise it. Left uncaught the exception would close the socket
    unanswered, and Prometheus would report a target that is down without ever
    saying why; the traceback goes to the journal, and the scrape gets a 500.
    """
    try:
        return route()
    except Exception:  # noqa: BLE001 - one bad file must not silence the scrape
        traceback.print_exc()
        return json_response({"error": "render_failed"}, 500)


def run_server(bind: str, port: int, routes: dict[str, Callable[[], Response]]) -> int:
    """Serve the read-only exporter routes until interrupted."""

    class Handler(BaseHTTPRequestHandler):
        timeout = REQUEST_TIMEOUT_SECONDS

        def do_GET(self) -> None:
            route = routes.get(self.path)
            if route is None:
                self.write_response(json_response({"error": "not_found"}, 404))
                return
            self.write_response(render_route(route))

        def write_response(self, response: Response) -> None:
            self.send_response(response.status)
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Type", response.content_type)
            self.send_header("Content-Length", str(len(response.body)))
            self.end_headers()
            self.wfile.write(response.body)

        def log_message(self, _format: str, *_args: object) -> None:
            return

    server = PooledHTTPServer((bind, port), Handler, WORKER_THREADS)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 130
    finally:
        server.server_close()
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


def decode_json(text: str) -> dict[str, Any]:
    """The object this text holds, or nothing at all if it holds anything else.

    A probe reads JSON that another program wrote -- a transcript line, the answer
    of a command -- and a scrape must survive every one of them being malformed.
    """
    try:
        data = json.loads(text)
    except json.JSONDecodeError:
        return {}
    return data if isinstance(data, dict) else {}


def utc_now() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
