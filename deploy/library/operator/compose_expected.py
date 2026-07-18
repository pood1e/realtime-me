from __future__ import annotations

from typing import Any


def build_expected_configuration(
    *,
    postgres_environment: dict[str, Any],
    migrate_environment: dict[str, Any],
    api_environment: dict[str, Any],
    worker_environment: dict[str, Any],
    app_image: str,
    postgres_image: str,
    project_directory: str,
    data_directory: str,
    postgres_directory: str,
) -> dict[str, Any]:
    build = {"context": project_directory, "dockerfile": "services/library/Dockerfile"}
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
            "edge": {"name": "realtime-me-edge", "ipam": {}, "external": True},
            "provider-egress": {"name": "cloud-drive_provider-egress", "ipam": {}},
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
                "networks": {"backend": None, "edge": {"aliases": ["library-api"]}},
                "restart": "unless-stopped",
                "volumes": data_volume,
            },
            "migrate": {
                "build": build,
                "command": ["/usr/local/bin/library-migrate"],
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
                "command": ["/usr/local/bin/library-worker"],
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
            "build": {"context": ".", "dockerfile": "services/library/Dockerfile"},
            "image": app_image,
        },
    }
