#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../../.." && pwd -P)
ENV_FILE=/etc/cloud-drive/runtime.env
COMPOSE_FILE=
MIN_FREE_BYTES=$((20 * 1024 * 1024 * 1024))

usage() {
  cat <<'USAGE'
Usage: deploy.sh [options]

Builds and starts PostgreSQL, migrate, API, and worker on the host. The
runtime environment file is never copied into the repository or printed.

Options:
  --repo-dir PATH   Installed realtime-me release tree (default: script's repository)
  --env-file PATH   Root-only Compose environment file (default: /etc/cloud-drive/runtime.env)
  --compose-file PATH
                    Compose file (default: <repo-dir>/deploy/library/compose.yaml)
  -h, --help        Show this help
USAGE
}

while (($# > 0)); do
  case "$1" in
    --repo-dir)
      REPO_DIR=${2:?missing value for --repo-dir}
      shift 2
      ;;
    --env-file)
      ENV_FILE=${2:?missing value for --env-file}
      shift 2
      ;;
    --compose-file)
      COMPOSE_FILE=${2:?missing value for --compose-file}
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown option: $1"
      ;;
  esac
done

require_root
require_command docker
require_command env
require_command mountpoint
require_command df
require_command awk

REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
if [[ -z "$COMPOSE_FILE" ]]; then
  COMPOSE_FILE="$REPO_DIR/deploy/library/compose.yaml"
fi

require_regular_file "$COMPOSE_FILE"
require_regular_file "$REPO_DIR/services/library/Dockerfile"
require_secure_root_file "$ENV_FILE"
docker compose version >/dev/null

POSTGRES_USER=$(require_env_value "$ENV_FILE" POSTGRES_USER)
POSTGRES_PASSWORD=$(require_env_value "$ENV_FILE" POSTGRES_PASSWORD)
POSTGRES_DB=$(require_env_value "$ENV_FILE" POSTGRES_DB)
PASSWORD_HASH_BASE64=$(require_env_value "$ENV_FILE" PASSWORD_HASH_BASE64)
SESSION_SECRET=$(require_env_value "$ENV_FILE" SESSION_SECRET)
MUSIC_PROVIDER_CREDENTIAL_KEY=$(require_env_value "$ENV_FILE" MUSIC_PROVIDER_CREDENTIAL_KEY)
PRIVATE_APP_ORIGINS=$(require_env_value "$ENV_FILE" PRIVATE_APP_ORIGINS)
PUBLIC_APP_ORIGINS=$(require_env_value "$ENV_FILE" PUBLIC_APP_ORIGINS)
SHARE_APP_ORIGIN=$(require_env_value "$ENV_FILE" SHARE_APP_ORIGIN)
MUSIC_APP_ORIGIN=$(require_env_value "$ENV_FILE" MUSIC_APP_ORIGIN)
PRIVATE_API_HOST=$(require_env_value "$ENV_FILE" PRIVATE_API_HOST)
PUBLIC_API_HOST=$(require_env_value "$ENV_FILE" PUBLIC_API_HOST)
SPOTIFY_CLIENT_ID=$(read_env_value "$ENV_FILE" SPOTIFY_CLIENT_ID)
SPOTIFY_CLIENT_SECRET=$(read_env_value "$ENV_FILE" SPOTIFY_CLIENT_SECRET)
VOLUME_MOUNT_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_VOLUME_MOUNT_DIR)
DATA_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_DATA_DIR)
POSTGRES_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_POSTGRES_DIR)
BACKUP_STAGING_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_BACKUP_STAGING_DIR)

# Keep secrets in variables only long enough to validate that they are safe for
# Compose's postgres:// URL. Do not print any runtime value.
[[ "$POSTGRES_PASSWORD" =~ ^[A-Za-z0-9._~-]{32,}$ ]] || die 'POSTGRES_PASSWORD must be URL-safe and at least 32 characters'
[[ "$POSTGRES_USER" =~ ^[A-Za-z0-9_]+$ ]] || die 'POSTGRES_USER must contain only letters, digits, or underscores'
[[ "$POSTGRES_DB" =~ ^[A-Za-z0-9_]+$ ]] || die 'POSTGRES_DB must contain only letters, digits, or underscores'
[[ "$PASSWORD_HASH_BASE64" =~ ^[A-Za-z0-9+/]+={0,2}$ ]] || die 'PASSWORD_HASH_BASE64 must be padded Base64 without whitespace'
[[ "$SESSION_SECRET" =~ ^[A-Fa-f0-9]{64}$ ]] || die 'SESSION_SECRET must contain exactly 64 hexadecimal characters'
[[ "$MUSIC_PROVIDER_CREDENTIAL_KEY" =~ ^[A-Za-z0-9+/]{43}=$ ]] ||
  die 'MUSIC_PROVIDER_CREDENTIAL_KEY must be padded Base64 containing exactly 32 bytes'
if [[ -n "$SPOTIFY_CLIENT_ID" || -n "$SPOTIFY_CLIENT_SECRET" ]]; then
  [[ -n "$SPOTIFY_CLIENT_ID" && -n "$SPOTIFY_CLIENT_SECRET" ]] ||
    die 'SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be configured together'
fi
[[ "$VOLUME_MOUNT_DIR" == /* && "$DATA_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_DATA_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'
[[ "$POSTGRES_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_POSTGRES_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'
[[ "$BACKUP_STAGING_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_BACKUP_STAGING_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'

require_mountpoint "$VOLUME_MOUNT_DIR"
[[ -f "$VOLUME_MOUNT_DIR/.cloud-drive-volume" ]] || die "missing volume marker: $VOLUME_MOUNT_DIR/.cloud-drive-volume"
[[ -d "$DATA_DIR" ]] || die "data directory does not exist: $DATA_DIR"
[[ -d "$POSTGRES_DIR" ]] || die "Postgres directory does not exist: $POSTGRES_DIR"
[[ -d "$BACKUP_STAGING_DIR" ]] || die "backup staging directory does not exist: $BACKUP_STAGING_DIR"

FREE_BYTES=$(available_bytes "$VOLUME_MOUNT_DIR")
[[ "$FREE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine free capacity for $VOLUME_MOUNT_DIR"
((FREE_BYTES >= MIN_FREE_BYTES)) || die "refusing to start with less than 20 GiB free on $VOLUME_MOUNT_DIR"

# Avoid shellcheck's unused-variable warning while keeping all required values
# secret and validating each one before Compose receives the environment file.
: "$PASSWORD_HASH_BASE64" "$SESSION_SECRET" "$MUSIC_PROVIDER_CREDENTIAL_KEY"
: "$PRIVATE_APP_ORIGINS" "$PUBLIC_APP_ORIGINS" "$SHARE_APP_ORIGIN" "$MUSIC_APP_ORIGIN"
: "$PRIVATE_API_HOST" "$PUBLIC_API_HOST"
: "$SPOTIFY_CLIENT_ID" "$SPOTIFY_CLIENT_SECRET"

compose() {
  env -i PATH="$PATH" HOME=/root docker compose \
    --project-directory "$REPO_DIR" \
    --project-name cloud-drive \
    --env-file "$ENV_FILE" \
    --file "$COMPOSE_FILE" \
    "$@"
}

note 'validating Docker Compose configuration'
compose config --quiet

note 'building and starting cloud-drive services'
compose up --build --detach --wait --wait-timeout 120 --remove-orphans
compose ps --status running
note 'deployment completed'
