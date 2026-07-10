#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../.." && pwd -P)
SYSTEMD_DIR=/etc/systemd/system

usage() {
  cat <<'USAGE'
Usage: install-backup-timer.sh [--repo-dir PATH]

Installs and enables the cloud-drive daily plain-snapshot backup timer. The unit uses
the canonical production checkout at /opt/cloud-drive.
USAGE
}

while (($# > 0)); do
  case "$1" in
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
require_command install
require_command systemctl
REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
[[ "$REPO_DIR" == '/opt/cloud-drive' ]] || die 'install the production checkout at /opt/cloud-drive before installing this timer'

install -o root -g root -m 0644 \
  "$REPO_DIR/ops/systemd/cloud-drive-backup.service" \
  "$SYSTEMD_DIR/cloud-drive-backup.service"
install -o root -g root -m 0644 \
  "$REPO_DIR/ops/systemd/cloud-drive-backup.timer" \
  "$SYSTEMD_DIR/cloud-drive-backup.timer"

systemctl daemon-reload
systemctl enable --now cloud-drive-backup.timer
systemctl status --no-pager cloud-drive-backup.timer
