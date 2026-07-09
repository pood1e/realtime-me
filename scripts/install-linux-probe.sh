#!/usr/bin/env bash
set -euo pipefail

# Installs the realtime-me Prometheus exporters on Linux. The host is scraped
# (pull) and stays unaware of the gateway: this only installs node_exporter and
# status-device-reporter.py as systemd services bound to the LAN.
# Register the host centrally with scripts/operator/register-device.py; Prometheus stamps the
# device identity via service discovery, so the exporters need no token or
# gateway address.

INSTALL_DIR=${INSTALL_DIR:-/opt/realtime-me}
DEVICE_NAME=${STATUS_DEVICE_NAME:-$(hostname 2>/dev/null || echo linux)}
DEVICE_KIND=${STATUS_DEVICE_KIND:-host}
DEVICE_ROLE=${STATUS_DEVICE_ROLE:-desktop}
EXPORTER_BIND=${STATUS_EXPORTER_BIND:-0.0.0.0}
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

log() { echo "[realtime-me] $*" >&2; }

set_proxy_env_if_missing() {
  local name=$1 value=$2
  case "$name" in
    http_proxy|https_proxy|all_proxy|no_proxy|HTTP_PROXY|HTTPS_PROXY|ALL_PROXY|NO_PROXY) ;;
    *) return 0 ;;
  esac
  [[ -n $value ]] || return 0
  # Use an if-block, not `[[ ]] && export`: when the proxy var is already set the
  # test is false and a trailing `&&` would make this function return non-zero,
  # which `set -e` turns into a silent early exit.
  if [[ -z ${!name:-} ]]; then
    export "$name=$value"
  fi
}

import_proxy_env_from_process() {
  local environ_file=/proc/$1/environ entry name value
  [[ -r $environ_file ]] || return 0
  while IFS= read -r -d '' entry; do
    name=${entry%%=*}; value=${entry#*=}
    set_proxy_env_if_missing "$name" "$value"
  done <"$environ_file"
}

inherit_proxy_env() {
  [[ -d /proc ]] || return 0
  local pid=$PPID depth=0
  while [[ $pid =~ ^[0-9]+$ && $pid -gt 1 && $depth -lt 8 ]]; do
    import_proxy_env_from_process "$pid"
    pid=$(awk '/^PPid:/ { print $2 }' "/proc/$pid/status" 2>/dev/null || true)
    [[ -n $pid ]] || return 0
    depth=$((depth + 1))
  done
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1" >&2; exit 2; }
}

curl_download_args() {
  local timeout=$1
  local args=(-fsSL --connect-timeout 5 --max-time "$timeout" --speed-time 5 --speed-limit 1 --retry 0)
  [[ $CURL_FORCE_IPV4 == 1 ]] && args=(-4 "${args[@]}")
  printf '%s\0' "${args[@]}"
}

remove_unit_file() {
  systemctl disable --now "$1" >/dev/null 2>&1 || true
  rm -f "/etc/systemd/system/$1"
}

remove_legacy_units() {
  log "Removing legacy push services"
  remove_unit_file realtime-me-status-device.timer
  remove_unit_file realtime-me-status-device.service
  remove_unit_file realtime-me-agent.timer
  remove_unit_file realtime-me-agent.service
  [[ $INSTALL_AGENT != 1 ]] && remove_unit_file realtime-me-agent-exporter.service
  systemctl daemon-reload
  systemctl reset-failed >/dev/null 2>&1 || true
}

download_file() {
  local name=$1 destination=$2 temporary curl_error args
  temporary=$(mktemp); curl_error=$(mktemp)
  mapfile -d '' -t args < <(curl_download_args "$DOWNLOAD_TIMEOUT_SECONDS")
  log "Downloading $name"
  for base_url in "${RAW_BASE_URLS[@]}"; do
    if "$CURL_BIN" "${args[@]}" "${base_url%/}/$name" -o "$temporary" 2>"$curl_error"; then
      rm -f "$curl_error"; mv "$temporary" "$destination"; return
    fi
    log "Download mirror failed; trying next mirror"
  done
  rm -f "$temporary" "$curl_error"
  echo "Could not download $name from configured mirrors. If this host uses a proxy, preserve proxy env when running sudo." >&2
  exit 1
}

# The reporters and the status_common module they import must move as one set.
# Everything is downloaded into a staging directory first, so a mirror that fails
# on the second file cannot leave a new shared module beside an old reporter.
download_exporters() {
  local staging
  staging=$(mktemp -d)

  download_file probe/status_common.py "$staging/status_common.py"
  download_file probe/status-device-reporter.py "$staging/status-device-reporter.py"
  if [[ $INSTALL_AGENT == 1 ]]; then
    download_file probe/agent-status-reporter.py "$staging/agent-status-reporter.py"
  fi

  install -d -m 755 "$INSTALL_DIR"
  install -m 644 "$staging/status_common.py" "$INSTALL_DIR/status_common.py"
  install -m 755 "$staging/status-device-reporter.py" "$INSTALL_DIR/status-device-reporter.py"
  if [[ $INSTALL_AGENT == 1 ]]; then
    install -m 755 "$staging/agent-status-reporter.py" "$INSTALL_DIR/agent-status-reporter.py"
  fi
  rm -rf "$staging"
}

node_exporter_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    armv7l|armv7*) echo armv7 ;;
    armv6l|armv6*) echo armv6 ;;
    *) echo "Unsupported CPU architecture: $(uname -m)" >&2; exit 2 ;;
  esac
}

install_node_exporter() {
  local arch archive workdir source_url filename args
  # -wF, because the version is a literal and "version 1.11.1" is a prefix of
  # "version 1.11.10": a substring match would call the pin already satisfied.
  if [[ -x $NODE_EXPORTER_BIN ]] && "$NODE_EXPORTER_BIN" --version 2>&1 | grep -qwF "version $NODE_EXPORTER_VERSION"; then
    log "node_exporter v$NODE_EXPORTER_VERSION already installed"; return
  fi
  arch=$(node_exporter_arch)
  filename=node_exporter-${NODE_EXPORTER_VERSION}.linux-${arch}.tar.gz
  source_url=${NODE_EXPORTER_URL:-https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/$filename}
  archive=$(mktemp); workdir=$(mktemp -d)
  mapfile -d '' -t args < <(curl_download_args "$NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS")
  log "Downloading node_exporter v$NODE_EXPORTER_VERSION"
  if ! "$CURL_BIN" "${args[@]}" "$source_url" -o "$archive"; then
    rm -rf "$archive" "$workdir"; echo "Could not download node_exporter." >&2; exit 1
  fi
  verify_node_exporter_archive "$archive" "$filename" "$workdir"
  tar -xzf "$archive" -C "$workdir"
  install -m 755 "$workdir"/node_exporter-*.linux-*/node_exporter "$NODE_EXPORTER_BIN"
  rm -rf "$archive" "$workdir"
}

verify_node_exporter_archive() {
  local archive=$1 filename=$2 workdir=$3 expected actual
  expected=$(node_exporter_expected_sha256 "$filename")
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
  local filename=$1 sums checksum_url args
  if [[ -n $NODE_EXPORTER_SHA256 ]]; then printf '%s' "$NODE_EXPORTER_SHA256"; return; fi
  checksum_url=${NODE_EXPORTER_SHA256SUMS_URL:-https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/sha256sums.txt}
  sums=$(mktemp)
  mapfile -d '' -t args < <(curl_download_args "$DOWNLOAD_TIMEOUT_SECONDS")
  if ! "$CURL_BIN" "${args[@]}" "$checksum_url" -o "$sums" 2>/dev/null; then
    rm -f "$sums"; return
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
ExecStart=$PYTHON_BIN $INSTALL_DIR/status-device-reporter.py --bind $EXPORTER_BIND --port $DEVICE_EXPORTER_PORT
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
ExecStart=$PYTHON_BIN $INSTALL_DIR/agent-status-reporter.py --bind $EXPORTER_BIND --port $AGENT_EXPORTER_PORT
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
  if [[ -n ${EXPORTER_HOST:-} ]]; then printf '%s' "$EXPORTER_HOST"; return; fi
  local route_host
  route_host=$(ip route get 1.1.1.1 2>/dev/null | awk '/src/ { for (i = 1; i <= NF; i++) if ($i == "src") { print $(i + 1); exit } }' || true)
  if [[ -n $route_host ]]; then printf '%s' "$route_host"; return; fi
  hostname -I 2>/dev/null | awk '{ print $1; exit }'
}

detect_exporter_host() {
  EXPORTER_HOST=$(auto_exporter_host)
  [[ -n $EXPORTER_HOST ]] || { echo "Could not detect exporter host. Set STATUS_EXPORTER_HOST." >&2; exit 2; }
}

print_summary() {
  echo "Installed Realtime Me Linux exporters."
  echo "Device:          $DEVICE_NAME"
  echo "Node exporter:   $EXPORTER_HOST:$NODE_EXPORTER_PORT"
  echo "Device exporter: $EXPORTER_HOST:$DEVICE_EXPORTER_PORT"
  echo "Agent:           $([[ $INSTALL_AGENT == 1 ]] && printf '%s:%s' "$EXPORTER_HOST" "$AGENT_EXPORTER_PORT" || echo disabled)"
  echo "Check:           systemctl status realtime-me-node-exporter.service --no-pager"
  echo
  echo "Now register this host centrally (where you can reach the gateway):"
  echo "  STATUS_INGEST_TOKEN=... python3 scripts/operator/register-device.py --url <GATEWAY_URL> \\"
  echo "    --host $EXPORTER_HOST --name \"$DEVICE_NAME\" --kind $DEVICE_KIND --role $DEVICE_ROLE$([[ $INSTALL_AGENT == 1 ]] && printf ' --install-agent')"
}

main() {
  require_root
  require_command "$CURL_BIN"
  require_command "$PYTHON_BIN"
  require_command systemctl
  require_command tar
  require_command sha256sum
  inherit_proxy_env
  detect_exporter_host
  remove_legacy_units
  download_exporters
  install_node_exporter
  log "Installing systemd units"
  write_systemd_units
  log "Starting exporters"
  start_units
  print_summary
}

main "$@"
