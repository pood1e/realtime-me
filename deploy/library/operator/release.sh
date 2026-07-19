#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

INCOMING_SOURCE_DIR=/var/lib/cloud-drive-release/incoming-source
SEALED_SOURCE_DIR=/var/lib/cloud-drive-release/sealed-source
PREVIOUS_SOURCE_DIR=/var/lib/cloud-drive-release/previous-source
INSTALL_SOURCE_DIR=/opt/cloud-drive/services/library
INCOMING_COMPOSE_FILE=/var/lib/cloud-drive-release/incoming-compose/compose.yaml
INSTALL_COMPOSE_FILE=/opt/cloud-drive/deploy/library/compose.yaml
DEPLOY_SCRIPT=/opt/cloud-drive/deploy/library/scripts/deploy.sh
BACKUP_SCRIPT=/opt/cloud-drive/deploy/library/scripts/backup.sh
COMPOSE_VALIDATOR=/usr/local/libexec/cloud-drive-operator/validate-release-compose.sh
LOCK_FILE=/run/lock/cloud-drive-release.lock
MAX_COMPOSE_BYTES=$((128 * 1024))

require_root
require_no_arguments "$@"
for command in cmp find flock install mktemp rm rsync stat; do
  require_command "$command"
done
for file in \
  /opt/cloud-drive/.dockerignore \
  /opt/cloud-drive/go.mod \
  /opt/cloud-drive/go.sum \
  /opt/cloud-drive/vendor/modules.txt \
  /opt/cloud-drive/gen/go/realtime/me/auth/v1/permission.pb.go \
  /opt/cloud-drive/libs/go/authn/verifier.go \
  /opt/cloud-drive/libs/go/serviceauth/key.go \
  /opt/cloud-drive/services/library/Dockerfile \
  /opt/cloud-drive/services/library/public.Caddyfile \
  "$BACKUP_SCRIPT" \
  "$DEPLOY_SCRIPT" \
  "$INSTALL_COMPOSE_FILE" \
  "$COMPOSE_VALIDATOR"; do
  require_root_controlled_file "$file"
done
require_root_controlled_tree /opt/cloud-drive/vendor
require_root_controlled_tree /opt/cloud-drive/gen/go/realtime/me/auth
require_root_controlled_tree /opt/cloud-drive/gen/go/realtime/me/library
require_root_controlled_tree /opt/cloud-drive/libs/go/authn
require_root_controlled_tree /opt/cloud-drive/libs/go/serviceauth
[[ -d "$INCOMING_SOURCE_DIR" && ! -L "$INCOMING_SOURCE_DIR" ]] ||
  die "incoming Library source directory is unavailable: $INCOMING_SOURCE_DIR"
[[ -f "$INCOMING_COMPOSE_FILE" && ! -L "$INCOMING_COMPOSE_FILE" ]] ||
  die 'stage compose.yaml before releasing'
COMPOSE_BYTES=$(stat --format='%s' "$INCOMING_COMPOSE_FILE")
[[ "$COMPOSE_BYTES" =~ ^[0-9]+$ ]] || die 'could not determine staged Compose size'
((COMPOSE_BYTES > 0 && COMPOSE_BYTES <= MAX_COMPOSE_BYTES)) ||
  die 'staged Compose file has an invalid size'

exec 9>"$LOCK_FILE"
flock -n 9 || die 'another cloud-drive release is already running'

WORK_DIR=$(mktemp -d /var/lib/cloud-drive-release/release.XXXXXXXX)
cleanup() {
  rm -rf -- "$WORK_DIR" "${SEALED_SOURCE_DIR:?}"
}
trap cleanup EXIT

rm -rf -- "${SEALED_SOURCE_DIR:?}"
install -d -o root -g root -m 0700 "$SEALED_SOURCE_DIR"
rsync -a --delete --exclude=/Dockerfile --chown=root:root --chmod=Dgo-w,Fgo-w \
  "$INCOMING_SOURCE_DIR/" "$SEALED_SOURCE_DIR/"
INVALID_ENTRY=$(find -P "$SEALED_SOURCE_DIR" -mindepth 1 ! -type d ! -type f -print -quit)
[[ -z "$INVALID_ENTRY" ]] || die "release contains an unsupported filesystem entry: $INVALID_ENTRY"
for file in cmd/migrate/main.go cmd/server/main.go cmd/worker/main.go; do
  [[ -f "$SEALED_SOURCE_DIR/$file" ]] || die "release is missing required Library source: $file"
done
cmp -s "$SEALED_SOURCE_DIR/public.Caddyfile" /opt/cloud-drive/services/library/public.Caddyfile ||
  die 'release cannot change the root-controlled public API allowlist'

CANDIDATE_COMPOSE_FILE="$WORK_DIR/compose.yaml"
PREVIOUS_COMPOSE_FILE="$WORK_DIR/compose.previous.yaml"
install -o root -g root -m 0600 "$INCOMING_COMPOSE_FILE" "$CANDIDATE_COMPOSE_FILE"
"$COMPOSE_VALIDATOR" "$CANDIDATE_COMPOSE_FILE" "$WORK_DIR"

note 'creating a pre-release backup'
"$BACKUP_SCRIPT"

rm -rf -- "${PREVIOUS_SOURCE_DIR:?}"
install -d -o root -g root -m 0700 "$PREVIOUS_SOURCE_DIR"
rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w "$INSTALL_SOURCE_DIR/" "$PREVIOUS_SOURCE_DIR/"
install -o root -g root -m 0644 "$INSTALL_COMPOSE_FILE" "$PREVIOUS_COMPOSE_FILE"

restore_previous() {
  rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w "$PREVIOUS_SOURCE_DIR/" "$INSTALL_SOURCE_DIR/"
  install -o root -g root -m 0644 "$PREVIOUS_COMPOSE_FILE" "$INSTALL_COMPOSE_FILE"
}

if ! rsync -a --delete --exclude=/Dockerfile --chown=root:root --chmod=Dgo-w,Fgo-w \
  "$SEALED_SOURCE_DIR/" "$INSTALL_SOURCE_DIR/" ||
  ! install -o root -g root -m 0644 "$CANDIDATE_COMPOSE_FILE" "$INSTALL_COMPOSE_FILE"; then
  restore_previous
  die 'installing the staged release failed; the previous files were restored'
fi

if "$DEPLOY_SCRIPT"; then
  note 'operator release completed'
  exit 0
fi

die 'deployment failed after the migration boundary; the candidate was retained for a forward fix'
