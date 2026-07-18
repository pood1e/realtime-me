from __future__ import annotations

from typing import Any


class PolicyError(ValueError):
    pass


def first_difference(actual: Any, expected: Any, path: str = "configuration") -> str | None:
    if type(actual) is not type(expected):
        return path
    if isinstance(expected, dict):
        if set(actual) != set(expected):
            return path
        for key in expected:
            difference = first_difference(actual[key], expected[key], f"{path}.{key}")
            if difference is not None:
                return difference
        return None
    if isinstance(expected, list):
        if len(actual) != len(expected):
            return path
        for index, expected_item in enumerate(expected):
            difference = first_difference(actual[index], expected_item, f"{path}[{index}]")
            if difference is not None:
                return difference
        return None
    return None if actual == expected else path
