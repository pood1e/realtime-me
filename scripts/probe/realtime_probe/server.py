"""Bounded HTTP server and scrape cache for the probe."""

from __future__ import annotations

import json
import socket
import threading
import time
import traceback
from collections.abc import Callable, Mapping
from concurrent.futures import ThreadPoolExecutor
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Any, NamedTuple
from urllib.parse import urlsplit

JSON_CONTENT_TYPE = "application/json; charset=utf-8"
METRICS_CONTENT_TYPE = "text/plain; version=0.0.4; charset=utf-8"
METRICS_CACHE_TTL_SECONDS = 5.0
REQUEST_TIMEOUT_SECONDS = 10
WORKER_THREADS = 8
MAX_QUEUED_REQUESTS = 32


class Response(NamedTuple):
    body: bytes
    content_type: str
    status: int = 200


def json_response(payload: object, status: int = 200) -> Response:
    return Response(
        json.dumps(payload, ensure_ascii=False).encode(), JSON_CONTENT_TYPE, status
    )


def metrics_response(payload: str) -> Response:
    return Response(payload.encode(), METRICS_CONTENT_TYPE)


def cached(
    render: Callable[[], Response], ttl_seconds: float = METRICS_CACHE_TTL_SECONDS
) -> Callable[[], Response]:
    lock = threading.Lock()
    expires_at = 0.0
    response: Response | None = None

    def serve() -> Response:
        nonlocal expires_at, response
        with lock:
            if response is not None and time.monotonic() < expires_at:
                return response
            response = render()
            expires_at = time.monotonic() + ttl_seconds
            return response

    return serve


class PooledHTTPServer(HTTPServer):
    allow_reuse_address = True
    request_queue_size = MAX_QUEUED_REQUESTS

    def __init__(
        self, address: tuple[str, int], handler: type[BaseHTTPRequestHandler]
    ) -> None:
        super().__init__(address, handler)
        self.pool = ThreadPoolExecutor(
            max_workers=WORKER_THREADS, thread_name_prefix="probe"
        )
        self.slots = threading.BoundedSemaphore(WORKER_THREADS + MAX_QUEUED_REQUESTS)

    def process_request(self, request: Any, client_address: Any) -> None:
        self.slots.acquire()
        try:
            self.pool.submit(self._handle_request, request, client_address)
        except RuntimeError:
            self.slots.release()
            self.shutdown_request(request)

    def _handle_request(self, request: Any, client_address: Any) -> None:
        try:
            self.finish_request(request, client_address)
        except Exception:  # Executor futures otherwise swallow the failure.
            self.handle_error(request, client_address)
        finally:
            try:
                self.shutdown_request(request)
            finally:
                self.slots.release()

    def server_close(self) -> None:
        super().server_close()
        self.pool.shutdown(wait=False)


class IPv6PooledHTTPServer(PooledHTTPServer):
    address_family = socket.AF_INET6


def run_server(
    bind: str, port: int, routes: Mapping[str, Callable[[], Response]]
) -> int:
    class Handler(BaseHTTPRequestHandler):
        timeout = REQUEST_TIMEOUT_SECONDS

        def do_GET(self) -> None:
            route = routes.get(urlsplit(self.path).path)
            response = (
                json_response({"error": "not_found"}, 404)
                if route is None
                else render_route(route)
            )
            self.send_response(response.status)
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Type", response.content_type)
            self.send_header("Content-Length", str(len(response.body)))
            self.end_headers()
            self.wfile.write(response.body)

        def log_message(self, format: str, *args: object) -> None:
            del format, args
            return

    server_type = IPv6PooledHTTPServer if ":" in bind else PooledHTTPServer
    server = server_type((bind, port), Handler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 130
    finally:
        server.server_close()
    return 0


def render_route(route: Callable[[], Response]) -> Response:
    try:
        return route()
    except Exception:  # Return a diagnosable HTTP response.
        traceback.print_exc()
        return json_response({"error": "render_failed"}, 500)
