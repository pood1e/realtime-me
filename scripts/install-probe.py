#!/usr/bin/env python3
"""Install the unified Realtime Me probe on Linux, macOS, or Windows."""

from __future__ import annotations

import argparse
import base64
import getpass
import hashlib
import hmac
import ipaddress
import json
import math
import os
import platform
import plistlib
import re
import shlex
import shutil
import socket
import subprocess
import sys
import tempfile
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path, PurePosixPath

try:
    import pwd
except ImportError:  # Windows has no POSIX account database.
    pwd = None

MINIMUM_PYTHON = (3, 10)
DEFAULT_PORT = 18082
MAX_ARTIFACT_BYTES = 2 * 1024 * 1024
COMMAND_TIMEOUT_SECONDS = 300
SERVICE_STOP_TIMEOUT_SECONDS = 30
RELEASE_PATTERN = re.compile(r"[0-9a-f]{40}")
SHA256_PATTERN = re.compile(r"[0-9a-f]{64}")
# Updated by scripts/probe/generate-integrity.py.
INTEGRITY_SHA256 = "7f64fde11a9c9e42706eb4efeeb992c6f5247e45fc92f2322c9051084c76abd7"
LINUX_SERVICE = "realtime-me-probe.service"
MACOS_LABEL = "me.realtime.probe"
WINDOWS_TASK = "Realtime Me Probe"
PRIVATE_NETWORKS = (
    ipaddress.ip_network("10.0.0.0/8"),
    ipaddress.ip_network("172.16.0.0/12"),
    ipaddress.ip_network("192.168.0.0/16"),
    ipaddress.ip_network("fd00::/8"),
)


@dataclass(frozen=True)
class Settings:
    system: str
    install_dir: Path
    python: Path
    bind: str
    port: int
    active_window_seconds: int
    codex_homes: str
    claude_home: str
    user: str
    device_name: str
    exporter_host: str
    device_kind: str
    device_role: str
    timeout_seconds: float
    base_urls: tuple[str, ...]

    @property
    def runtime_dir(self) -> Path:
        return self.install_dir / "runtime"


class Installer:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings

    def install(self) -> None:
        prepare_install_directory(self.settings)
        staging = self._stage_runtime()
        controller = service_controller(self.settings)
        backup: Path | None = None
        try:
            controller.stop()
            controller.remove_legacy()
            backup = activate_runtime(self.settings.runtime_dir, staging)
            controller.configure()
            secure_installation(self.settings)
            controller.start()
        except Exception:
            if backup is not None and backup.exists():
                controller.stop()
                restore_runtime(self.settings.runtime_dir, backup)
                controller.configure()
                secure_installation(self.settings)
                controller.start()
            raise
        finally:
            if staging.exists():
                remove_path(staging)
        if backup is not None:
            remove_path(backup)
        try:
            remove_obsolete_probe_install(self.settings)
        except OSError as error:
            log(f"Could not remove inactive legacy probe files: {error}")
        self._print_summary()

    def _stage_runtime(self) -> Path:
        self.settings.install_dir.mkdir(parents=True, exist_ok=True)
        staging = Path(
            tempfile.mkdtemp(prefix=".runtime-", dir=self.settings.install_dir)
        )
        try:
            integrity_payload = self._download_verified(
                "probe/integrity.json", INTEGRITY_SHA256
            )
            files = parse_integrity_manifest(integrity_payload)
            (staging / "integrity.json").write_bytes(integrity_payload)
            for relative, digest in files.items():
                destination = staging / relative
                destination.parent.mkdir(parents=True, exist_ok=True)
                destination.write_bytes(
                    self._download_verified(f"probe/{relative.as_posix()}", digest)
                )
            requirements = staging / "requirements.txt"
            vendor = staging / "vendor"
            log("Installing pinned Python dependencies")
            run(
                [
                    str(self.settings.python),
                    "-I",
                    "-m",
                    "pip",
                    "install",
                    "--disable-pip-version-check",
                    "--isolated",
                    "--index-url",
                    "https://pypi.org/simple",
                    "--no-input",
                    "--no-cache-dir",
                    "--no-compile",
                    "--no-deps",
                    "--only-binary=:all:",
                    "--require-hashes",
                    "--retries",
                    "2",
                    "--target",
                    str(vendor),
                    "--timeout",
                    str(self.settings.timeout_seconds),
                    "--requirement",
                    str(requirements),
                ]
            )
            make_readable(staging)
            return staging
        except Exception:
            remove_path(staging)
            raise

    def _download_verified(self, relative: str, expected_sha256: str) -> bytes:
        last_error: Exception | None = None
        for base_url in self.settings.base_urls:
            url = f"{base_url.rstrip('/')}/{relative}"
            try:
                request = urllib.request.Request(
                    url, headers={"User-Agent": "realtime-me-probe-installer/1"}
                )
                with urllib.request.urlopen(
                    request, timeout=self.settings.timeout_seconds
                ) as response:
                    if urllib.parse.urlsplit(response.geturl()).scheme != "https":
                        raise RuntimeError(f"insecure redirect for {relative}")
                    declared_size = response.headers.get("Content-Length")
                    if declared_size and int(declared_size) > MAX_ARTIFACT_BYTES:
                        raise RuntimeError(f"artifact is too large: {relative}")
                    payload = response.read(MAX_ARTIFACT_BYTES + 1)
                if len(payload) > MAX_ARTIFACT_BYTES:
                    raise RuntimeError(f"artifact is too large: {relative}")
                actual_sha256 = hashlib.sha256(payload).hexdigest()
                if not hmac.compare_digest(actual_sha256, expected_sha256):
                    raise RuntimeError(f"SHA-256 mismatch for {relative}")
                return payload
            except (OSError, ValueError, RuntimeError) as error:
                last_error = error
                log(f"Rejected download mirror {base_url}: {error}")
        raise RuntimeError(f"could not download {relative}: {last_error}")

    def _print_summary(self) -> None:
        settings = self.settings
        print(f"Installed Realtime Me probe for {settings.system}.")
        print(f"Probe endpoint: {settings.exporter_host}:{settings.port}")
        print(
            f"Health check:   http://{url_host(settings.exporter_host)}:{settings.port}/healthz"
        )
        if settings.system == "windows":
            print(
                "Allow the selected Python executable through Windows Firewall when prompted."
            )
        print("\nRegister this host centrally:")
        if settings.system == "windows":
            command = (
                "$env:STATUS_INGEST_TOKEN='...'; py -3 scripts/operator/register-device.py "
                f"--url {ps_literal('<GATEWAY_URL>')} --host {ps_literal(settings.exporter_host)} "
                f"--name {ps_literal(settings.device_name)} --kind {settings.device_kind} "
                f"--role {settings.device_role}"
            )
        else:
            command = (
                "STATUS_INGEST_TOKEN=... python3 scripts/operator/register-device.py "
                f"--url {shell_word('<GATEWAY_URL>')} --host {shell_word(settings.exporter_host)} "
                f"--name {shell_word(settings.device_name)} --kind {settings.device_kind} "
                f"--role {settings.device_role}"
            )
        print(f"  {command}")


class LinuxServiceController:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        self.unit = Path("/etc/systemd/system") / LINUX_SERVICE

    def stop(self) -> None:
        subprocess.run(
            ["systemctl", "disable", "--now", LINUX_SERVICE],
            check=False,
            stdout=subprocess.DEVNULL,
            timeout=SERVICE_STOP_TIMEOUT_SECONDS,
        )

    def remove_legacy(self) -> None:
        for service in (
            "realtime-me-node-exporter.service",
            "realtime-me-device-exporter.service",
            "realtime-me-agent-exporter.service",
            "realtime-me-status-device.service",
            "realtime-me-status-device.timer",
            "realtime-me-agent.service",
            "realtime-me-agent.timer",
        ):
            subprocess.run(
                ["systemctl", "disable", "--now", service],
                check=False,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                timeout=SERVICE_STOP_TIMEOUT_SECONDS,
            )
            (Path("/etc/systemd/system") / service).unlink(missing_ok=True)
        remove_legacy_payload(self.settings.install_dir)
        run(["systemctl", "daemon-reload"])

    def configure(self) -> None:
        runner = write_unix_runner(self.settings)
        if pwd is None:
            raise RuntimeError("POSIX account database is unavailable")
        user = pwd.getpwnam(self.settings.user)
        directives = (
            f"User={self.settings.user}\n"
            f"Environment=XDG_RUNTIME_DIR=/run/user/{user.pw_uid}\n"
            f"Environment=DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/{user.pw_uid}/bus\n"
        )
        atomic_write(
            self.unit,
            "[Unit]\n"
            "Description=Realtime Me cross-platform probe\n"
            "Wants=network-online.target\n"
            "After=network-online.target\n\n"
            "[Service]\n"
            "Type=simple\n"
            f"{directives}"
            f"ExecStart={runner}\n"
            "Restart=always\n"
            "RestartSec=5s\n"
            "UMask=0077\n"
            "NoNewPrivileges=true\n"
            "CapabilityBoundingSet=\n"
            "LockPersonality=true\n"
            "MemoryDenyWriteExecute=true\n"
            "PrivateDevices=true\n"
            "PrivateTmp=true\n"
            "ProtectClock=true\n"
            "ProtectControlGroups=true\n"
            "ProtectHome=read-only\n"
            "ProtectHostname=true\n"
            "ProtectKernelLogs=true\n"
            "ProtectKernelModules=true\n"
            "ProtectKernelTunables=true\n"
            "ProtectProc=invisible\n"
            "ProtectSystem=strict\n"
            "RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6\n"
            "RestrictNamespaces=true\n"
            "RestrictRealtime=true\n"
            "RestrictSUIDSGID=true\n"
            "SystemCallArchitectures=native\n\n"
            "[Install]\n"
            "WantedBy=multi-user.target\n",
            mode=0o644,
        )
        run(["systemctl", "daemon-reload"])

    def start(self) -> None:
        run(["systemctl", "enable", "--now", LINUX_SERVICE])


class MacOSServiceController:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        if pwd is None:
            raise RuntimeError("POSIX account database is unavailable")
        self.user = pwd.getpwnam(settings.user)
        self.domain = f"gui/{self.user.pw_uid}"
        self.agents_dir = Path(self.user.pw_dir) / "Library" / "LaunchAgents"
        self.logs_dir = Path("/var/log/realtime-me-probe")
        self.log_file = self.logs_dir / f"{self.user.pw_uid}.log"
        self.plist = self.agents_dir / f"{MACOS_LABEL}.plist"

    def stop(self) -> None:
        subprocess.run(
            ["launchctl", "bootout", f"{self.domain}/{MACOS_LABEL}"],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            timeout=SERVICE_STOP_TIMEOUT_SECONDS,
        )

    def remove_legacy(self) -> None:
        for label in (
            "space.pood1e.realtime-me.node-exporter",
            "space.pood1e.realtime-me.device-exporter",
            "space.pood1e.realtime-me.agent-exporter",
        ):
            subprocess.run(
                ["launchctl", "bootout", f"{self.domain}/{label}"],
                check=False,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                timeout=SERVICE_STOP_TIMEOUT_SECONDS,
            )
            (self.agents_dir / f"{label}.plist").unlink(missing_ok=True)
        (self.agents_dir / f"{MACOS_LABEL}.plist").unlink(missing_ok=True)
        remove_legacy_payload(self.settings.install_dir)

    def configure(self) -> None:
        runner = write_unix_runner(self.settings)
        if self.agents_dir.is_symlink():
            raise RuntimeError(
                f"LaunchAgents directory cannot be a symlink: {self.agents_dir}"
            )
        self.agents_dir.mkdir(parents=True, exist_ok=True)
        if self.logs_dir.is_symlink():
            raise RuntimeError(
                f"probe log directory cannot be a symlink: {self.logs_dir}"
            )
        self.logs_dir.mkdir(parents=True, exist_ok=True)
        os.chown(self.logs_dir, 0, 0)
        self.logs_dir.chmod(0o755)
        atomic_write(self.log_file, b"", mode=0o600)
        os.chown(self.log_file, self.user.pw_uid, self.user.pw_gid)
        self.log_file.chmod(0o600)
        payload = {
            "Label": MACOS_LABEL,
            "ProgramArguments": [str(runner)],
            "KeepAlive": True,
            "RunAtLoad": True,
            "StandardErrorPath": str(self.log_file),
            "StandardOutPath": str(self.log_file),
        }
        atomic_write(self.plist, plistlib.dumps(payload, sort_keys=False), mode=0o644)

    def start(self) -> None:
        run(["launchctl", "bootstrap", self.domain, str(self.plist)])
        subprocess.run(
            ["launchctl", "enable", f"{self.domain}/{MACOS_LABEL}"],
            check=False,
            timeout=SERVICE_STOP_TIMEOUT_SECONDS,
        )
        run(["launchctl", "kickstart", "-k", f"{self.domain}/{MACOS_LABEL}"])


class WindowsServiceController:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        self.runner = settings.install_dir / "run-probe.cmd"

    def stop(self) -> None:
        subprocess.run(
            ["schtasks.exe", "/End", "/TN", WINDOWS_TASK],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            timeout=SERVICE_STOP_TIMEOUT_SECONDS,
        )
        subprocess.run(
            ["schtasks.exe", "/Delete", "/F", "/TN", WINDOWS_TASK],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            timeout=SERVICE_STOP_TIMEOUT_SECONDS,
        )

    def remove_legacy(self) -> None:
        remove_legacy_payload(self.settings.install_dir)

    def configure(self) -> None:
        bootstrap = write_python_runner(self.settings)
        atomic_write(
            self.runner,
            "@echo off\n"
            "chcp 65001 >nul\n"
            f"{probe_command(self.settings, bootstrap, windows=True)}\n"
            "exit /b %ERRORLEVEL%\n",
            encoding="utf-8",
            newline="\r\n",
        )
        identity = self.settings.user
        task_arguments = ps_literal(f'/d /c ""{self.runner}""')
        task_identity = ps_literal(identity)
        task_name = ps_literal(WINDOWS_TASK)
        script = (
            "$ErrorActionPreference = 'Stop'\n"
            f"$action = New-ScheduledTaskAction -Execute $env:ComSpec -Argument {task_arguments}\n"
            f"$trigger = New-ScheduledTaskTrigger -AtLogOn -User {task_identity}\n"
            f"$principal = New-ScheduledTaskPrincipal -UserId {task_identity} "
            "-LogonType Interactive -RunLevel Limited\n"
            "$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries "
            "-DontStopIfGoingOnBatteries -StartWhenAvailable "
            "-ExecutionTimeLimit ([TimeSpan]::Zero) -MultipleInstances IgnoreNew "
            "-RestartCount 255 -RestartInterval (New-TimeSpan -Minutes 1)\n"
            f"Register-ScheduledTask -TaskName {task_name} -Action $action -Trigger $trigger "
            "-Principal $principal -Settings $settings -Force | Out-Null\n"
        )
        run(
            [
                "powershell.exe",
                "-NoLogo",
                "-NoProfile",
                "-NonInteractive",
                "-EncodedCommand",
                encoded_powershell(script),
            ]
        )

    def start(self) -> None:
        run(
            [
                "powershell.exe",
                "-NoLogo",
                "-NoProfile",
                "-NonInteractive",
                "-Command",
                f"Start-ScheduledTask -TaskName {ps_literal(WINDOWS_TASK)}",
            ]
        )


def main() -> int:
    require_python_version()
    try:
        settings = parse_settings()
        validate_environment(settings)
        Installer(settings).install()
    except (
        OSError,
        RuntimeError,
        subprocess.CalledProcessError,
        subprocess.TimeoutExpired,
    ) as error:
        print(f"installation failed: {error}", file=sys.stderr)
        return 1
    return 0


def parse_settings() -> Settings:
    system = platform.system().lower()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--install-dir",
        default=os.getenv("REALTIME_PROBE_INSTALL_DIR", default_install_dir(system)),
    )
    parser.add_argument("--bind", default=os.getenv("REALTIME_PROBE_BIND", ""))
    parser.add_argument(
        "--port", type=port, default=os.getenv("REALTIME_PROBE_PORT", str(DEFAULT_PORT))
    )
    parser.add_argument(
        "--active-window-seconds",
        type=positive_integer,
        default=os.getenv("REALTIME_PROBE_ACTIVE_WINDOW_SECONDS", "300"),
    )
    parser.add_argument("--codex-homes", default=os.getenv("REALTIME_CODEX_HOMES", ""))
    parser.add_argument("--claude-home", default=os.getenv("REALTIME_CLAUDE_HOME", ""))
    parser.add_argument(
        "--user",
        default=os.getenv("REALTIME_PROBE_USER", default_service_user(system)),
    )
    parser.add_argument(
        "--name",
        default=os.getenv("REALTIME_PROBE_DEVICE_NAME", platform.node() or system),
    )
    parser.add_argument("--host", default=os.getenv("REALTIME_PROBE_HOST", ""))
    parser.add_argument(
        "--kind",
        choices=("host", "virtual_machine"),
        default=os.getenv("REALTIME_PROBE_DEVICE_KIND", "host"),
    )
    parser.add_argument(
        "--role",
        choices=("server", "desktop", "vm"),
        default=os.getenv("REALTIME_PROBE_DEVICE_ROLE", "desktop"),
    )
    parser.add_argument(
        "--timeout-seconds",
        type=positive_float,
        default=os.getenv("REALTIME_PROBE_DOWNLOAD_TIMEOUT_SECONDS", "20"),
    )
    args = parser.parse_args()
    base_urls = configured_base_urls()
    exporter_host = args.host.strip().strip("[]") or detect_lan_address()
    return Settings(
        system=system,
        install_dir=Path(os.path.abspath(Path(args.install_dir).expanduser())),
        python=Path(sys.executable).resolve(),
        bind=args.bind.strip().strip("[]") or exporter_host,
        port=args.port,
        active_window_seconds=args.active_window_seconds,
        codex_homes=args.codex_homes,
        claude_home=args.claude_home,
        user=args.user,
        device_name=args.name,
        exporter_host=exporter_host,
        device_kind=args.kind,
        device_role=args.role,
        timeout_seconds=args.timeout_seconds,
        base_urls=base_urls,
    )


def configured_base_urls() -> tuple[str, ...]:
    configured = tuple(os.getenv("REALTIME_PROBE_SOURCE_URLS", "").split())
    if configured:
        return tuple(validated_base_url(url) for url in configured)

    release = os.getenv("REALTIME_PROBE_RELEASE", "").strip().lower()
    if RELEASE_PATTERN.fullmatch(release) is None:
        raise RuntimeError(
            "REALTIME_PROBE_RELEASE must be the reviewed 40-character Git commit"
        )
    return (
        f"https://cdn.jsdelivr.net/gh/pood1e/realtime-me@{release}/scripts",
        f"https://raw.githubusercontent.com/pood1e/realtime-me/{release}/scripts",
    )


def default_install_dir(system: str) -> str:
    if system == "linux":
        return "/opt/realtime-me-probe"
    if system == "darwin":
        return "/Library/Application Support/RealtimeMeProbe"
    if system == "windows":
        return str(
            Path(os.getenv("PROGRAMFILES", r"C:\Program Files")) / "RealtimeMeProbe"
        )
    return ""


def validate_environment(settings: Settings) -> None:
    if settings.system not in {"linux", "darwin", "windows"}:
        raise RuntimeError(f"unsupported operating system: {platform.system()}")
    if settings.system == "linux":
        if os.geteuid() != 0:
            raise RuntimeError("Linux installation must run as root (use sudo)")
        if pwd is None:
            raise RuntimeError("POSIX account database is unavailable")
        try:
            user = pwd.getpwnam(settings.user)
        except KeyError as error:
            raise RuntimeError(
                f"Linux probe user does not exist: {settings.user}"
            ) from error
        if user.pw_uid == 0:
            raise RuntimeError("Linux probe must run as a non-root user; pass --user")
        require_command("systemctl")
    elif settings.system == "darwin":
        if os.geteuid() != 0:
            raise RuntimeError("macOS installation must run as root (use sudo)")
        if pwd is None:
            raise RuntimeError("POSIX account database is unavailable")
        try:
            user = pwd.getpwnam(settings.user)
        except KeyError as error:
            raise RuntimeError(
                f"macOS probe user does not exist: {settings.user}"
            ) from error
        if user.pw_uid == 0:
            raise RuntimeError("macOS probe must target the login user; pass --user")
        require_command("launchctl")
    else:
        if not windows_is_administrator():
            raise RuntimeError(
                "Windows installation requires an Administrator terminal"
            )
        if not settings.user:
            raise RuntimeError("Windows probe user is empty; pass --user")
        require_command("icacls.exe")
        require_command("powershell.exe")
        require_command("schtasks.exe")
    if not settings.exporter_host:
        raise RuntimeError("could not detect a LAN address; pass --host")
    if not is_private_address(settings.exporter_host):
        raise RuntimeError("probe host must be a private LAN or VPN IP address")
    run([str(settings.python), "-I", "-m", "pip", "--version"])


def service_controller(
    settings: Settings,
) -> LinuxServiceController | MacOSServiceController | WindowsServiceController:
    if settings.system == "linux":
        return LinuxServiceController(settings)
    if settings.system == "darwin":
        return MacOSServiceController(settings)
    return WindowsServiceController(settings)


def prepare_install_directory(settings: Settings) -> None:
    install_dir = settings.install_dir
    if install_dir.is_symlink():
        raise RuntimeError(
            f"probe install directory cannot be a symlink: {install_dir}"
        )
    install_dir.mkdir(parents=True, exist_ok=True)
    if not install_dir.is_dir():
        raise RuntimeError(f"probe install path is not a directory: {install_dir}")
    if settings.system == "windows":
        harden_windows_installation(settings)
        return
    os.chown(install_dir, 0, 0)
    install_dir.chmod(0o755)


def secure_installation(settings: Settings) -> None:
    if settings.system == "windows":
        harden_windows_installation(settings)
        return
    for directory, directories, files in os.walk(settings.install_dir):
        current = Path(directory)
        if current.is_symlink():
            raise RuntimeError(f"installed probe contains a symlink: {current}")
        os.chown(current, 0, 0)
        current.chmod(0o755)
        for name in (*directories, *files):
            path = current / name
            if path.is_symlink():
                raise RuntimeError(f"installed probe contains a symlink: {path}")
            os.chown(path, 0, 0)
            path.chmod(0o755 if path.is_dir() or path.name == "run-probe" else 0o644)


def harden_windows_installation(settings: Settings) -> None:
    path = str(settings.install_dir)
    run(["icacls.exe", path, "/reset", "/T", "/C", "/Q"])
    run(
        [
            "icacls.exe",
            path,
            "/setowner",
            "*S-1-5-32-544",
            "/T",
            "/C",
            "/Q",
        ]
    )
    run(
        [
            "icacls.exe",
            path,
            "/inheritance:r",
            "/grant:r",
            "*S-1-5-18:(OI)(CI)F",
            "*S-1-5-32-544:(OI)(CI)F",
            f"{settings.user}:(OI)(CI)RX",
            "/T",
            "/C",
            "/Q",
        ]
    )
    run(["icacls.exe", path, "/verify", "/T", "/C", "/Q"])


def activate_runtime(runtime: Path, staging: Path) -> Path | None:
    backup = runtime.with_name("runtime.previous")
    remove_path(backup)
    if runtime.exists():
        os.replace(runtime, backup)
    try:
        os.replace(staging, runtime)
    except Exception:
        if backup.exists():
            os.replace(backup, runtime)
        raise
    return backup if backup.exists() else None


def restore_runtime(runtime: Path, backup: Path) -> None:
    remove_path(runtime)
    os.replace(backup, runtime)


def write_unix_runner(settings: Settings) -> Path:
    bootstrap = write_python_runner(settings)
    runner = settings.install_dir / "run-probe"
    atomic_write(
        runner,
        "#!/bin/sh\n" f"exec {probe_command(settings, bootstrap)}\n",
        mode=0o755,
    )
    return runner


def write_python_runner(settings: Settings) -> Path:
    runner = settings.install_dir / "run-probe.py"
    atomic_write(
        runner,
        "import runpy\n"
        "import sys\n\n"
        f"sys.path.insert(0, {json.dumps(str(settings.runtime_dir))})\n"
        f"sys.path.append({json.dumps(str(settings.runtime_dir / 'vendor'))})\n"
        'runpy.run_module("realtime_probe", run_name="__main__")\n',
    )
    return runner


def probe_arguments(settings: Settings) -> list[str]:
    arguments = [
        "--bind",
        settings.bind,
        "--port",
        str(settings.port),
        "--active-window-seconds",
        str(settings.active_window_seconds),
    ]
    if settings.codex_homes:
        arguments.extend(("--codex-homes", settings.codex_homes))
    if settings.claude_home:
        arguments.extend(("--claude-home", settings.claude_home))
    return arguments


def probe_command(settings: Settings, bootstrap: Path, windows: bool = False) -> str:
    command = [
        str(settings.python),
        "-I",
        "-S",
        "-B",
        "-X",
        "utf8",
        str(bootstrap),
        *probe_arguments(settings),
    ]
    return subprocess.list2cmdline(command) if windows else shlex.join(command)


def remove_legacy_payload(install_dir: Path) -> None:
    for name in (
        "status_common.py",
        "status-device-reporter.py",
        "agent-status-reporter.py",
        "node_exporter",
    ):
        path = install_dir / name
        remove_path(path)


def remove_obsolete_probe_install(settings: Settings) -> None:
    obsolete: Path | None = None
    if settings.system == "linux":
        obsolete = Path("/opt/realtime-me")
    elif settings.system == "darwin" and pwd is not None:
        obsolete = Path(pwd.getpwnam(settings.user).pw_dir) / ".realtime-me"
    elif settings.system == "windows":
        local_app_data = os.getenv("LOCALAPPDATA", "").strip()
        if local_app_data:
            obsolete = Path(local_app_data) / "RealtimeMe"
    if obsolete is None or obsolete == settings.install_dir:
        return

    for name in ("runtime", "runtime.previous"):
        runtime = obsolete / name
        if is_probe_runtime(runtime):
            remove_path(runtime)
    for name in ("run-probe", "run-probe.cmd"):
        runner = obsolete / name
        if is_probe_runner(runner):
            runner.unlink(missing_ok=True)


def is_probe_runtime(path: Path) -> bool:
    return (path / "realtime_probe" / "__main__.py").is_file()


def is_probe_runner(path: Path) -> bool:
    if not path.is_file():
        return False
    try:
        return "-m realtime_probe" in path.read_text(encoding="utf-8")
    except (OSError, UnicodeError):
        return False


def remove_path(path: Path) -> None:
    if path.is_symlink() or path.is_file():
        path.unlink(missing_ok=True)
    elif path.is_dir():
        shutil.rmtree(path)


def parse_integrity_manifest(payload: bytes) -> dict[PurePosixPath, str]:
    try:
        document = json.loads(payload.decode("utf-8"), object_pairs_hook=unique_object)
    except (UnicodeError, json.JSONDecodeError) as error:
        raise RuntimeError("invalid probe integrity manifest") from error
    if not isinstance(document, dict) or set(document) != {"version", "files"}:
        raise RuntimeError("invalid probe integrity manifest fields")
    if type(document["version"]) is not int or document["version"] != 1:
        raise RuntimeError("unsupported probe integrity manifest version")
    raw_files = document["files"]
    if not isinstance(raw_files, dict) or not raw_files:
        raise RuntimeError("probe integrity manifest has no files")

    files: dict[PurePosixPath, str] = {}
    for name, digest in raw_files.items():
        if not isinstance(name, str) or not isinstance(digest, str):
            raise RuntimeError("invalid probe integrity manifest entry")
        path = safe_manifest_path(name)
        if path != PurePosixPath("requirements.txt") and (
            path.parts[0] != "realtime_probe" or path.suffix != ".py"
        ):
            raise RuntimeError(f"unexpected probe runtime file: {path}")
        if SHA256_PATTERN.fullmatch(digest) is None:
            raise RuntimeError(f"invalid SHA-256 for {path}")
        if path in files:
            raise RuntimeError(f"duplicate probe runtime path: {path}")
        files[path] = digest

    required = {
        PurePosixPath("requirements.txt"),
        PurePosixPath("realtime_probe/__init__.py"),
        PurePosixPath("realtime_probe/__main__.py"),
    }
    if not required.issubset(files):
        raise RuntimeError("probe integrity manifest is incomplete")
    return dict(sorted(files.items(), key=lambda item: item[0].as_posix()))


def unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if key in result:
            raise RuntimeError(f"duplicate JSON field: {key}")
        result[key] = value
    return result


def safe_manifest_path(value: str) -> PurePosixPath:
    normalized = value.strip()
    path = PurePosixPath(normalized)
    if (
        not normalized
        or normalized != value
        or "\\" in normalized
        or ":" in normalized
        or not path.parts
        or path.as_posix() != normalized
        or path.is_absolute()
        or ".." in path.parts
    ):
        raise RuntimeError(f"invalid probe manifest path: {value}")
    return path


def make_readable(root: Path) -> None:
    for directory, directories, files in os.walk(root):
        current = Path(directory)
        if current.is_symlink():
            raise RuntimeError(f"downloaded probe contains a symlink: {current}")
        current.chmod(0o755)
        for name in directories:
            path = current / name
            if path.is_symlink():
                raise RuntimeError(f"downloaded probe contains a symlink: {path}")
            path.chmod(0o755)
        for name in files:
            path = current / name
            if path.is_symlink():
                raise RuntimeError(f"downloaded probe contains a symlink: {path}")
            path.chmod(0o644)


def atomic_write(
    path: Path,
    content: str | bytes,
    *,
    mode: int = 0o644,
    encoding: str = "utf-8",
    newline: str | None = None,
) -> None:
    payload = content
    if isinstance(payload, str):
        if newline is not None:
            payload = payload.replace("\n", newline)
        payload = payload.encode(encoding)
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, temporary_name = tempfile.mkstemp(
        prefix=f".{path.name}.", dir=path.parent
    )
    temporary = Path(temporary_name)
    try:
        with os.fdopen(descriptor, "wb") as file:
            file.write(payload)
            file.flush()
            os.fsync(file.fileno())
        temporary.chmod(mode)
        os.replace(temporary, path)
    finally:
        temporary.unlink(missing_ok=True)


def validated_base_url(value: str) -> str:
    normalized = value.rstrip("/")
    parsed = urllib.parse.urlsplit(normalized)
    if (
        parsed.scheme != "https"
        or parsed.hostname is None
        or parsed.username is not None
        or parsed.password is not None
        or parsed.query
        or parsed.fragment
    ):
        raise RuntimeError(f"probe source must be an HTTPS base URL: {value}")
    return normalized


def default_service_user(system: str) -> str:
    if system == "windows":
        return windows_identity()
    return os.getenv("SUDO_USER", "").strip() or getpass.getuser()


def windows_is_administrator() -> bool:
    if platform.system().lower() != "windows":
        return False
    try:
        import ctypes

        return bool(ctypes.windll.shell32.IsUserAnAdmin())
    except (AttributeError, OSError):
        return False


def detect_lan_address() -> str:
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as connection:
            connection.connect(("192.0.2.1", 80))
            address = str(connection.getsockname()[0])
            if is_private_address(address):
                return address
    except OSError:
        pass
    try:
        return next(
            address
            for address in socket.gethostbyname_ex(socket.gethostname())[2]
            if is_private_address(address)
        )
    except (OSError, StopIteration):
        return ""


def is_private_address(value: str) -> bool:
    try:
        address = ipaddress.ip_address(value.strip("[]"))
    except ValueError:
        return False
    return any(address in network for network in PRIVATE_NETWORKS)


def windows_identity() -> str:
    domain = os.getenv("USERDOMAIN", "").strip()
    username = os.getenv("USERNAME", "").strip() or getpass.getuser()
    return f"{domain}\\{username}" if domain else username


def ps_literal(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def encoded_powershell(value: str) -> str:
    return base64.b64encode(value.encode("utf-16-le")).decode("ascii")


def shell_word(value: str) -> str:
    return shlex.quote(value)


def url_host(value: str) -> str:
    return f"[{value}]" if ":" in value and not value.startswith("[") else value


def port(value: str) -> int:
    number = int(value)
    if not 1 <= number <= 65535:
        raise argparse.ArgumentTypeError("port must be between 1 and 65535")
    return number


def positive_integer(value: str) -> int:
    number = int(value)
    if number <= 0:
        raise argparse.ArgumentTypeError("value must be positive")
    return number


def positive_float(value: str) -> float:
    number = float(value)
    if not math.isfinite(number) or number <= 0:
        raise argparse.ArgumentTypeError("value must be a positive finite number")
    return number


def require_python_version() -> None:
    if sys.version_info < MINIMUM_PYTHON:
        version = ".".join(map(str, MINIMUM_PYTHON))
        raise SystemExit(f"Python {version} or newer is required")


def require_command(command: str) -> None:
    if shutil.which(command) is None:
        raise RuntimeError(f"missing required command: {command}")


def run(command: list[str]) -> None:
    log("Running " + shell_word(command[0]))
    subprocess.run(command, check=True, timeout=COMMAND_TIMEOUT_SECONDS)


def log(message: str) -> None:
    print(f"[realtime-me] {message}", file=sys.stderr)


if __name__ == "__main__":
    raise SystemExit(main())
