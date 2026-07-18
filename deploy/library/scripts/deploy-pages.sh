#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

REPO_DIR=$(cd -- "$SCRIPT_DIR/../../.." && pwd -P)
PAGES_ENV_FILE=
BRANCH=main
BUILD=true

usage() {
  cat <<'USAGE'
Usage: deploy-pages.sh [options]

Builds and deploys the authentication, drive, books, music, images,
wallpapers, and share applications to their Cloudflare Pages projects.

Options:
  --repo-dir PATH   Checked-out cloud-drive repository
  --env-file PATH   Local Pages config (default: deploy/library/pages.env)
  --branch NAME     Pages branch (default: main)
  --skip-build      Deploy existing dist directories
  -h, --help        Show this help
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
[[ -f "$REPO_DIR/package.json" ]] || die "package.json not found: $REPO_DIR"
PAGES_ENV_FILE=${PAGES_ENV_FILE:-$REPO_DIR/deploy/library/pages.env}
require_regular_file "$PAGES_ENV_FILE"
[[ "$BRANCH" =~ ^[A-Za-z0-9._/-]+$ ]] || die 'branch name contains unsupported characters'

apps=(auth drive books music images wallpapers share)
project_keys=(
  AUTH_PAGES_PROJECT DRIVE_PAGES_PROJECT BOOKS_PAGES_PROJECT
  MUSIC_PAGES_PROJECT IMAGES_PAGES_PROJECT WALLPAPERS_PAGES_PROJECT
  SHARE_PAGES_PROJECT
)
projects=()
for key in "${project_keys[@]}"; do
  value=$(require_env_value "$PAGES_ENV_FILE" "$key")
  [[ "$value" =~ ^[a-z0-9-]+$ ]] || die "$key must contain lowercase letters, digits, or hyphens"
  projects+=("$value")
done

vite_keys=(
  VITE_PRIVATE_API_BASE VITE_PUBLIC_API_BASE VITE_AUTH_APP_ORIGIN
  VITE_DRIVE_APP_ORIGIN VITE_BOOKS_APP_ORIGIN VITE_MUSIC_APP_ORIGIN
  VITE_IMAGES_APP_ORIGIN VITE_WALLPAPERS_APP_ORIGIN VITE_SHARE_APP_ORIGIN
  VITE_DEFAULT_RETURN_URL
)
for key in "${vite_keys[@]}"; do
  value=$(require_env_value "$PAGES_ENV_FILE" "$key")
  [[ "$value" =~ ^https://[^/]+$ ]] || die "$key must be an exact HTTPS origin"
  export "$key=$value"
done

cd -- "$REPO_DIR"
WRANGLER="$REPO_DIR/node_modules/.bin/wrangler"
[[ -x "$WRANGLER" ]] || die 'Wrangler is not installed; run pnpm install first'
"$WRANGLER" whoami >/dev/null

if [[ "$BUILD" == true ]]; then
  note 'building seven Cloudflare Pages applications'
  pnpm --filter './apps/web/library/**' --if-present build
fi

for index in "${!apps[@]}"; do
  app=${apps[$index]}
  project=${projects[$index]}
  artifact="$REPO_DIR/apps/web/library/$app/dist"
  [[ -d "$artifact" ]] || die "Pages artifact is missing: $artifact"
  note "deploying $app to Cloudflare Pages project $project"
  "$WRANGLER" pages deploy "$artifact" --project-name "$project" --branch "$BRANCH"
done

note 'all Pages deployments completed'
