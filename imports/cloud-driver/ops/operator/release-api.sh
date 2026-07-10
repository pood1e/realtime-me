#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

INCOMING_DIR=/var/lib/cloud-drive-release/incoming-api
SEALED_DIR=/var/lib/cloud-drive-release/sealed-api
PREVIOUS_DIR=/var/lib/cloud-drive-release/previous-api
INSTALL_DIR=/opt/cloud-drive/api
DEPLOY_SCRIPT=/opt/cloud-drive/ops/scripts/deploy.sh
LOCK_FILE=/run/lock/cloud-drive-release.lock

require_root
require_no_arguments "$@"
for command in flock find install rsync rm stat; do
  require_command "$command"
done
for file in \
  /opt/cloud-drive/.dockerignore \
  /opt/cloud-drive/api/Dockerfile \
  /opt/cloud-drive/ops/docker-compose.yml \
  "$DEPLOY_SCRIPT"; do
  require_root_controlled_file "$file"
done
[[ -d "$INCOMING_DIR" && ! -L "$INCOMING_DIR" ]] || die "incoming API directory is unavailable: $INCOMING_DIR"

exec 9>"$LOCK_FILE"
flock -n 9 || die 'another cloud-drive release is already running'

cleanup() {
  rm -rf -- "${SEALED_DIR:?}"
}
trap cleanup EXIT

rm -rf -- "${SEALED_DIR:?}"
install -d -o root -g root -m 0700 "$SEALED_DIR"
rsync -a --delete --exclude=/Dockerfile --chown=root:root --chmod=Dgo-w,Fgo-w \
  "$INCOMING_DIR/" "$SEALED_DIR/"

INVALID_ENTRY=$(find -P "$SEALED_DIR" -mindepth 1 ! -type d ! -type f -print -quit)
[[ -z "$INVALID_ENTRY" ]] || die "release contains an unsupported filesystem entry: $INVALID_ENTRY"
for file in go.mod cmd/server/main.go cmd/worker/main.go vendor/modules.txt; do
  [[ -f "$SEALED_DIR/$file" ]] || die "release is missing required API source: $file"
done

rm -rf -- "${PREVIOUS_DIR:?}"
install -d -o root -g root -m 0700 "$PREVIOUS_DIR"
rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w "$INSTALL_DIR/" "$PREVIOUS_DIR/"

restore_previous() {
  rsync -a --delete --chown=root:root --chmod=Dgo-w,Fgo-w "$PREVIOUS_DIR/" "$INSTALL_DIR/"
}

if ! rsync -a --delete --exclude=/Dockerfile --chown=root:root --chmod=Dgo-w,Fgo-w \
  "$SEALED_DIR/" "$INSTALL_DIR/"; then
  restore_previous
  die 'installing the staged API release failed; the previous source was restored'
fi

if "$DEPLOY_SCRIPT"; then
  note 'operator API release completed'
  exit 0
fi

note 'new API release failed; restoring the previous source'
restore_previous
if ! "$DEPLOY_SCRIPT"; then
  die 'the release and automatic rollback both failed'
fi
die 'the release failed and was rolled back successfully'
