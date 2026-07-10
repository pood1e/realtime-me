#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../.." && pwd -P)
OPERATOR_USER=
OPERATOR_GROUP=cloud-drive-operators
LIBEXEC_DIR=/usr/local/libexec/cloud-drive-operator
SBIN_DIR=/usr/local/sbin
STATE_DIR=/var/lib/cloud-drive-release
INCOMING_DIR="$STATE_DIR/incoming-api"
SUDOERS_FILE=/etc/sudoers.d/cloud-drive-operators

usage() {
  cat <<'USAGE'
Usage: install-operator-access.sh --user USER [--repo-dir PATH]

Grants one existing login user access only to root-owned cloud-drive operator
gateways. It does not grant Docker socket access or general passwordless sudo.
USAGE
}

while (($# > 0)); do
  case "$1" in
    --user)
      OPERATOR_USER=${2:?missing value for --user}
      shift 2
      ;;
    --repo-dir)
      REPO_DIR=${2:?missing value for --repo-dir}
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
for command in getent groupadd id install mktemp usermod visudo; do
  require_command "$command"
done
[[ -n "$OPERATOR_USER" ]] || die '--user is required'
id "$OPERATOR_USER" >/dev/null 2>&1 || die "operator user does not exist: $OPERATOR_USER"
REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
[[ "$REPO_DIR" == '/opt/cloud-drive' ]] || die 'install operator access from the canonical /opt/cloud-drive checkout'

if ! getent group "$OPERATOR_GROUP" >/dev/null; then
  groupadd --system "$OPERATOR_GROUP"
fi
usermod --append --groups "$OPERATOR_GROUP" "$OPERATOR_USER"

install -d -o root -g root -m 0755 "$LIBEXEC_DIR" "$STATE_DIR"
install -d -o root -g "$OPERATOR_GROUP" -m 2770 "$INCOMING_DIR"
install -o root -g root -m 0644 "$REPO_DIR/ops/operator/lib.sh" "$LIBEXEC_DIR/lib.sh"
install -o root -g root -m 0755 "$REPO_DIR/ops/operator/release-api.sh" "$SBIN_DIR/cloud-drive-release-api"
install -o root -g root -m 0755 "$REPO_DIR/ops/operator/backup-now.sh" "$SBIN_DIR/cloud-drive-backup-now"
install -o root -g root -m 0755 "$REPO_DIR/ops/operator/status.sh" "$SBIN_DIR/cloud-drive-status"
install -o root -g root -m 0755 "$REPO_DIR/ops/operator/logs.sh" "$SBIN_DIR/cloud-drive-logs"

SUDOERS_TEMP=$(mktemp)
cleanup() {
  rm -f -- "$SUDOERS_TEMP"
}
trap cleanup EXIT
cat >"$SUDOERS_TEMP" <<EOF
Cmnd_Alias CLOUD_DRIVE_OPERATOR_COMMANDS = $SBIN_DIR/cloud-drive-release-api, $SBIN_DIR/cloud-drive-backup-now, $SBIN_DIR/cloud-drive-status, $SBIN_DIR/cloud-drive-logs
%$OPERATOR_GROUP ALL=(root) NOPASSWD: CLOUD_DRIVE_OPERATOR_COMMANDS
EOF
visudo -cf "$SUDOERS_TEMP"
install -o root -g root -m 0440 "$SUDOERS_TEMP" "$SUDOERS_FILE"
visudo -cf "$SUDOERS_FILE"

note "cloud-drive operator access installed for $OPERATOR_USER"
note 'start a new login session before using the operator gateways'
