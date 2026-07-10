#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../.." && pwd -P)
ENV_FILE=/etc/cloud-drive/runtime.env
BACKUP_ENV_FILE=/etc/cloud-drive/backup.env
COMPOSE_FILE=
MIN_FREE_BYTES=$((20 * 1024 * 1024 * 1024))
RESTIC_CHECK_WEEKDAY=7

usage() {
  cat <<'USAGE'
Usage: backup.sh [options]

Creates a consistent PostgreSQL dump, backs up file data and the dump to an
encrypted restic repository, retains 30 daily snapshots, and runs a repository
check every Sunday. A missing USB mount or insufficient capacity is a failure.

Options:
  --repo-dir PATH          Checked-out cloud-drive repository
  --env-file PATH          Root-only Compose environment file
  --backup-env-file PATH   Root-only restic configuration file
  --compose-file PATH      Compose file
  -h, --help               Show this help
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
    --backup-env-file)
      BACKUP_ENV_FILE=${2:?missing value for --backup-env-file}
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
for command in docker restic mountpoint flock date du df awk rm tr; do
  require_command "$command"
done

REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
if [[ -z "$COMPOSE_FILE" ]]; then
  COMPOSE_FILE="$REPO_DIR/ops/docker-compose.yml"
fi

require_regular_file "$COMPOSE_FILE"
require_secure_root_file "$ENV_FILE"
require_secure_root_file "$BACKUP_ENV_FILE"
docker compose version >/dev/null

POSTGRES_USER=$(require_env_value "$ENV_FILE" POSTGRES_USER)
POSTGRES_DB=$(require_env_value "$ENV_FILE" POSTGRES_DB)
TUNNEL_TOKEN=$(read_cloudflare_tunnel_token "$ENV_FILE")
VOLUME_MOUNT_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_VOLUME_MOUNT_DIR)
DATA_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_DATA_DIR)
BACKUP_STAGING_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_BACKUP_STAGING_DIR)
BACKUP_MOUNT_DIR=$(require_env_value "$BACKUP_ENV_FILE" BACKUP_MOUNT_DIR)
RESTIC_REPOSITORY=$(require_env_value "$BACKUP_ENV_FILE" RESTIC_REPOSITORY)
RESTIC_PASSWORD_FILE=$(require_env_value "$BACKUP_ENV_FILE" RESTIC_PASSWORD_FILE)

[[ "$VOLUME_MOUNT_DIR" == /* && "$DATA_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_DATA_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'
[[ "$BACKUP_STAGING_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_BACKUP_STAGING_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'
[[ "$BACKUP_MOUNT_DIR" == /* ]] || die 'BACKUP_MOUNT_DIR must be an absolute path'
[[ "$RESTIC_REPOSITORY" == "$BACKUP_MOUNT_DIR"/* ]] || die 'RESTIC_REPOSITORY must be located on BACKUP_MOUNT_DIR'

require_mountpoint "$VOLUME_MOUNT_DIR"
require_mountpoint "$BACKUP_MOUNT_DIR"
require_root_owned_nonwritable_directory "$BACKUP_MOUNT_DIR"
[[ -f "$VOLUME_MOUNT_DIR/.cloud-drive-volume" ]] || die "missing volume marker: $VOLUME_MOUNT_DIR/.cloud-drive-volume"
[[ -d "$DATA_DIR" ]] || die "data directory does not exist: $DATA_DIR"
IMMUTABLE_BLOBS_DIR="$DATA_DIR/blobs"
[[ -d "$IMMUTABLE_BLOBS_DIR" ]] || die "immutable blob directory does not exist: $IMMUTABLE_BLOBS_DIR"
[[ -d "$BACKUP_STAGING_DIR" ]] || die "backup staging directory does not exist: $BACKUP_STAGING_DIR"
require_secure_root_file "$RESTIC_PASSWORD_FILE"
[[ -s "$RESTIC_PASSWORD_FILE" ]] || die "restic password file is empty: $RESTIC_PASSWORD_FILE"

PRIMARY_FREE_BYTES=$(available_bytes "$VOLUME_MOUNT_DIR")
BACKUP_FREE_BYTES=$(available_bytes "$BACKUP_MOUNT_DIR")
[[ "$PRIMARY_FREE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine free capacity for $VOLUME_MOUNT_DIR"
[[ "$BACKUP_FREE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine free capacity for $BACKUP_MOUNT_DIR"
((PRIMARY_FREE_BYTES >= MIN_FREE_BYTES)) || die "less than 20 GiB free on primary data volume: $VOLUME_MOUNT_DIR"
((BACKUP_FREE_BYTES >= MIN_FREE_BYTES)) || die "less than 20 GiB free on USB backup volume: $BACKUP_MOUNT_DIR"

# Reject clearly impossible first backups before writing a database dump. Restic
# still reports a non-zero failure if a later snapshot exhausts remaining space.
SOURCE_BYTES=$(du --summarize --block-size=1 "$IMMUTABLE_BLOBS_DIR" | awk '{ print $1 }')
[[ "$SOURCE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine immutable blob size: $IMMUTABLE_BLOBS_DIR"
((BACKUP_FREE_BYTES >= SOURCE_BYTES + MIN_FREE_BYTES)) || die 'USB backup volume lacks source-size capacity plus the 20 GiB safety margin'

LOCK_FILE=/run/lock/cloud-drive-backup.lock
exec 9>"$LOCK_FILE"
flock -n 9 || die 'another cloud-drive backup is already running'

compose() {
  TUNNEL_TOKEN="$TUNNEL_TOKEN" docker compose \
    --project-directory "$REPO_DIR" \
    --project-name cloud-drive \
    --env-file "$ENV_FILE" \
    --file "$COMPOSE_FILE" \
    "$@"
}

DUMP_FILE="$BACKUP_STAGING_DIR/postgres.dump"
cleanup() {
  rm -f -- "$DUMP_FILE"
}
trap cleanup EXIT

umask 0077
note 'creating PostgreSQL backup dump'
compose exec --no-TTY postgres pg_dump --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --format=custom >"$DUMP_FILE"
unset TUNNEL_TOKEN
[[ -s "$DUMP_FILE" ]] || die 'PostgreSQL dump is empty'

note 'backing up immutable blobs and PostgreSQL dump with restic'
RESTIC_REPOSITORY="$RESTIC_REPOSITORY" RESTIC_PASSWORD_FILE="$RESTIC_PASSWORD_FILE" \
  restic backup --tag cloud-drive "$IMMUTABLE_BLOBS_DIR" "$DUMP_FILE"
RESTIC_REPOSITORY="$RESTIC_REPOSITORY" RESTIC_PASSWORD_FILE="$RESTIC_PASSWORD_FILE" \
  restic forget --tag cloud-drive --group-by host,tags --keep-daily 30 --prune

if [[ $(date +%u) -eq $RESTIC_CHECK_WEEKDAY ]]; then
  note 'running scheduled restic repository check'
  RESTIC_REPOSITORY="$RESTIC_REPOSITORY" RESTIC_PASSWORD_FILE="$RESTIC_PASSWORD_FILE" restic check
fi

note 'backup completed'
