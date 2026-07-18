"""Read-only coding-agent collector.

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

import json
import os
import re
import sqlite3
import time
from collections import Counter
from collections.abc import Iterable
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any

from .metrics import MetricFamily, Sample
from .processes import ProcessSnapshot, open_files, running_processes

UUID_PATTERN = re.compile(
    r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", re.IGNORECASE
)
CODEX_FAILED_GOAL_STATES = {"blocked", "usage_limited", "budget_limited"}
# Codex brackets every turn with a start record and an end record, and persists
# both whatever else the thread's history mode keeps. It still serialises them
# under their v1 names -- protocol.rs renames TurnStarted to task_started and
# TurnComplete to task_complete -- and accepts the v2 names only when reading,
# so a rollout may carry either spelling depending on who wrote it.
CODEX_TURN_START_EVENTS = {"task_started", "turn_started"}
CODEX_TURN_END_EVENTS = {"task_complete", "turn_complete", "turn_aborted"}
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
# A workflow's agents live one directory below the session's other sub-agents, one
# directory per run, and the journal in there is what brackets each of them.
CLAUDE_WORKFLOW_RUNS_DIR = "workflows"
CLAUDE_WORKFLOW_JOURNAL = "journal.jsonl"
CLAUDE_WORKFLOW_AGENT_RETURNED_EVENT = "result"
# /proc/<pid>/stat holds the process's start tick in its 22nd field, counting the
# pid and the bracketed comm the fields before it are read past.
CLAUDE_PROC_STAT_START_TIME_INDEX = 19

CODEX_KIND = "codex"
CLAUDE_KIND = "claude"


@dataclass(frozen=True)
class AgentSnapshot:
    kind: str
    state: str
    model: str = ""
    # one entry per top-level agent working right now, holding the model it runs.
    # A host can have several of one kind out at once -- three codex sessions in
    # three terminals are three agents, not one -- and the page draws each.
    agent_models: tuple[str, ...] = ()
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


class AgentCollector:
    name = "agents"

    def __init__(
        self, codex_homes: list[Path], claude_home: Path, active_window_seconds: int
    ) -> None:
        self._codex_homes = codex_homes
        self._claude_home = claude_home
        self._active_window_seconds = active_window_seconds

    def collect(self) -> tuple[MetricFamily, ...]:
        now = time.time()
        processes = running_processes()
        snapshots = (
            codex_snapshot(
                self._codex_homes, processes, now, self._active_window_seconds
            ),
            claude_snapshot(
                self._claude_home, processes, now, self._active_window_seconds
            ),
        )
        return agent_metric_families(snapshots)


# --- Codex ------------------------------------------------------------------
#
# Codex holds each thread's rollout .jsonl open for the thread's whole life and
# flushes every record, so the set of open rollouts is the set of live threads.
# A thread it spawned names its parent in the session_meta record that heads its
# own rollout, which is what separates a sub-agent from a session the user drove
# themselves.


def codex_snapshot(
    homes: list[Path],
    processes: dict[int, ProcessSnapshot],
    now: float,
    active_window_seconds: int,
) -> AgentSnapshot:
    open_threads = codex_open_threads(homes, processes)
    subagent_ids = codex_subagent_ids(open_threads)
    candidates = codex_candidates(homes, open_threads, now, active_window_seconds)
    agents = [item for item in candidates if item.thread_id not in subagent_ids]
    if not agents:
        return AgentSnapshot(kind=CODEX_KIND, state="idle")

    agents.sort(
        key=lambda item: (item.state != "idle", item.updated_at_seconds), reverse=True
    )
    top = agents[0]
    return AgentSnapshot(
        kind=CODEX_KIND,
        state=top.state,
        model=top.model,
        agent_models=tuple(sorted(working_models(agents))),
        subagent_models=tuple(
            sorted(
                working_models(
                    item for item in candidates if item.thread_id in subagent_ids
                )
            )
        ),
        budget_remaining_percent=top.budget_remaining_percent,
    )


def working_models(candidates: Iterable[CodexCandidate]) -> list[str]:
    """The model each thread that is mid-turn runs, one entry per thread.

    A sub-agent is held to the same test as the agent that spawned it. Codex
    keeps a finished thread's rollout open for a while after its last turn ends,
    so being alive is not being at work: only an open turn bracket is, and a
    thread the window has retired has stopped writing either way.
    """
    return [item.model for item in candidates if item.state == "running"]


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
            payload = record.get("payload")
            if not isinstance(payload, dict):
                return ""
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
    now: float,
    active_window_seconds: int,
) -> list[CodexCandidate]:
    """Every live thread, agent and sub-agent alike, with the state it is in."""
    threads = codex_threads(homes)
    goals: dict[str, CodexGoal] = {}
    for home in homes:
        for database in sorted(home.glob(CODEX_GOALS_DATABASES)):
            goals.update(read_codex_goals(database))

    candidates: list[CodexCandidate] = []
    for thread_id, rollout in open_threads.items():
        # An open rollout is proof enough of a live thread. The databases only
        # add the model and the goal, and a schema generation this exporter has
        # never seen must not make Codex vanish from the page.
        thread = threads.get(thread_id)
        goal = goals.get(thread_id)
        working, updated_at = codex_activity(rollout, now, active_window_seconds)
        updated_at = max(
            thread.updated_at_seconds if thread else 0,
            goal.updated_at_seconds if goal else 0,
            updated_at,
        )
        candidates.append(
            CodexCandidate(
                thread_id=thread_id,
                state=codex_state(goal.status if goal else "", working),
                model=codex_model(thread, rollout),
                updated_at_seconds=updated_at,
                budget_remaining_percent=goal.budget_remaining_percent
                if goal
                else None,
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
    rows = query_sqlite(
        database,
        "select id, model, updated_at from threads order by updated_at desc limit 50",
    )
    if not rows:
        rows = query_sqlite(
            database,
            "select id, updated_at from threads order by updated_at desc limit 50",
        )
    return {
        str(row["id"]).lower(): CodexThread(
            thread_id=str(row["id"]).lower(),
            model=str(row["model"] or "") if "model" in row else "",
            updated_at_seconds=epoch_seconds(row["updated_at"]),
        )
        for row in rows
    }


def read_codex_goals(database: Path) -> dict[str, CodexGoal]:
    rows = query_sqlite(
        database,
        "select thread_id, status, token_budget, tokens_used, updated_at_ms "
        "from thread_goals order by updated_at_ms desc limit 50",
    )
    return {
        str(row["thread_id"]).lower(): CodexGoal(
            thread_id=str(row["thread_id"]).lower(),
            status=str(row["status"] or ""),
            updated_at_seconds=epoch_seconds(row["updated_at_ms"]),
            budget_remaining_percent=budget_remaining(
                row["token_budget"], row["tokens_used"]
            ),
        )
        for row in rows
    }


def codex_subagent_ids(open_threads: dict[str, Path]) -> set[str]:
    """The live threads that another thread spawned, working or not.

    A spawned thread names its parent in the rollout it holds open, so the file
    that proves it is alive also says whose worker it is. The state database
    keeps a spawn edge `open` while a finished child is merely resumable, and
    records no edge at all for an ephemeral one, so the rollout is both the
    narrower answer and the one that survives a schema this exporter has never
    seen. Only the parent's id is read: a sub-agent's nickname, its role, and
    what it was asked to do stay on the host.
    """
    return {
        thread_id
        for thread_id, rollout in open_threads.items()
        if codex_rollout_parent(rollout)
    }


def codex_rollout_parent(rollout: Path) -> str:
    """The thread that spawned this one, named by the rollout's first record."""
    record = first_record(rollout)
    if record.get("type") != "session_meta":
        return ""
    payload = record.get("payload") if isinstance(record.get("payload"), dict) else {}
    return (
        str(payload.get("parent_thread_id") or "") if isinstance(payload, dict) else ""
    )


def codex_open_threads(
    homes: list[Path], processes: dict[int, ProcessSnapshot]
) -> dict[str, Path]:
    """Map each live Codex thread id to the rollout its process holds open."""
    threads = {}
    for process_id, process in processes.items():
        if "codex" not in process.command_line:
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


def codex_activity(
    rollout: Path, now: float, active_window_seconds: int
) -> tuple[bool, float]:
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
    events = [
        event for event in map(decode_json, reversed(tail_lines(rollout))) if event
    ]
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


def claude_snapshot(
    home: Path,
    processes: dict[int, ProcessSnapshot],
    now: float,
    active_window_seconds: int,
) -> AgentSnapshot:
    sessions = claude_live_sessions(home, processes)
    if not sessions:
        return AgentSnapshot(kind=CLAUDE_KIND, state="idle")

    subagents = tuple(
        model for session in sessions for model in claude_subagent_models(session, now)
    )
    # `busy` is what a session claims; its transcript is what it did. A background
    # session goes on claiming `busy` while it waits for its next instruction, so
    # the claim alone lights the mascot for as long as the job lives. A turn writes
    # its thinking, its prose and every tool call as it goes, so a session silent
    # for the whole window is not computing -- the window that retires a Codex turn
    # wedged on a tool call retires this one too, and for the same reason.
    working = [
        session
        for session in sessions
        if session.busy and now - session.written_at_seconds <= active_window_seconds
    ]
    newest = max(working or sessions, key=lambda session: session.written_at_seconds)
    state = "running" if working or subagents else "idle"
    model = last_model(newest.transcript)
    # Each working session is its own agent: two Claude sessions in two terminals
    # are two agents. A session that has gone quiet while its sub-agents work is
    # still one agent out, so a running kind always names at least one.
    agent_models = tuple(sorted(last_model(session.transcript) for session in working))
    if not agent_models and state == "running":
        agent_models = (model,)
    return AgentSnapshot(
        kind=CLAUDE_KIND,
        state=state,
        model=model,
        agent_models=agent_models,
        subagent_models=subagents,
    )


def claude_live_sessions(
    home: Path, processes: dict[int, ProcessSnapshot]
) -> list[ClaudeSession]:
    sessions_dir = home / "sessions"
    if not sessions_dir.exists():
        return []
    sessions = []
    for path in sessions_dir.glob("*.json"):
        data = read_json_object(path)
        process_id = read_int(data.get("pid"))
        process = processes.get(process_id)
        if process is None or "claude" not in process.command_line:
            continue  # the session file outlives the CLI that wrote it
        if not started_the_process(process_id, str(data.get("procStart") or "")):
            continue  # the pid is live, but the kernel has since handed it to someone else
        transcript = claude_transcript(
            home, str(data.get("cwd") or ""), str(data.get("sessionId") or "")
        )
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


def started_the_process(process_id: int, started_at_ticks: str) -> bool:
    """Whether the process holding this pid is the one the session wrote about.

    A pid is only on loan. A session names both the pid and the tick the kernel
    started it on, so a pid recycled to another program -- and on a machine that
    runs Claude Code, the next program is often another `claude` -- is caught by
    the tick, which no two processes on one boot can share. Linux publishes it as
    the 22nd field of /proc/<pid>/stat, behind a comm that may itself hold spaces
    and brackets. Where the kernel does not publish it, the pid has to stand.
    """
    if not started_at_ticks:
        return True
    try:
        fields = (
            Path(f"/proc/{process_id}/stat").read_text().rpartition(") ")[2].split()
        )
    except OSError:
        return True
    if len(fields) <= CLAUDE_PROC_STAT_START_TIME_INDEX:
        return True
    return fields[CLAUDE_PROC_STAT_START_TIME_INDEX] == started_at_ticks


def claude_transcript(home: Path, cwd: str, session_id: str) -> Path | None:
    if not cwd or not session_id:
        return None
    return home / "projects" / re.sub(r"[^A-Za-z0-9]", "-", cwd) / f"{session_id}.jsonl"


def claude_subagent_models(session: ClaudeSession, now: float) -> list[str]:
    """The model each of this session's working sub-agents runs, one per agent.

    A session puts its sub-agents out two ways and files them apart. One it drives
    from its own turn, and that transcript sits straight in the subagents
    directory. The others a workflow script drives, and they sit one level down in
    `workflows/<run>/`, beside the journal that runs them. Both are sub-agents and
    the page draws a mascot for each, so both are counted -- and each is retired by
    whichever record brackets it, which is not the same record for the two.

    A sub-agent can be given a different model from the session that spawned it,
    so its own transcript is the only place its model is written.
    """
    if not session.subagents_dir.exists():
        return []
    working = claude_working_task_subagents(
        session, now
    ) + claude_working_workflow_subagents(session, now)
    return [last_model(transcript) for transcript in sorted(working)]


def claude_working_task_subagents(session: ClaudeSession, now: float) -> list[Path]:
    """The sub-agents this session drove from its own turn and has not yet stopped."""
    candidates = fresh_subagent_transcripts(session.subagents_dir, now)
    if not candidates:
        return []
    stopped_at = claude_stopped_subagents(session.transcript)
    return [
        transcript
        for transcript in candidates
        if not claude_subagent_finished(transcript, stopped_at)
    ]


def claude_working_workflow_subagents(session: ClaudeSession, now: float) -> list[Path]:
    """The agents of this session's workflows that have yet to return.

    A workflow spawns its agents from a script rather than from the session's turn,
    so the session never announces them, and their transcripts do not close on an
    assistant's last word -- they end on the tool result that fed it. Neither test
    that retires a task sub-agent can retire one of these. The journal beside them
    is what brackets each: `started` when the script spawns it, `result` when it
    returns.
    """
    workflows_dir = session.subagents_dir / CLAUDE_WORKFLOW_RUNS_DIR
    if not workflows_dir.exists():
        return []
    working: list[Path] = []
    for run in sorted(workflows_dir.iterdir()):
        if not run.is_dir():
            continue
        returned = claude_returned_workflow_agents(run / CLAUDE_WORKFLOW_JOURNAL)
        working += [
            t
            for t in fresh_subagent_transcripts(run, now)
            if subagent_id(t) not in returned
        ]
    return working


def claude_returned_workflow_agents(journal: Path) -> set[str]:
    """The agents this journal has a result for: the ones that have returned.

    Only a record's type and the agent it names are read. A result also carries
    what that agent returned, which is the workflow's own work and stays here.
    """
    returned: set[str] = set()
    try:
        with journal.open("rb") as file:
            for raw in file:
                record = decode_json(raw.decode(errors="ignore"))
                if record.get("type") == CLAUDE_WORKFLOW_AGENT_RETURNED_EVENT:
                    returned.add(str(record.get("agentId") or ""))
    except OSError:
        return set()
    return returned


def fresh_subagent_transcripts(directory: Path, now: float) -> list[Path]:
    """The sub-agent transcripts in one directory, minus the ones that went quiet.

    An agent wedged on a tool call stops writing, and the window retires it exactly
    as it retires a Codex turn that wedged the same way.
    """
    return [
        transcript
        for transcript in directory.glob("agent-*.jsonl")
        if now - modified_at(transcript) <= CLAUDE_SUBAGENT_STALE_AFTER_SECONDS
    ]


def subagent_id(transcript: Path) -> str:
    """The id a sub-agent is known by, which its transcript is named after."""
    return transcript.name[len("agent-") : -len(".jsonl")]


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
                for task_id in CLAUDE_TASK_ID_PATTERN.findall(
                    raw.decode(errors="ignore")
                ):
                    stopped[task_id] = max(stopped.get(task_id, 0), announced_at)
    except OSError:
        return {}
    return stopped


def claude_subagent_finished(transcript: Path, stopped_at: dict[str, float]) -> bool:
    announced_at = stopped_at.get(subagent_id(transcript))
    if announced_at:
        # A write after the announcement means the sub-agent was resumed.
        return (
            modified_at(transcript)
            <= announced_at + CLAUDE_SUBAGENT_RESUME_EPSILON_SECONDS
        )
    record = last_record(transcript)
    return (
        record.get("type") == "assistant"
        and (record.get("message") or {}).get("stop_reason") == "end_turn"
    )


def last_model(transcript: Path) -> str:
    models = re.findall(
        r'"model":"([^"]+)"', tail_text(transcript, CLAUDE_MODEL_TAIL_BYTES)
    )
    return models[-1] if models else ""


# --- shared -----------------------------------------------------------------


def query_sqlite(
    database: Path, statement: str, parameters: tuple[object, ...] = ()
) -> list[sqlite3.Row]:
    if not database.exists():
        return []
    try:
        connection = sqlite3.connect(
            f"{database.resolve().as_uri()}?mode=ro", uri=True, timeout=0.5
        )
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


def decode_json(text: str) -> dict[str, Any]:
    try:
        data = json.loads(text)
    except json.JSONDecodeError:
        return {}
    return data if isinstance(data, dict) else {}


def read_int(value: Any) -> int:
    """A whole number from a field that was only ever meant to hold one."""
    try:
        return int(value)
    except (TypeError, ValueError):
        return 0


def first_record(path: Path) -> dict[str, Any]:
    """Decode the first JSON record, which for a rollout is its session_meta."""
    try:
        with path.open("rb") as handle:
            return decode_json(handle.readline().decode("utf-8", "replace"))
    except OSError:
        return {}


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


def agent_metric_families(
    snapshots: Iterable[AgentSnapshot],
) -> tuple[MetricFamily, ...]:
    state_samples: list[Sample] = []
    info_samples: list[Sample] = []
    running_samples: list[Sample] = []
    subagent_samples: list[Sample] = []
    budget_samples: list[Sample] = []
    for snapshot in snapshots:
        labels = {"agent_kind": snapshot.kind}
        state_samples.extend(
            Sample(int(snapshot.state == state), {**labels, "state": state})
            for state in ("idle", "running", "failed")
        )
        if snapshot.model:
            info_samples.append(Sample(1, {**labels, "model": snapshot.model}))
        # One series per model, so a sub-agent running a different model than the
        # agent that spawned it is still counted under its own name. With none
        # working the bare series still reports zero, which clears the last scrape.
        # Agents of one kind are counted the same way: a host can have several out
        # at once, and each is drawn on the page.
        for samples, models in (
            (running_samples, snapshot.agent_models),
            (subagent_samples, snapshot.subagent_models),
        ):
            counts = Counter(models)
            for model in sorted(counts) or [""]:
                samples.append(Sample(counts[model], {**labels, "model": model}))
        if snapshot.budget_remaining_percent is not None:
            budget_samples.append(
                Sample(
                    max(0, min(100, snapshot.budget_remaining_percent)) / 100, labels
                )
            )
    return (
        MetricFamily(
            "realtime_agent_state",
            "Agent state labelled by state. OpenTelemetry name: realtime.agent.state.",
            state_samples,
            unit="1",
        ),
        MetricFamily(
            "realtime_agent_info",
            "Agent metadata as labels, always 1. OpenTelemetry name: realtime.agent.info.",
            info_samples,
            unit="1",
        ),
        MetricFamily(
            "realtime_agent_running_count",
            "Top-level agents working now, grouped by kind and model. "
            "OpenTelemetry name: realtime.agent.running.count.",
            running_samples,
            unit="1",
        ),
        MetricFamily(
            "realtime_agent_subagents_running",
            "Sub-agents working now, grouped by kind and model. OpenTelemetry name: realtime.agent.subagents.running.",
            subagent_samples,
            unit="1",
        ),
        MetricFamily(
            "realtime_agent_budget_remaining_ratio",
            "Agent budget remaining as a fraction. OpenTelemetry name: realtime.agent.budget.remaining.",
            budget_samples,
            unit="1",
        ),
    )


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
