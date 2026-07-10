#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../.." && pwd -P)
PAGES_ENV_FILE=${PAGES_ENV_FILE:-}
PRIVATE_PAGES_PROJECT=${PRIVATE_PAGES_PROJECT:-}
SHARE_PAGES_PROJECT=${SHARE_PAGES_PROJECT:-}
VITE_PRIVATE_API_BASE=${VITE_PRIVATE_API_BASE:-}
VITE_SHARE_API_BASE=${VITE_SHARE_API_BASE:-}
BRANCH=main
BUILD=true

usage() {
  cat <<'USAGE'
Usage: deploy-pages.sh [options]

Builds and manually deploys the private and public-share static applications to
Cloudflare Pages. Authenticate Wrangler before running it; this script neither
accepts nor prints a Cloudflare token.

Options:
  --repo-dir PATH          Checked-out cloud-drive repository
  --env-file PATH          Local Pages config (default: ops/pages.env)
  --private-project NAME   Private Pages project
  --share-project NAME     Public-share Pages project
  --branch NAME            Pages branch (default: main)
  --skip-build             Deploy existing dist directories
  -h, --help               Show this help

Environment overrides:
  PAGES_ENV_FILE, PRIVATE_PAGES_PROJECT, SHARE_PAGES_PROJECT,
  VITE_PRIVATE_API_BASE, VITE_SHARE_API_BASE
USAGE
}

while (($# > 0)); do
  case "$1" in
    --repo-dir)
      REPO_DIR=${2:?missing value for --repo-dir}
      shift 2
      ;;
    --env-file)
      PAGES_ENV_FILE=${2:?missing value for --env-file}
      shift 2
      ;;
    --private-project)
      PRIVATE_PAGES_PROJECT=${2:?missing value for --private-project}
      shift 2
      ;;
    --share-project)
      SHARE_PAGES_PROJECT=${2:?missing value for --share-project}
      shift 2
      ;;
    --branch)
      BRANCH=${2:?missing value for --branch}
      shift 2
      ;;
    --skip-build)
      BUILD=false
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown option: $1"
      ;;
  esac
done

require_command pnpm
REPO_DIR=$(cd -- "$REPO_DIR" && pwd -P)
[[ -f "$REPO_DIR/package.json" ]] || die "package.json not found in repository: $REPO_DIR"
PAGES_ENV_FILE=${PAGES_ENV_FILE:-$REPO_DIR/ops/pages.env}
if [[ -f "$PAGES_ENV_FILE" ]]; then
  [[ -n "$PRIVATE_PAGES_PROJECT" ]] || PRIVATE_PAGES_PROJECT=$(require_env_value "$PAGES_ENV_FILE" PRIVATE_PAGES_PROJECT)
  [[ -n "$SHARE_PAGES_PROJECT" ]] || SHARE_PAGES_PROJECT=$(require_env_value "$PAGES_ENV_FILE" SHARE_PAGES_PROJECT)
  [[ -n "$VITE_PRIVATE_API_BASE" ]] || VITE_PRIVATE_API_BASE=$(require_env_value "$PAGES_ENV_FILE" VITE_PRIVATE_API_BASE)
  [[ -n "$VITE_SHARE_API_BASE" ]] || VITE_SHARE_API_BASE=$(require_env_value "$PAGES_ENV_FILE" VITE_SHARE_API_BASE)
fi

[[ -n "$PRIVATE_PAGES_PROJECT" ]] || die 'private Pages project is required; configure ops/pages.env or PRIVATE_PAGES_PROJECT'
[[ -n "$SHARE_PAGES_PROJECT" ]] || die 'share Pages project is required; configure ops/pages.env or SHARE_PAGES_PROJECT'
[[ "$PRIVATE_PAGES_PROJECT" =~ ^[a-z0-9-]+$ ]] || die 'private Pages project name must contain lowercase letters, digits, or hyphens'
[[ "$SHARE_PAGES_PROJECT" =~ ^[a-z0-9-]+$ ]] || die 'share Pages project name must contain lowercase letters, digits, or hyphens'
[[ "$BRANCH" =~ ^[A-Za-z0-9._/-]+$ ]] || die 'branch name contains unsupported characters'

cd -- "$REPO_DIR"
WRANGLER="$REPO_DIR/node_modules/.bin/wrangler"
[[ -x "$WRANGLER" ]] || die 'Wrangler is not installed; run pnpm install first'
"$WRANGLER" --version >/dev/null
"$WRANGLER" whoami >/dev/null

if [[ "$BUILD" == true ]]; then
  [[ -n "$VITE_PRIVATE_API_BASE" ]] || die 'VITE_PRIVATE_API_BASE is required when building Pages'
  [[ -n "$VITE_SHARE_API_BASE" ]] || die 'VITE_SHARE_API_BASE is required when building Pages'
  export VITE_PRIVATE_API_BASE VITE_SHARE_API_BASE
  note 'building Cloudflare Pages artifacts'
  pnpm build:web
fi

PRIVATE_DIST="$REPO_DIR/web/apps/private/dist"
SHARE_DIST="$REPO_DIR/web/apps/share/dist"
[[ -d "$PRIVATE_DIST" ]] || die "private Pages artifact is missing: $PRIVATE_DIST"
[[ -d "$SHARE_DIST" ]] || die "share Pages artifact is missing: $SHARE_DIST"

note "deploying private Pages artifact to $PRIVATE_PAGES_PROJECT"
"$WRANGLER" pages deploy "$PRIVATE_DIST" --project-name "$PRIVATE_PAGES_PROJECT" --branch "$BRANCH"
note "deploying share Pages artifact to $SHARE_PAGES_PROJECT"
"$WRANGLER" pages deploy "$SHARE_DIST" --project-name "$SHARE_PAGES_PROJECT" --branch "$BRANCH"
note 'Pages deployments completed'
