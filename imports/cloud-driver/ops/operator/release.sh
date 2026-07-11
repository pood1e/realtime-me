#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

INCOMING_API_DIR=/var/lib/cloud-drive-release/incoming-api
SEALED_API_DIR=/var/lib/cloud-drive-release/sealed-api
PREVIOUS_API_DIR=/var/lib/cloud-drive-release/previous-api
INSTALL_API_DIR=/opt/cloud-drive/api
INCOMING_COMPOSE_FILE=/var/lib/cloud-drive-release/incoming-compose/docker-compose.yml
INSTALL_COMPOSE_FILE=/opt/cloud-drive/ops/docker-compose.yml
DEPLOY_SCRIPT=/opt/cloud-drive/ops/scripts/deploy.sh
BACKUP_SCRIPT=/opt/cloud-drive/ops/scripts/backup.sh
COMPOSE_VALIDATOR=/usr/local/libexec/cloud-drive-operator/validate-release-compose.sh
LOCK_FILE=/run/lock/cloud-drive-release.lock
MAX_COMPOSE_BYTES=$((128 * 1024))

require_root
require_no_arguments "$@"
for command in find flock install mktemp rm rsync stat; do
  require_command "$command"
done
for file in \
  /opt/cloud-drive/.dockerignore \
  /opt/cloud-drive/api/Dockerfile \
  "$BACKUP_SCRIPT" \
  "$DEPLOY_SCRIPT" \
  "$INSTALL_COMPOSE_FILE" \
  "$COMPOSE_VALIDATOR"; do
  require_root_controlled_file "$file"
done
[[ -d "$INCOMING_API_DIR" && ! -L "$INCOMING_API_DIR" ]] ||
  die "incoming API directory is unavailable: $INCOMING_API_DIR"
[[ -f "$INCOMING_COMPOSE_FILE" && ! -L "$INCOMING_COMPOSE_FILE" ]] ||
  die 'stage docker-compose.yml before releasing'
COMPOSE_BYTES=$(stat --format='%s' "$INCOMING_COMPOSE_FILE")
[[ "$COMPOSE_BYTES" =~ ^[0-9]+$ ]] || die 'could not determine staged Compose size'
((COMPOSE_BYTES > 0 && COMPOSE_BYTES <= MAX_COMPOSE_BYTES)) ||
  die 'staged Compose file has an invalid size'

exec 9>"$LOCK_FILE"
flock -n 9 || die 'another cloud-drive release is already running'

WORK_DIR=$(mktemp -d /var/lib/cloud-drive-release/release.XXXXXXXX)
cleanup() {
  rm -rf -- "$WORK_DIR" "${SEALED_API_DIR:?}"
}
trap cleanup EXIT

rm -rf -- "${SEALED_API_DIR:?}"
install -d -o root -g root -m 0700 "$SEALED_API_DIR"
rsync -a --delete --exclude=/Dockerfile --chown=root:root --chmod=Dgo-w,Fgo-w \
  "$INCOMING_API_DIR/" "$SEALED_API_DIR/"
INVALID_ENTRY=$(find -P "$SEALED_API_DIR" -mindepth 1 ! -type d ! -type f -print -quit)
[[ -z "$INVALID_ENTRY" ]] || die "release contains an unsupported filesystem entry: $INVALID_ENTRY"
for file in go.mod cmd/migrate/main.go cmd/server/main.go cmd/worker/main.go vendor/modules.txt; do
  [[ -f "$SEALED_API_DIR/$file" ]] || die "release is missing required API source: $file"
done

CANDIDATE_COMPOSE_FILE="$WORK_DIR/docker-compose.yml"
PREVIOUS_COMPOSE_FILE="$WORK_DIR/docker-compose.previous.yml"
install -o root -g root -m 0600 "$INCOMING_COMPOSE_FILE" "$CANDIDATE_COMPOSE_FILE"
"$COMPOSE_VALIDATOR" "$CANDIDATE_COMPOSE_FILE" "$WORK_DIR"

note 'creating a pre-release backup'
"$BACKUP_SCRIPT"

rm -rf -- "${PREVIOUS_API_DIR:?}"
install -d -o root -g root -m 0700 "$PREVIOUS_API_DIR"
rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w "$INSTALL_API_DIR/" "$PREVIOUS_API_DIR/"
install -o root -g root -m 0644 "$INSTALL_COMPOSE_FILE" "$PREVIOUS_COMPOSE_FILE"

restore_previous() {
  rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w "$PREVIOUS_API_DIR/" "$INSTALL_API_DIR/"
  install -o root -g root -m 0644 "$PREVIOUS_COMPOSE_FILE" "$INSTALL_COMPOSE_FILE"
}

if ! rsync -a --delete --exclude=/Dockerfile --chown=root:root --chmod=Dgo-w,Fgo-w \
  "$SEALED_API_DIR/" "$INSTALL_API_DIR/" ||
  ! install -o root -g root -m 0644 "$CANDIDATE_COMPOSE_FILE" "$INSTALL_COMPOSE_FILE"; then
  restore_previous
  die 'installing the staged release failed; the previous files were restored'
fi

if "$DEPLOY_SCRIPT"; then
  note 'operator release completed'
  exit 0
fi

die 'deployment failed after the migration boundary; the candidate was retained for a forward fix'
