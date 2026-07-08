#!/usr/bin/env bash
set -euo pipefail

# Installs the realtime-me Prometheus exporters on macOS. The host is scraped
# (pull), not pushed, and stays unaware of the gateway: this only runs
# node_exporter (darwin build) for host metrics and status-device-reporter.py
# --serve for media/Bluetooth, both as LaunchAgents under the logged-in user
# (so media/Bluetooth are visible). Register the host centrally with
# register-device.py; Prometheus stamps the device identity via service
# discovery, so the exporters need no token or gateway address.

INSTALL_DIR=${INSTALL_DIR:-$HOME/.realtime-me}
DEVICE_NAME=${STATUS_DEVICE_NAME:-$(scutil --get ComputerName 2>/dev/null || hostname -s 2>/dev/null || echo mac)}
DEVICE_KIND=${STATUS_DEVICE_KIND:-host}
DEVICE_ROLE=${STATUS_DEVICE_ROLE:-desktop}
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
    <key>KeepAlive</key><true/>
    <key>RunAtLoad</key><true/>
    <key>StandardErrorPath</key><string>$INSTALL_DIR/agent-exporter.log</string>
    <key>StandardOutPath</key><string>$INSTALL_DIR/agent-exporter.log</string>
</dict>
</plist>
PLIST
}

load_agent() {
  local label=$1 plist=$2 domain target attempt
  domain="gui/$(id -u)"
  target="$domain/$label"
  # bootout is asynchronous; bootstrap fails with "5: Input/output error" while
  # a prior instance is still torn down. Unload (modern + legacy), wait until the
  # label is really gone, then bootstrap with a few retries.
  launchctl bootout "$target" 2>/dev/null || true
  launchctl unload "$plist" 2>/dev/null || true
  for attempt in 1 2 3 4 5; do
    launchctl print "$target" >/dev/null 2>&1 || break
    sleep 1
  done
  for attempt in 1 2 3; do
    if launchctl bootstrap "$domain" "$plist" 2>/dev/null; then
      launchctl enable "$target" 2>/dev/null || true
      return 0
    fi
    sleep 1
  done
  # Final attempt surfaces the real error if it still will not load.
  launchctl bootstrap "$domain" "$plist"
  launchctl enable "$target" 2>/dev/null || true
}

print_summary() {
  echo "Installed Realtime Me macOS exporters."
  echo "Node exporter:   $EXPORTER_HOST:$NODE_EXPORTER_PORT"
  echo "Device exporter: $EXPORTER_HOST:$DEVICE_EXPORTER_PORT"
  echo "Agent:           $([[ $INSTALL_AGENT == 1 ]] && printf '%s:%s' "$EXPORTER_HOST" "$AGENT_EXPORTER_PORT" || echo disabled)"
  echo "Check:           launchctl list | grep $LABEL_PREFIX   (logs in $INSTALL_DIR/*.log)"
  echo "If the gateway cannot scrape, allow node_exporter/python3 through the macOS firewall."
  echo
  echo "Now register this host centrally (where you can reach the gateway):"
  echo "  STATUS_INGEST_TOKEN=... python3 register-device.py --url <GATEWAY_URL> \\"
  echo "    --host $EXPORTER_HOST --name \"$DEVICE_NAME\" --kind $DEVICE_KIND --role $DEVICE_ROLE$([[ $INSTALL_AGENT == 1 ]] && printf ' --install-agent')"
}

main() {
  require_not_root
  require_command "$CURL_BIN"
  require_command "$PYTHON_BIN"
  require_command tar
  require_command shasum
  detect_exporter_host
  mkdir -p "$INSTALL_DIR" "$LAUNCH_AGENTS_DIR"
  download_exporters
  install_node_exporter
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
