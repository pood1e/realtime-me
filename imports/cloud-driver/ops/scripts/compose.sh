#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../.." && pwd -P)
ENV_FILE=/etc/cloud-drive/runtime.env
COMPOSE_FILE=

usage() {
  cat <<'USAGE'
Usage: compose.sh [--repo-dir PATH] [--env-file PATH] [--compose-file PATH] -- <docker compose arguments>

Runs Docker Compose with the Cloudflare Tunnel token supplied only as a
short-lived environment-sourced Docker secret. The token is never put in argv.
USAGE
}

while (($# > 0)); do
  case "$1" in
    --repo-dir)
      REPO_DIR=${2:?missing value for --repo-dir}
      shift 2
      ;;
    --env-file)
      ENV_FILE=${2:?missing value for --env-file}
      shift 2
      ;;
    --compose-file)
      COMPOSE_FILE=${2:?missing value for --compose-file}
      shift 2
      ;;
    --)
      shift
      break
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      break
      ;;
  esac
done

(($# > 0)) || die 'provide Docker Compose arguments after --'
require_root
require_command docker
require_command tr
REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
if [[ -z "$COMPOSE_FILE" ]]; then
  COMPOSE_FILE="$REPO_DIR/ops/docker-compose.yml"
fi
require_regular_file "$COMPOSE_FILE"
require_secure_root_file "$ENV_FILE"
TUNNEL_TOKEN=$(read_cloudflare_tunnel_token "$ENV_FILE")

export TUNNEL_TOKEN
exec docker compose \
  --project-directory "$REPO_DIR" \
  --project-name cloud-drive \
  --env-file "$ENV_FILE" \
  --file "$COMPOSE_FILE" \
  "$@"
