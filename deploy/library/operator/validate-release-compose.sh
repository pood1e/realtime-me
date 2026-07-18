#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# shellcheck source=lib.sh
source /usr/local/libexec/cloud-drive-operator/lib.sh

VALIDATOR=/usr/local/libexec/cloud-drive-operator/validate-compose.py
HOST_LIB=/opt/cloud-drive/deploy/library/scripts/lib.sh
RUNTIME_ENV=/etc/cloud-drive/runtime.env

require_root
[[ $# -eq 2 ]] || die 'validate-release-compose.sh requires candidate and work directory'
CANDIDATE_FILE=$1
WORK_DIR=$2

for command in docker env timeout; do
  require_command "$command"
done
for file in \
  "$HOST_LIB" \
  /usr/local/libexec/cloud-drive-operator/compose_expected.py \
  /usr/local/libexec/cloud-drive-operator/compose_policy.py \
  /usr/local/libexec/cloud-drive-operator/compose_rendered_policy.py \
  /usr/local/libexec/cloud-drive-operator/compose_source_policy.py \
  "$VALIDATOR"; do
  require_root_controlled_file "$file"
done
require_root_controlled_file "$CANDIDATE_FILE"
[[ -d "$WORK_DIR" && ! -L "$WORK_DIR" ]] || die 'release validation work directory is invalid'

DUMMY_ENV="$WORK_DIR/validation.env"
DUMMY_RENDERED="$WORK_DIR/validation.json"
RUNTIME_RENDERED="$WORK_DIR/runtime.json"

"$VALIDATOR" source "$CANDIDATE_FILE"
cat >"$DUMMY_ENV" <<'EOF'
POSTGRES_DB=cloud_drive_validation
POSTGRES_USER=cloud_drive_validation
POSTGRES_PASSWORD=validation-password-0123456789abcdef
MUSIC_PROVIDER_CREDENTIAL_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=
PUBLIC_SITE_ORIGIN=https://site.invalid
CONSOLE_ORIGIN=https://console.invalid
OIDC_ISSUER=https://identity.invalid/realms/realtime-me
LIBRARY_AUTH_AUDIENCE=realtime-me
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
  local output_file=$2

  env -i PATH="$PATH" HOME=/root \
    timeout 20s docker compose \
      --project-directory /opt/cloud-drive \
      --project-name cloud-drive \
      --env-file "$environment_file" \
      --file "$CANDIDATE_FILE" \
      config --format json >"$output_file"
}

if ! render_compose "$DUMMY_ENV" "$DUMMY_RENDERED" 2>/dev/null; then
  die 'staged Compose file could not be rendered safely'
fi
"$VALIDATOR" rendered "$DUMMY_RENDERED" \
  --project-directory /opt/cloud-drive \
  --data-directory /var/lib/cloud-drive-validation/data \
  --postgres-directory /var/lib/cloud-drive-validation/postgres

# shellcheck source=../scripts/lib.sh
source "$HOST_LIB"
require_secure_root_file "$RUNTIME_ENV"
DATA_DIRECTORY=$(require_env_value "$RUNTIME_ENV" CLOUD_DRIVE_DATA_DIR)
POSTGRES_DIRECTORY=$(require_env_value "$RUNTIME_ENV" CLOUD_DRIVE_POSTGRES_DIR)
if ! render_compose "$RUNTIME_ENV" "$RUNTIME_RENDERED" 2>/dev/null; then
  die 'staged Compose file is incompatible with the runtime configuration'
fi
"$VALIDATOR" rendered "$RUNTIME_RENDERED" \
  --project-directory /opt/cloud-drive \
  --data-directory "$DATA_DIRECTORY" \
  --postgres-directory "$POSTGRES_DIRECTORY"
