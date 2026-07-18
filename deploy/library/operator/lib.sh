#!/usr/bin/env bash

PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
export PATH

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

note() {
  printf '%s\n' "$*"
}

require_root() {
  [[ ${EUID} -eq 0 ]] || die 'this operator gateway must run through sudo'
}

require_no_arguments() {
  (($# == 0)) || die 'this operator gateway does not accept arguments'
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command is unavailable: $1"
}

require_root_controlled_file() {
  local file=$1
  local mode owner permissions

  [[ -f "$file" && ! -L "$file" ]] || die "required control file is missing: $file"
  owner=$(stat --format='%u' "$file")
  mode=$(stat --format='%a' "$file")
  permissions=${mode: -3}
  [[ "$owner" == '0' ]] || die "control file must be owned by root: $file"
  [[ ${permissions:1:1} != [2367] && ${permissions:2:1} != [2367] ]] ||
    die "control file must not be writable by group or others: $file"
}

require_root_controlled_tree() {
  local directory=$1
  local invalid unsafe

  [[ -d "$directory" && ! -L "$directory" ]] || die "required control tree is missing: $directory"
  invalid=$(find -P "$directory" -mindepth 1 ! -type d ! -type f -print -quit)
  [[ -z "$invalid" ]] || die "control tree contains an unsupported entry: $invalid"
  unsafe=$(find -P "$directory" \( ! -user root -o -perm /022 \) -print -quit)
  [[ -z "$unsafe" ]] || die "control tree is not root-controlled: $unsafe"
}
