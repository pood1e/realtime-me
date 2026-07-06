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
import xml.etree.ElementTree as ET
from pathlib import Path

CPU_CORES = "system.cpu.logical.count"
CPU_USAGE = "system.cpu.utilization"
MEMORY_USAGE = "system.memory.usage"
MEMORY_LIMIT = "system.memory.limit"
FILESYSTEM_USAGE = "system.filesystem.usage"
FILESYSTEM_LIMIT = "system.filesystem.limit"
FILESYSTEM_UTILIZATION = "system.filesystem.utilization"
OS_HINTS = (
    ("kali", "Kali Linux"),
    ("ubuntu", "Ubuntu"),
    ("debian", "Debian"),
    ("fedora", "Fedora"),
    ("centos", "CentOS"),
    ("arch", "Arch Linux"),
    ("windows", "Windows"),
)


def main() -> int:
    args = parse_args()
    payload = build_payload(args)
    if args.print:
        print(json.dumps(payload, indent=2))
        return 0

    token = args.token or os.getenv("STATUS_INGEST_TOKEN")
    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2

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
    parser.add_argument("--state-file", default=os.getenv("STATUS_DEVICE_REPORTER_STATE_FILE", str(default_state_file())))
    parser.add_argument("--timeout-seconds", type=float, default=5)
    parser.add_argument("--print", action="store_true")
    return parser.parse_args()


def build_payload(args: argparse.Namespace) -> dict:
    state = ReporterState.load(Path(args.state_file))
    machines = virtual_machines(args.vm_name_contains, state)
    state.save()
    return {
        "device_id": args.device_id,
        "device_name": args.device_name,
        "device_model": args.device_model,
        "kind": args.kind,
        "role": args.role,
        "state": "online",
        "updated_at": utc_now(),
        "metrics": metrics(),
        "children": machines,
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


def virtual_machines(name_contains: str, reporter_state: "ReporterState") -> list[dict]:
    if not shutil.which("virsh"):
        return []
    output = run(["virsh", "list", "--all"])
    machines = []
    for line in output.splitlines():
        match = re.match(r"\s*(?:\d+|-)\s+(\S+)\s+(.+?)\s*$", line)
        if not match:
            continue
        name, vm_state = match.group(1), re.sub(r"\s+", " ", match.group(2).strip())
        if name.lower() in {"name", "----"}:
            continue
        if name_contains and name_contains.lower() not in name.lower():
            continue
        machines.append({
            "device_id": f"vm-{name}",
            "device_name": name,
            "device_model": virtual_machine_model(name),
            "kind": "virtual_machine",
            "state": vm_state,
            "updated_at": utc_now(),
            "metrics": virtual_machine_metrics(name, reporter_state),
        })
    return machines


def virtual_machine_metrics(name: str, state: "ReporterState") -> list[dict]:
    info = virsh_dominfo(name)
    memory = virsh_dommemstat(name)
    stats = virsh_domstats(name)
    metrics = []
    cpus = info.get("CPU(s)")
    cpu_count = integer(cpus)
    used_memory, max_memory = virtual_machine_memory(info, memory)
    disk_usage, disk_limit = virtual_machine_disk(stats)
    cpu_ratio = state.vm_cpu_ratio(name, nanoseconds(stats.get("cpu.time")), cpu_count)
    if cpu_count is not None:
        metrics.append(sample(CPU_CORES, "{cpu}", float(cpu_count)))
    if cpu_ratio is not None:
        metrics.append(sample(CPU_USAGE, "1", cpu_ratio))
    if used_memory is not None:
        metrics.append(sample(MEMORY_USAGE, "By", float(used_memory), {"system.memory.state": "used"}))
    if max_memory is not None:
        metrics.append(sample(MEMORY_LIMIT, "By", float(max_memory)))
    if disk_usage is not None:
        metrics.append(sample(FILESYSTEM_USAGE, "By", float(disk_usage), {"device": "virtual_disk"}))
    if disk_limit is not None:
        metrics.append(sample(FILESYSTEM_LIMIT, "By", float(disk_limit), {"device": "virtual_disk"}))
    if disk_usage is not None and disk_limit is not None and disk_limit > 0:
        metrics.append(sample(FILESYSTEM_UTILIZATION, "1", clamp_ratio(disk_usage / disk_limit), {"device": "virtual_disk"}))
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


def virsh_dommemstat(name: str) -> dict[str, int]:
    output = run(["virsh", "dommemstat", name])
    stats = {}
    for line in output.splitlines():
        values = line.split()
        if len(values) != 2:
            continue
        try:
            stats[values[0]] = int(values[1])
        except ValueError:
            continue
    return stats


def virsh_domstats(name: str) -> dict[str, str]:
    output = run(["virsh", "domstats", "--cpu-total", "--balloon", "--block", name])
    stats = {}
    for line in output.splitlines():
        if "=" not in line:
            continue
        key, value = line.strip().split("=", 1)
        stats[key] = value
    return stats


def virtual_machine_model(name: str) -> str:
    output = run(["virsh", "dumpxml", name])
    if not output:
        return ""
    try:
        root = ET.fromstring(output)
    except ET.ParseError:
        return ""
    return " · ".join(unique(model_parts(name, root)))[:120]


def model_parts(name: str, root: ET.Element) -> list[str]:
    parts = []
    os_name = libosinfo_name(root) or guest_os_hint(name, root)
    if os_name:
        parts.append(os_name)
    machine_type = root.find("./os/type")
    if machine_type is not None:
        machine = machine_type.attrib.get("machine", "").strip()
        architecture = machine_type.attrib.get("arch", "").strip()
        virtualization = (machine_type.text or "").strip()
        parts.extend([machine, architecture, virtualization])
    return [part for part in parts if part]


def libosinfo_name(root: ET.Element) -> str:
    for element in root.iter():
        if not element.tag.endswith("os"):
            continue
        os_id = element.attrib.get("id", "")
        if "/libosinfo.org/" not in os_id:
            continue
        return os_id.rstrip("/").rsplit("/", 1)[-1].replace("-", " ").title()
    return ""


def guest_os_hint(name: str, root: ET.Element) -> str:
    values = [name]
    for element in root.iter():
        if element.tag.endswith("title") or element.tag.endswith("description"):
            values.append(element.text or "")
        values.extend(element.attrib.values())
    text = " ".join(values).lower()
    for marker, label in OS_HINTS:
        if marker in text:
            return label
    return ""


def unique(values: list[str]) -> list[str]:
    result = []
    seen = set()
    for value in values:
        if value in seen:
            continue
        seen.add(value)
        result.append(value)
    return result


def virtual_machine_memory(info: dict[str, str], memory: dict[str, int]) -> tuple[int | None, int | None]:
    actual = kibibytes_value(memory.get("actual"))
    unused = kibibytes_value(memory.get("unused"))
    if actual is not None and unused is not None:
        return max(0, actual - unused), actual
    return kibibytes(info.get("Used memory")), kibibytes(info.get("Max memory"))


def virtual_machine_disk(stats: dict[str, str]) -> tuple[int | None, int | None]:
    allocation = 0
    capacity = 0
    for index in range(integer(stats.get("block.count")) or 0):
        path = stats.get(f"block.{index}.path", "")
        if path.lower().endswith(".iso"):
            continue
        block_allocation = integer(stats.get(f"block.{index}.allocation"))
        block_capacity = integer(stats.get(f"block.{index}.capacity"))
        if block_allocation is None or block_capacity is None or block_capacity <= 0:
            continue
        allocation += block_allocation
        capacity += block_capacity
    if capacity <= 0:
        return None, None
    return allocation, capacity


def kibibytes(value: str | None) -> int | None:
    if not value:
        return None
    match = re.search(r"(\d+)", value)
    if not match:
        return None
    return int(match.group(1)) * 1024


def kibibytes_value(value: int | None) -> int | None:
    return None if value is None else value * 1024


def nanoseconds(value: str | None) -> int | None:
    return integer(value)


def integer(value: str | None) -> int | None:
    if value is None:
        return None
    try:
        return int(value)
    except ValueError:
        return None


def device_model() -> str:
    system = platform.system().lower()
    if system == "darwin":
        os_name = " ".join(part for part in [
            run(["sw_vers", "-productName"]).strip(),
            run(["sw_vers", "-productVersion"]).strip(),
        ] if part)
        model = run(["sysctl", "-n", "hw.model"]).strip()
        cpu = run(["sysctl", "-n", "machdep.cpu.brand_string"]).strip()
        return " · ".join(part for part in [os_name, model, cpu] if part)[:120]
    if system == "linux":
        parts = [read_first(paths) for paths in [
            ["/sys/class/dmi/id/sys_vendor", "/sys/class/dmi/id/board_vendor"],
            ["/sys/class/dmi/id/product_name", "/sys/class/dmi/id/board_name"],
        ]]
        hardware = " ".join(part for part in parts if part and not part.startswith("Default"))
        return " · ".join(part for part in [linux_os_name(), hardware] if part)[:120] or platform.machine()
    return platform.platform()


def linux_os_name() -> str:
    try:
        for line in Path("/etc/os-release").read_text().splitlines():
            if line.startswith("PRETTY_NAME="):
                return line.split("=", 1)[1].strip().strip('"')
    except OSError:
        return ""
    return ""


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


def default_state_file() -> Path:
    return Path.home() / ".cache" / "realtime-me" / "status-device-reporter.json"


class ReporterState:
    def __init__(self, path: Path, data: dict) -> None:
        self.path = path
        self.data = data
        self.now = time.time()

    @classmethod
    def load(cls, path: Path) -> "ReporterState":
        try:
            data = json.loads(path.read_text())
        except (OSError, json.JSONDecodeError):
            data = {}
        return cls(path, data if isinstance(data, dict) else {})

    def save(self) -> None:
        try:
            self.path.parent.mkdir(parents=True, exist_ok=True)
            self.path.write_text(json.dumps(self.data), encoding="utf-8")
        except OSError:
            return

    def vm_cpu_ratio(self, name: str, cpu_time_ns: int | None, cpu_count: int | None) -> float | None:
        if cpu_time_ns is None or cpu_count is None or cpu_count < 1:
            return None
        key = f"vm:{name}:cpu"
        previous = self.data.get(key)
        self.data[key] = {"time": self.now, "cpu_time_ns": cpu_time_ns}
        if not isinstance(previous, dict):
            return None
        previous_time = previous.get("time")
        previous_cpu_time = previous.get("cpu_time_ns")
        if not isinstance(previous_time, (int, float)) or not isinstance(previous_cpu_time, int):
            return None
        elapsed_seconds = self.now - previous_time
        elapsed_cpu_seconds = (cpu_time_ns - previous_cpu_time) / 1_000_000_000
        if elapsed_seconds <= 0 or elapsed_cpu_seconds < 0:
            return None
        return clamp_ratio(elapsed_cpu_seconds / elapsed_seconds / cpu_count)


if __name__ == "__main__":
    raise SystemExit(main())
