#!/usr/bin/env python3
"""Read-only coding-agent exporter for Prometheus.

This host is unaware of the gateway. It exposes each agent's run state and
remaining budget on /metrics; Prometheus discovers it through the gateway's HTTP
service discovery and stamps the gateway-minted device uid and name onto every
series as target labels. The exporter holds no identity, no token, and no
gateway address. Task titles are never collected: nothing consumes them, and
they are the most sensitive thing this host can see.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sqlite3
import time
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any

from status_common import (
    json_response,
    label_set,
    run,
    run_server,
    text_response,
)

UUID_PATTERN = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", re.IGNORECASE)
CODEX_FAILED_GOAL_STATES = {"blocked", "usage_limited", "budget_limited"}
CODEX_RUNNING_GOAL_STATES = {"active"}
CLAUDE_ACTIVE_TASK_STATES = {"in_progress"}
CLAUDE_ACTIVE_JOB_STATES = {"working"}
CLAUDE_ACTIVE_JOB_TEMPOS = {"active", "working"}
CLAUDE_BUSY_SESSION_STATES = {"busy"}
CLAUDE_IN_FLIGHT_JOB_STATE = "working"

CODEX_KIND = "codex"
CLAUDE_KIND = "claude"


@dataclass(frozen=True)
class AgentSnapshot:
    kind: str
    state: str
    budget_remaining_percent: int | None = None


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


def main() -> int:
    args = parse_args()
    routes = {
        "/healthz": lambda: json_response({"ok": True}),
        "/metrics": lambda: text_response(render_prometheus_metrics(build_snapshots(args))),
    }
    return run_server(args.bind, args.port, routes)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Serve local Claude Code and Codex agent state for Prometheus.")
    parser.add_argument("--active-window-seconds", type=int, default=int(os.getenv("STATUS_AGENT_ACTIVE_WINDOW_SECONDS", "300")))
    parser.add_argument("--codex-homes", default=os.getenv("STATUS_CODEX_HOMES", "~/.codex-api:~/.codex"))
    parser.add_argument(
        "--claude-sessions-dir",
        default=os.getenv("STATUS_CLAUDE_SESSIONS_DIR", "~/.claude/sessions"),
    )
    parser.add_argument("--claude-tasks-dir", default=os.getenv("STATUS_CLAUDE_TASKS_DIR", "~/.claude/tasks"))
    parser.add_argument("--claude-jobs-dir", default=os.getenv("STATUS_CLAUDE_JOBS_DIR", "~/.claude/jobs"))
    parser.add_argument("--bind", default=os.getenv("STATUS_AGENT_EXPORTER_BIND", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("STATUS_AGENT_EXPORTER_PORT", "18082")))
    return parser.parse_args()


def build_snapshots(args: argparse.Namespace) -> list[AgentSnapshot]:
    now = time.time()
    codex_homes = [expand_path(value) for value in re.split(r"[:;,]", args.codex_homes) if value.strip()]
    return [
        codex_snapshot(codex_homes, now, args.active_window_seconds),
        claude_snapshot(
            expand_path(args.claude_sessions_dir),
            expand_path(args.claude_tasks_dir),
            expand_path(args.claude_jobs_dir),
            now,
            args.active_window_seconds,
        ),
    ]


def codex_snapshot(homes: list[Path], now: float, active_window_seconds: int) -> AgentSnapshot:
    open_sessions = codex_open_sessions(homes)
    candidates = codex_candidates(homes, open_sessions, now, active_window_seconds)
    if not candidates:
        return AgentSnapshot(kind=CODEX_KIND, state="idle")

    candidates.sort(key=lambda item: (item.state != "idle", item.updated_at_seconds), reverse=True)
    top = candidates[0]
    return AgentSnapshot(
        kind=CODEX_KIND,
        state=top.state,
        budget_remaining_percent=top.budget_remaining_percent,
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
        candidates.append(CodexCandidate(thread_id, state, updated_at, goal.budget_remaining_percent if goal else None))

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


def claude_snapshot(sessions_dir: Path, tasks_dir: Path, jobs_dir: Path, now: float, active_window_seconds: int) -> AgentSnapshot:
    sessions = read_claude_sessions(sessions_dir)
    jobs = read_claude_jobs(jobs_dir)
    if not sessions and not jobs:
        return AgentSnapshot(kind=CLAUDE_KIND, state="idle")

    sessions.sort(key=lambda item: item.updated_at_seconds, reverse=True)
    session = sessions[0] if sessions else None
    task = claude_open_task(tasks_dir / session.cli_session_id) if session and session.cli_session_id else None
    active_job = latest_item([job for job in jobs if claude_job_active(job, now, active_window_seconds)])
    active = claude_active(session, task, active_job, jobs, now, active_window_seconds)
    return AgentSnapshot(kind=CLAUDE_KIND, state="running" if active else "idle")


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
        if not data or status not in CLAUDE_ACTIVE_TASK_STATES:
            continue
        title = str(data.get("activeForm") or data.get("subject") or "")
        if title:
            candidates.append(ClaudeTask(title, status, path.stat().st_mtime))
    candidates.sort(key=lambda item: (item.status in CLAUDE_ACTIVE_TASK_STATES, item.updated_at_seconds), reverse=True)
    return candidates[0] if candidates else None


def read_claude_jobs(jobs_dir: Path) -> list[ClaudeTask]:
    if not jobs_dir.exists():
        return []
    candidates: list[ClaudeTask] = []
    for path in sorted(jobs_dir.glob("*/state.json"), key=lambda item: item.stat().st_mtime, reverse=True)[:20]:
        data = read_json_object(path)
        if not data:
            continue
        title = str(data.get("name") or data.get("intent") or data.get("sessionId") or "")
        if not title:
            continue
        tempo = str(data.get("tempo") or "")
        state = str(data.get("state") or "")
        if has_in_flight_work(data.get("inFlight")) or tempo in CLAUDE_ACTIVE_JOB_TEMPOS:
            state = CLAUDE_IN_FLIGHT_JOB_STATE
        candidates.append(ClaudeTask(title, state, timestamp_seconds(data.get("updatedAt") or path.stat().st_mtime)))
    candidates.sort(key=lambda item: item.updated_at_seconds, reverse=True)
    return candidates


def claude_active(
    session: ClaudeSession | None,
    task: ClaudeTask | None,
    active_job: ClaudeTask | None,
    jobs: list[ClaudeTask],
    now: float,
    active_window_seconds: int,
) -> bool:
    if task and task.status in CLAUDE_ACTIVE_TASK_STATES and recent(task.updated_at_seconds, now, active_window_seconds):
        return True
    if active_job:
        return True
    if not jobs and session and session.status in CLAUDE_BUSY_SESSION_STATES:
        return recent(session.updated_at_seconds, now, active_window_seconds)
    return False


def claude_job_active(job: ClaudeTask, now: float, active_window_seconds: int) -> bool:
    return job.status in CLAUDE_ACTIVE_JOB_STATES and recent(job.updated_at_seconds, now, active_window_seconds)


def latest_item(items: list[ClaudeTask]) -> ClaudeTask | None:
    if not items:
        return None
    return max(items, key=lambda item: item.updated_at_seconds)


def has_in_flight_work(value: Any) -> bool:
    if value is True:
        return True
    if not isinstance(value, dict):
        return False
    for key in ("tasks", "queued"):
        try:
            if int(value.get(key) or 0) > 0:
                return True
        except (TypeError, ValueError):
            continue
    return bool(value.get("kinds"))


def recent(timestamp_seconds: float, now: float, window_seconds: int) -> bool:
    return 0 <= now - timestamp_seconds <= window_seconds


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


def render_prometheus_metrics(snapshots: list[AgentSnapshot]) -> str:
    lines = [
        "# HELP realtime_agent_state Agent state labelled by state. OpenTelemetry name: realtime.agent.state.",
        "# TYPE realtime_agent_state gauge",
        "# UNIT realtime_agent_state 1",
        "# HELP realtime_agent_budget_remaining_ratio Agent budget remaining as a fraction. OpenTelemetry name: realtime.agent.budget.remaining.",
        "# TYPE realtime_agent_budget_remaining_ratio gauge",
        "# UNIT realtime_agent_budget_remaining_ratio 1",
    ]
    for snapshot in snapshots:
        labels = {"agent_kind": snapshot.kind}
        for state in ("idle", "running", "failed"):
            lines.append(f"realtime_agent_state{label_set({**labels, 'state': state})} {1 if snapshot.state == state else 0}")
        if snapshot.budget_remaining_percent is not None:
            lines.append(
                f"realtime_agent_budget_remaining_ratio{label_set(labels)} "
                f"{max(0, min(100, snapshot.budget_remaining_percent)) / 100}"
            )
    lines.append("")
    return "\n".join(lines)


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
        if isinstance(value, str):
            return parse_event_timestamp(value, 0)
        return 0
    return numeric / 1000 if numeric > 10_000_000_000 else numeric


def parse_event_timestamp(value: Any, fallback: float) -> float:
    if not isinstance(value, str):
        return fallback
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00")).timestamp()
    except ValueError:
        return fallback


def expand_path(value: str) -> Path:
    return Path(value.strip()).expanduser().resolve()


def is_relative_to(path: Path, parent: Path) -> bool:
    try:
        path.resolve().relative_to(parent.resolve())
        return True
    except ValueError:
        return False


if __name__ == "__main__":
    raise SystemExit(main())
