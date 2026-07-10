#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

require_root
require_no_arguments "$@"
require_command systemctl

systemctl start cloud-drive-backup.service
systemctl show cloud-drive-backup.service --property=Result --property=ExecMainStatus
systemctl is-active cloud-drive-backup.timer
