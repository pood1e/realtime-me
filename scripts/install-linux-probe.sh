#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR=${INSTALL_DIR:-/opt/realtime-me}
ENV_FILE=${ENV_FILE:-/etc/realtime-me.env}
GATEWAY_URL=${STATUS_GATEWAY_URL:-}
DEVICE_ID=${STATUS_DEVICE_ID:-$(hostname -s 2>/dev/null || hostname)}
DEVICE_NAME=${STATUS_DEVICE_NAME:-$(hostname 2>/dev/null || echo linux)}
DEVICE_KIND=${STATUS_DEVICE_KIND:-host}
DEVICE_ROLE=${STATUS_DEVICE_ROLE:-desktop}
INSTALL_AGENT=${INSTALL_AGENT:-0}
PROXY_ENV_NAMES=(
  http_proxy
  https_proxy
  all_proxy
  no_proxy
  HTTP_PROXY
  HTTPS_PROXY
  ALL_PROXY
  NO_PROXY
)
if [[ -n ${REALTIME_ME_RAW_BASE_URL:-} ]]; then
  RAW_BASE_URLS=("$REALTIME_ME_RAW_BASE_URL")
elif [[ -n ${REALTIME_ME_RAW_BASE_URLS:-} ]]; then
  read -r -a RAW_BASE_URLS <<<"$REALTIME_ME_RAW_BASE_URLS"
else
  RAW_BASE_URLS=(
    "https://cdn.jsdelivr.net/gh/pood1e/realtime-me@main/scripts"
    "https://raw.githubusercontent.com/pood1e/realtime-me/main/scripts"
  )
fi
PYTHON_BIN=${PYTHON_BIN:-/usr/bin/python3}
CURL_BIN=${CURL_BIN:-/usr/bin/curl}

require_root() {
  if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
    echo "Run with sudo or as root." >&2
    exit 2
  fi
}

log() {
  echo "[realtime-me] $*" >&2
}

set_proxy_env_if_missing() {
  local name=$1
  local value=$2
  case "$name" in
    http_proxy|https_proxy|all_proxy|no_proxy|HTTP_PROXY|HTTPS_PROXY|ALL_PROXY|NO_PROXY) ;;
    *) return ;;
  esac
  [[ -n $value ]] || return
  if [[ -z ${!name:-} ]]; then
    export "$name=$value"
  fi
}

import_proxy_env_from_process() {
  local pid=$1
  local environ_file=/proc/$pid/environ
  local entry
  local name
  local value
  [[ -r $environ_file ]] || return
  while IFS= read -r -d '' entry; do
    name=${entry%%=*}
    value=${entry#*=}
    set_proxy_env_if_missing "$name" "$value"
  done <"$environ_file"
}

inherit_proxy_env() {
  [[ -d /proc ]] || return
  local pid=$PPID
  local depth=0
  while [[ $pid =~ ^[0-9]+$ && $pid -gt 1 && $depth -lt 8 ]]; do
    import_proxy_env_from_process "$pid"
    pid=$(awk '/^PPid:/ { print $2 }' "/proc/$pid/status" 2>/dev/null || true)
    [[ -n $pid ]] || return
    depth=$((depth + 1))
  done
}

require_command() {
  local command=$1
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Missing required command: $command" >&2
    exit 2
  fi
}

read_gateway_url() {
  if [[ -n ${GATEWAY_URL:-} ]]; then
    normalize_url "$GATEWAY_URL"
    return
  fi
  if [[ ! -r /dev/tty ]]; then
    echo "Set STATUS_GATEWAY_URL or run from an interactive terminal." >&2
    exit 2
  fi
  local value
  read -rp "STATUS_GATEWAY_URL: " value </dev/tty
  if [[ -z $value ]]; then
    echo "STATUS_GATEWAY_URL is required." >&2
    exit 2
  fi
  normalize_url "$value"
}

normalize_url() {
  local value=${1%/}
  case "$value" in
    http://*|https://*) printf '%s' "$value" ;;
    *)
      echo "STATUS_GATEWAY_URL must start with http:// or https://." >&2
      exit 2
      ;;
  esac
}

read_token() {
  if [[ -n ${STATUS_INGEST_TOKEN:-} ]]; then
    printf '%s' "$STATUS_INGEST_TOKEN"
    return
  fi
  if [[ ! -r /dev/tty ]]; then
    echo "Set STATUS_INGEST_TOKEN or run from an interactive terminal." >&2
    exit 2
  fi
  local token
  read -rsp "STATUS_INGEST_TOKEN: " token </dev/tty
  echo >/dev/tty
  if [[ -z $token ]]; then
    echo "STATUS_INGEST_TOKEN is required." >&2
    exit 2
  fi
  printf '%s' "$token"
}

write_env_file() {
  local token=$1
  umask 077
  cat >"$ENV_FILE" <<ENV
STATUS_GATEWAY_URL=$GATEWAY_URL
STATUS_DEVICE_ID=$DEVICE_ID
STATUS_DEVICE_NAME=$DEVICE_NAME
STATUS_DEVICE_KIND=$DEVICE_KIND
STATUS_DEVICE_ROLE=$DEVICE_ROLE
STATUS_INGEST_TOKEN=$token
ENV
  append_proxy_env_file
  chmod 600 "$ENV_FILE"
}

append_proxy_env_file() {
  local name
  local value
  for name in "${PROXY_ENV_NAMES[@]}"; do
    value=${!name:-}
    [[ -n $value ]] || continue
    printf '%s=%s\n' "$name" "$value" >>"$ENV_FILE"
  done
}

download_reporters() {
  install -d -m 755 "$INSTALL_DIR"
  download_file status-device-reporter.py "$INSTALL_DIR/status-device-reporter.py"
  chmod 755 "$INSTALL_DIR/status-device-reporter.py"
  if [[ $INSTALL_AGENT == 1 ]]; then
    download_file agent-status-reporter.py "$INSTALL_DIR/agent-status-reporter.py"
    chmod 755 "$INSTALL_DIR/agent-status-reporter.py"
  fi
}

download_file() {
  local name=$1
  local destination=$2
  local temporary
  local curl_error
  temporary=$(mktemp)
  curl_error=$(mktemp)
  log "Downloading $name"
  for base_url in "${RAW_BASE_URLS[@]}"; do
    if "$CURL_BIN" -fsSL --connect-timeout 10 --max-time 60 "${base_url%/}/$name" -o "$temporary" 2>"$curl_error"; then
      rm -f "$curl_error"
      mv "$temporary" "$destination"
      return
    fi
    log "Download mirror failed; trying next mirror"
  done
  rm -f "$temporary" "$curl_error"
  echo "Could not download $name from configured mirrors." >&2
  exit 1
}

write_systemd_unit() {
  cat >/etc/systemd/system/realtime-me-status-device.service <<SERVICE
[Unit]
Description=Realtime Me Linux status reporter
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
EnvironmentFile=$ENV_FILE
ExecStart=$PYTHON_BIN $INSTALL_DIR/status-device-reporter.py
SERVICE

  cat >/etc/systemd/system/realtime-me-status-device.timer <<'TIMER'
[Unit]
Description=Run Realtime Me Linux status reporter

[Timer]
OnBootSec=15s
OnUnitActiveSec=10s
AccuracySec=1s

[Install]
WantedBy=timers.target
TIMER
}

write_agent_unit() {
  [[ $INSTALL_AGENT == 1 ]] || return
  cat >/etc/systemd/system/realtime-me-agent.service <<SERVICE
[Unit]
Description=Realtime Me local agent status reporter
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
EnvironmentFile=$ENV_FILE
ExecStart=$PYTHON_BIN $INSTALL_DIR/agent-status-reporter.py
SERVICE

  cat >/etc/systemd/system/realtime-me-agent.timer <<'TIMER'
[Unit]
Description=Run Realtime Me local agent status reporter

[Timer]
OnBootSec=20s
OnUnitActiveSec=10s
AccuracySec=1s

[Install]
WantedBy=timers.target
TIMER
}

start_units() {
  systemctl daemon-reload
  systemctl enable --now realtime-me-status-device.timer >/dev/null
  systemctl start realtime-me-status-device.service
  if [[ $INSTALL_AGENT == 1 ]]; then
    systemctl enable --now realtime-me-agent.timer >/dev/null
    systemctl start realtime-me-agent.service
  fi
}

print_summary() {
  echo "Installed Realtime Me probe."
  echo "Gateway: $GATEWAY_URL"
  echo "Device:  $DEVICE_NAME ($DEVICE_ID)"
  echo "Agent:   $([[ $INSTALL_AGENT == 1 ]] && echo enabled || echo disabled)"
  echo "Check:   systemctl status realtime-me-status-device.service --no-pager"
}

main() {
  require_root
  require_command "$CURL_BIN"
  require_command "$PYTHON_BIN"
  require_command systemctl
  inherit_proxy_env
  GATEWAY_URL=$(read_gateway_url)
  local token
  token=$(read_token)
  download_reporters
  log "Writing configuration"
  write_env_file "$token"
  log "Installing systemd units"
  write_systemd_unit
  write_agent_unit
  log "Starting systemd units"
  start_units
  print_summary
}

main "$@"
