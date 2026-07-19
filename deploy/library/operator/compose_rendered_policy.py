from __future__ import annotations

import ipaddress
import re
from pathlib import Path
from typing import Any

from compose_policy import PolicyError


def validate_rendered(
    config: Any,
    *,
    project_directory: str,
    data_directory: str,
    postgres_directory: str,
) -> None:
    root = _mapping(config, "configuration")
    _expect_keys(root, {"name", "networks", "secrets", "services", "x-app-image"}, "configuration")
    _expect_equal(root["name"], "cloud-drive", "project name")

    services = _mapping(root["services"], "services")
    _expect_keys(services, {"api", "library-public", "migrate", "postgres", "worker"}, "services")

    postgres = _service(services, "postgres")
    migrate = _service(services, "migrate")
    api = _service(services, "api")
    public = _service(services, "library-public")
    worker = _service(services, "worker")

    environments = {
        "postgres": _service_environment(postgres, "postgres"),
        "migrate": _service_environment(migrate, "migrate"),
        "api": _service_environment(api, "api"),
        "library-public": _service_environment(public, "library-public"),
        "worker": _service_environment(worker, "worker"),
    }
    _validate_environments(environments)

    app_image = _text(api.get("image"), "api image")
    postgres_image = _text(postgres.get("image"), "postgres image")
    if app_image != "cloud-drive:local":
        raise PolicyError("API image is outside the approved local build policy")
    if re.fullmatch(r"postgres:18\.\d+-alpine", postgres_image) is None:
        raise PolicyError("Postgres image is outside the approved policy")

    project_directory = str(Path(project_directory))
    data_directory = str(Path(data_directory))
    postgres_directory = str(Path(postgres_directory))
    build = {"context": project_directory, "dockerfile": "services/library/Dockerfile"}
    postgres_dependency = {"postgres": {"condition": "service_healthy", "required": True}}
    application_dependencies = {
        **postgres_dependency,
        "migrate": {"condition": "service_completed_successfully", "required": True},
    }
    data_volume = [_data_volume(data_directory)]

    _expect_service(
        postgres,
        {
            "command": None,
            "entrypoint": None,
            "environment": environments["postgres"],
            "healthcheck": _postgres_healthcheck(),
            "image": postgres_image,
            "networks": {"backend": None},
            "restart": "unless-stopped",
            "shm_size": "268435456",
            "volumes": [
                {
                    "type": "bind",
                    "source": postgres_directory,
                    "target": "/var/lib/postgresql",
                    "bind": {"create_host_path": False},
                }
            ],
        },
        "postgres",
    )
    _expect_service(
        migrate,
        {
            "build": build,
            "command": ["/usr/local/bin/library-migrate"],
            "depends_on": postgres_dependency,
            "entrypoint": None,
            "environment": environments["migrate"],
            "image": app_image,
            "networks": {"backend": None},
            "restart": "no",
            "volumes": data_volume,
        },
        "migrate",
    )
    _expect_service(
        worker,
        {
            "build": build,
            "command": ["/usr/local/bin/library-worker"],
            "depends_on": application_dependencies,
            "entrypoint": None,
            "environment": environments["worker"],
            "image": app_image,
            "networks": {"backend": None, "provider-egress": None},
            "restart": "unless-stopped",
            "volumes": data_volume,
        },
        "worker",
    )

    _expect_keys(
        api,
        {
            "build",
            "command",
            "depends_on",
            "entrypoint",
            "environment",
            "expose",
            "healthcheck",
            "image",
            "networks",
            "ports",
            "restart",
            "secrets",
            "volumes",
        },
        "api service",
    )
    _expect_equal(api["build"], build, "api build")
    _expect_equal(api["command"], None, "api command")
    _expect_equal(api["depends_on"], application_dependencies, "api dependencies")
    _expect_equal(api["entrypoint"], None, "api entrypoint")
    _expect_equal(api["expose"], ["8080"], "api exposure")
    _expect_equal(api["healthcheck"], _api_healthcheck(), "api healthcheck")
    _expect_equal(api["networks"], {"backend": None, "provider-egress": {"gw_priority": 1}}, "api networks")
    _expect_equal(api["restart"], "unless-stopped", "api restart policy")
    _expect_equal(
        api["secrets"],
        [{"source": "internal_api_key", "target": "/run/secrets/internal_api_key"}],
        "api secrets",
    )
    _expect_equal(api["volumes"], data_volume, "api volumes")
    lan_address = _validate_private_ports(api["ports"])
    _expect_equal(environments["api"]["PRIVATE_API_HOST"], lan_address, "private API host")

    _expect_keys(
        public,
        {
            "command",
            "depends_on",
            "entrypoint",
            "environment",
            "healthcheck",
            "image",
            "networks",
            "restart",
            "volumes",
        },
        "library-public service",
    )
    _expect_equal(public["command"], None, "library-public command")
    _expect_equal(
        public["depends_on"],
        {"api": {"condition": "service_healthy", "required": True}},
        "library-public dependencies",
    )
    _expect_equal(public["entrypoint"], None, "library-public entrypoint")
    _expect_equal(public["healthcheck"], _public_healthcheck(), "library-public healthcheck")
    _expect_equal(public["image"], "caddy:2.11.4-alpine", "library-public image")
    _expect_equal(
        public["networks"],
        {"backend": None, "edge": {"aliases": ["library-public"]}},
        "library-public networks",
    )
    _expect_equal(public["restart"], "unless-stopped", "library-public restart policy")
    _expect_equal(
        public["volumes"],
        [
            {
                "type": "bind",
                "source": str(Path(project_directory, "services/library/public.Caddyfile")),
                "target": "/etc/caddy/Caddyfile",
                "read_only": True,
                "bind": {"create_host_path": False},
            }
        ],
        "library-public volumes",
    )

    _validate_networks(root["networks"])
    _validate_secrets(root["secrets"])
    _expect_equal(
        root["x-app-image"],
        {
            "build": {"context": ".", "dockerfile": "services/library/Dockerfile"},
            "image": app_image,
        },
        "application image extension",
    )


def _validate_environments(environments: dict[str, dict[str, Any]]) -> None:
    postgres = environments["postgres"]
    migrate = environments["migrate"]
    api = environments["api"]
    public = environments["library-public"]
    worker = environments["worker"]
    _expect_keys(migrate, {"DATABASE_URL", "DATA_ROOT"}, "migrate environment")
    _expect_keys(
        postgres,
        {"PGDATA", "POSTGRES_DB", "POSTGRES_INITDB_ARGS", "POSTGRES_PASSWORD", "POSTGRES_USER"},
        "postgres environment",
    )
    _expect_keys(
        api,
        {
            "CONSOLE_ORIGIN",
            "DATABASE_URL",
            "DATA_ROOT",
            "INTERNAL_API_KEY_FILE",
            "LIBRARY_AUTH_AUDIENCE",
            "LISTEN_ADDR",
            "MUSIC_PROVIDER_CREDENTIAL_KEY",
            "OIDC_ISSUER",
            "PRIVATE_API_HOST",
            "PUBLIC_API_HOST",
            "PUBLIC_SITE_ORIGIN",
            "RESERVED_FREE_BYTES",
            "SPOTIFY_CLIENT_ID",
            "SPOTIFY_CLIENT_SECRET",
        },
        "api environment",
    )
    _expect_keys(public, {"PUBLIC_API_HOST"}, "library-public environment")
    _expect_keys(
        worker,
        {
            "CONSOLE_ORIGIN",
            "DATABASE_URL",
            "DATA_ROOT",
            "MUSIC_PROVIDER_CREDENTIAL_KEY",
            "RESERVED_FREE_BYTES",
            "SPOTIFY_CLIENT_ID",
            "SPOTIFY_CLIENT_SECRET",
        },
        "worker environment",
    )

    database = _matching_text(postgres["POSTGRES_DB"], r"[A-Za-z0-9_]+", "database name")
    username = _matching_text(postgres["POSTGRES_USER"], r"[A-Za-z0-9_]+", "database user")
    password = _matching_text(postgres["POSTGRES_PASSWORD"], r"[A-Za-z0-9._~-]{32,}", "database password")
    database_url = f"postgres://{username}:{password}@postgres:5432/{database}?sslmode=disable"
    for service in (api, migrate, worker):
        _expect_equal(service["DATABASE_URL"], database_url, "database URL")

    credential_key = _matching_text(
        api["MUSIC_PROVIDER_CREDENTIAL_KEY"],
        r"[A-Za-z0-9+/]{43}=",
        "music credential key",
    )
    _expect_equal(worker["MUSIC_PROVIDER_CREDENTIAL_KEY"], credential_key, "worker credential key")
    reserved_bytes = _matching_text(api["RESERVED_FREE_BYTES"], r"[0-9]+", "reserved bytes")
    _expect_equal(worker["RESERVED_FREE_BYTES"], reserved_bytes, "worker reserved bytes")
    _expect_equal(api["INTERNAL_API_KEY_FILE"], "/run/secrets/internal_api_key", "internal API key path")
    _expect_equal(worker["CONSOLE_ORIGIN"], api["CONSOLE_ORIGIN"], "worker Console origin")
    _expect_equal(public["PUBLIC_API_HOST"], api["PUBLIC_API_HOST"], "public API host")

    for key in (
        "CONSOLE_ORIGIN",
        "LIBRARY_AUTH_AUDIENCE",
        "OIDC_ISSUER",
        "PRIVATE_API_HOST",
        "PUBLIC_API_HOST",
        "PUBLIC_SITE_ORIGIN",
    ):
        _bounded_text(api[key], key)

    spotify_client_id = _text(api["SPOTIFY_CLIENT_ID"], "Spotify client ID")
    spotify_client_secret = _text(api["SPOTIFY_CLIENT_SECRET"], "Spotify client secret")
    if bool(spotify_client_id) != bool(spotify_client_secret):
        raise PolicyError("Spotify credentials must be configured together")
    _expect_equal(worker["SPOTIFY_CLIENT_ID"], spotify_client_id, "worker Spotify client ID")
    _expect_equal(worker["SPOTIFY_CLIENT_SECRET"], spotify_client_secret, "worker Spotify client secret")


def _validate_networks(value: Any) -> None:
    networks = _mapping(value, "networks")
    _expect_equal(
        networks,
        {
            "backend": {"name": "cloud-drive_backend", "ipam": {}, "internal": True},
            "edge": {"name": "realtime-me-edge", "ipam": {}, "external": True},
            "provider-egress": {"name": "cloud-drive_provider-egress", "ipam": {}},
        },
        "networks",
    )


def _validate_secrets(value: Any) -> None:
    secrets = _mapping(value, "secrets")
    _expect_keys(secrets, {"internal_api_key"}, "secrets")
    secret = _mapping(secrets["internal_api_key"], "internal API key secret")
    _expect_keys(secret, {"file", "name"}, "internal API key secret")
    _expect_equal(secret["name"], "cloud-drive_internal_api_key", "internal API key secret name")
    path = Path(_bounded_text(secret["file"], "internal API key file"))
    if not path.is_absolute():
        raise PolicyError("internal API key file must be absolute")


def _validate_private_ports(value: Any) -> str:
    if not isinstance(value, list) or len(value) != 2:
        raise PolicyError("api must publish exactly one LAN and one VPN binding")
    addresses: dict[ipaddress.IPv4Address, str] = {}
    for index, value_item in enumerate(value):
        port = _mapping(value_item, f"api port {index}")
        _expect_keys(port, {"host_ip", "mode", "protocol", "published", "target"}, "api port")
        _expect_equal(port["mode"], "ingress", "api port mode")
        _expect_equal(port["protocol"], "tcp", "api port protocol")
        _expect_equal(port["published"], "18081", "api published port")
        _expect_equal(port["target"], 8080, "api target port")
        host = _text(port["host_ip"], "api bind address")
        try:
            address = ipaddress.ip_address(host)
        except ValueError as error:
            raise PolicyError("api bind address must be a literal IPv4 address") from error
        if not isinstance(address, ipaddress.IPv4Address) or address in addresses:
            raise PolicyError("api bind addresses must be distinct IPv4 addresses")
        addresses[address] = host

    vpn_address = ipaddress.ip_address("10.66.0.11")
    if vpn_address not in addresses:
        raise PolicyError("api VPN bind must use 10.66.0.11")
    lan_addresses = [
        address
        for address in addresses
        if address in ipaddress.ip_network("192.168.0.0/24")
        and address not in {ipaddress.ip_address("192.168.0.0"), ipaddress.ip_address("192.168.0.255")}
    ]
    if len(lan_addresses) != 1:
        raise PolicyError("api must bind one usable 192.168.0.0/24 LAN address")
    return addresses[lan_addresses[0]]


def _expect_service(actual: dict[str, Any], expected: dict[str, Any], name: str) -> None:
    _expect_keys(actual, set(expected), f"{name} service")
    for key, value in expected.items():
        _expect_equal(actual[key], value, f"{name} {key}")


def _service(services: dict[str, Any], name: str) -> dict[str, Any]:
    return _mapping(services.get(name), f"{name} service")


def _service_environment(service: dict[str, Any], name: str) -> dict[str, Any]:
    return _mapping(service.get("environment"), f"{name} environment")


def _data_volume(source: str) -> dict[str, Any]:
    return {
        "type": "bind",
        "source": source,
        "target": "/var/lib/cloud-drive",
        "bind": {"create_host_path": False},
    }


def _postgres_healthcheck() -> dict[str, Any]:
    return {
        "test": ["CMD-SHELL", 'pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'],
        "timeout": "5s",
        "interval": "10s",
        "retries": 10,
        "start_period": "10s",
    }


def _api_healthcheck() -> dict[str, Any]:
    return {
        "test": ["CMD-SHELL", "wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1"],
        "timeout": "5s",
        "interval": "10s",
        "retries": 12,
        "start_period": "15s",
    }


def _public_healthcheck() -> dict[str, Any]:
    return {
        "test": ["CMD", "wget", "-q", "--spider", "http://127.0.0.1:8080/healthz"],
        "timeout": "5s",
        "interval": "15s",
        "retries": 5,
        "start_period": "5s",
    }


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
