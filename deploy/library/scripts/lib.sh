#!/usr/bin/env bash

# Shared helpers for host-side cloud-drive administration scripts.

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

note() {
  printf '%s\n' "$*"
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command is unavailable: $1"
}

require_root() {
  [[ ${EUID} -eq 0 ]] || die 'run this host administration script as root'
}

require_regular_file() {
  [[ -f "$1" ]] || die "required file does not exist: $1"
}

require_secure_root_file() {
  local file=$1
  local owner mode

  require_regular_file "$file"
  owner=$(stat --format='%u' "$file")
  mode=$(stat --format='%a' "$file")
  [[ "$owner" == '0' ]] || die "file must be owned by root: $file"
  case "$mode" in
    400|600)
      ;;
    *)
      die "file must have mode 0400 or 0600: $file"
      ;;
  esac
}

require_root_owned_nonwritable_directory() {
  local directory=$1
  local mode owner permissions

  [[ -d "$directory" ]] || die "required directory does not exist: $directory"
  owner=$(stat --format='%u' "$directory")
  mode=$(stat --format='%a' "$directory")
  permissions=${mode: -3}
  [[ "$owner" == '0' ]] || die "directory must be owned by root: $directory"
  [[ ${permissions:1:1} != [2367] && ${permissions:2:1} != [2367] ]] ||
    die "directory must not be writable by group or others: $directory"
}

is_placeholder() {
  case "$1" in
    *REPLACE_*|*CHANGE_*|*EXAMPLE_*|*YOUR_*|*'<'*'>'*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

read_env_value() {
  local file=$1
  local key=$2

  awk -v key="$key" '
    $0 ~ "^[[:space:]]*" key "=" {
      sub("^[[:space:]]*" key "=", "", $0)
      sub(/\r$/, "", $0)
      print
      exit
    }
  ' "$file"
}

require_env_value() {
  local file=$1
  local key=$2
  local value

  value=$(read_env_value "$file" "$key")
  [[ -n "$value" ]] || die "missing $key in $file"
  is_placeholder "$value" && die "$key in $file still contains a placeholder"
  printf '%s' "$value"
}

require_mountpoint() {
  local path=$1

  [[ -d "$path" ]] || die "required directory does not exist: $path"
  mountpoint -q "$path" || die "required path is not a mounted filesystem: $path"
}

available_bytes() {
  local path=$1

  df --block-size=1 --output=avail "$path" | awk 'NR == 2 { print $1 }'
}
