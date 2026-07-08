#!/usr/bin/env bash
set -euo pipefail

# Installs the realtime-me Prometheus probe on macOS. Like the Linux probe, the
# host is scraped (pull), not pushed: node_exporter (darwin build) publishes host
# metrics and status-device-reporter.py --serve publishes media/Bluetooth. Both
# run as LaunchAgents under the logged-in user (so media/Bluetooth are visible)
# and are registered as gateway scrape targets. Only the phone uses push.

INSTALL_DIR=${INSTALL_DIR:-$HOME/.realtime-me}
GATEWAY_URL=${STATUS_GATEWAY_URL:-}
REGISTER_TARGETS=${STATUS_REGISTER_TARGETS:-0}
DEVICE_NAME=${STATUS_DEVICE_NAME:-$(scutil --get ComputerName 2>/dev/null || hostname -s 2>/dev/null || echo mac)}
DEVICE_MODEL=${STATUS_DEVICE_MODEL:-$(sysctl -n hw.model 2>/dev/null || echo "")}
DEVICE_KIND=${STATUS_DEVICE_KIND:-host}
DEVICE_ROLE=${STATUS_DEVICE_ROLE:-desktop}
IDENTITY_FILE=${STATUS_IDENTITY_FILE:-$INSTALL_DIR/identity.json}
EXPORTER_BIND=${STATUS_EXPORTER_BIND:-0.0.0.0}
EXPORTER_HOST=${STATUS_EXPORTER_HOST:-}
NODE_EXPORTER_PORT=${STATUS_NODE_EXPORTER_PORT:-9100}
DEVICE_EXPORTER_PORT=${STATUS_DEVICE_EXPORTER_PORT:-18083}
AGENT_EXPORTER_PORT=${STATUS_AGENT_EXPORTER_PORT:-18082}
INSTALL_AGENT=${INSTALL_AGENT:-0}
NODE_EXPORTER_VERSION=${NODE_EXPORTER_VERSION:-1.11.1}
NODE_EXPORTER_VERSION=${NODE_EXPORTER_VERSION#v}
NODE_EXPORTER_BIN=$INSTALL_DIR/node_exporter
NODE_EXPORTER_SHA256=${NODE_EXPORTER_SHA256:-}
DOWNLOAD_TIMEOUT_SECONDS=${STATUS_DOWNLOAD_TIMEOUT_SECONDS:-15}
NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS=${STATUS_NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS:-90}
LABEL_PREFIX=space.pood1e.realtime-me
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
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
PYTHON_BIN=${PYTHON_BIN:-$(command -v python3 2>/dev/null || echo /usr/bin/python3)}
CURL_BIN=${CURL_BIN:-/usr/bin/curl}

log() { echo "[realtime-me] $*" >&2; }

require_not_root() {
  if [[ ${EUID:-$(id -u)} -eq 0 ]]; then
    echo "Run as your normal login user (not sudo): LaunchAgents and media/Bluetooth access are per-user." >&2
    exit 2
  fi
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1" >&2; exit 2; }
}

xml_escape() {
  local value=$1
  value=${value//&/&amp;}
  value=${value//</&lt;}
  value=${value//>/&gt;}
  printf '%s' "$value"
}

normalize_url() {
  local value=${1%/}
  case "$value" in
    http://*|https://*) printf '%s' "$value" ;;
    *) echo "STATUS_GATEWAY_URL must start with http:// or https://." >&2; exit 2 ;;
  esac
}

configure_gateway_url() {
  [[ -z ${GATEWAY_URL:-} ]] && return 0
  normalize_url "$GATEWAY_URL"
}

read_token() {
  if [[ -n ${STATUS_INGEST_TOKEN:-} ]]; then printf '%s' "$STATUS_INGEST_TOKEN"; return; fi
  if [[ ! -r /dev/tty ]]; then echo "Set STATUS_INGEST_TOKEN or run from an interactive terminal." >&2; exit 2; fi
  local token
  read -rsp "STATUS_INGEST_TOKEN: " token </dev/tty; echo >/dev/tty
  [[ -n $token ]] || { echo "STATUS_INGEST_TOKEN is required." >&2; exit 2; }
  printf '%s' "$token"
}

curl_download() {
  local url=$1 dest=$2 timeout=$3
  "$CURL_BIN" -fsSL -4 --connect-timeout 5 --max-time "$timeout" --retry 0 "$url" -o "$dest"
}

download_file() {
  local name=$1 dest=$2 tmp
  tmp=$(mktemp)
  log "Downloading $name"
  for base in "${RAW_BASE_URLS[@]}"; do
    if curl_download "${base%/}/$name" "$tmp" "$DOWNLOAD_TIMEOUT_SECONDS" 2>/dev/null; then
      mv "$tmp" "$dest"; return
    fi
    log "Download mirror failed; trying next mirror"
  done
  rm -f "$tmp"; echo "Could not download $name from configured mirrors." >&2; exit 1
}

download_exporters() {
  mkdir -p "$INSTALL_DIR"
  download_file status_common.py "$INSTALL_DIR/status_common.py"
  download_file status-device-reporter.py "$INSTALL_DIR/status-device-reporter.py"
  chmod 755 "$INSTALL_DIR/status-device-reporter.py"
  [[ $INSTALL_AGENT == 1 ]] || return 0
  download_file agent-status-reporter.py "$INSTALL_DIR/agent-status-reporter.py"
  chmod 755 "$INSTALL_DIR/agent-status-reporter.py"
}

node_exporter_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    arm64|aarch64) echo arm64 ;;
    *) echo "Unsupported CPU architecture: $(uname -m)" >&2; exit 2 ;;
  esac
}

node_exporter_expected_sha256() {
  local filename=$1 sums
  if [[ -n $NODE_EXPORTER_SHA256 ]]; then printf '%s' "$NODE_EXPORTER_SHA256"; return; fi
  sums=$(mktemp)
  if ! curl_download "https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/sha256sums.txt" "$sums" "$DOWNLOAD_TIMEOUT_SECONDS" 2>/dev/null; then
    rm -f "$sums"; return
  fi
  awk -v f="$filename" '$2 == f { print $1 }' "$sums"
  rm -f "$sums"
}

install_node_exporter() {
  if [[ -x $NODE_EXPORTER_BIN ]] && "$NODE_EXPORTER_BIN" --version 2>&1 | grep -q "version $NODE_EXPORTER_VERSION"; then
    log "node_exporter v$NODE_EXPORTER_VERSION already installed"; return
  fi
  local arch filename url archive workdir expected actual
  arch=$(node_exporter_arch)
  filename=node_exporter-${NODE_EXPORTER_VERSION}.darwin-${arch}.tar.gz
  url=${NODE_EXPORTER_URL:-https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXPORTER_VERSION}/$filename}
  archive=$(mktemp); workdir=$(mktemp -d)
  log "Downloading node_exporter v$NODE_EXPORTER_VERSION ($arch)"
  if ! curl_download "$url" "$archive" "$NODE_EXPORTER_DOWNLOAD_TIMEOUT_SECONDS"; then
    rm -rf "$archive" "$workdir"; echo "Could not download node_exporter." >&2; exit 1
  fi
  expected=$(node_exporter_expected_sha256 "$filename")
  [[ -n $expected ]] || { rm -rf "$archive" "$workdir"; echo "Could not determine node_exporter checksum; set NODE_EXPORTER_SHA256." >&2; exit 1; }
  actual=$(shasum -a 256 "$archive" | awk '{print $1}')
  [[ $actual == "$expected" ]] || { rm -rf "$archive" "$workdir"; echo "node_exporter checksum mismatch (expected $expected, got $actual)." >&2; exit 1; }
  log "Verified node_exporter SHA-256 checksum"
  tar -xzf "$archive" -C "$workdir"
  install -m 755 "$workdir"/node_exporter-*.darwin-*/node_exporter "$NODE_EXPORTER_BIN"
  xattr -dr com.apple.quarantine "$NODE_EXPORTER_BIN" 2>/dev/null || true
  rm -rf "$archive" "$workdir"
}

auto_exporter_host() {
  if [[ -n ${EXPORTER_HOST:-} ]]; then printf '%s' "$EXPORTER_HOST"; return; fi
  local iface ip=""
  iface=$(route -n get default 2>/dev/null | awk '/interface:/{print $2; exit}')
  [[ -n $iface ]] && ip=$(ipconfig getifaddr "$iface" 2>/dev/null || true)
  [[ -z $ip ]] && ip=$(ipconfig getifaddr en0 2>/dev/null || true)
  printf '%s' "$ip"
}

detect_exporter_host() {
  EXPORTER_HOST=$(auto_exporter_host)
  [[ -n $EXPORTER_HOST ]] || { echo "Could not detect LAN IP. Set STATUS_EXPORTER_HOST." >&2; exit 2; }
}

should_register_targets() { [[ $REGISTER_TARGETS == 1 || -n ${GATEWAY_URL:-} ]]; }

enroll_device() {
  local token=$1 device_uid
  device_uid=$(
    STATUS_GATEWAY_URL=$GATEWAY_URL STATUS_INGEST_TOKEN=$token STATUS_IDENTITY_FILE=$IDENTITY_FILE \
    STATUS_DEVICE_NAME=$DEVICE_NAME STATUS_DEVICE_MODEL=$DEVICE_MODEL \
    STATUS_DEVICE_KIND=$DEVICE_KIND STATUS_DEVICE_ROLE=$DEVICE_ROLE \
    "$PYTHON_BIN" "$INSTALL_DIR/status-device-reporter.py" --print-uid
  ) || true
  [[ -n $device_uid ]] || { echo "Device enrollment failed. Check STATUS_GATEWAY_URL and STATUS_INGEST_TOKEN." >&2; exit 1; }
  printf '%s' "$device_uid"
}

json_payload() {
  local host=$1 install_agent=$2 device_uid=$3
  "$PYTHON_BIN" - "$host" "$install_agent" "$device_uid" "$INSTALL_DIR" <<'PY'
import json, os, sys

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
  local endpoint=$1 token=$2 payload=$3 config
  config=$(mktemp); chmod 600 "$config"
  {
    printf 'header = "Accept: application/json"\n'
    printf 'header = "Authorization: Bearer %s"\n' "$token"
    printf 'header = "Connect-Protocol-Version: 1"\n'
    printf 'header = "Content-Type: application/json"\n'
  } >"$config"
  if ! "$CURL_BIN" -fsSL --connect-timeout 10 --max-time 30 --config "$config" --data-binary "$payload" "$endpoint" >/dev/null; then
    rm -f "$config"; echo "Could not register Prometheus scrape targets." >&2; exit 1
  fi
  rm -f "$config"
}

register_prometheus_targets() {
  if ! should_register_targets; then
    log "Skipping device enrollment and Prometheus target registration"
    return
  fi
  [[ -n ${GATEWAY_URL:-} ]] || { echo "Set STATUS_GATEWAY_URL when STATUS_REGISTER_TARGETS=1." >&2; exit 2; }
  local token device_uid payload
  token=$(read_token)
  log "Enrolling device"
  device_uid=$(enroll_device "$token")
  export STATUS_NODE_EXPORTER_PORT=$NODE_EXPORTER_PORT
  export STATUS_DEVICE_EXPORTER_PORT=$DEVICE_EXPORTER_PORT
  export STATUS_AGENT_EXPORTER_PORT=$AGENT_EXPORTER_PORT
  export STATUS_DEVICE_NAME=$DEVICE_NAME STATUS_DEVICE_MODEL=$DEVICE_MODEL
  export STATUS_DEVICE_KIND=$DEVICE_KIND STATUS_DEVICE_ROLE=$DEVICE_ROLE
  payload=$(json_payload "$EXPORTER_HOST" "$INSTALL_AGENT" "$device_uid")
  log "Registering Prometheus scrape targets"
  post_connect "$GATEWAY_URL/realtime.me.v1.IngestService/RegisterScrapeTargets" "$token" "$payload"
}

plist_path() { printf '%s/%s.%s.plist' "$LAUNCH_AGENTS_DIR" "$LABEL_PREFIX" "$1"; }

write_node_exporter_agent() {
  cat >"$(plist_path node-exporter)" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>$LABEL_PREFIX.node-exporter</string>
    <key>ProgramArguments</key>
    <array>
        <string>$NODE_EXPORTER_BIN</string>
        <string>--web.listen-address=$EXPORTER_BIND:$NODE_EXPORTER_PORT</string>
    </array>
    <key>KeepAlive</key><true/>
    <key>RunAtLoad</key><true/>
    <key>StandardErrorPath</key><string>$INSTALL_DIR/node-exporter.log</string>
    <key>StandardOutPath</key><string>$INSTALL_DIR/node-exporter.log</string>
</dict>
</plist>
PLIST
}

write_device_exporter_agent() {
  cat >"$(plist_path device-exporter)" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>$LABEL_PREFIX.device-exporter</string>
    <key>ProgramArguments</key>
    <array>
        <string>$PYTHON_BIN</string>
        <string>$INSTALL_DIR/status-device-reporter.py</string>
        <string>--serve</string>
        <string>--bind</string><string>$EXPORTER_BIND</string>
        <string>--port</string><string>$DEVICE_EXPORTER_PORT</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>STATUS_DEVICE_NAME</key><string>$(xml_escape "$DEVICE_NAME")</string>
        <key>STATUS_DEVICE_MODEL</key><string>$(xml_escape "$DEVICE_MODEL")</string>
        <key>STATUS_DEVICE_KIND</key><string>$DEVICE_KIND</string>
        <key>STATUS_DEVICE_ROLE</key><string>$DEVICE_ROLE</string>
        <key>STATUS_IDENTITY_FILE</key><string>$IDENTITY_FILE</string>
    </dict>
    <key>KeepAlive</key><true/>
    <key>RunAtLoad</key><true/>
    <key>StandardErrorPath</key><string>$INSTALL_DIR/device-exporter.log</string>
    <key>StandardOutPath</key><string>$INSTALL_DIR/device-exporter.log</string>
</dict>
</plist>
PLIST
}

write_agent_exporter_agent() {
  cat >"$(plist_path agent-exporter)" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key><string>$LABEL_PREFIX.agent-exporter</string>
    <key>ProgramArguments</key>
    <array>
        <string>$PYTHON_BIN</string>
        <string>$INSTALL_DIR/agent-status-reporter.py</string>
        <string>--serve</string>
        <string>--bind</string><string>$EXPORTER_BIND</string>
        <string>--port</string><string>$AGENT_EXPORTER_PORT</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>STATUS_IDENTITY_FILE</key><string>$IDENTITY_FILE</string>
    </dict>
    <key>KeepAlive</key><true/>
    <key>RunAtLoad</key><true/>
    <key>StandardErrorPath</key><string>$INSTALL_DIR/agent-exporter.log</string>
    <key>StandardOutPath</key><string>$INSTALL_DIR/agent-exporter.log</string>
</dict>
</plist>
PLIST
}

load_agent() {
  local label=$1 plist=$2 domain
  domain="gui/$(id -u)"
  launchctl bootout "$domain/$label" 2>/dev/null || true
  launchctl bootstrap "$domain" "$plist"
  launchctl enable "$domain/$label" 2>/dev/null || true
}

print_summary() {
  echo "Installed Realtime Me macOS probe."
  [[ -n ${GATEWAY_URL:-} ]] && echo "Gateway:         $GATEWAY_URL"
  echo "Device:          $DEVICE_NAME"
  echo "Node exporter:   $EXPORTER_HOST:$NODE_EXPORTER_PORT"
  echo "Device exporter: $EXPORTER_HOST:$DEVICE_EXPORTER_PORT"
  echo "Agent:           $([[ $INSTALL_AGENT == 1 ]] && printf '%s:%s' "$EXPORTER_HOST" "$AGENT_EXPORTER_PORT" || echo disabled)"
  echo "Check:           launchctl list | grep $LABEL_PREFIX"
  echo "Logs:            $INSTALL_DIR/*.log"
  echo "If the gateway cannot scrape, allow node_exporter/python3 through the macOS firewall (System Settings > Network > Firewall)."
}

main() {
  require_not_root
  require_command "$CURL_BIN"
  require_command "$PYTHON_BIN"
  require_command tar
  require_command shasum
  GATEWAY_URL=$(configure_gateway_url)
  detect_exporter_host
  mkdir -p "$INSTALL_DIR" "$LAUNCH_AGENTS_DIR"
  download_exporters
  install_node_exporter
  register_prometheus_targets
  log "Writing LaunchAgents"
  write_node_exporter_agent
  write_device_exporter_agent
  [[ $INSTALL_AGENT == 1 ]] && write_agent_exporter_agent
  log "Loading LaunchAgents"
  load_agent "$LABEL_PREFIX.node-exporter" "$(plist_path node-exporter)"
  load_agent "$LABEL_PREFIX.device-exporter" "$(plist_path device-exporter)"
  [[ $INSTALL_AGENT == 1 ]] && load_agent "$LABEL_PREFIX.agent-exporter" "$(plist_path agent-exporter)"
  print_summary
}

main "$@"
