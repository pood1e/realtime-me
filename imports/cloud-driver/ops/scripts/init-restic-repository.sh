#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

BACKUP_ENV_FILE=/etc/cloud-drive/backup.env

usage() {
  cat <<'USAGE'
Usage: init-restic-repository.sh [--backup-env-file PATH]

Initializes the encrypted restic repository on the mounted USB disk. It refuses
to initialize over an existing repository that cannot be opened with the given
password file.
USAGE
}

while (($# > 0)); do
  case "$1" in
    --backup-env-file)
      BACKUP_ENV_FILE=${2:?missing value for --backup-env-file}
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
require_command restic
require_command mountpoint
require_command install
require_secure_root_file "$BACKUP_ENV_FILE"

BACKUP_MOUNT_DIR=$(require_env_value "$BACKUP_ENV_FILE" BACKUP_MOUNT_DIR)
RESTIC_REPOSITORY=$(require_env_value "$BACKUP_ENV_FILE" RESTIC_REPOSITORY)
RESTIC_PASSWORD_FILE=$(require_env_value "$BACKUP_ENV_FILE" RESTIC_PASSWORD_FILE)

[[ "$BACKUP_MOUNT_DIR" == /* ]] || die 'BACKUP_MOUNT_DIR must be an absolute path'
[[ "$RESTIC_REPOSITORY" == "$BACKUP_MOUNT_DIR"/* ]] || die 'RESTIC_REPOSITORY must be located on BACKUP_MOUNT_DIR'
require_mountpoint "$BACKUP_MOUNT_DIR"
require_root_owned_nonwritable_directory "$BACKUP_MOUNT_DIR"
require_secure_root_file "$RESTIC_PASSWORD_FILE"
[[ -s "$RESTIC_PASSWORD_FILE" ]] || die "restic password file is empty: $RESTIC_PASSWORD_FILE"

if [[ -e "$RESTIC_REPOSITORY/config" ]]; then
  RESTIC_REPOSITORY="$RESTIC_REPOSITORY" RESTIC_PASSWORD_FILE="$RESTIC_PASSWORD_FILE" restic snapshots >/dev/null
  note 'restic repository already initialized and accessible'
  exit 0
fi

install -d -m 0700 "$(dirname -- "$RESTIC_REPOSITORY")"
RESTIC_REPOSITORY="$RESTIC_REPOSITORY" RESTIC_PASSWORD_FILE="$RESTIC_PASSWORD_FILE" restic init
note 'encrypted restic repository initialized'
