#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../.." && pwd -P)
ENV_FILE=/etc/cloud-drive/runtime.env
COMPOSE_FILE=
SNAPSHOT=
CONFIRM=false

usage() {
  cat <<'USAGE'
Usage: restore.sh --snapshot PATH --confirm-destroy [options]

Destructively restores PostgreSQL and immutable content objects from one plain
backup snapshot. Derived artifacts are discarded and queued for regeneration.

Options:
  --snapshot PATH    Snapshot containing postgres.dump and objects/ (required)
  --confirm-destroy  Confirm replacement of current application data (required)
  --repo-dir PATH    Checked-out cloud-drive repository
  --env-file PATH    Root-only runtime environment file
  --compose-file PATH
  -h, --help         Show this help
USAGE
}

while (($# > 0)); do
  case "$1" in
    --snapshot)
      SNAPSHOT=${2:?missing value for --snapshot}
      shift 2
      ;;
    --confirm-destroy)
      CONFIRM=true
      shift
      ;;
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
for command in docker mountpoint rsync install rm; do
  require_command "$command"
done
[[ "$CONFIRM" == true ]] || die '--confirm-destroy is required'
[[ -n "$SNAPSHOT" ]] || die '--snapshot is required'
SNAPSHOT=$(cd -- "$SNAPSHOT" && pwd -P)
[[ -f "$SNAPSHOT/postgres.dump" ]] || die "missing snapshot database dump: $SNAPSHOT/postgres.dump"
[[ -d "$SNAPSHOT/objects" ]] || die "missing snapshot objects: $SNAPSHOT/objects"

REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
COMPOSE_FILE=${COMPOSE_FILE:-$REPO_DIR/ops/docker-compose.yml}
require_regular_file "$COMPOSE_FILE"
require_secure_root_file "$ENV_FILE"

POSTGRES_USER=$(require_env_value "$ENV_FILE" POSTGRES_USER)
POSTGRES_DB=$(require_env_value "$ENV_FILE" POSTGRES_DB)
VOLUME_MOUNT_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_VOLUME_MOUNT_DIR)
DATA_DIR=$(require_env_value "$ENV_FILE" CLOUD_DRIVE_DATA_DIR)
TUNNEL_TOKEN=$(read_cloudflare_tunnel_token "$ENV_FILE")
require_mountpoint "$VOLUME_MOUNT_DIR"
[[ -f "$VOLUME_MOUNT_DIR/.cloud-drive-volume" ]] || die 'primary volume marker is missing'
[[ "$DATA_DIR" == "$VOLUME_MOUNT_DIR"/* ]] || die 'data directory is outside the primary volume'

compose() {
  TUNNEL_TOKEN="$TUNNEL_TOKEN" docker compose \
    --project-directory "$REPO_DIR" \
    --project-name cloud-drive \
    --env-file "$ENV_FILE" \
    --file "$COMPOSE_FILE" \
    "$@"
}

note 'stopping request and processing services'
compose stop api worker cloudflared || true
compose up --detach --wait postgres

note 'replacing immutable content objects'
install -d -o 10001 -g 10001 -m 0700 "$DATA_DIR/objects"
rsync --archive --delete --numeric-ids "$SNAPSHOT/objects/" "$DATA_DIR/objects/"
rm -rf -- "$DATA_DIR/artifacts" "$DATA_DIR/uploads" "$DATA_DIR/work"
install -d -o 10001 -g 10001 -m 0700 \
  "$DATA_DIR/artifacts" "$DATA_DIR/uploads" "$DATA_DIR/work"

note 'restoring PostgreSQL metadata'
compose exec --no-TTY postgres dropdb --username "$POSTGRES_USER" --if-exists --force "$POSTGRES_DB"
compose exec --no-TTY postgres createdb --username "$POSTGRES_USER" --owner "$POSTGRES_USER" "$POSTGRES_DB"
compose exec --no-TTY postgres pg_restore \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --exit-on-error <"$SNAPSHOT/postgres.dump"

note 'queuing deterministic artifact regeneration'
compose exec --no-TTY postgres psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --set ON_ERROR_STOP=1 <<'SQL'
BEGIN;
TRUNCATE content_artifacts;
DELETE FROM processing_jobs;
DELETE FROM worker_state;
UPDATE books SET processing_status = 'pending';
UPDATE tracks SET processing_status = 'pending';
UPDATE images SET processing_status = 'pending';
INSERT INTO processing_jobs (uid, kind, resource_uid, status)
SELECT gen_random_uuid()::text, 'book', uid, 'pending' FROM books
UNION ALL
SELECT gen_random_uuid()::text, 'track', uid, 'pending' FROM tracks
UNION ALL
SELECT gen_random_uuid()::text, 'image', uid, 'pending' FROM images
UNION ALL
SELECT gen_random_uuid()::text, 'wallpaper', uid, 'pending' FROM wallpapers;
COMMIT;
SQL

note 'starting the restored service suite'
compose up --build --detach --wait --wait-timeout 180 --remove-orphans
compose ps --status running
unset TUNNEL_TOKEN
note 'restore completed; derived artifacts are rebuilding in the worker'
