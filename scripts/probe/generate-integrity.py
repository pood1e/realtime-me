#!/usr/bin/env python3
"""Generate the probe's canonical SHA-256 integrity manifest."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import sys
from pathlib import Path

PROBE_DIR = Path(__file__).resolve().parent
INSTALLER = PROBE_DIR.parent / "install-probe.py"
INTEGRITY = PROBE_DIR / "integrity.json"
CONSTANT_PATTERN = re.compile(r'(?m)^INTEGRITY_SHA256 = "[0-9a-f]{64}"$')


def source_files() -> list[Path]:
    package_files = []
    for path in sorted((PROBE_DIR / "realtime_probe").rglob("*")):
        if "__pycache__" in path.parts or path.is_dir():
            continue
        if path.is_symlink() or path.suffix != ".py":
            raise RuntimeError(f"unsupported probe runtime file: {path}")
        package_files.append(path)
    requirements = PROBE_DIR / "requirements.txt"
    if requirements.is_symlink() or not requirements.is_file():
        raise RuntimeError(f"invalid probe requirements file: {requirements}")
    return [requirements, *package_files]


def render_integrity() -> bytes:
    files = {
        path.relative_to(PROBE_DIR)
        .as_posix(): hashlib.sha256(path.read_bytes())
        .hexdigest()
        for path in source_files()
    }
    return (
        json.dumps({"version": 1, "files": files}, indent=2, sort_keys=True) + "\n"
    ).encode()


def installer_with_digest(digest: str) -> str:
    source = INSTALLER.read_bytes().decode()
    updated, replacements = CONSTANT_PATTERN.subn(
        f'INTEGRITY_SHA256 = "{digest}"', source
    )
    if replacements != 1:
        raise RuntimeError("installer integrity constant is missing or duplicated")
    return updated


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()

    integrity = render_integrity()
    installer = installer_with_digest(hashlib.sha256(integrity).hexdigest())
    if args.check:
        stale = []
        if not INTEGRITY.is_file() or INTEGRITY.read_bytes() != integrity:
            stale.append(INTEGRITY)
        if INSTALLER.read_bytes() != installer.encode():
            stale.append(INSTALLER)
        if stale:
            print(
                "stale probe integrity: " + ", ".join(map(str, stale)),
                file=sys.stderr,
            )
            return 1
        return 0

    INTEGRITY.write_bytes(integrity)
    INSTALLER.write_bytes(installer.encode())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
