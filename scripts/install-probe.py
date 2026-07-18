#!/usr/bin/env python3
"""Install the unified Realtime Me probe on Linux, macOS, or Windows."""

from __future__ import annotations

import argparse
import base64
import getpass
import math
import os
import platform
import plistlib
import shlex
import shutil
import socket
import subprocess
import sys
import tempfile
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path, PurePosixPath

try:
    import pwd
except ImportError:  # Windows has no POSIX account database.
    pwd = None

MINIMUM_PYTHON = (3, 10)
DEFAULT_PORT = 18082
DEFAULT_BASE_URLS = (
    "https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts",
    "https://raw.githubusercontent.com/pood1e/realtime-me/main/scripts",
)
LINUX_SERVICE = "realtime-me-probe.service"
MACOS_LABEL = "me.realtime.probe"
WINDOWS_TASK = "Realtime Me Probe"


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

    @property
    def python_path(self) -> str:
        separator = ";" if self.system == "windows" else ":"
        return separator.join((str(self.runtime_dir), str(self.runtime_dir / "vendor")))


class Installer:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings

    def install(self) -> None:
        staging = self._stage_runtime()
        controller = service_controller(self.settings)
        backup: Path | None = None
        try:
            controller.stop()
            controller.remove_legacy()
            backup = activate_runtime(self.settings.runtime_dir, staging)
            controller.configure()
            controller.start()
        except Exception:
            if backup is not None and backup.exists():
                controller.stop()
                restore_runtime(self.settings.runtime_dir, backup)
                controller.configure()
                controller.start()
            raise
        finally:
            if staging.exists():
                shutil.rmtree(staging, ignore_errors=True)
        if backup is not None:
            shutil.rmtree(backup, ignore_errors=True)
        self._print_summary()

    def _stage_runtime(self) -> Path:
        self.settings.install_dir.mkdir(parents=True, exist_ok=True)
        staging = Path(
            tempfile.mkdtemp(prefix=".runtime-", dir=self.settings.install_dir)
        )
        try:
            manifest = self._download("probe/manifest.txt").decode().splitlines()
            files = [safe_manifest_path(line) for line in manifest if line.strip()]
            if not files:
                raise RuntimeError("probe manifest is empty")
            requirements = staging / "requirements.txt"
            requirements.write_bytes(self._download("probe/requirements.txt"))
            for relative in files:
                destination = staging / relative
                destination.parent.mkdir(parents=True, exist_ok=True)
                destination.write_bytes(self._download(f"probe/{relative.as_posix()}"))
            vendor = staging / "vendor"
            log("Installing pinned Python dependencies")
            run(
                [
                    str(self.settings.python),
                    "-m",
                    "pip",
                    "install",
                    "--disable-pip-version-check",
                    "--no-input",
                    "--upgrade",
                    "--target",
                    str(vendor),
                    "--requirement",
                    str(requirements),
                ]
            )
            make_readable(staging)
            return staging
        except Exception:
            shutil.rmtree(staging, ignore_errors=True)
            raise

    def _download(self, relative: str) -> bytes:
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
                    return response.read()
            except (OSError, urllib.error.URLError) as error:
                last_error = error
                log(f"Download mirror failed: {base_url}")
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
            )
            (Path("/etc/systemd/system") / service).unlink(missing_ok=True)
        remove_legacy_payload(self.settings.install_dir)
        run(["systemctl", "daemon-reload"])

    def configure(self) -> None:
        runner = write_unix_runner(self.settings)
        directives = ""
        if self.settings.user != "root":
            if pwd is None:
                raise RuntimeError("POSIX account database is unavailable")
            user = pwd.getpwnam(self.settings.user)
            directives = (
                f"User={self.settings.user}\n"
                f"Environment=XDG_RUNTIME_DIR=/run/user/{user.pw_uid}\n"
                f"Environment=DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/{user.pw_uid}/bus\n"
            )
        self.unit.write_text(
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
            "NoNewPrivileges=true\n"
            "PrivateTmp=true\n\n"
            "[Install]\n"
            "WantedBy=multi-user.target\n"
        )
        run(["systemctl", "daemon-reload"])

    def start(self) -> None:
        run(["systemctl", "enable", "--now", LINUX_SERVICE])


class MacOSServiceController:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        self.domain = f"gui/{os.getuid()}"
        self.agents_dir = Path.home() / "Library" / "LaunchAgents"
        self.plist = self.agents_dir / f"{MACOS_LABEL}.plist"

    def stop(self) -> None:
        subprocess.run(
            ["launchctl", "bootout", f"{self.domain}/{MACOS_LABEL}"],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
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
            )
            (self.agents_dir / f"{label}.plist").unlink(missing_ok=True)
        remove_legacy_payload(self.settings.install_dir)

    def configure(self) -> None:
        runner = write_unix_runner(self.settings)
        self.agents_dir.mkdir(parents=True, exist_ok=True)
        payload = {
            "Label": MACOS_LABEL,
            "ProgramArguments": [str(runner)],
            "KeepAlive": True,
            "RunAtLoad": True,
            "StandardErrorPath": str(self.settings.install_dir / "probe.log"),
            "StandardOutPath": str(self.settings.install_dir / "probe.log"),
        }
        with self.plist.open("wb") as file:
            plistlib.dump(payload, file, sort_keys=False)

    def start(self) -> None:
        run(["launchctl", "bootstrap", self.domain, str(self.plist)])
        subprocess.run(
            ["launchctl", "enable", f"{self.domain}/{MACOS_LABEL}"], check=False
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
        )
        subprocess.run(
            ["schtasks.exe", "/Delete", "/F", "/TN", WINDOWS_TASK],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )

    def remove_legacy(self) -> None:
        remove_legacy_payload(self.settings.install_dir)

    def configure(self) -> None:
        command_arguments = probe_arguments(self.settings, windows=True)
        self.runner.write_text(
            "@echo off\n"
            "chcp 65001 >nul\n"
            "set PYTHONUTF8=1\n"
            f'set "PYTHONPATH={self.settings.python_path}"\n'
            f'"{self.settings.python}" -m realtime_probe {command_arguments}\n'
            "exit /b %ERRORLEVEL%\n",
            encoding="utf-8",
            newline="\r\n",
        )
        identity = windows_identity()
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
    except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
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
    parser.add_argument("--bind", default=os.getenv("REALTIME_PROBE_BIND", "0.0.0.0"))
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
        default=os.getenv(
            "REALTIME_PROBE_USER", os.getenv("SUDO_USER", "") or getpass.getuser()
        ),
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
    return Settings(
        system=system,
        install_dir=Path(args.install_dir).expanduser().resolve(),
        python=Path(sys.executable).resolve(),
        bind=args.bind,
        port=args.port,
        active_window_seconds=args.active_window_seconds,
        codex_homes=args.codex_homes,
        claude_home=args.claude_home,
        user=args.user,
        device_name=args.name,
        exporter_host=args.host or detect_lan_address(),
        device_kind=args.kind,
        device_role=args.role,
        timeout_seconds=args.timeout_seconds,
        base_urls=base_urls,
    )


def configured_base_urls() -> tuple[str, ...]:
    single = os.getenv("REALTIME_ME_RAW_BASE_URL", "").strip()
    if single:
        return (single,)
    multiple = tuple(os.getenv("REALTIME_ME_RAW_BASE_URLS", "").split())
    return multiple or DEFAULT_BASE_URLS


def default_install_dir(system: str) -> str:
    if system == "linux":
        return "/opt/realtime-me"
    if system == "windows":
        return str(Path(os.getenv("LOCALAPPDATA", str(Path.home()))) / "RealtimeMe")
    return str(Path.home() / ".realtime-me")


def validate_environment(settings: Settings) -> None:
    if settings.system not in {"linux", "darwin", "windows"}:
        raise RuntimeError(f"unsupported operating system: {platform.system()}")
    if settings.system == "linux":
        if os.geteuid() != 0:
            raise RuntimeError("Linux installation must run as root (use sudo)")
        if pwd is None:
            raise RuntimeError("POSIX account database is unavailable")
        try:
            pwd.getpwnam(settings.user)
        except KeyError as error:
            raise RuntimeError(
                f"Linux probe user does not exist: {settings.user}"
            ) from error
        require_command("systemctl")
    elif settings.system == "darwin":
        if os.geteuid() == 0:
            raise RuntimeError(
                "macOS installation must run as the logged-in user, not sudo"
            )
        require_command("launchctl")
    else:
        require_command("powershell.exe")
        require_command("schtasks.exe")
    if not settings.exporter_host:
        raise RuntimeError("could not detect a LAN address; pass --host")
    run([str(settings.python), "-m", "pip", "--version"])


def service_controller(
    settings: Settings,
) -> LinuxServiceController | MacOSServiceController | WindowsServiceController:
    if settings.system == "linux":
        return LinuxServiceController(settings)
    if settings.system == "darwin":
        return MacOSServiceController(settings)
    return WindowsServiceController(settings)


def activate_runtime(runtime: Path, staging: Path) -> Path | None:
    backup = runtime.with_name("runtime.previous")
    shutil.rmtree(backup, ignore_errors=True)
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
    shutil.rmtree(runtime, ignore_errors=True)
    os.replace(backup, runtime)


def write_unix_runner(settings: Settings) -> Path:
    runner = settings.install_dir / "run-probe"
    runner.write_text(
        "#!/bin/sh\n"
        f"export PYTHONPATH={shlex.quote(settings.python_path)}\n"
        f"exec {shlex.quote(str(settings.python))} -m realtime_probe {probe_arguments(settings)}\n"
    )
    runner.chmod(0o755)
    return runner


def probe_arguments(settings: Settings, windows: bool = False) -> str:
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
    if windows:
        return subprocess.list2cmdline(arguments)
    return shlex.join(arguments)


def remove_legacy_payload(install_dir: Path) -> None:
    for name in (
        "status_common.py",
        "status-device-reporter.py",
        "agent-status-reporter.py",
        "node_exporter",
    ):
        path = install_dir / name
        if path.is_dir():
            shutil.rmtree(path, ignore_errors=True)
        else:
            path.unlink(missing_ok=True)


def safe_manifest_path(value: str) -> PurePosixPath:
    normalized = value.strip()
    path = PurePosixPath(normalized)
    if (
        not normalized
        or "\\" in normalized
        or ":" in normalized
        or path.is_absolute()
        or ".." in path.parts
    ):
        raise RuntimeError(f"invalid probe manifest path: {value}")
    return path


def make_readable(root: Path) -> None:
    for directory, directories, files in os.walk(root):
        Path(directory).chmod(0o755)
        for name in directories:
            (Path(directory) / name).chmod(0o755)
        for name in files:
            (Path(directory) / name).chmod(0o644)


def detect_lan_address() -> str:
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as connection:
            connection.connect(("192.0.2.1", 80))
            address = str(connection.getsockname()[0])
            if address and not address.startswith("127."):
                return address
    except OSError:
        pass
    try:
        return next(
            address
            for address in socket.gethostbyname_ex(socket.gethostname())[2]
            if not address.startswith("127.")
        )
    except (OSError, StopIteration):
        return ""


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
    subprocess.run(command, check=True)


def log(message: str) -> None:
    print(f"[realtime-me] {message}", file=sys.stderr)


if __name__ == "__main__":
    raise SystemExit(main())
