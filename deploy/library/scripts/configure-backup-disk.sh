#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

USB_UUID=
MOUNT_DIR=/mnt/cloud-drive-backup

usage() {
  cat <<'USAGE'
Usage: configure-backup-disk.sh --uuid UUID [--mount-dir PATH]

Adds a non-destructive UUID-based ext4 mount to /etc/fstab for the existing USB
backup filesystem. It never formats the selected disk.
USAGE
}

while (($# > 0)); do
  case "$1" in
    --uuid)
      USB_UUID=${2:?missing value for --uuid}
      shift 2
      ;;
    --mount-dir)
      MOUNT_DIR=${2:?missing value for --mount-dir}
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
for command in blkid findmnt mount install awk chown chmod; do
  require_command "$command"
done

[[ -n "$USB_UUID" ]] || die '--uuid is required'
[[ "$USB_UUID" =~ ^[A-Fa-f0-9-]+$ ]] || die '--uuid must be a filesystem UUID'
[[ "$MOUNT_DIR" == /* ]] || die '--mount-dir must be an absolute path'

DEVICE=$(blkid -U "$USB_UUID" 2>/dev/null || true)
[[ -n "$DEVICE" ]] || die "no block device found for UUID: $USB_UUID"
FILESYSTEM_TYPE=$(blkid -s TYPE -o value "$DEVICE")
[[ "$FILESYSTEM_TYPE" == 'ext4' ]] || die "USB backup filesystem must be ext4, found: $FILESYSTEM_TYPE"

install -d -m 0750 "$MOUNT_DIR"
EXISTING_TARGET=$(findmnt --noheadings --output TARGET --target "$MOUNT_DIR" 2>/dev/null | awk '{$1=$1; print}' || true)
[[ "$EXISTING_TARGET" != "$MOUNT_DIR" ]] || die "mount point is already mounted: $MOUNT_DIR"
EXISTING_DEVICE_TARGET=$(findmnt --noheadings --output TARGET --source "$DEVICE" 2>/dev/null | awk '{$1=$1; print}' || true)
[[ -z "$EXISTING_DEVICE_TARGET" ]] || die "USB backup device is already mounted at: $EXISTING_DEVICE_TARGET"
if awk -v mount_dir="$MOUNT_DIR" '$1 !~ /^#/ && $2 == mount_dir { found = 1 } END { exit !found }' /etc/fstab; then
  die "an /etc/fstab entry already targets $MOUNT_DIR"
fi

printf 'UUID=%s %s ext4 defaults,nodev,nosuid,noexec,nofail,x-systemd.device-timeout=30s 0 2\n' "$USB_UUID" "$MOUNT_DIR" >>/etc/fstab
mount "$MOUNT_DIR"
require_mountpoint "$MOUNT_DIR"
chown root:root "$MOUNT_DIR"
chmod 0711 "$MOUNT_DIR"
note "USB backup disk mounted at $MOUNT_DIR"
