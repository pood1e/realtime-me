#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR=${INSTALL_DIR:-/opt/realtime-me}
ENV_FILE=${ENV_FILE:-/etc/realtime-me.env}
GATEWAY_URL=${STATUS_GATEWAY_URL:-http://192.168.0.126:18080}
DEVICE_ID=${STATUS_DEVICE_ID:-$(hostname -s 2>/dev/null || hostname)}
DEVICE_NAME=${STATUS_DEVICE_NAME:-$(hostname 2>/dev/null || echo linux)}
DEVICE_KIND=${STATUS_DEVICE_KIND:-host}
DEVICE_ROLE=${STATUS_DEVICE_ROLE:-desktop}
INSTALL_AGENT=${INSTALL_AGENT:-0}
RAW_BASE_URL=${REALTIME_ME_RAW_BASE_URL:-https://raw.githubusercontent.com/pood1e/realtime-me/main/scripts}
PYTHON_BIN=${PYTHON_BIN:-/usr/bin/python3}
CURL_BIN=${CURL_BIN:-/usr/bin/curl}

require_root() {
  if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
    echo "Run with sudo or as root." >&2
    exit 2
  fi
}

require_command() {
  local command=$1
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Missing required command: $command" >&2
    exit 2
  fi
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
  chmod 600 "$ENV_FILE"
}

download_reporters() {
  install -d -m 755 "$INSTALL_DIR"
  "$CURL_BIN" -fsSL "$RAW_BASE_URL/status-device-reporter.py" -o "$INSTALL_DIR/status-device-reporter.py"
  chmod 755 "$INSTALL_DIR/status-device-reporter.py"
  if [[ $INSTALL_AGENT == 1 ]]; then
    "$CURL_BIN" -fsSL "$RAW_BASE_URL/agent-status-reporter.py" -o "$INSTALL_DIR/agent-status-reporter.py"
    chmod 755 "$INSTALL_DIR/agent-status-reporter.py"
  fi
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
  local token
  token=$(read_token)
  download_reporters
  write_env_file "$token"
  write_systemd_unit
  write_agent_unit
  start_units
  print_summary
}

main "$@"
