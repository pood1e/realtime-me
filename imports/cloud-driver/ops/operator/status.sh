#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

COMPOSE_SCRIPT=/opt/cloud-drive/ops/scripts/compose.sh

require_root
require_no_arguments "$@"
require_command findmnt
require_command systemctl
require_root_controlled_file "$COMPOSE_SCRIPT"

findmnt --noheadings --output SOURCE,TARGET /srv/cloud-drive/data
findmnt --noheadings --output SOURCE,TARGET /mnt/cloud-drive-backup
"$COMPOSE_SCRIPT" -- ps
systemctl show cloud-drive-backup.service --property=Result --property=ExecMainStatus
systemctl list-timers --no-pager cloud-drive-backup.timer
