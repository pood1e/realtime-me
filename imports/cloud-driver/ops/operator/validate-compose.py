#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Sequence

from compose_policy import PolicyError
from compose_rendered_policy import validate_rendered
from compose_source_policy import validate_source


def parse_arguments(arguments: Sequence[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate a staged cloud-drive Compose release")
    commands = parser.add_subparsers(dest="command", required=True)

    source = commands.add_parser("source", help="validate the Compose source before rendering")
    source.add_argument("path", type=Path)

    rendered = commands.add_parser("rendered", help="validate rendered Compose JSON")
    rendered.add_argument("path", type=Path)
    rendered.add_argument("--data-directory", required=True)
    rendered.add_argument("--postgres-directory", required=True)
    return parser.parse_args(arguments)


def load_json(path: Path) -> object:
    with path.open("r", encoding="utf-8") as stream:
        return json.load(stream)


def main(arguments: Sequence[str]) -> int:
    options = parse_arguments(arguments)
    try:
        if options.command == "source":
            validate_source(options.path)
        else:
            validate_rendered(
                load_json(options.path),
                data_directory=options.data_directory,
                postgres_directory=options.postgres_directory,
            )
    except (OSError, UnicodeError, json.JSONDecodeError, PolicyError) as error:
        print(f"error: Compose release rejected: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
