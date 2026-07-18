"""Minimal ConnectRPC client used by operator-side status commands."""

from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
from pathlib import Path

JSON_CONTENT_TYPE = "application/json; charset=utf-8"


class ConnectError(Exception):
    def __init__(self, code: str) -> None:
        super().__init__(code)
        self.code = code


def connect_post(
    base_url: str,
    service: str,
    method: str,
    token: str,
    message: dict[str, object],
    timeout_seconds: float = 5,
) -> dict[str, object]:
    endpoint = f"{base_url.rstrip('/')}/realtime.me.status.v1.{service}/{method}"
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
    try:
        payload = json.loads(body) if body.strip() else {}
    except json.JSONDecodeError as error:
        raise ConnectError("invalid_response") from error
    return payload if isinstance(payload, dict) else {}


def connect_error_code(error: urllib.error.HTTPError) -> str:
    try:
        payload = json.loads(error.read().decode())
    except (OSError, ValueError):
        payload = {}
    code = payload.get("code") if isinstance(payload, dict) else None
    return str(code) if code else f"http_{error.code}"


def ensure_device_uid(
    base_url: str,
    token: str,
    identity_file: Path,
    kind: str,
    role: str,
    display_name: str,
    model: str,
    timeout_seconds: float = 5,
) -> str:
    cached = read_cached_uid(identity_file)
    if cached:
        return cached
    response = connect_post(
        base_url,
        "EnrollmentService",
        "EnrollDevice",
        token,
        {
            "kind": proto_enum("DEVICE_KIND", kind),
            "role": proto_enum("DEVICE_ROLE", role),
            "displayName": display_name,
            "model": model,
        },
        timeout_seconds,
    )
    uid = str(response.get("deviceUid") or "")
    if uid:
        write_cached_uid(identity_file, uid)
    return uid


def proto_enum(prefix: str, value: str) -> str:
    return f"{prefix}_{value.upper()}"


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
