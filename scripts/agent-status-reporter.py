#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import re
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
    task: str = ""
    updated_at: str = ""
    budget_remaining_percent: int | None = None

    def payload(self) -> dict[str, Any]:
        data: dict[str, Any] = {
            "agent_id": self.agent_id,
            "state": self.state,
        }
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
class ClaudeSession:
    session_id: str
    cli_session_id: str
    title: str
    updated_at_seconds: float
    archived: bool


def main() -> int:
    args = parse_args()
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
    parser.add_argument("--timeout-seconds", type=float, default=5)
    parser.add_argument("--active-window-seconds", type=int, default=int(os.getenv("STATUS_AGENT_ACTIVE_WINDOW_SECONDS", "300")))
    parser.add_argument("--codex-homes", default=os.getenv("STATUS_CODEX_HOMES", "~/.codex-api:~/.codex"))
    parser.add_argument(
        "--claude-sessions-dir",
        default=os.getenv("STATUS_CLAUDE_SESSIONS_DIR", "~/Library/Application Support/Claude/claude-code-sessions"),
    )
    parser.add_argument("--claude-tasks-dir", default=os.getenv("STATUS_CLAUDE_TASKS_DIR", "~/.claude/tasks"))
    parser.add_argument("--print", action="store_true")
    return parser.parse_args()


def build_snapshots(args: argparse.Namespace) -> list[AgentSnapshot]:
    now = time.time()
    codex_homes = [expand_path(value) for value in re.split(r"[:;,]", args.codex_homes) if value.strip()]
    return [
        *codex_snapshots(codex_homes, now, args.active_window_seconds),
        claude_snapshot(expand_path(args.claude_sessions_dir), expand_path(args.claude_tasks_dir), now, args.active_window_seconds),
    ]


def codex_snapshots(homes: list[Path], now: float, active_window_seconds: int) -> list[AgentSnapshot]:
    open_thread_ids = codex_open_thread_ids(homes)
    candidates = codex_candidates(homes, open_thread_ids, now, active_window_seconds)
    if not candidates:
        return [AgentSnapshot(agent_id="codex", state="idle", updated_at=utc_now())]

    candidates.sort(key=lambda item: (item.state != "idle", item.updated_at_seconds), reverse=True)
    return [codex_agent_snapshot(candidate) for candidate in candidates]


def codex_agent_snapshot(candidate: CodexCandidate) -> AgentSnapshot:
    return AgentSnapshot(
        agent_id=f"codex:{candidate.thread_id[:8]}",
        state=candidate.state,
        task=sanitize_task(candidate.task),
        updated_at=iso_time(candidate.updated_at_seconds),
        budget_remaining_percent=candidate.budget_remaining_percent,
    )


def codex_candidates(homes: list[Path], open_thread_ids: set[str], now: float, active_window_seconds: int) -> list[CodexCandidate]:
    threads: dict[str, CodexThread] = {}
    goals: dict[str, CodexGoal] = {}
    for home in homes:
        threads.update(read_codex_threads(home / "state_5.sqlite"))
        goals.update(read_codex_goals(home / "goals_1.sqlite"))

    candidates: list[CodexCandidate] = []
    for thread_id in open_thread_ids:
        thread = threads.get(thread_id)
        goal = goals.get(thread_id)
        if not thread and not goal:
            continue
        updated_at = max(thread.updated_at_seconds if thread else 0, goal.updated_at_seconds if goal else 0)
        active = now - updated_at <= active_window_seconds
        state = codex_state(goal.status if goal else "", active)
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


def codex_open_thread_ids(homes: list[Path]) -> set[str]:
    session_paths = codex_open_session_paths(homes)
    thread_ids = set()
    for path in session_paths:
        match = UUID_PATTERN.search(path.name)
        if match:
            thread_ids.add(match.group(0))
    return thread_ids


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


def claude_snapshot(sessions_dir: Path, tasks_dir: Path, now: float, active_window_seconds: int) -> AgentSnapshot:
    sessions = read_claude_sessions(sessions_dir)
    if not sessions:
        return AgentSnapshot(agent_id="claude-code", state="idle", updated_at=utc_now())

    sessions.sort(key=lambda item: item.updated_at_seconds, reverse=True)
    session = sessions[0]
    open_task = claude_open_task(tasks_dir / session.cli_session_id) if session.cli_session_id else ""
    active = now - session.updated_at_seconds <= active_window_seconds
    return AgentSnapshot(
        agent_id="claude-code",
        state="running" if active else "idle",
        task=sanitize_task(open_task or session.title),
        updated_at=iso_time(session.updated_at_seconds),
    )


def read_claude_sessions(sessions_dir: Path) -> list[ClaudeSession]:
    if not sessions_dir.exists():
        return []
    sessions: list[ClaudeSession] = []
    for path in sorted(sessions_dir.rglob("local_*.json"), key=lambda item: item.stat().st_mtime, reverse=True)[:100]:
        data = read_json_object(path)
        if not data:
            continue
        title = str(data.get("title") or "")
        if not title:
            continue
        updated_at = timestamp_seconds(data.get("lastActivityAt") or data.get("lastFocusedAt") or path.stat().st_mtime)
        sessions.append(
            ClaudeSession(
                session_id=str(data.get("sessionId") or path.stem),
                cli_session_id=str(data.get("cliSessionId") or ""),
                title=title,
                updated_at_seconds=updated_at,
                archived=bool(data.get("isArchived")),
            )
        )
    return [session for session in sessions if not session.archived]


def claude_open_task(task_dir: Path) -> str:
    if not task_dir.exists():
        return ""
    candidates: list[tuple[float, str]] = []
    for path in sorted(task_dir.glob("*.json"), key=lambda item: item.stat().st_mtime, reverse=True)[:50]:
        data = read_json_object(path)
        if not data or str(data.get("status") or "") not in CLAUDE_OPEN_TASK_STATES:
            continue
        title = str(data.get("activeForm") or data.get("subject") or "")
        if title:
            candidates.append((path.stat().st_mtime, title))
    candidates.sort(reverse=True)
    return candidates[0][1] if candidates else ""


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
