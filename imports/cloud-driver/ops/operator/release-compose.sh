#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

INCOMING_DIR=/var/lib/cloud-drive-release/incoming-compose
INCOMING_FILE="$INCOMING_DIR/docker-compose.yml"
INSTALL_FILE=/opt/cloud-drive/ops/docker-compose.yml
DEPLOY_SCRIPT=/opt/cloud-drive/ops/scripts/deploy.sh
HOST_LIB=/opt/cloud-drive/ops/scripts/lib.sh
VALIDATOR=/usr/local/libexec/cloud-drive-operator/validate-compose.py
POLICY_DIR=/usr/local/libexec/cloud-drive-operator
RUNTIME_ENV=/etc/cloud-drive/runtime.env
LOCK_FILE=/run/lock/cloud-drive-release.lock
MAX_SOURCE_BYTES=$((128 * 1024))

require_root
require_no_arguments "$@"
for command in docker env flock install mktemp rm stat timeout; do
  require_command "$command"
done
for file in \
  /opt/cloud-drive/.dockerignore \
  /opt/cloud-drive/api/Dockerfile \
  "$DEPLOY_SCRIPT" \
  "$HOST_LIB" \
  "$INSTALL_FILE" \
  "$POLICY_DIR/compose_expected.py" \
  "$POLICY_DIR/compose_policy.py" \
  "$POLICY_DIR/compose_rendered_policy.py" \
  "$POLICY_DIR/compose_source_policy.py" \
  "$VALIDATOR"; do
  require_root_controlled_file "$file"
done
[[ -d "$INCOMING_DIR" && ! -L "$INCOMING_DIR" ]] ||
  die "incoming Compose directory is unavailable: $INCOMING_DIR"
[[ -f "$INCOMING_FILE" && ! -L "$INCOMING_FILE" ]] ||
  die "stage docker-compose.yml in $INCOMING_DIR"
SOURCE_BYTES=$(stat --format='%s' "$INCOMING_FILE")
[[ "$SOURCE_BYTES" =~ ^[0-9]+$ ]] || die 'could not determine staged Compose size'
((SOURCE_BYTES > 0 && SOURCE_BYTES <= MAX_SOURCE_BYTES)) ||
  die 'staged Compose file has an invalid size'

exec 9>"$LOCK_FILE"
flock -n 9 || die 'another cloud-drive release is already running'

WORK_DIR=$(mktemp -d /var/lib/cloud-drive-release/compose.XXXXXX)
cleanup() {
  rm -rf -- "$WORK_DIR"
}
trap cleanup EXIT

CANDIDATE_FILE="$WORK_DIR/docker-compose.yml"
PREVIOUS_FILE="$WORK_DIR/docker-compose.previous.yml"
DUMMY_ENV="$WORK_DIR/validation.env"
DUMMY_RENDERED="$WORK_DIR/validation.json"
RUNTIME_RENDERED="$WORK_DIR/runtime.json"

install -o root -g root -m 0600 "$INCOMING_FILE" "$CANDIDATE_FILE"
"$VALIDATOR" source "$CANDIDATE_FILE"

cat >"$DUMMY_ENV" <<'EOF'
POSTGRES_DB=cloud_drive_validation
POSTGRES_USER=cloud_drive_validation
POSTGRES_PASSWORD=validation-password-0123456789abcdef
PASSWORD_HASH_BASE64=dmFsaWRhdGlvbi1wYXNzd29yZA==
SESSION_SECRET=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
MUSIC_PROVIDER_CREDENTIAL_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
PRIVATE_APP_ORIGINS=https://private.invalid
PUBLIC_APP_ORIGINS=https://public.invalid
SHARE_APP_ORIGIN=https://share.invalid
MUSIC_APP_ORIGIN=https://music.invalid
PRIVATE_API_HOST=private.invalid
PUBLIC_API_HOST=public.invalid
SPOTIFY_CLIENT_ID=
SPOTIFY_CLIENT_SECRET=
CLOUD_DRIVE_DATA_DIR=/var/lib/cloud-drive-validation/data
CLOUD_DRIVE_POSTGRES_DIR=/var/lib/cloud-drive-validation/postgres
EOF
chmod 0600 "$DUMMY_ENV"

render_compose() {
  local environment_file=$1
  local tunnel_token=$2
  local output_file=$3

  env -i PATH="$PATH" HOME=/root TUNNEL_TOKEN="$tunnel_token" \
    timeout 20s docker compose \
      --project-directory /opt/cloud-drive \
      --project-name cloud-drive \
      --env-file "$environment_file" \
      --file "$CANDIDATE_FILE" \
      config --format json >"$output_file"
}

if ! render_compose "$DUMMY_ENV" validation-token "$DUMMY_RENDERED" 2>/dev/null; then
  die 'staged Compose file could not be rendered safely'
fi
"$VALIDATOR" rendered "$DUMMY_RENDERED" \
  --data-directory /var/lib/cloud-drive-validation/data \
  --postgres-directory /var/lib/cloud-drive-validation/postgres

# shellcheck source=../scripts/lib.sh
source "$HOST_LIB"
require_secure_root_file "$RUNTIME_ENV"
DATA_DIRECTORY=$(require_env_value "$RUNTIME_ENV" CLOUD_DRIVE_DATA_DIR)
POSTGRES_DIRECTORY=$(require_env_value "$RUNTIME_ENV" CLOUD_DRIVE_POSTGRES_DIR)
TUNNEL_TOKEN=$(read_cloudflare_tunnel_token "$RUNTIME_ENV")
if ! render_compose "$RUNTIME_ENV" "$TUNNEL_TOKEN" "$RUNTIME_RENDERED" 2>/dev/null; then
  unset TUNNEL_TOKEN
  die 'staged Compose file is incompatible with the runtime configuration'
fi
unset TUNNEL_TOKEN
"$VALIDATOR" rendered "$RUNTIME_RENDERED" \
  --data-directory "$DATA_DIRECTORY" \
  --postgres-directory "$POSTGRES_DIRECTORY"

install -o root -g root -m 0644 "$INSTALL_FILE" "$PREVIOUS_FILE"
install -o root -g root -m 0644 "$CANDIDATE_FILE" "$INSTALL_FILE"

restore_previous() {
  install -o root -g root -m 0644 "$PREVIOUS_FILE" "$INSTALL_FILE"
}

if "$DEPLOY_SCRIPT"; then
  note 'operator Compose release completed'
  exit 0
fi

note 'new Compose release failed; restoring the previous configuration'
restore_previous
if ! "$DEPLOY_SCRIPT"; then
  die 'the Compose release and automatic rollback both failed'
fi
die 'the Compose release failed and was rolled back successfully'
