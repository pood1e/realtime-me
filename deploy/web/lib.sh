#!/usr/bin/env bash

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

require_regular_file() {
  [[ -f "$1" ]] || die "required file does not exist: $1"
}

require_env_value() {
  local file=$1
  local key=$2
  local value

  value=$(awk -v key="$key" '
    $0 ~ "^[[:space:]]*" key "=" {
      sub("^[[:space:]]*" key "=", "", $0)
      sub(/\r$/, "", $0)
      print
      exit
    }
  ' "$file")
  [[ -n "$value" ]] || die "missing $key in $file"
  [[ "$value" != *REPLACE_* ]] || die "$key in $file still contains a placeholder"
  printf '%s' "$value"
}
