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
SNAPSHOT_RETENTION_COUNT=30

usage() {
  cat <<'USAGE'
Usage: backup.sh [options]

Creates a consistent PostgreSQL dump and a plain rsync snapshot containing the
dump and immutable file data. Unchanged files are hard-linked from the previous
snapshot, and the latest 30 successful snapshots are retained. A missing USB
mount or insufficient capacity is a failure.

Options:
  --repo-dir PATH          Checked-out cloud-drive repository
  --env-file PATH          Root-only Compose environment file
  --backup-env-file PATH   Root-only backup configuration file
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
for command in docker mountpoint flock date du df awk rm tr rsync install mktemp find sort tail mv ln sync; do
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
BACKUP_ROOT_DIR=$(require_env_value "$BACKUP_ENV_FILE" BACKUP_ROOT_DIR)

[[ "$VOLUME_MOUNT_DIR" == /* && "$DATA_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_DATA_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'
[[ "$BACKUP_STAGING_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'CLOUD_DRIVE_BACKUP_STAGING_DIR must be below CLOUD_DRIVE_VOLUME_MOUNT_DIR'
[[ "$BACKUP_MOUNT_DIR" == /* ]] || die 'BACKUP_MOUNT_DIR must be an absolute path'
[[ "$BACKUP_ROOT_DIR" == "$BACKUP_MOUNT_DIR"/* ]] || die 'BACKUP_ROOT_DIR must be located below BACKUP_MOUNT_DIR'

require_mountpoint "$VOLUME_MOUNT_DIR"
require_mountpoint "$BACKUP_MOUNT_DIR"
require_root_owned_nonwritable_directory "$BACKUP_MOUNT_DIR"
[[ -f "$VOLUME_MOUNT_DIR/.cloud-drive-volume" ]] || die "missing volume marker: $VOLUME_MOUNT_DIR/.cloud-drive-volume"
[[ -d "$DATA_DIR" ]] || die "data directory does not exist: $DATA_DIR"
IMMUTABLE_BLOBS_DIR="$DATA_DIR/blobs"
[[ -d "$IMMUTABLE_BLOBS_DIR" ]] || die "immutable blob directory does not exist: $IMMUTABLE_BLOBS_DIR"
[[ -d "$BACKUP_STAGING_DIR" ]] || die "backup staging directory does not exist: $BACKUP_STAGING_DIR"

PRIMARY_FREE_BYTES=$(available_bytes "$VOLUME_MOUNT_DIR")
BACKUP_FREE_BYTES=$(available_bytes "$BACKUP_MOUNT_DIR")
[[ "$PRIMARY_FREE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine free capacity for $VOLUME_MOUNT_DIR"
[[ "$BACKUP_FREE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine free capacity for $BACKUP_MOUNT_DIR"
((PRIMARY_FREE_BYTES >= MIN_FREE_BYTES)) || die "less than 20 GiB free on primary data volume: $VOLUME_MOUNT_DIR"
((BACKUP_FREE_BYTES >= MIN_FREE_BYTES)) || die "less than 20 GiB free on USB backup volume: $BACKUP_MOUNT_DIR"

LOCK_FILE=/run/lock/cloud-drive-backup.lock
exec 9>"$LOCK_FILE"
flock -n 9 || die 'another cloud-drive backup is already running'

SNAPSHOTS_DIR="$BACKUP_ROOT_DIR/snapshots"
install -d -o root -g root -m 0700 "$BACKUP_ROOT_DIR" "$SNAPSHOTS_DIR"
require_root_owned_nonwritable_directory "$BACKUP_ROOT_DIR"
require_root_owned_nonwritable_directory "$SNAPSHOTS_DIR"

list_snapshot_names() {
  local name

  find "$SNAPSHOTS_DIR" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' |
    while IFS= read -r name; do
      if [[ "$name" =~ ^[0-9]{8}T[0-9]{6}Z$ ]]; then
        printf '%s\n' "$name"
      fi
    done |
    LC_ALL=C sort
}

prune_snapshots() {
  local -a snapshots
  local index remove_count

  mapfile -t snapshots < <(list_snapshot_names)
  remove_count=$((${#snapshots[@]} - SNAPSHOT_RETENTION_COUNT))
  if ((remove_count <= 0)); then
    return 0
  fi

  for ((index = 0; index < remove_count; index++)); do
    rm -rf -- "${SNAPSHOTS_DIR:?}/${snapshots[$index]}"
  done
}

PREVIOUS_SNAPSHOT_NAME=$(list_snapshot_names | tail --lines=1)
if [[ -z "$PREVIOUS_SNAPSHOT_NAME" ]]; then
  SOURCE_BYTES=$(du --summarize --block-size=1 "$IMMUTABLE_BLOBS_DIR" | awk '{ print $1 }')
  [[ "$SOURCE_BYTES" =~ ^[0-9]+$ ]] || die "could not determine immutable blob size: $IMMUTABLE_BLOBS_DIR"
  ((BACKUP_FREE_BYTES >= SOURCE_BYTES + MIN_FREE_BYTES)) ||
    die 'USB backup volume lacks source-size capacity plus the 20 GiB safety margin'
fi

compose() {
  TUNNEL_TOKEN="$TUNNEL_TOKEN" docker compose \
    --project-directory "$REPO_DIR" \
    --project-name cloud-drive \
    --env-file "$ENV_FILE" \
    --file "$COMPOSE_FILE" \
    "$@"
}

DUMP_FILE="$BACKUP_STAGING_DIR/postgres.dump"
INCOMPLETE_SNAPSHOT=
TEMPORARY_LATEST_LINK=
cleanup() {
  rm -f -- "$DUMP_FILE"
  [[ -z "$INCOMPLETE_SNAPSHOT" ]] || rm -rf -- "$INCOMPLETE_SNAPSHOT"
  [[ -z "$TEMPORARY_LATEST_LINK" ]] || rm -f -- "$TEMPORARY_LATEST_LINK"
}
trap cleanup EXIT

umask 0077
note 'creating PostgreSQL backup dump'
compose exec --no-TTY postgres pg_dump --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --format=custom >"$DUMP_FILE"
unset TUNNEL_TOKEN
[[ -s "$DUMP_FILE" ]] || die 'PostgreSQL dump is empty'

SNAPSHOT_NAME=$(date --utc +%Y%m%dT%H%M%SZ)
FINAL_SNAPSHOT="$SNAPSHOTS_DIR/$SNAPSHOT_NAME"
[[ ! -e "$FINAL_SNAPSHOT" ]] || die "snapshot already exists: $FINAL_SNAPSHOT"
INCOMPLETE_SNAPSHOT=$(mktemp --directory "$SNAPSHOTS_DIR/.incomplete.XXXXXXXX")
install -d -o root -g root -m 0700 "$INCOMPLETE_SNAPSHOT/blobs"

RSYNC_ARGUMENTS=(--archive --delete --numeric-ids)
if [[ -n "$PREVIOUS_SNAPSHOT_NAME" ]]; then
  RSYNC_ARGUMENTS+=(--link-dest="$SNAPSHOTS_DIR/$PREVIOUS_SNAPSHOT_NAME/blobs")
fi

note 'copying immutable blobs into a plain incremental snapshot'
rsync "${RSYNC_ARGUMENTS[@]}" "$IMMUTABLE_BLOBS_DIR/" "$INCOMPLETE_SNAPSHOT/blobs/"
install -o root -g root -m 0600 "$DUMP_FILE" "$INCOMPLETE_SNAPSHOT/postgres.dump"
sync --file-system "$INCOMPLETE_SNAPSHOT"

mv --no-target-directory "$INCOMPLETE_SNAPSHOT" "$FINAL_SNAPSHOT"
INCOMPLETE_SNAPSHOT=
TEMPORARY_LATEST_LINK="$BACKUP_ROOT_DIR/.latest.$SNAPSHOT_NAME"
ln --symbolic "snapshots/$SNAPSHOT_NAME" "$TEMPORARY_LATEST_LINK"
mv --force --no-target-directory "$TEMPORARY_LATEST_LINK" "$BACKUP_ROOT_DIR/latest"
TEMPORARY_LATEST_LINK=

prune_snapshots
sync --file-system "$BACKUP_ROOT_DIR"
note "plain backup completed: $FINAL_SNAPSHOT"
