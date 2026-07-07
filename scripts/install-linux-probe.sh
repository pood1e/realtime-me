#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR=${INSTALL_DIR:-/opt/realtime-me}
ENV_FILE=${ENV_FILE:-/etc/realtime-me.env}
GATEWAY_URL=${STATUS_GATEWAY_URL:-}
REGISTER_TARGETS=${STATUS_REGISTER_TARGETS:-0}
DEVICE_NAME=${STATUS_DEVICE_NAME:-$(hostname 2>/dev/null || echo linux)}
DEVICE_MODEL=${STATUS_DEVICE_MODEL:-}
DEVICE_KIND=${STATUS_DEVICE_KIND:-host}
DEVICE_ROLE=${STATUS_DEVICE_ROLE:-desktop}
IDENTITY_FILE=${STATUS_IDENTITY_FILE:-$INSTALL_DIR/identity.json}
EXPORTER_BIND=${STATUS_EXPORTER_BIND:-127.0.0.1}
EXPORTER_HOST=${STATUS_EXPORTER_HOST:-}
NODE_EXPORTER_PORT=${STATUS_NODE_EXPORTER_PORT:-9100}
DEVICE_EXPORTER_PORT=${STATUS_DEVICE_EXPORTER_PORT:-18083}
AGENT_EXPORTER_PORT=${STATUS_AGENT_EXPORTER_PORT:-18082}
INSTALL_AGENT=${INSTALL_AGENT:-0}
PROBE_USER=${STATUS_PROBE_USER:-${SUDO_USER:-}}
NODE_EXPORTER_VERSION=${NODE_EXPORTER_VERSION:-1.11.1}
NODE_EXPORTER_VERSION=${NODE_EXPORTER_VERSION#v}
NODE_EXPORTER_BIN=${NODE_EXPORTER_BIN:-/usr/local/bin/node_exporter}
NODE_EXPORTER_SHA256=${NODE_EXPORTER_SHA256:-}
DOWNLOAD_TIMEOUT_SECONDS=${STATUS_DOWNLOAD_TIMEOUT_SECONDS:-15}
NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS=${STATUS_NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS:-90}
CURL_FORCE_IPV4=${STATUS_CURL_FORCE_IPV4:-1}
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
    *) return 0 ;;
  esac
  [[ -n $value ]] || return 0
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
  [[ -r $environ_file ]] || return 0
  while IFS= read -r -d '' entry; do
    name=${entry%%=*}
    value=${entry#*=}
    set_proxy_env_if_missing "$name" "$value"
  done <"$environ_file"
}

inherit_proxy_env() {
  [[ -d /proc ]] || return 0
  local pid=$PPID
  local depth=0
  while [[ $pid =~ ^[0-9]+$ && $pid -gt 1 && $depth -lt 8 ]]; do
    import_proxy_env_from_process "$pid"
    pid=$(awk '/^PPid:/ { print $2 }' "/proc/$pid/status" 2>/dev/null || true)
    [[ -n $pid ]] || return 0
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

curl_download_args() {
  local timeout=$1
  local args=(-fsSL --connect-timeout 5 --max-time "$timeout" --speed-time 5 --speed-limit 1 --retry 0)
  if [[ $CURL_FORCE_IPV4 == 1 ]]; then
    args=(-4 "${args[@]}")
  fi
  printf '%s\0' "${args[@]}"
}

remove_unit_file() {
  local unit=$1
  systemctl disable --now "$unit" >/dev/null 2>&1 || true
  rm -f "/etc/systemd/system/$unit"
}

remove_legacy_units() {
  log "Removing legacy push services"
  remove_unit_file realtime-me-status-device.timer
  remove_unit_file realtime-me-status-device.service
  remove_unit_file realtime-me-agent.timer
  remove_unit_file realtime-me-agent.service
  if [[ $INSTALL_AGENT != 1 ]]; then
    remove_unit_file realtime-me-agent-exporter.service
  fi
  systemctl daemon-reload
  systemctl reset-failed >/dev/null 2>&1 || true
}

configure_gateway_url() {
  if [[ -z ${GATEWAY_URL:-} ]]; then
    return 0
  fi
  normalize_url "$GATEWAY_URL"
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
  umask 077
  cat >"$ENV_FILE" <<ENV
STATUS_DEVICE_NAME=$DEVICE_NAME
STATUS_DEVICE_MODEL=$DEVICE_MODEL
STATUS_DEVICE_KIND=$DEVICE_KIND
STATUS_DEVICE_ROLE=$DEVICE_ROLE
STATUS_IDENTITY_FILE=$IDENTITY_FILE
STATUS_DEVICE_EXPORTER_BIND=$EXPORTER_BIND
STATUS_DEVICE_EXPORTER_PORT=$DEVICE_EXPORTER_PORT
STATUS_AGENT_EXPORTER_BIND=$EXPORTER_BIND
STATUS_AGENT_EXPORTER_PORT=$AGENT_EXPORTER_PORT
ENV
  chmod 644 "$ENV_FILE"
}

download_file() {
  local name=$1
  local destination=$2
  local temporary
  local curl_error
  local args
  temporary=$(mktemp)
  curl_error=$(mktemp)
  mapfile -d '' -t args < <(curl_download_args "$DOWNLOAD_TIMEOUT_SECONDS")
  log "Downloading $name"
  for base_url in "${RAW_BASE_URLS[@]}"; do
    log "Trying ${base_url%/}/$name"
    if "$CURL_BIN" "${args[@]}" "${base_url%/}/$name" -o "$temporary" 2>"$curl_error"; then
      rm -f "$curl_error"
      mv "$temporary" "$destination"
      return
    fi
    log "Download mirror failed; trying next mirror"
  done
  rm -f "$temporary" "$curl_error"
  echo "Could not download $name from configured mirrors. If this host uses a proxy, preserve proxy env when running sudo." >&2
  exit 1
}

download_exporters() {
  install -d -m 755 "$INSTALL_DIR"
  download_file status_common.py "$INSTALL_DIR/status_common.py"
  chmod 644 "$INSTALL_DIR/status_common.py"
  download_file status-device-reporter.py "$INSTALL_DIR/status-device-reporter.py"
  chmod 755 "$INSTALL_DIR/status-device-reporter.py"
  [[ $INSTALL_AGENT == 1 ]] || return 0
  download_file agent-status-reporter.py "$INSTALL_DIR/agent-status-reporter.py"
  chmod 755 "$INSTALL_DIR/agent-status-reporter.py"
}

node_exporter_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    armv7l|armv7*) echo armv7 ;;
    armv6l|armv6*) echo armv6 ;;
    *)
      echo "Unsupported CPU architecture: $(uname -m)" >&2
      exit 2
      ;;
  esac
}

install_node_exporter() {
  local arch
  local archive
  local workdir
  local source_url
  local filename
  local args
  if [[ -x $NODE_EXPORTER_BIN ]] && "$NODE_EXPORTER_BIN" --version 2>&1 | grep -q "version $NODE_EXPORTER_VERSION"; then
    log "node_exporter v$NODE_EXPORTER_VERSION already installed"
    return
  fi
  arch=$(node_exporter_arch)
  filename=node_exporter-${NODE_EXPORTER_VERSION}.linux-${arch}.tar.gz
  source_url=${NODE_EXPORTER_URL:-https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/$filename}
  archive=$(mktemp)
  workdir=$(mktemp -d)
  mapfile -d '' -t args < <(curl_download_args "$NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS")
  log "Downloading node_exporter v$NODE_EXPORTER_VERSION"
  if ! "$CURL_BIN" "${args[@]}" "$source_url" -o "$archive"; then
    rm -rf "$archive" "$workdir"
    echo "Could not download node_exporter. If this host uses a proxy, preserve proxy env when running sudo." >&2
    exit 1
  fi
  verify_node_exporter_archive "$archive" "$filename" "$arch" "$workdir"
  tar -xzf "$archive" -C "$workdir"
  install -m 755 "$workdir"/node_exporter-*.linux-*/node_exporter "$NODE_EXPORTER_BIN"
  rm -rf "$archive" "$workdir"
}

verify_node_exporter_archive() {
  local archive=$1
  local filename=$2
  local arch=$3
  local workdir=$4
  local expected
  local actual
  expected=$(node_exporter_expected_sha256 "$filename" "$arch")
  if [[ -z $expected ]]; then
    rm -rf "$archive" "$workdir"
    echo "Could not determine node_exporter SHA-256 checksum. Set NODE_EXPORTER_SHA256 to verify manually." >&2
    exit 1
  fi
  actual=$(sha256sum "$archive" | awk '{ print $1 }')
  if [[ $actual != "$expected" ]]; then
    rm -rf "$archive" "$workdir"
    echo "node_exporter checksum mismatch (expected $expected, got $actual)." >&2
    exit 1
  fi
  log "Verified node_exporter SHA-256 checksum"
}

node_exporter_expected_sha256() {
  local filename=$1
  local arch=$2
  local sums
  local checksum_url
  local args
  if [[ -n $NODE_EXPORTER_SHA256 ]]; then
    printf '%s' "$NODE_EXPORTER_SHA256"
    return
  fi
  checksum_url=${NODE_EXPORTER_SHA256SUMS_URL:-https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/sha256sums.txt}
  sums=$(mktemp)
  mapfile -d '' -t args < <(curl_download_args "$DOWNLOAD_TIMEOUT_SECONDS")
  if ! "$CURL_BIN" "${args[@]}" "$checksum_url" -o "$sums" 2>/dev/null; then
    rm -f "$sums"
    return
  fi
  awk -v f="$filename" '$2 == f { print $1 }' "$sums"
  rm -f "$sums"
}

probe_service_directives() {
  [[ -n ${PROBE_USER:-} && $PROBE_USER != root ]] || return 0
  local uid
  uid=$(id -u "$PROBE_USER" 2>/dev/null || true)
  [[ -n $uid ]] || return 0
  printf 'User=%s\n' "$PROBE_USER"
  printf 'Environment=XDG_RUNTIME_DIR=/run/user/%s\n' "$uid"
  printf 'Environment=DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%s/bus\n' "$uid"
}

write_systemd_units() {
  local user_directives
  user_directives=$(probe_service_directives)
  cat >/etc/systemd/system/realtime-me-node-exporter.service <<SERVICE
[Unit]
Description=Realtime Me node_exporter
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecStart=$NODE_EXPORTER_BIN --web.listen-address=$EXPORTER_BIND:$NODE_EXPORTER_PORT
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
SERVICE

  cat >/etc/systemd/system/realtime-me-device-exporter.service <<SERVICE
[Unit]
Description=Realtime Me device exporter
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
$user_directives
EnvironmentFile=$ENV_FILE
ExecStart=$PYTHON_BIN $INSTALL_DIR/status-device-reporter.py --serve --bind $EXPORTER_BIND --port $DEVICE_EXPORTER_PORT
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
SERVICE

  [[ $INSTALL_AGENT == 1 ]] || return 0
  cat >/etc/systemd/system/realtime-me-agent-exporter.service <<SERVICE
[Unit]
Description=Realtime Me agent exporter
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
$user_directives
EnvironmentFile=$ENV_FILE
ExecStart=$PYTHON_BIN $INSTALL_DIR/agent-status-reporter.py --serve --bind $EXPORTER_BIND --port $AGENT_EXPORTER_PORT
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
SERVICE
}

start_units() {
  systemctl daemon-reload
  systemctl enable realtime-me-node-exporter.service >/dev/null
  systemctl restart realtime-me-node-exporter.service
  systemctl enable realtime-me-device-exporter.service >/dev/null
  systemctl restart realtime-me-device-exporter.service
  if [[ $INSTALL_AGENT == 1 ]]; then
    systemctl enable realtime-me-agent-exporter.service >/dev/null
    systemctl restart realtime-me-agent-exporter.service
  fi
}

auto_exporter_host() {
  if [[ -n ${EXPORTER_HOST:-} ]]; then
    printf '%s' "$EXPORTER_HOST"
    return
  fi
  local route_host
  route_host=$(ip route get 1.1.1.1 2>/dev/null | awk '/src/ { for (i = 1; i <= NF; i++) if ($i == "src") { print $(i + 1); exit } }' || true)
  if [[ -n $route_host ]]; then
    printf '%s' "$route_host"
    return
  fi
  hostname -I 2>/dev/null | awk '{ print $1; exit }'
}

json_payload() {
  local host=$1
  local install_agent=$2
  local device_uid=$3
  "$PYTHON_BIN" - "$host" "$install_agent" "$device_uid" "$INSTALL_DIR" <<'PY'
import json
import os
import sys

host, install_agent_arg, device_uid, install_dir = sys.argv[1:5]
sys.path.insert(0, install_dir)
from status_common import device_kind_enum, device_role_enum

install_agent = install_agent_arg == "1"
node_port = os.environ["STATUS_NODE_EXPORTER_PORT"]
device_port = os.environ["STATUS_DEVICE_EXPORTER_PORT"]
agent_port = os.environ["STATUS_AGENT_EXPORTER_PORT"]
device_kind = os.environ["STATUS_DEVICE_KIND"]
device_role = os.environ["STATUS_DEVICE_ROLE"]
common = {
    "deviceUid": device_uid,
    "displayName": os.environ["STATUS_DEVICE_NAME"],
    "model": os.environ.get("STATUS_DEVICE_MODEL", ""),
    "kind": device_kind_enum(device_kind),
    "role": device_role_enum(device_role),
}
node_job = "SCRAPE_JOB_VM_NODE_EXPORTER" if device_kind == "virtual_machine" or device_role == "vm" else "SCRAPE_JOB_NODE_EXPORTER"
targets = [
    dict(common, job=node_job, target=f"{host}:{node_port}"),
    dict(common, job="SCRAPE_JOB_DEVICE_EXPORTER", target=f"{host}:{device_port}"),
]
if install_agent:
    targets.append(dict(common, job="SCRAPE_JOB_AGENT_EXPORTER", target=f"{host}:{agent_port}"))
print(json.dumps({"targets": targets}, separators=(",", ":")))
PY
}

post_connect() {
  local endpoint=$1
  local token=$2
  local payload=$3
  local config
  config=$(mktemp)
  trap 'rm -f "$config"' EXIT
  chmod 600 "$config"
  {
    printf 'header = "Accept: application/json"\n'
    printf 'header = "Authorization: Bearer %s"\n' "$token"
    printf 'header = "Connect-Protocol-Version: 1"\n'
    printf 'header = "Content-Type: application/json"\n'
  } >"$config"
  if ! "$CURL_BIN" -fsSL --connect-timeout 10 --max-time 30 --config "$config" --data-binary "$payload" "$endpoint" >/dev/null; then
    echo "Could not register Prometheus scrape targets." >&2
    exit 1
  fi
}

detect_exporter_host() {
  EXPORTER_HOST=$(auto_exporter_host)
  if [[ -z $EXPORTER_HOST ]]; then
    echo "Could not detect exporter host. Set STATUS_EXPORTER_HOST." >&2
    exit 2
  fi
}

should_register_targets() {
  [[ $REGISTER_TARGETS == 1 || -n ${GATEWAY_URL:-} ]]
}

chown_identity_file() {
  [[ -n ${PROBE_USER:-} && $PROBE_USER != root ]] || return 0
  chown "$PROBE_USER" "$IDENTITY_FILE" 2>/dev/null || true
}

enroll_device() {
  local token=$1
  local device_uid
  device_uid=$(
    STATUS_GATEWAY_URL=$GATEWAY_URL \
    STATUS_INGEST_TOKEN=$token \
    STATUS_IDENTITY_FILE=$IDENTITY_FILE \
    STATUS_DEVICE_NAME=$DEVICE_NAME \
    STATUS_DEVICE_MODEL=$DEVICE_MODEL \
    STATUS_DEVICE_KIND=$DEVICE_KIND \
    STATUS_DEVICE_ROLE=$DEVICE_ROLE \
    "$PYTHON_BIN" "$INSTALL_DIR/status-device-reporter.py" --print-uid
  ) || true
  if [[ -z $device_uid ]]; then
    echo "Device enrollment failed. Check STATUS_GATEWAY_URL and STATUS_INGEST_TOKEN." >&2
    exit 1
  fi
  chown_identity_file
  printf '%s' "$device_uid"
}

register_prometheus_targets() {
  if ! should_register_targets; then
    log "Skipping device enrollment and Prometheus target registration"
    return
  fi
  if [[ -z ${GATEWAY_URL:-} ]]; then
    echo "Set STATUS_GATEWAY_URL when STATUS_REGISTER_TARGETS=1." >&2
    exit 2
  fi
  local token
  local device_uid
  local payload
  token=$(read_token)
  log "Enrolling device"
  device_uid=$(enroll_device "$token")
  export STATUS_NODE_EXPORTER_PORT=$NODE_EXPORTER_PORT
  export STATUS_DEVICE_EXPORTER_PORT=$DEVICE_EXPORTER_PORT
  export STATUS_AGENT_EXPORTER_PORT=$AGENT_EXPORTER_PORT
  export STATUS_DEVICE_NAME=$DEVICE_NAME
  export STATUS_DEVICE_MODEL=$DEVICE_MODEL
  export STATUS_DEVICE_KIND=$DEVICE_KIND
  export STATUS_DEVICE_ROLE=$DEVICE_ROLE
  payload=$(json_payload "$EXPORTER_HOST" "$INSTALL_AGENT" "$device_uid")
  log "Registering Prometheus scrape targets"
  post_connect "$GATEWAY_URL/realtime.me.v1.IngestService/RegisterScrapeTargets" "$token" "$payload"
}

print_summary() {
  echo "Installed Realtime Me Prometheus probe."
  if [[ -n ${GATEWAY_URL:-} ]]; then
    echo "Gateway:         $GATEWAY_URL"
  fi
  echo "Device:          $DEVICE_NAME"
  echo "Node exporter:   $EXPORTER_HOST:$NODE_EXPORTER_PORT"
  echo "Device exporter: $EXPORTER_HOST:$DEVICE_EXPORTER_PORT"
  echo "Agent:    $([[ $INSTALL_AGENT == 1 ]] && printf '%s:%s' "$EXPORTER_HOST" "$AGENT_EXPORTER_PORT" || echo disabled)"
  echo "Check:    systemctl status realtime-me-node-exporter.service --no-pager"
}

main() {
  require_root
  require_command "$CURL_BIN"
  require_command "$PYTHON_BIN"
  require_command systemctl
  require_command tar
  require_command sha256sum
  inherit_proxy_env
  GATEWAY_URL=$(configure_gateway_url)
  detect_exporter_host
  remove_legacy_units
  log "Writing configuration"
  write_env_file
  download_exporters
  install_node_exporter
  log "Installing systemd units"
  write_systemd_units
  register_prometheus_targets
  log "Starting exporters"
  start_units
  print_summary
}

main "$@"
