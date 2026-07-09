#!/usr/bin/env python3
"""Read-only coding-agent exporter for Prometheus.

This host is unaware of the gateway. It exposes each agent's run state, model,
remaining budget and live sub-agent count on /metrics; Prometheus discovers it
through the gateway's HTTP service discovery and stamps the gateway-minted
device uid and name onto every series as target labels. The exporter holds no
identity, no token, and no gateway address.

Nothing here asks an agent anything: both agents are read entirely from files
they already write. Task titles, prompts, objectives and sub-agent descriptions
are never collected -- nothing consumes them, and they are the most sensitive
thing this host can see.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sqlite3
import time
from collections import Counter
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any

from status_common import (
    json_response,
    label_set,
    run,
    cached,
    run_server,
    text_response,
)

UUID_PATTERN = re.compile(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", re.IGNORECASE)
PROCESS_LINE_PATTERN = re.compile(r"\s*(\d+)\s+(.+)$")

CODEX_FAILED_GOAL_STATES = {"blocked", "usage_limited", "budget_limited"}
# Codex brackets every turn with a start record and an end record, and persists
# both whatever else the thread's history mode keeps. It still serialises them
# under their v1 names -- protocol.rs renames TurnStarted to task_started and
# TurnComplete to task_complete -- and accepts the v2 names only when reading,
# so a rollout may carry either spelling depending on who wrote it.
CODEX_TURN_START_EVENTS = {"task_started", "turn_started"}
CODEX_TURN_END_EVENTS = {"task_complete", "turn_complete", "turn_aborted"}
CODEX_SPAWN_EDGE_OPEN = "open"
# The numeric suffix is a schema generation Codex bumps on a breaking rebuild, so
# match the family rather than pinning the generation this exporter was written
# against. A newer one must not silently take Codex off the page.
CODEX_STATE_DATABASES = "state_*.sqlite"
CODEX_GOALS_DATABASES = "goals_*.sqlite"

CLAUDE_BUSY_SESSION_STATES = {"busy"}
# A sub-agent that is killed outright is never announced as stopped, so it would
# read as working forever. Retire one that has not written in this long.
CLAUDE_SUBAGENT_STALE_AFTER_SECONDS = 900
# A sub-agent's last write lands just before its session announces it stopped,
# but the two can tie, so tolerate a write a moment either side. Only a write
# well after the announcement means the sub-agent was resumed.
CLAUDE_SUBAGENT_RESUME_EPSILON_SECONDS = 5
# The id is whatever the session names between the tags; never assume a charset.
CLAUDE_TASK_ID_PATTERN = re.compile(r"<task-id>([^<]+)</task-id>")
CLAUDE_MODEL_TAIL_BYTES = 96 * 1024

CODEX_KIND = "codex"
CLAUDE_KIND = "claude"


@dataclass(frozen=True)
class AgentSnapshot:
    kind: str
    state: str
    model: str = ""
    # one entry per sub-agent working right now, holding the model it runs
    subagent_models: tuple[str, ...] = ()
    budget_remaining_percent: int | None = None


@dataclass(frozen=True)
class CodexThread:
    thread_id: str
    model: str
    updated_at_seconds: float


@dataclass(frozen=True)
class CodexGoal:
    thread_id: str
    status: str
    updated_at_seconds: float
    budget_remaining_percent: int | None


@dataclass(frozen=True)
class CodexCandidate:
    thread_id: str
    state: str
    model: str
    updated_at_seconds: float
    budget_remaining_percent: int | None


@dataclass(frozen=True)
class ClaudeSession:
    transcript: Path
    subagents_dir: Path
    busy: bool
    written_at_seconds: float


def main() -> int:
    args = parse_args()
    routes = {
        "/healthz": lambda: json_response({"ok": True}),
        "/metrics": cached(lambda: text_response(render_prometheus_metrics(build_snapshots(args)))),
    }
    return run_server(args.bind, args.port, routes)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Serve local Claude Code and Codex agent state for Prometheus.")
    parser.add_argument("--active-window-seconds", type=int, default=int(os.getenv("STATUS_AGENT_ACTIVE_WINDOW_SECONDS", "300")))
    parser.add_argument("--codex-homes", default=os.getenv("STATUS_CODEX_HOMES", "~/.codex-api:~/.codex"))
    parser.add_argument("--claude-home", default=os.getenv("STATUS_CLAUDE_HOME", "~/.claude"))
    parser.add_argument("--bind", default=os.getenv("STATUS_AGENT_EXPORTER_BIND", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("STATUS_AGENT_EXPORTER_PORT", "18082")))
    return parser.parse_args()


def build_snapshots(args: argparse.Namespace) -> list[AgentSnapshot]:
    now = time.time()
    processes = running_processes()
    codex_homes = [expand_path(value) for value in re.split(r"[:;,]", args.codex_homes) if value.strip()]
    return [
        codex_snapshot(codex_homes, processes, now, args.active_window_seconds),
        claude_snapshot(expand_path(args.claude_home), processes, now, args.active_window_seconds),
    ]


def running_processes() -> dict[int, str]:
    """Map every visible process id to its lowercased command line."""
    processes = {}
    for line in run(["ps", "ax", "-o", "pid=,args="]).splitlines():
        match = PROCESS_LINE_PATTERN.match(line)
        if not match:
            continue
        args = match.group(2).lower()
        if "agent-status-reporter" in args:
            continue
        processes[int(match.group(1))] = args
    return processes


# --- Codex ------------------------------------------------------------------
#
# Codex holds each thread's rollout .jsonl open for the thread's whole life and
# flushes every record, so the set of open rollouts is the set of live threads.
# A thread it spawned is an ordinary thread with an `open` edge back to its
# parent in the state database, which is what separates a sub-agent from a
# session the user drove themselves.


def codex_snapshot(homes: list[Path], processes: dict[int, str], now: float, active_window_seconds: int) -> AgentSnapshot:
    open_threads = codex_open_threads(homes, processes)
    subagent_ids = codex_open_subagent_ids(homes) & set(open_threads)
    candidates = codex_candidates(homes, open_threads, subagent_ids, now, active_window_seconds)
    if not candidates:
        return AgentSnapshot(kind=CODEX_KIND, state="idle")

    candidates.sort(key=lambda item: (item.state != "idle", item.updated_at_seconds), reverse=True)
    top = candidates[0]
    threads = codex_threads(homes)
    return AgentSnapshot(
        kind=CODEX_KIND,
        state=top.state,
        model=top.model,
        subagent_models=tuple(codex_model(threads.get(uid), open_threads[uid]) for uid in sorted(subagent_ids)),
        budget_remaining_percent=top.budget_remaining_percent,
    )


def codex_model(thread: CodexThread | None, rollout: Path) -> str:
    """The model a thread runs.

    The state database only projects it, and only since the generation that added
    the column. The thread's own rollout is what that projection is extracted
    from, so it is the authority whenever the projection is absent.
    """
    if thread and thread.model:
        return thread.model
    return codex_rollout_model(rollout)


def codex_rollout_model(rollout: Path) -> str:
    """The model named by the newest turn_context record, written once a turn."""
    for size in (256 * 1024, 4 * 1024 * 1024):
        for line in reversed(tail_text(rollout, size).splitlines()):
            record = decode_json(line)
            if record.get("type") != "turn_context":
                continue
            payload = record.get("payload") if isinstance(record.get("payload"), dict) else {}
            return str(payload.get("model") or "")
        if size >= file_size(rollout):
            break
    return ""


def codex_threads(homes: list[Path]) -> dict[str, CodexThread]:
    threads: dict[str, CodexThread] = {}
    for home in homes:
        for database in sorted(home.glob(CODEX_STATE_DATABASES)):
            threads.update(read_codex_threads(database))
    return threads


def codex_candidates(
    homes: list[Path],
    open_threads: dict[str, Path],
    subagent_ids: set[str],
    now: float,
    active_window_seconds: int,
) -> list[CodexCandidate]:
    threads = codex_threads(homes)
    goals: dict[str, CodexGoal] = {}
    for home in homes:
        for database in sorted(home.glob(CODEX_GOALS_DATABASES)):
            goals.update(read_codex_goals(database))

    candidates: list[CodexCandidate] = []
    for thread_id, rollout in open_threads.items():
        if thread_id in subagent_ids:
            continue  # a sub-agent is counted, not reported as the agent itself
        # An open rollout is proof enough of a live thread. The databases only
        # add the model and the goal, and a schema generation this exporter has
        # never seen must not make Codex vanish from the page.
        thread = threads.get(thread_id)
        goal = goals.get(thread_id)
        working, updated_at = codex_activity(rollout, now, active_window_seconds)
        updated_at = max(thread.updated_at_seconds if thread else 0, goal.updated_at_seconds if goal else 0, updated_at)
        candidates.append(
            CodexCandidate(
                thread_id=thread_id,
                state=codex_state(goal.status if goal else "", working),
                model=codex_model(thread, rollout),
                updated_at_seconds=updated_at,
                budget_remaining_percent=goal.budget_remaining_percent if goal else None,
            )
        )
    return candidates


def codex_state(goal_status: str, working: bool) -> str:
    """A goal stays active for its whole life, so it says nothing about whether
    Codex is computing right now; only the turn does. A goal it cannot pursue is
    the one thing the turn cannot show."""
    if goal_status in CODEX_FAILED_GOAL_STATES:
        return "failed"
    return "running" if working else "idle"


def read_codex_threads(database: Path) -> dict[str, CodexThread]:
    # `model` arrived in a later schema than `threads` itself, and selecting a
    # column that is not there would drop every thread rather than its model.
    rows = query_sqlite(database, "select id, model, updated_at from threads order by updated_at desc limit 50")
    if not rows:
        rows = query_sqlite(database, "select id, updated_at from threads order by updated_at desc limit 50")
    return {
        str(row["id"]).lower(): CodexThread(
            thread_id=str(row["id"]).lower(),
            model=str(row["model"] or "") if "model" in row.keys() else "",
            updated_at_seconds=epoch_seconds(row["updated_at"]),
        )
        for row in rows
    }


def read_codex_goals(database: Path) -> dict[str, CodexGoal]:
    rows = query_sqlite(
        database,
        "select thread_id, status, token_budget, tokens_used, updated_at_ms from thread_goals order by updated_at_ms desc limit 50",
    )
    return {
        str(row["thread_id"]).lower(): CodexGoal(
            thread_id=str(row["thread_id"]).lower(),
            status=str(row["status"] or ""),
            updated_at_seconds=epoch_seconds(row["updated_at_ms"]),
            budget_remaining_percent=budget_remaining(row["token_budget"], row["tokens_used"]),
        )
        for row in rows
    }


def codex_open_subagent_ids(homes: list[Path]) -> set[str]:
    """Thread ids Codex spawned from another thread and has not yet closed."""
    children = set()
    for home in homes:
        for database in sorted(home.glob(CODEX_STATE_DATABASES)):
            rows = query_sqlite(
                database,
                "select child_thread_id from thread_spawn_edges where status = ?",
                (CODEX_SPAWN_EDGE_OPEN,),
            )
            children.update(str(row["child_thread_id"]).lower() for row in rows)
    return children


def codex_open_threads(homes: list[Path], processes: dict[int, str]) -> dict[str, Path]:
    """Map each live Codex thread id to the rollout its process holds open."""
    threads = {}
    for process_id, args in processes.items():
        if "codex" not in args:
            continue
        for path in open_files(process_id):
            if path.suffix != ".jsonl" or "/sessions/" not in path.as_posix():
                continue
            if not any(is_relative_to(path, home) for home in homes):
                continue
            match = UUID_PATTERN.search(path.name)
            if match:
                # Thread ids are lowercase in the databases these are joined to.
                threads[match.group(0).lower()] = path
    return threads


def codex_activity(rollout: Path, now: float, active_window_seconds: int) -> tuple[bool, float]:
    """Whether this thread is mid-turn, and when it last wrote.

    The newest bracket decides: a turn whose start has no end after it is still
    running. What lies between the brackets is named after whichever tool the
    model reached for, and Codex keeps or drops those records depending on how
    the thread stores its history, so nothing between them is worth reading.

    A turn long enough to bury its own start beyond the tail is still running,
    which is why an absent bracket reads the same as an open one. Either way a
    thread that wedges on a tool call stops writing, and the window retires it.
    """
    written_at = modified_at(rollout)
    events = [event for event in map(decode_json, reversed(tail_lines(rollout))) if event]
    if events:
        written_at = parse_event_timestamp(events[0].get("timestamp"), written_at)
    for event in events:
        payload_type = event_payload_type(event)
        if payload_type in CODEX_TURN_END_EVENTS:
            return False, written_at
        if payload_type in CODEX_TURN_START_EVENTS:
            break
    return now - written_at <= active_window_seconds, written_at


def event_payload_type(event: dict[str, Any]) -> str:
    payload = event.get("payload")
    return str(payload.get("type") or "") if isinstance(payload, dict) else ""


def open_files(process_id: int) -> list[Path]:
    paths = []
    for line in run(["lsof", "-Fn", "-p", str(process_id)]).splitlines():
        if line.startswith("n/"):
            paths.append(Path(line[1:]))
    return paths


# --- Claude Code ------------------------------------------------------------
#
# Claude Code appends to its transcripts and closes them again, so no file is
# ever held open and there is nothing for lsof to find. Instead every live CLI
# writes sessions/<pid>.json, which points at its transcript and carries its own
# busy flag. Each sub-agent gets its own transcript under the session's own
# directory, `<transcript without .jsonl>/subagents/agent-<id>.jsonl`, and the
# session announces the sub-agent's id when it stops. Neither a tool result nor a
# closing record can stand in for that announcement: a background sub-agent is
# answered the moment it launches, and it writes its thinking and its prose as
# separate records, so the tail holds no tool call long before it is done.


def claude_snapshot(home: Path, processes: dict[int, str], now: float, active_window_seconds: int) -> AgentSnapshot:
    sessions = claude_live_sessions(home, processes)
    if not sessions:
        return AgentSnapshot(kind=CLAUDE_KIND, state="idle")

    subagents = tuple(model for session in sessions for model in claude_subagent_models(session, now))
    # `busy` is what a session claims; its transcript is what it did. A background
    # session goes on claiming `busy` while it waits for its next instruction, so
    # the claim alone lights the mascot for as long as the job lives. A turn writes
    # its thinking, its prose and every tool call as it goes, so a session silent
    # for the whole window is not computing -- the window that retires a Codex turn
    # wedged on a tool call retires this one too, and for the same reason.
    working = [session for session in sessions if session.busy and now - session.written_at_seconds <= active_window_seconds]
    newest = max(working or sessions, key=lambda session: session.written_at_seconds)
    return AgentSnapshot(
        kind=CLAUDE_KIND,
        state="running" if working or subagents else "idle",
        model=last_model(newest.transcript),
        subagent_models=subagents,
    )


def claude_live_sessions(home: Path, processes: dict[int, str]) -> list[ClaudeSession]:
    sessions_dir = home / "sessions"
    if not sessions_dir.exists():
        return []
    sessions = []
    for path in sessions_dir.glob("*.json"):
        data = read_json_object(path)
        process_id = int(data.get("pid") or 0)
        if "claude" not in processes.get(process_id, ""):
            continue  # the session file outlives the CLI that wrote it
        transcript = claude_transcript(home, str(data.get("cwd") or ""), str(data.get("sessionId") or ""))
        if not transcript or not transcript.exists():
            continue
        sessions.append(
            ClaudeSession(
                transcript=transcript,
                subagents_dir=transcript.with_suffix("") / "subagents",
                busy=str(data.get("status") or "") in CLAUDE_BUSY_SESSION_STATES,
                written_at_seconds=modified_at(transcript),
            )
        )
    return sessions


def claude_transcript(home: Path, cwd: str, session_id: str) -> Path | None:
    if not cwd or not session_id:
        return None
    return home / "projects" / re.sub(r"[^A-Za-z0-9]", "-", cwd) / f"{session_id}.jsonl"


def claude_subagent_models(session: ClaudeSession, now: float) -> list[str]:
    """The model each of this session's working sub-agents runs, one per agent.

    A sub-agent can be given a different model from the session that spawned it,
    so its own transcript is the only place its model is written.
    """
    if not session.subagents_dir.exists():
        return []
    candidates = [
        transcript
        for transcript in session.subagents_dir.glob("agent-*.jsonl")
        if now - modified_at(transcript) <= CLAUDE_SUBAGENT_STALE_AFTER_SECONDS
    ]
    if not candidates:
        return []
    stopped_at = claude_stopped_subagents(session.transcript)
    return [last_model(t) for t in sorted(candidates) if not claude_subagent_finished(t, stopped_at)]


def claude_stopped_subagents(transcript: Path) -> dict[str, float]:
    """When each sub-agent was last announced as stopped by its session."""
    stopped: dict[str, float] = {}
    try:
        with transcript.open("rb") as file:
            for raw in file:
                if b"<task-id>" not in raw:
                    continue
                record = decode_json(raw.decode(errors="ignore"))
                announced_at = parse_event_timestamp(record.get("timestamp"), 0)
                for task_id in CLAUDE_TASK_ID_PATTERN.findall(raw.decode(errors="ignore")):
                    stopped[task_id] = max(stopped.get(task_id, 0), announced_at)
    except OSError:
        return {}
    return stopped


def claude_subagent_finished(transcript: Path, stopped_at: dict[str, float]) -> bool:
    announced_at = stopped_at.get(transcript.name[len("agent-") : -len(".jsonl")])
    if announced_at:
        # A write after the announcement means the sub-agent was resumed.
        return modified_at(transcript) <= announced_at + CLAUDE_SUBAGENT_RESUME_EPSILON_SECONDS
    record = last_record(transcript)
    return record.get("type") == "assistant" and (record.get("message") or {}).get("stop_reason") == "end_turn"


def last_model(transcript: Path) -> str:
    models = re.findall(r'"model":"([^"]+)"', tail_text(transcript, CLAUDE_MODEL_TAIL_BYTES))
    return models[-1] if models else ""


# --- shared -----------------------------------------------------------------


def query_sqlite(database: Path, statement: str, parameters: tuple[object, ...] = ()) -> list[sqlite3.Row]:
    if not database.exists():
        return []
    try:
        connection = sqlite3.connect(f"file:{database}?mode=ro", uri=True, timeout=0.5)
        connection.row_factory = sqlite3.Row
        connection.execute("pragma busy_timeout=500")
        try:
            return list(connection.execute(statement, parameters))
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


def last_record(path: Path) -> dict[str, Any]:
    """Decode the last complete JSON record, growing the read until one parses."""
    for size in (64 * 1024, 1024 * 1024, 8 * 1024 * 1024):
        lines = tail_text(path, size).splitlines()
        for line in reversed(lines):
            record = decode_json(line)
            if record:
                return record
        if size >= file_size(path):
            break
    return {}


def tail_text(path: Path, max_bytes: int) -> str:
    try:
        with path.open("rb") as file:
            file.seek(0, os.SEEK_END)
            file.seek(max(0, file.tell() - max_bytes))
            return file.read().decode(errors="ignore")
    except OSError:
        return ""


def tail_lines(path: Path, max_bytes: int = 131_072) -> list[str]:
    return tail_text(path, max_bytes).splitlines()


def file_size(path: Path) -> int:
    try:
        return path.stat().st_size
    except OSError:
        return 0


def modified_at(path: Path) -> float:
    try:
        return path.stat().st_mtime
    except OSError:
        return 0


def render_prometheus_metrics(snapshots: list[AgentSnapshot]) -> str:
    lines = [
        "# HELP realtime_agent_state Agent state labelled by state. OpenTelemetry name: realtime.agent.state.",
        "# TYPE realtime_agent_state gauge",
        "# UNIT realtime_agent_state 1",
        "# HELP realtime_agent_info Agent metadata as labels, always 1. OpenTelemetry name: realtime.agent.info.",
        "# TYPE realtime_agent_info gauge",
        "# UNIT realtime_agent_info 1",
        "# HELP realtime_agent_subagents_running Sub-agents the agent has working right now, by the model each runs. OpenTelemetry name: realtime.agent.subagents.running.",
        "# TYPE realtime_agent_subagents_running gauge",
        "# UNIT realtime_agent_subagents_running 1",
        "# HELP realtime_agent_budget_remaining_ratio Agent budget remaining as a fraction. OpenTelemetry name: realtime.agent.budget.remaining.",
        "# TYPE realtime_agent_budget_remaining_ratio gauge",
        "# UNIT realtime_agent_budget_remaining_ratio 1",
    ]
    for snapshot in snapshots:
        labels = {"agent_kind": snapshot.kind}
        for state in ("idle", "running", "failed"):
            lines.append(f"realtime_agent_state{label_set({**labels, 'state': state})} {1 if snapshot.state == state else 0}")
        if snapshot.model:
            lines.append(f"realtime_agent_info{label_set({**labels, 'model': snapshot.model})} 1")
        # One series per model, so a sub-agent running a different model than the
        # agent that spawned it is still counted under its own name. With none
        # working the bare series still reports zero, which clears the last scrape.
        counts = Counter(snapshot.subagent_models)
        for model in sorted(counts) or [""]:
            lines.append(f"realtime_agent_subagents_running{label_set({**labels, 'model': model})} {counts[model]}")
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
    return max(0, min(100, round(100 - used * 100 / budget)))


def epoch_seconds(value: Any) -> float:
    """Codex stores some columns in seconds and their newer twins in milliseconds."""
    try:
        numeric = float(value or 0)
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
