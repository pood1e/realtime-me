#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

COMPOSE_SCRIPT=/opt/cloud-drive/ops/scripts/compose.sh

require_root
(($# == 1)) || die 'usage: cloud-drive-logs api|worker|migrate|postgres|cloudflared|backup'
case "$1" in
  api|worker|migrate|postgres|cloudflared)
    require_root_controlled_file "$COMPOSE_SCRIPT"
    exec "$COMPOSE_SCRIPT" -- logs --no-color --tail=200 "$1"
    ;;
  backup)
    require_command journalctl
    exec journalctl --unit=cloud-drive-backup.service --lines=200 --no-pager
    ;;
  *)
    die 'usage: cloud-drive-logs api|worker|migrate|postgres|cloudflared|backup'
    ;;
esac
