from __future__ import annotations

import re
from collections import Counter
from pathlib import Path

from compose_policy import PolicyError

_MAX_SOURCE_BYTES = 128 * 1024
_VARIABLE_PATTERN = re.compile(r"\$(?:\{[^}\n]*\}|\$[A-Za-z_][A-Za-z0-9_]*)")
_YAML_REFERENCE_PATTERN = re.compile(r"[&*][A-Za-z][A-Za-z0-9_-]*")
_FORBIDDEN_KEY_PATTERN = re.compile(
    r"(?m)^\s*(?:include|extends|env_file|label_file|file)\s*:"
)
_BLOCK_SCALAR_PATTERN = re.compile(r"(?m):\s*[>|][-+]?\s*(?:#.*)?$")

_EXPECTED_VARIABLES = Counter(
    {
        "$$POSTGRES_DB": 1,
        "$$POSTGRES_USER": 1,
        "${CLOUD_DRIVE_DATA_DIR:?CLOUD_DRIVE_DATA_DIR is required}": 3,
        "${CLOUD_DRIVE_IMAGE:-cloud-drive:local}": 1,
        "${CLOUD_DRIVE_POSTGRES_DIR:?CLOUD_DRIVE_POSTGRES_DIR is required}": 1,
        "${CONSOLE_ORIGIN:?CONSOLE_ORIGIN is required}": 1,
        "${LIBRARY_AUTH_AUDIENCE:?LIBRARY_AUTH_AUDIENCE is required}": 1,
        "${MUSIC_PROVIDER_CREDENTIAL_KEY:?MUSIC_PROVIDER_CREDENTIAL_KEY is required}": 2,
        "${OIDC_ISSUER:?OIDC_ISSUER is required}": 1,
        "${POSTGRES_DB:?POSTGRES_DB is required}": 4,
        "${POSTGRES_IMAGE:-postgres:18.4-alpine}": 1,
        "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}": 4,
        "${POSTGRES_USER:?POSTGRES_USER is required}": 4,
        "${PRIVATE_API_HOST:?PRIVATE_API_HOST is required}": 2,
        "${PUBLIC_API_HOST:?PUBLIC_API_HOST is required}": 1,
        "${PUBLIC_SITE_ORIGIN:?PUBLIC_SITE_ORIGIN is required}": 1,
        "${RESERVED_FREE_BYTES:-21474836480}": 2,
        "${SPOTIFY_CLIENT_ID:-}": 2,
        "${SPOTIFY_CLIENT_SECRET:-}": 2,
    }
)
_EXPECTED_YAML_REFERENCES = Counter({"&app-image": 1, "*app-image": 3})


def validate_source(path: Path) -> None:
    content = path.read_bytes()
    if not content or len(content) > _MAX_SOURCE_BYTES:
        raise PolicyError("Compose source has an invalid size")

    try:
        source = content.decode("utf-8")
    except UnicodeDecodeError as error:
        raise PolicyError("Compose source must be UTF-8") from error

    if "\x00" in source or "\t" in source or "\r" in source:
        raise PolicyError("Compose source contains unsupported control characters")
    if "!" in source or _BLOCK_SCALAR_PATTERN.search(source):
        raise PolicyError("Compose source contains unsupported YAML syntax")
    if _FORBIDDEN_KEY_PATTERN.search(source):
        raise PolicyError("Compose source may not load external files")
    if source.count("<<:") != 3:
        raise PolicyError("Compose source has an unexpected merge structure")

    variables = Counter(_VARIABLE_PATTERN.findall(source))
    if variables != _EXPECTED_VARIABLES:
        raise PolicyError("Compose source has unexpected variable interpolation")
    if "$" in _VARIABLE_PATTERN.sub("", source):
        raise PolicyError("Compose source contains an unsupported dollar expression")

    references = Counter(_YAML_REFERENCE_PATTERN.findall(source))
    if references != _EXPECTED_YAML_REFERENCES:
        raise PolicyError("Compose source has an unexpected YAML reference")
    without_references = _YAML_REFERENCE_PATTERN.sub("", source)
    if "&" in without_references or "*" in without_references:
        raise PolicyError("Compose source contains an unsupported YAML reference")
