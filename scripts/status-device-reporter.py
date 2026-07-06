#!/usr/bin/env python3
from __future__ import annotations
import argparse
import json
import os
import platform
import re
import shutil
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request
from pathlib import Path

CPU_CORES = "system.cpu.logical.count"
CPU_USAGE = "system.cpu.utilization"
MEMORY_USAGE = "system.memory.usage"
MEMORY_LIMIT = "system.memory.limit"
FILESYSTEM_USAGE = "system.filesystem.usage"
FILESYSTEM_LIMIT = "system.filesystem.limit"
FILESYSTEM_UTILIZATION = "system.filesystem.utilization"


def main() -> int:
    args = parse_args()
    token = args.token or os.getenv("STATUS_INGEST_TOKEN")
    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2

    payload = build_payload(args)
    if args.print:
        print(json.dumps(payload, indent=2))
        return 0

    endpoint = args.url.rstrip("/") + "/api/ingest/host"
    request = urllib.request.Request(
        endpoint,
        data=json.dumps(payload).encode(),
        method="POST",
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json; charset=utf-8",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=args.timeout_seconds) as response:
            if response.status < 200 or response.status > 299:
                print(f"gateway rejected device status: HTTP {response.status}", file=sys.stderr)
                return 1
    except urllib.error.HTTPError as error:
        print(f"gateway rejected device status: HTTP {error.code}", file=sys.stderr)
        return 1
    except OSError as error:
        print(f"gateway device status push failed: {error.__class__.__name__}", file=sys.stderr)
        return 1
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Publish host device status to realtime-me gateway.")
    parser.add_argument("--url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--token", default="")
    parser.add_argument("--device-id", default=os.getenv("STATUS_DEVICE_ID", socket.gethostname()))
    parser.add_argument("--device-name", default=os.getenv("STATUS_DEVICE_NAME", socket.gethostname()))
    parser.add_argument("--device-model", default=os.getenv("STATUS_DEVICE_MODEL", device_model()))
    parser.add_argument("--kind", default=os.getenv("STATUS_DEVICE_KIND", "host"))
    parser.add_argument("--role", default=os.getenv("STATUS_DEVICE_ROLE", "desktop"))
    parser.add_argument("--vm-name-contains", default=os.getenv("STATUS_VM_NAME_CONTAINS", ""))
    parser.add_argument("--timeout-seconds", type=float, default=5)
    parser.add_argument("--print", action="store_true")
    return parser.parse_args()


def build_payload(args: argparse.Namespace) -> dict:
    return {
        "device_id": args.device_id,
        "device_name": args.device_name,
        "device_model": args.device_model,
        "kind": args.kind,
        "role": args.role,
        "state": "online",
        "updated_at": utc_now(),
        "metrics": metrics(),
        "children": virtual_machines(args.vm_name_contains),
    }


def metrics() -> list[dict]:
    disk = shutil.disk_usage("/")
    memory = memory_usage()
    result = [
        sample(CPU_CORES, "{cpu}", float(os.cpu_count() or 0)),
        sample(CPU_USAGE, "1", cpu_usage()),
        sample(MEMORY_USAGE, "By", float(memory[0]), {"system.memory.state": "used"}),
        sample(MEMORY_LIMIT, "By", float(memory[1])),
        sample(FILESYSTEM_USAGE, "By", float(disk.used), {"mountpoint": "/"}),
        sample(FILESYSTEM_LIMIT, "By", float(disk.total), {"mountpoint": "/"}),
        sample(FILESYSTEM_UTILIZATION, "1", disk.used / disk.total if disk.total else 0, {"mountpoint": "/"}),
    ]
    return [item for item in result if item["value"] >= 0]


def sample(name: str, unit: str, value: float, attributes: dict[str, str] | None = None) -> dict:
    payload = {"name": name, "unit": unit, "value": round(value, 6)}
    if attributes:
        payload["attributes"] = attributes
    return payload


def cpu_usage() -> float:
    system = platform.system().lower()
    if system == "linux":
        first = linux_cpu_times()
        time.sleep(0.25)
        second = linux_cpu_times()
        total = sum(second) - sum(first)
        idle = second[3] - first[3]
        return clamp_ratio(1 - idle / total) if total > 0 else 0
    if system == "darwin":
        output = run(["top", "-l", "2", "-n", "0", "-s", "1"])
        matches = re.findall(r"CPU usage: .*?, .*?, ([0-9.]+)% idle", output)
        if matches:
            return clamp_ratio(1 - float(matches[-1]) / 100)
    return 0


def linux_cpu_times() -> list[int]:
    values = Path("/proc/stat").read_text().splitlines()[0].split()[1:]
    return [int(value) for value in values]


def memory_usage() -> tuple[int, int]:
    system = platform.system().lower()
    if system == "linux":
        values = {}
        for line in Path("/proc/meminfo").read_text().splitlines():
            key, value = line.split(":", 1)
            values[key] = int(value.strip().split()[0]) * 1024
        total = values.get("MemTotal", 0)
        available = values.get("MemAvailable", 0)
        return max(0, total - available), total
    if system == "darwin":
        total = int(run(["sysctl", "-n", "hw.memsize"]).strip() or "0")
        stats = run(["vm_stat"])
        page_size_match = re.search(r"page size of (\d+) bytes", stats)
        page_size = int(page_size_match.group(1)) if page_size_match else 4096
        pages = {}
        for line in stats.splitlines():
            if ":" not in line:
                continue
            key, value = line.split(":", 1)
            pages[key.strip()] = int(re.sub(r"[^0-9]", "", value) or "0")
        free = pages.get("Pages free", 0) + pages.get("Pages speculative", 0)
        return max(0, total - free * page_size), total
    return 0, 0


def virtual_machines(name_contains: str) -> list[dict]:
    if not shutil.which("virsh"):
        return []
    output = run(["virsh", "list", "--all"])
    machines = []
    for line in output.splitlines():
        match = re.match(r"\s*(?:\d+|-)\s+(\S+)\s+(.+?)\s*$", line)
        if not match:
            continue
        name, state = match.group(1), re.sub(r"\s+", " ", match.group(2).strip())
        if name.lower() in {"name", "----"}:
            continue
        if name_contains and name_contains.lower() not in name.lower():
            continue
        machines.append({
            "device_id": f"vm-{name}",
            "device_name": name,
            "kind": "virtual_machine",
            "state": state,
            "updated_at": utc_now(),
            "metrics": virtual_machine_metrics(name),
        })
    return machines


def virtual_machine_metrics(name: str) -> list[dict]:
    info = virsh_dominfo(name)
    metrics = []
    cpus = info.get("CPU(s)")
    used_memory = kibibytes(info.get("Used memory"))
    max_memory = kibibytes(info.get("Max memory"))
    if cpus is not None:
        metrics.append(sample(CPU_CORES, "{cpu}", float(cpus)))
    if used_memory is not None:
        metrics.append(sample(MEMORY_USAGE, "By", float(used_memory), {"system.memory.state": "used"}))
    if max_memory is not None:
        metrics.append(sample(MEMORY_LIMIT, "By", float(max_memory)))
    return metrics


def virsh_dominfo(name: str) -> dict[str, str]:
    output = run(["virsh", "dominfo", name])
    info = {}
    for line in output.splitlines():
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        info[key.strip()] = value.strip()
    return info


def kibibytes(value: str | None) -> int | None:
    if not value:
        return None
    match = re.search(r"(\d+)", value)
    if not match:
        return None
    return int(match.group(1)) * 1024


def device_model() -> str:
    system = platform.system().lower()
    if system == "darwin":
        model = run(["sysctl", "-n", "hw.model"]).strip()
        cpu = run(["sysctl", "-n", "machdep.cpu.brand_string"]).strip()
        return " · ".join(part for part in [model, cpu] if part)
    if system == "linux":
        parts = [read_first(paths) for paths in [
            ["/sys/class/dmi/id/sys_vendor", "/sys/class/dmi/id/board_vendor"],
            ["/sys/class/dmi/id/product_name", "/sys/class/dmi/id/board_name"],
        ]]
        return " ".join(part for part in parts if part and not part.startswith("Default")) or platform.machine()
    return platform.platform()


def read_first(paths: list[str]) -> str:
    for path in paths:
        try:
            value = Path(path).read_text().strip()
        except OSError:
            continue
        if value:
            return value
    return ""


def run(command: list[str]) -> str:
    try:
        return subprocess.check_output(command, text=True, stderr=subprocess.DEVNULL, timeout=5)
    except (OSError, subprocess.SubprocessError):
        return ""


def utc_now() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def clamp_ratio(value: float) -> float:
    return max(0, min(1, value))


if __name__ == "__main__":
    raise SystemExit(main())
