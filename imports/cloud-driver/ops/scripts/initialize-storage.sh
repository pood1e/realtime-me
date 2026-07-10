#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

VG_NAME=
LV_NAME=cloud-drive-data
LV_SIZE=
MOUNT_DIR=/srv/cloud-drive/data
APP_UID=10001
APP_GID=10001
POSTGRES_UID=70
POSTGRES_GID=70

usage() {
  cat <<'USAGE'
Usage: initialize-storage.sh [options]

Creates one new ext4 logical volume for cloud-drive. This is intentionally a
one-time, destructive operation and refuses to reuse an existing logical volume.

Options:
  --vg-name NAME          Existing volume group (required)
  --lv-name NAME          New logical volume name (default: cloud-drive-data)
  --size SIZE             Logical volume size accepted by lvcreate (required)
  --mount-dir PATH        Mount point (default: /srv/cloud-drive/data)
  --app-uid UID           API container owner UID (default: 10001)
  --app-gid GID           API container owner GID (default: 10001)
  --postgres-uid UID      Postgres Alpine container owner UID (default: 70)
  --postgres-gid GID      Postgres Alpine container owner GID (default: 70)
  -h, --help              Show this help
USAGE
}

while (($# > 0)); do
  case "$1" in
    --vg-name)
      VG_NAME=${2:?missing value for --vg-name}
      shift 2
      ;;
    --lv-name)
      LV_NAME=${2:?missing value for --lv-name}
      shift 2
      ;;
    --size)
      LV_SIZE=${2:?missing value for --size}
      shift 2
      ;;
    --mount-dir)
      MOUNT_DIR=${2:?missing value for --mount-dir}
      shift 2
      ;;
    --app-uid)
      APP_UID=${2:?missing value for --app-uid}
      shift 2
      ;;
    --app-gid)
      APP_GID=${2:?missing value for --app-gid}
      shift 2
      ;;
    --postgres-uid)
      POSTGRES_UID=${2:?missing value for --postgres-uid}
      shift 2
      ;;
    --postgres-gid)
      POSTGRES_GID=${2:?missing value for --postgres-gid}
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
for command in vgs lvs lvcreate mkfs.ext4 blkid findmnt mount install chown awk; do
  require_command "$command"
done

[[ -n "$VG_NAME" ]] || die '--vg-name is required'
[[ -n "$LV_SIZE" ]] || die '--size is required'
[[ "$MOUNT_DIR" == /* ]] || die '--mount-dir must be an absolute path'
[[ "$APP_UID" =~ ^[0-9]+$ && "$APP_GID" =~ ^[0-9]+$ ]] || die 'API UID/GID must be numeric'
[[ "$POSTGRES_UID" =~ ^[0-9]+$ && "$POSTGRES_GID" =~ ^[0-9]+$ ]] || die 'Postgres UID/GID must be numeric'

LV_PATH="/dev/${VG_NAME}/${LV_NAME}"
vgs "$VG_NAME" >/dev/null
if lvs "$LV_PATH" >/dev/null 2>&1; then
  die "logical volume already exists; refusing to format or reuse it: $LV_PATH"
fi

install -d -m 0755 "$MOUNT_DIR"
EXISTING_TARGET=$(findmnt --noheadings --output TARGET --target "$MOUNT_DIR" 2>/dev/null | awk '{$1=$1; print}' || true)
[[ "$EXISTING_TARGET" != "$MOUNT_DIR" ]] || die "mount point is already mounted: $MOUNT_DIR"
if awk -v mount_dir="$MOUNT_DIR" '$1 !~ /^#/ && $2 == mount_dir { found = 1 } END { exit !found }' /etc/fstab; then
  die "an /etc/fstab entry already targets $MOUNT_DIR"
fi

note "creating ${LV_SIZE} logical volume ${LV_PATH}"
lvcreate --yes --size "$LV_SIZE" --name "$LV_NAME" "$VG_NAME"
mkfs.ext4 -F -L cloud-drive-data "$LV_PATH"

VOLUME_UUID=$(blkid -s UUID -o value "$LV_PATH")
[[ -n "$VOLUME_UUID" ]] || die "could not read ext4 UUID for $LV_PATH"
printf 'UUID=%s %s ext4 defaults,nodev,nosuid,noexec,noatime,x-systemd.device-timeout=30s 0 2\n' "$VOLUME_UUID" "$MOUNT_DIR" >>/etc/fstab
mount "$MOUNT_DIR"
require_mountpoint "$MOUNT_DIR"

install -d -m 0750 "$MOUNT_DIR/files"
chown "$APP_UID:$APP_GID" "$MOUNT_DIR/files"
install -d -m 0700 "$MOUNT_DIR/postgres"
chown "$POSTGRES_UID:$POSTGRES_GID" "$MOUNT_DIR/postgres"
install -d -o root -g root -m 0700 "$MOUNT_DIR/backup-staging"
printf 'cloud-drive-volume-v1\n' >"$MOUNT_DIR/.cloud-drive-volume"
chmod 0600 "$MOUNT_DIR/.cloud-drive-volume"

note "storage initialized at $MOUNT_DIR"
