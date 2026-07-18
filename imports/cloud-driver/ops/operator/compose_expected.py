from __future__ import annotations

from typing import Any


def build_expected_configuration(
    *,
    postgres_environment: dict[str, Any],
    migrate_environment: dict[str, Any],
    api_environment: dict[str, Any],
    worker_environment: dict[str, Any],
    cloudflared_environment: dict[str, Any],
    app_image: str,
    postgres_image: str,
    cloudflared_image: str,
    data_directory: str,
    postgres_directory: str,
) -> dict[str, Any]:
    build = {"context": "/opt/cloud-drive", "dockerfile": "api/Dockerfile"}
    postgres_dependency = {"postgres": {"condition": "service_healthy", "required": True}}
    application_dependencies = {
        **postgres_dependency,
        "migrate": {"condition": "service_completed_successfully", "required": True},
    }
    data_volume = [
        {
            "type": "bind",
            "source": data_directory,
            "target": "/var/lib/cloud-drive",
            "bind": {"create_host_path": False},
        }
    ]
    return {
        "name": "cloud-drive",
        "networks": {
            "backend": {"name": "cloud-drive_backend", "ipam": {}, "internal": True},
            "edge": {"name": "cloud-drive_edge", "ipam": {}},
            "provider-egress": {"name": "cloud-drive_provider-egress", "ipam": {}},
        },
        "secrets": {
            "cloudflare_tunnel_token": {
                "name": "cloud-drive_cloudflare_tunnel_token",
                "environment": "TUNNEL_TOKEN",
            }
        },
        "services": {
            "api": {
                "build": build,
                "command": None,
                "depends_on": application_dependencies,
                "entrypoint": None,
                "environment": api_environment,
                "expose": ["8080"],
                "healthcheck": {
                    "test": [
                        "CMD-SHELL",
                        "wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1",
                    ],
                    "timeout": "5s",
                    "interval": "10s",
                    "retries": 12,
                    "start_period": "15s",
                },
                "image": app_image,
                "networks": {"backend": None, "edge": None},
                "restart": "unless-stopped",
                "volumes": data_volume,
            },
            "cloudflared": {
                "command": ["tunnel", "--metrics", "127.0.0.1:2000", "run"],
                "depends_on": {"api": {"condition": "service_healthy", "required": True}},
                "entrypoint": None,
                "environment": cloudflared_environment,
                "healthcheck": {
                    "test": [
                        "CMD",
                        "cloudflared",
                        "tunnel",
                        "--metrics",
                        "127.0.0.1:2000",
                        "ready",
                    ],
                    "timeout": "5s",
                    "interval": "15s",
                    "retries": 8,
                    "start_period": "15s",
                },
                "image": cloudflared_image,
                "networks": {"edge": None},
                "restart": "unless-stopped",
                "secrets": [
                    {
                        "source": "cloudflare_tunnel_token",
                        "target": "cloudflare_tunnel_token",
                        "uid": "65532",
                        "gid": "65532",
                        "mode": "0400",
                    }
                ],
            },
            "migrate": {
                "build": build,
                "command": ["/usr/local/bin/cloud-drive-migrate"],
                "depends_on": postgres_dependency,
                "entrypoint": None,
                "environment": migrate_environment,
                "image": app_image,
                "networks": {"backend": None},
                "restart": "no",
                "volumes": data_volume,
            },
            "postgres": {
                "command": None,
                "entrypoint": None,
                "environment": postgres_environment,
                "healthcheck": {
                    "test": [
                        "CMD-SHELL",
                        'pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"',
                    ],
                    "timeout": "5s",
                    "interval": "10s",
                    "retries": 10,
                    "start_period": "10s",
                },
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
            "worker": {
                "build": build,
                "command": ["/usr/local/bin/cloud-drive-worker"],
                "depends_on": application_dependencies,
                "entrypoint": None,
                "environment": worker_environment,
                "image": app_image,
                "networks": {"backend": None, "provider-egress": None},
                "restart": "unless-stopped",
                "volumes": data_volume,
            },
        },
        "x-app-image": {
            "build": {"context": ".", "dockerfile": "api/Dockerfile"},
            "image": app_image,
        },
    }
