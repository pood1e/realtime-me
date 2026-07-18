from __future__ import annotations

import re
from typing import Any

from compose_expected import build_expected_configuration
from compose_policy import PolicyError, first_difference


def validate_rendered(
    config: Any,
    *,
    project_directory: str,
    data_directory: str,
    postgres_directory: str,
) -> None:
    root = _mapping(config, "configuration")
    services = _mapping(root.get("services"), "services")
    _expect_keys(services, {"api", "migrate", "postgres", "worker"}, "services")

    postgres_environment = _service_environment(services, "postgres")
    migrate_environment = _service_environment(services, "migrate")
    api_environment = _service_environment(services, "api")
    worker_environment = _service_environment(services, "worker")
    _validate_environments(
        postgres_environment,
        migrate_environment,
        api_environment,
        worker_environment,
    )

    app_image = _text(_mapping(services["api"], "api").get("image"), "api image")
    postgres_image = _text(
        _mapping(services["postgres"], "postgres").get("image"),
        "postgres image",
    )
    if app_image != "cloud-drive:local":
        raise PolicyError("API image is outside the approved local build policy")
    if not re.fullmatch(r"postgres:18\.\d+-alpine", postgres_image):
        raise PolicyError("Postgres image is outside the approved policy")

    expected = build_expected_configuration(
        postgres_environment=postgres_environment,
        migrate_environment=migrate_environment,
        api_environment=api_environment,
        worker_environment=worker_environment,
        app_image=app_image,
        postgres_image=postgres_image,
        project_directory=project_directory,
        data_directory=data_directory,
        postgres_directory=postgres_directory,
    )
    difference = first_difference(root, expected)
    if difference is not None:
        raise PolicyError(f"rendered Compose configuration violates policy at {difference}")


def _validate_environments(
    postgres: dict[str, Any],
    migrate: dict[str, Any],
    api: dict[str, Any],
    worker: dict[str, Any],
) -> None:
    _expect_keys(
        migrate,
        {"DATABASE_URL", "DATA_ROOT"},
        "migrate environment",
    )
    _expect_keys(
        postgres,
        {"PGDATA", "POSTGRES_DB", "POSTGRES_INITDB_ARGS", "POSTGRES_PASSWORD", "POSTGRES_USER"},
        "postgres environment",
    )
    _expect_keys(
        api,
        {
            "DATABASE_URL",
            "DATA_ROOT",
            "LISTEN_ADDR",
            "MUSIC_APP_ORIGIN",
            "MUSIC_PROVIDER_CREDENTIAL_KEY",
            "PASSWORD_HASH_BASE64",
            "PRIVATE_API_HOST",
            "PRIVATE_APP_ORIGINS",
            "PUBLIC_API_HOST",
            "PUBLIC_APP_ORIGINS",
            "RESERVED_FREE_BYTES",
            "SESSION_SECRET",
            "SHARE_APP_ORIGIN",
            "SPOTIFY_CLIENT_ID",
            "SPOTIFY_CLIENT_SECRET",
        },
        "api environment",
    )
    _expect_keys(
        worker,
        {
            "DATABASE_URL",
            "DATA_ROOT",
            "MUSIC_PROVIDER_CREDENTIAL_KEY",
            "PRIVATE_API_HOST",
            "RESERVED_FREE_BYTES",
            "SPOTIFY_CLIENT_ID",
            "SPOTIFY_CLIENT_SECRET",
        },
        "worker environment",
    )
    database = _matching_text(postgres["POSTGRES_DB"], r"[A-Za-z0-9_]+", "database name")
    username = _matching_text(postgres["POSTGRES_USER"], r"[A-Za-z0-9_]+", "database user")
    password = _matching_text(
        postgres["POSTGRES_PASSWORD"],
        r"[A-Za-z0-9._~-]{32,}",
        "database password",
    )
    database_url = f"postgres://{username}:{password}@postgres:5432/{database}?sslmode=disable"
    _expect_equal(api["DATABASE_URL"], database_url, "api database URL")
    _expect_equal(migrate["DATABASE_URL"], database_url, "migrate database URL")
    _expect_equal(worker["DATABASE_URL"], database_url, "worker database URL")

    _matching_text(api["PASSWORD_HASH_BASE64"], r"[A-Za-z0-9+/]+={0,2}", "password hash")
    _matching_text(api["SESSION_SECRET"], r"[A-Fa-f0-9]{64}", "session secret")
    credential_key = _matching_text(
        api["MUSIC_PROVIDER_CREDENTIAL_KEY"],
        r"[A-Za-z0-9+/]{43}=",
        "music credential key",
    )
    _expect_equal(worker["MUSIC_PROVIDER_CREDENTIAL_KEY"], credential_key, "worker credential key")

    reserved_bytes = _matching_text(api["RESERVED_FREE_BYTES"], r"[0-9]+", "reserved bytes")
    _expect_equal(worker["RESERVED_FREE_BYTES"], reserved_bytes, "worker reserved bytes")
    for key in (
        "MUSIC_APP_ORIGIN",
        "PRIVATE_API_HOST",
        "PRIVATE_APP_ORIGINS",
        "PUBLIC_API_HOST",
        "PUBLIC_APP_ORIGINS",
        "SHARE_APP_ORIGIN",
    ):
        _bounded_text(api[key], key)

    spotify_client_id = _text(api["SPOTIFY_CLIENT_ID"], "Spotify client ID")
    spotify_client_secret = _text(api["SPOTIFY_CLIENT_SECRET"], "Spotify client secret")
    if bool(spotify_client_id) != bool(spotify_client_secret):
        raise PolicyError("Spotify credentials must be configured together")
    _expect_equal(worker["PRIVATE_API_HOST"], api["PRIVATE_API_HOST"], "worker private API host")
    _expect_equal(worker["SPOTIFY_CLIENT_ID"], spotify_client_id, "worker Spotify client ID")
    _expect_equal(worker["SPOTIFY_CLIENT_SECRET"], spotify_client_secret, "worker Spotify client secret")


def _service_environment(services: dict[str, Any], name: str) -> dict[str, Any]:
    service = _mapping(services.get(name), f"{name} service")
    return _mapping(service.get("environment"), f"{name} environment")


def _mapping(value: Any, label: str) -> dict[str, Any]:
    if not isinstance(value, dict) or not all(isinstance(key, str) for key in value):
        raise PolicyError(f"{label} must be a mapping")
    return value


def _text(value: Any, label: str) -> str:
    if not isinstance(value, str):
        raise PolicyError(f"{label} must be text")
    return value


def _bounded_text(value: Any, label: str) -> str:
    text = _text(value, label)
    if not text or len(text) > 4096 or any(ord(character) < 32 for character in text):
        raise PolicyError(f"{label} has an invalid value")
    return text


def _matching_text(value: Any, pattern: str, label: str) -> str:
    text = _text(value, label)
    if re.fullmatch(pattern, text) is None:
        raise PolicyError(f"{label} has an invalid format")
    return text


def _expect_keys(value: dict[str, Any], expected: set[str], label: str) -> None:
    if set(value) != expected:
        raise PolicyError(f"{label} has unexpected fields")


def _expect_equal(actual: Any, expected: Any, label: str) -> None:
    if actual != expected:
        raise PolicyError(f"{label} does not match policy")
