#!/usr/bin/env python3
from __future__ import annotations

import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
import re
import socket
import sqlite3
import subprocess
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

UUID_PATTERN = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", re.IGNORECASE)
URL_PATTERN = re.compile(r"https?://\S+")
EMAIL_PATTERN = re.compile(r"\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b", re.IGNORECASE)
TOKEN_PATTERN = re.compile(r"\b[A-Za-z0-9][A-Za-z0-9_-]{31,}\b")
CODEX_FAILED_GOAL_STATES = {"blocked", "usage_limited", "budget_limited"}
CODEX_RUNNING_GOAL_STATES = {"active"}
CLAUDE_OPEN_TASK_STATES = {"in_progress", "pending"}


@dataclass(frozen=True)
class AgentSnapshot:
    agent_id: str
    state: str
    device_id: str = ""
    device_name: str = ""
    task: str = ""
    updated_at: str = ""
    budget_remaining_percent: int | None = None

    def payload(self) -> dict[str, Any]:
        data: dict[str, Any] = {
            "agent_id": self.agent_id,
            "state": self.state,
        }
        if self.device_id:
            data["device_id"] = self.device_id
        if self.device_name:
            data["device_name"] = self.device_name
        if self.task:
            data["task"] = self.task
        if self.updated_at:
            data["updated_at"] = self.updated_at
        if self.budget_remaining_percent is not None:
            data["budget_remaining_percent"] = self.budget_remaining_percent
        return data


@dataclass(frozen=True)
class CodexThread:
    thread_id: str
    title: str
    updated_at_seconds: float


@dataclass(frozen=True)
class CodexGoal:
    thread_id: str
    objective: str
    status: str
    updated_at_seconds: float
    budget_remaining_percent: int | None


@dataclass(frozen=True)
class CodexCandidate:
    thread_id: str
    task: str
    state: str
    updated_at_seconds: float
    budget_remaining_percent: int | None


@dataclass(frozen=True)
class CodexActivity:
    working: bool
    updated_at_seconds: float


@dataclass(frozen=True)
class ClaudeSession:
    session_id: str
    cli_session_id: str
    title: str
    status: str
    updated_at_seconds: float
    archived: bool


@dataclass(frozen=True)
class ClaudeTask:
    title: str
    status: str
    updated_at_seconds: float


@dataclass(frozen=True)
class DeviceIdentity:
    device_id: str
    device_name: str


def main() -> int:
    args = parse_args()
    if args.serve:
        return serve(args)

    snapshots = build_snapshots(args)
    if args.print:
        print(json.dumps([snapshot.payload() for snapshot in snapshots], indent=2, ensure_ascii=False))
        return 0

    token = args.token or os.getenv("STATUS_INGEST_TOKEN")
    if not token:
        print("STATUS_INGEST_TOKEN is required", file=sys.stderr)
        return 2

    endpoint = args.url.rstrip("/") + "/api/ingest/agent"
    for snapshot in snapshots:
        if not post_agent(endpoint, token, snapshot.payload(), args.timeout_seconds):
            return 1
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Publish local Claude Code and Codex task titles to realtime-me gateway.")
    parser.add_argument("--url", default=os.getenv("STATUS_GATEWAY_URL", "http://127.0.0.1:18080"))
    parser.add_argument("--token", default="")
    parser.add_argument("--device-id", default=os.getenv("STATUS_DEVICE_ID", socket.gethostname()))
    parser.add_argument("--device-name", default=os.getenv("STATUS_DEVICE_NAME", socket.gethostname()))
    parser.add_argument("--timeout-seconds", type=float, default=5)
    parser.add_argument("--active-window-seconds", type=int, default=int(os.getenv("STATUS_AGENT_ACTIVE_WINDOW_SECONDS", "300")))
    parser.add_argument("--codex-homes", default=os.getenv("STATUS_CODEX_HOMES", "~/.codex-api:~/.codex"))
    parser.add_argument(
        "--claude-sessions-dir",
        default=os.getenv("STATUS_CLAUDE_SESSIONS_DIR", "~/Library/Application Support/Claude/claude-code-sessions"),
    )
    parser.add_argument("--claude-tasks-dir", default=os.getenv("STATUS_CLAUDE_TASKS_DIR", "~/.claude/tasks"))
    parser.add_argument("--claude-jobs-dir", default=os.getenv("STATUS_CLAUDE_JOBS_DIR", "~/.claude/jobs"))
    parser.add_argument("--serve", action="store_true")
    parser.add_argument("--bind", default=os.getenv("STATUS_AGENT_EXPORTER_BIND", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("STATUS_AGENT_EXPORTER_PORT", "18082")))
    parser.add_argument("--print", action="store_true")
    return parser.parse_args()


def build_snapshots(args: argparse.Namespace) -> list[AgentSnapshot]:
    now = time.time()
    device = DeviceIdentity(args.device_id, args.device_name)
    codex_homes = [expand_path(value) for value in re.split(r"[:;,]", args.codex_homes) if value.strip()]
    return [
        *codex_snapshots(codex_homes, now, args.active_window_seconds, device),
        claude_snapshot(
            expand_path(args.claude_sessions_dir),
            expand_path(args.claude_tasks_dir),
            expand_path(args.claude_jobs_dir),
            now,
            args.active_window_seconds,
            device,
        ),
    ]


def serve(args: argparse.Namespace) -> int:
    class Handler(BaseHTTPRequestHandler):
        def do_GET(self) -> None:
            if self.path == "/healthz":
                self.write_json({"ok": True})
                return
            if self.path == "/api/agent-status":
                self.write_json([snapshot.payload() for snapshot in build_snapshots(args)])
                return
            self.write_json({"error": "not_found"}, 404)

        def log_message(self, _format: str, *_args: object) -> None:
            return

        def write_json(self, payload: object, status: int = 200) -> None:
            data = json.dumps(payload, ensure_ascii=False).encode()
            self.send_response(status)
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

    server = ThreadingHTTPServer((args.bind, args.port), Handler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        return 130
    return 0


def codex_snapshots(homes: list[Path], now: float, active_window_seconds: int, device: DeviceIdentity) -> list[AgentSnapshot]:
    open_sessions = codex_open_sessions(homes)
    candidates = codex_candidates(homes, open_sessions, now, active_window_seconds)
    if not candidates:
        return [AgentSnapshot(agent_id="codex", state="idle", device_id=device.device_id, device_name=device.device_name, updated_at=utc_now())]

    candidates.sort(key=lambda item: (item.state != "idle", item.updated_at_seconds), reverse=True)
    return [codex_agent_snapshot(candidate, device) for candidate in candidates]


def codex_agent_snapshot(candidate: CodexCandidate, device: DeviceIdentity) -> AgentSnapshot:
    return AgentSnapshot(
        agent_id=f"codex:{candidate.thread_id[:8]}",
        state=candidate.state,
        device_id=device.device_id,
        device_name=device.device_name,
        task=sanitize_task(candidate.task),
        updated_at=iso_time(candidate.updated_at_seconds),
        budget_remaining_percent=candidate.budget_remaining_percent,
    )


def codex_candidates(homes: list[Path], open_sessions: dict[str, Path], now: float, active_window_seconds: int) -> list[CodexCandidate]:
    threads: dict[str, CodexThread] = {}
    goals: dict[str, CodexGoal] = {}
    for home in homes:
        threads.update(read_codex_threads(home / "state_5.sqlite"))
        goals.update(read_codex_goals(home / "goals_1.sqlite"))

    candidates: list[CodexCandidate] = []
    for thread_id, session_path in open_sessions.items():
        thread = threads.get(thread_id)
        goal = goals.get(thread_id)
        if not thread and not goal:
            continue
        activity = codex_activity(session_path, now, active_window_seconds)
        updated_at = max(
            thread.updated_at_seconds if thread else 0,
            goal.updated_at_seconds if goal else 0,
            activity.updated_at_seconds,
        )
        state = codex_state(goal.status if goal else "", activity.working)
        task = goal.objective if goal and goal.status in CODEX_RUNNING_GOAL_STATES | CODEX_FAILED_GOAL_STATES else thread.title if thread else goal.objective
        candidates.append(CodexCandidate(thread_id, task, state, updated_at, goal.budget_remaining_percent if goal else None))

    return candidates


def codex_state(goal_status: str, active: bool) -> str:
    if goal_status in CODEX_FAILED_GOAL_STATES:
        return "failed"
    if active or goal_status in CODEX_RUNNING_GOAL_STATES:
        return "running"
    return "idle"


def read_codex_threads(database: Path) -> dict[str, CodexThread]:
    if not database.exists():
        return {}
    rows = query_sqlite(
        database,
        "select id, title, updated_at from threads order by updated_at desc limit 50",
    )
    threads = {}
    for row in rows:
        thread_id = str(row["id"])
        threads[thread_id] = CodexThread(
            thread_id=thread_id,
            title=str(row["title"] or ""),
            updated_at_seconds=float(row["updated_at"] or 0),
        )
    return threads


def read_codex_goals(database: Path) -> dict[str, CodexGoal]:
    if not database.exists():
        return {}
    rows = query_sqlite(
        database,
        "select thread_id, objective, status, token_budget, tokens_used, updated_at_ms from thread_goals order by updated_at_ms desc limit 50",
    )
    goals = {}
    for row in rows:
        thread_id = str(row["thread_id"])
        goals[thread_id] = CodexGoal(
            thread_id=thread_id,
            objective=str(row["objective"] or ""),
            status=str(row["status"] or ""),
            updated_at_seconds=float(row["updated_at_ms"] or 0) / 1000,
            budget_remaining_percent=budget_remaining(row["token_budget"], row["tokens_used"]),
        )
    return goals


def codex_open_sessions(homes: list[Path]) -> dict[str, Path]:
    session_paths = codex_open_session_paths(homes)
    sessions = {}
    for path in session_paths:
        match = UUID_PATTERN.search(path.name)
        if match:
            sessions[match.group(0)] = path
    return sessions


def codex_activity(session_path: Path, now: float, active_window_seconds: int) -> CodexActivity:
    fallback = session_path.stat().st_mtime
    for line in reversed(tail_lines(session_path)):
        event = decode_json(line)
        if not event:
            continue
        timestamp = parse_event_timestamp(event.get("timestamp"), fallback)
        payload = event.get("payload") if isinstance(event.get("payload"), dict) else {}
        payload_type = str(payload.get("type") or "")
        event_type = str(event.get("type") or "")
        if payload_type == "token_count":
            continue
        if payload_type in {"task_complete", "agent_message"}:
            return CodexActivity(False, timestamp)
        if event_type == "response_item" and payload_type == "message":
            return CodexActivity(False, timestamp)
        if payload_type == "function_call":
            return CodexActivity(True, timestamp)
        if payload_type in {"function_call_output", "reasoning", "user_message"}:
            return CodexActivity(now - timestamp <= active_window_seconds, timestamp)
    return CodexActivity(False, fallback)


def codex_open_session_paths(homes: list[Path]) -> set[Path]:
    process_ids = codex_process_ids()
    session_paths = set()
    for process_id in process_ids:
        for path in open_files(process_id):
            if path.suffix != ".jsonl" or "/sessions/" not in path.as_posix():
                continue
            if any(is_relative_to(path, home) for home in homes):
                session_paths.add(path)
    return session_paths


def codex_process_ids() -> list[int]:
    output = run(["ps", "ax", "-o", "pid=,args="])
    process_ids = []
    for line in output.splitlines():
        match = re.match(r"\s*(\d+)\s+(.+)$", line)
        if not match:
            continue
        args = match.group(2).lower()
        if "codex" not in args or "agent-status-reporter" in args:
            continue
        process_ids.append(int(match.group(1)))
    return process_ids


def open_files(process_id: int) -> list[Path]:
    output = run(["lsof", "-Fn", "-p", str(process_id)])
    paths = []
    for line in output.splitlines():
        if not line.startswith("n/"):
            continue
        paths.append(Path(line[1:]))
    return paths


def claude_snapshot(sessions_dir: Path, tasks_dir: Path, jobs_dir: Path, now: float, active_window_seconds: int, device: DeviceIdentity) -> AgentSnapshot:
    sessions = read_claude_sessions(sessions_dir)
    job = latest_claude_job(jobs_dir)
    if not sessions and job is None:
        return AgentSnapshot(agent_id="claude-code", state="idle", device_id=device.device_id, device_name=device.device_name, updated_at=utc_now())

    sessions.sort(key=lambda item: item.updated_at_seconds, reverse=True)
    session = sessions[0] if sessions else None
    task = claude_open_task(tasks_dir / session.cli_session_id) if session and session.cli_session_id else None
    updated_at = max(
        session.updated_at_seconds if session else 0,
        task.updated_at_seconds if task else 0,
        job.updated_at_seconds if job else 0,
    )
    active = claude_active(session, task, job, now, active_window_seconds)
    return AgentSnapshot(
        agent_id="claude-code",
        state="running" if active else "idle",
        device_id=device.device_id,
        device_name=device.device_name,
        task=sanitize_task((task.title if task else "") or (job.title if job else "") or (session.title if session else "")),
        updated_at=iso_time(updated_at),
    )


def read_claude_sessions(sessions_dir: Path) -> list[ClaudeSession]:
    if not sessions_dir.exists():
        return []
    sessions: list[ClaudeSession] = []
    for path in sorted(sessions_dir.rglob("*.json"), key=lambda item: item.stat().st_mtime, reverse=True)[:100]:
        data = read_json_object(path)
        if not data:
            continue
        title = str(data.get("title") or data.get("name") or "")
        session_id = str(data.get("sessionId") or path.stem)
        cli_session_id = str(data.get("cliSessionId") or data.get("sessionId") or "")
        if not title and not session_id:
            continue
        updated_at = timestamp_seconds(data.get("lastActivityAt") or data.get("lastFocusedAt") or data.get("updatedAt") or data.get("statusUpdatedAt") or path.stat().st_mtime)
        sessions.append(
            ClaudeSession(
                session_id=session_id,
                cli_session_id=cli_session_id,
                title=title or session_id[:8],
                status=str(data.get("status") or ""),
                updated_at_seconds=updated_at,
                archived=bool(data.get("isArchived")),
            )
        )
    return [session for session in sessions if not session.archived]


def claude_open_task(task_dir: Path) -> ClaudeTask | None:
    if not task_dir.exists():
        return None
    candidates: list[ClaudeTask] = []
    for path in sorted(task_dir.glob("*.json"), key=lambda item: item.stat().st_mtime, reverse=True)[:50]:
        data = read_json_object(path)
        status = str(data.get("status") or "")
        if not data or status not in CLAUDE_OPEN_TASK_STATES:
            continue
        title = str(data.get("activeForm") or data.get("subject") or "")
        if title:
            candidates.append(ClaudeTask(title, status, path.stat().st_mtime))
    candidates.sort(key=lambda item: item.updated_at_seconds, reverse=True)
    return candidates[0] if candidates else None


def latest_claude_job(jobs_dir: Path) -> ClaudeTask | None:
    if not jobs_dir.exists():
        return None
    candidates: list[ClaudeTask] = []
    for path in sorted(jobs_dir.glob("*/state.json"), key=lambda item: item.stat().st_mtime, reverse=True)[:20]:
        data = read_json_object(path)
        if not data:
            continue
        title = str(data.get("name") or data.get("intent") or data.get("sessionId") or "")
        if not title:
            continue
        state = str(data.get("state") or "")
        in_flight = "running" if data.get("inFlight") is True else state
        candidates.append(ClaudeTask(title, in_flight, timestamp_seconds(data.get("updatedAt") or path.stat().st_mtime)))
    candidates.sort(key=lambda item: item.updated_at_seconds, reverse=True)
    return candidates[0] if candidates else None


def claude_active(session: ClaudeSession | None, task: ClaudeTask | None, job: ClaudeTask | None, now: float, active_window_seconds: int) -> bool:
    if task and task.status in CLAUDE_OPEN_TASK_STATES and now - task.updated_at_seconds <= active_window_seconds:
        return True
    if job and job.status in {"running", "working", "busy", "in_progress"} and now - job.updated_at_seconds <= active_window_seconds:
        return True
    if not session:
        return False
    if session.status in {"running", "working", "busy", "in_progress"} and now - session.updated_at_seconds <= active_window_seconds:
        return True
    return False


def query_sqlite(database: Path, statement: str) -> list[sqlite3.Row]:
    try:
        connection = sqlite3.connect(f"file:{database}?mode=ro", uri=True, timeout=0.5)
        connection.row_factory = sqlite3.Row
        connection.execute("pragma busy_timeout=500")
        try:
            return list(connection.execute(statement))
        finally:
            connection.close()
    except sqlite3.Error:
        return []


def read_json_object(path: Path) -> dict[str, Any]:
    try:
        data = json.loads(path.read_text(errors="ignore"))
    except (OSError, json.JSONDecodeError):
        return {}
    return data if isinstance(data, dict) else {}


def decode_json(line: str) -> dict[str, Any]:
    try:
        data = json.loads(line)
    except json.JSONDecodeError:
        return {}
    return data if isinstance(data, dict) else {}


def tail_lines(path: Path, max_bytes: int = 131_072) -> list[str]:
    try:
        with path.open("rb") as file:
            file.seek(0, os.SEEK_END)
            size = file.tell()
            file.seek(max(0, size - max_bytes))
            return file.read().decode(errors="ignore").splitlines()
    except OSError:
        return []


def post_agent(endpoint: str, token: str, payload: dict[str, Any], timeout_seconds: float) -> bool:
    request = urllib.request.Request(
        endpoint,
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
                print(f"gateway rejected agent status: HTTP {response.status}", file=sys.stderr)
                return False
    except urllib.error.HTTPError as error:
        print(f"gateway rejected agent status: HTTP {error.code}", file=sys.stderr)
        return False
    except OSError as error:
        print(f"gateway agent status push failed: {error.__class__.__name__}", file=sys.stderr)
        return False
    return True


def sanitize_task(value: str) -> str:
    text = re.sub(r"[\x00-\x1f\x7f]", " ", value)
    text = URL_PATTERN.sub("[link]", text)
    text = EMAIL_PATTERN.sub("[email]", text)
    text = TOKEN_PATTERN.sub("[redacted]", text)
    text = re.sub(r"\s+", " ", text).strip()
    if len(text) <= 120:
        return text
    return text[:119].rstrip() + "…"


def budget_remaining(token_budget: Any, tokens_used: Any) -> int | None:
    try:
        budget = int(token_budget)
        used = int(tokens_used or 0)
    except (TypeError, ValueError):
        return None
    if budget <= 0:
        return None
    remaining = round(100 - used * 100 / budget)
    return max(0, min(100, remaining))


def timestamp_seconds(value: Any) -> float:
    try:
        numeric = float(value)
    except (TypeError, ValueError):
        return 0
    return numeric / 1000 if numeric > 10_000_000_000 else numeric


def parse_event_timestamp(value: Any, fallback: float) -> float:
    if not isinstance(value, str):
        return fallback
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00")).timestamp()
    except ValueError:
        return fallback


def iso_time(timestamp: float) -> str:
    if timestamp <= 0:
        return utc_now()
    return datetime.fromtimestamp(timestamp, timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def utc_now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def expand_path(value: str) -> Path:
    return Path(value.strip()).expanduser().resolve()


def is_relative_to(path: Path, parent: Path) -> bool:
    try:
        path.resolve().relative_to(parent.resolve())
        return True
    except ValueError:
        return False


def run(command: list[str]) -> str:
    try:
        return subprocess.run(command, check=False, capture_output=True, text=True, timeout=2).stdout
    except (OSError, subprocess.TimeoutExpired):
        return ""


if __name__ == "__main__":
    raise SystemExit(main())
