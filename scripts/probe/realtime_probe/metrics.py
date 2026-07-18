"""Small, deterministic Prometheus text renderer used by every collector."""

from __future__ import annotations

import re
from collections.abc import Mapping, Sequence
from dataclasses import dataclass, field

_LABEL_NAME_PATTERN = re.compile(r"[^A-Za-z0-9_]")


@dataclass(frozen=True)
class Sample:
    value: int | float
    labels: Mapping[str, str] = field(default_factory=dict)


@dataclass(frozen=True)
class MetricFamily:
    name: str
    help: str
    samples: Sequence[Sample]
    unit: str = ""
    type: str = "gauge"


def render(families: Sequence[MetricFamily]) -> str:
    lines: list[str] = []
    for family in families:
        lines.extend(
            (
                f"# HELP {family.name} {family.help}",
                f"# TYPE {family.name} {family.type}",
            )
        )
        if family.unit:
            lines.append(f"# UNIT {family.name} {family.unit}")
        lines.extend(
            f"{family.name}{label_set(sample.labels)} {number(sample.value)}"
            for sample in family.samples
        )
    lines.append("")
    return "\n".join(lines)


def label_set(labels: Mapping[str, str]) -> str:
    pairs = [
        f'{label_name(key)}="{escape_label_value(value)}"'
        for key, value in sorted(labels.items())
        if value
    ]
    return "{" + ",".join(pairs) + "}" if pairs else ""


def label_name(value: str) -> str:
    name = _LABEL_NAME_PATTERN.sub("_", value)
    if not name:
        return "label"
    return f"label_{name}" if name[0].isdigit() else name


def escape_label_value(value: str) -> str:
    return value.replace("\\", "\\\\").replace("\n", "\\n").replace('"', '\\"')


def number(value: int | float) -> str:
    return str(value) if isinstance(value, int) else format(value, ".15g")
