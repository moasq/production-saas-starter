#!/usr/bin/env sh
# Read-only diagnostics. Never source .env or print Docker/configuration errors.
set -eu
cd "$(dirname "$0")/.."
doctor_failures=0
doctor_incomplete=0
doctor_timeout=${DOCTOR_TIMEOUT_SECONDS:-20}
if ! printf '%s\n' "$doctor_timeout" | awk '/^[0-9]+$/ && $0 >= 1 && $0 <= 120 {valid=1} END {exit !valid}'; then
  echo 'FAIL DOCTOR_TIMEOUT_SECONDS must be an integer from 1 to 120.' >&2
  exit 1
fi
# Bound each read-only Docker call, including an unresponsive daemon. The
# watchdog owns and cleans up its sleep process; no global processes are killed.
docker() (
  command docker "$@" <&0 2>/dev/null &
  doctor_child=$!
  (
    sleep "$doctor_timeout" & doctor_timer=$!
    trap '[ -z "$doctor_timer" ] || kill "$doctor_timer" 2>/dev/null || true' 0
    trap 'exit 0' TERM INT
    wait "$doctor_timer" || exit 0
    doctor_timer=''
    kill -TERM "$doctor_child" 2>/dev/null || true
    sleep 1 & doctor_timer=$!
    wait "$doctor_timer" || exit 0
    doctor_timer=''
    kill -KILL "$doctor_child" 2>/dev/null || true
  ) >/dev/null 2>&1 &
  doctor_watchdog=$!
  doctor_status=0
  wait "$doctor_child" || doctor_status=$?
  kill -TERM "$doctor_watchdog" 2>/dev/null || true
  wait "$doctor_watchdog" 2>/dev/null || true
  exit "$doctor_status"
)
pass() { printf 'PASS %s\n' "$1"; }
fail() { printf 'FAIL %s\n' "$1"; doctor_failures=$((doctor_failures + 1)); }
warn() { printf 'WARN %s\n' "$1"; }
finish() {
  if [ "$doctor_failures" -gt 0 ]; then
    printf '\nDoctor found %s problem(s). Fix them and rerun ./setup.sh --doctor.\n' "$doctor_failures"
    exit 1
  fi
  if [ "$doctor_incomplete" -gt 0 ]; then
    printf '\nDoctor could not complete all checks; see WARN lines.\n'
    exit 2
  fi
  printf '\nSetup diagnostics passed. Browser journeys, external email and live billing require separate verification.\n'
}
case "$(uname -s)" in
  Darwin|Linux) pass 'Supported POSIX host (Linux, macOS, or Linux inside WSL2).' ;;
  *) fail 'Use Linux/macOS, or WSL2 with Docker Desktop integration; native Windows shells are unsupported.'; finish ;;
esac
# command -v would see the wrapper above, so resolve the executable in a fresh shell.
if ! sh -c 'command -v docker' >/dev/null 2>&1; then fail 'Install Docker with the Compose plugin.'; finish; fi
if ! docker compose version >/dev/null 2>&1; then fail 'Install Docker Compose v2 or newer with config, up --wait, and ps --all support.'; finish; fi
pass 'Docker CLI and Compose are available; local Go, Node and pnpm are optional.'
if ! docker info >/dev/null 2>&1; then fail 'Docker daemon is unavailable. Start Docker/Desktop and enable WSL integration if applicable.'; finish; fi
pass 'Docker daemon is reachable.'
if [ ! -f .env ]; then fail 'Missing .env. Run ./setup.sh to generate private local configuration; doctor does not create it.'; finish; fi
# Let Compose resolve dotenv quoting, interpolation and shell overrides. A small,
# input-only model avoids required-value errors masking which settings are absent.
# Hold the result only in memory and select known names. Never evaluate its text.
if ! doctor_environment=$(printf 'services:\n  doctor:\n    image: scratch\n' | docker compose --project-directory "$PWD" -f - config --environment 2>/dev/null); then
  fail 'Compose cannot read .env. Check dotenv syntax and COMPOSE_PROJECT_NAME; values are suppressed.'; finish
fi
value() { printf '%s\n' "$doctor_environment" | awk -v key="$1" 'index($0,key "=")==1 { print substr($0,length(key)+2); exit }'; }
for doctor_key in POSTGRES_PASSWORD APP_DATABASE_PASSWORD AUTH_DATABASE_PASSWORD BETTER_AUTH_SECRET AUTH_INTERNAL_SECRET; do
  doctor_secret=$(value "$doctor_key")
  if [ "${#doctor_secret}" -lt 32 ]; then fail "$doctor_key must contain at least 32 characters. Run setup for missing values; replace weak values deliberately."; else pass "$doctor_key is present with sufficient length (value hidden)."; fi
done
unset doctor_secret
# Unknown names are hints, not rejected extensions; applications may add their
# own settings. Only identifier-shaped names are displayed, never their values.
doctor_unknown=$(awk -F= '
  FNR==NR { if ($1 ~ /^[A-Z_][A-Z0-9_]*$/) known[$1]=1; next }
  { key=$1; sub(/^[[:space:]]*export[[:space:]]+/,"",key); gsub(/^[[:space:]]+|[[:space:]]+$/,"",key)
    if (key ~ /^[A-Z_][A-Z0-9_]*$/ && !known[key] && key !~ /^(COMPOSE_|DOCKER_|STYTCH_|NEXT_PUBLIC_STYTCH_)/) print key
  }' .env.example .env)
for doctor_key in $doctor_unknown; do warn "$doctor_key is not a starter setting in .env.example; check its spelling or your extension configuration."; done
# Name-only hints catch old provider setup without showing any values.
if awk -F= '/^[[:space:]]*(export[[:space:]]+)?(NEXT_PUBLIC_)?STYTCH_[A-Z0-9_]*[[:space:]]*=/ { found=1 } END { exit !found }' .env; then
  warn 'Legacy Stytch settings are unused. Follow docs/UPGRADING.md; Better Auth roles are defined in application code.'
fi
doctor_url=$(value APP_BASE_URL)
case "$doctor_url" in
  http://*|https://*)
    if ! printf '%s\n' "$doctor_url" | awk '/^https?:\/\/[A-Za-z0-9.:[\]-]+\/?$/ { valid=1 } END { exit !valid }'; then fail 'APP_BASE_URL must be an HTTP(S) origin without credentials, path, query or fragment.'; fi ;;
  *) fail 'Set APP_BASE_URL to the public HTTP(S) origin.' ;;
esac
doctor_profiles=$(value COMPOSE_PROFILES)
doctor_local=false
case ",$doctor_profiles," in *,local,*) doctor_local=true ;; esac
doctor_smtp=$(value SMTP_HOST)
if [ -z "$doctor_smtp" ] || [ "$doctor_smtp" = mailpit ]; then
  if [ "$doctor_local" = true ]; then pass 'Local Mailpit profile is selected for SMTP capture.'; else fail 'SMTP_HOST requires an external provider when COMPOSE_PROFILES excludes local.'; fi
else
  pass 'External SMTP host is configured; delivery and credentials are not tested.'
fi
case "$(value SMTP_SECURE)" in ''|true|false) ;; *) fail 'SMTP_SECURE must be true or false.' ;; esac
if [ "$(value APP_ENV)" = PROD ]; then
  case "$doctor_url" in https://*) ;; *) fail 'Production APP_BASE_URL must use HTTPS.' ;; esac
  if [ "$doctor_local" = true ]; then fail 'Disable the local Mailpit profile for production and configure external SMTP.'; fi
fi
case "$(value APP_ENV)" in ''|DEV|PROD) ;; *) fail 'APP_ENV must be DEV or PROD.' ;; esac
case "$(value BILLING_ENABLED)" in
  true)
    for doctor_key in POLAR_ACCESS_TOKEN POLAR_PRODUCT_ID; do
      if [ -z "$(value "$doctor_key")" ]; then fail "$doctor_key is required when billing is enabled (value hidden)."; fi
    done
    case "$(value POLAR_ENVIRONMENT)" in sandbox|production) ;; *) fail 'POLAR_ENVIRONMENT must be sandbox or production when billing is enabled.' ;; esac
    pass 'Billing configuration checked; no provider calls or payments attempted.' ;;
  ''|false) pass 'Billing is disabled and does not require provider credentials.' ;;
  *) fail 'BILLING_ENABLED must be true or false.' ;;
esac
if ! docker compose config --quiet >/dev/null 2>&1; then
  fail 'Compose configuration is invalid. Check setting names, port numbers and required values in .env.example; raw errors are suppressed.'
  finish
fi
pass 'Compose configuration resolves.'
# Ignore listeners owned by this exact Compose service; reruns must accept them.
# Host listener checks are meaningful only for a local Docker daemon.
doctor_local_daemon=true
doctor_endpoint=${DOCKER_HOST:-}
if [ -z "$doctor_endpoint" ]; then doctor_endpoint=$(docker context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true); fi
case "$doctor_endpoint" in unix://*) ;; *) doctor_local_daemon=false ;; esac
for doctor_spec in HTTP_PORT:3000:caddy:80 HTTPS_PORT:443:caddy:443 MAILPIT_PORT:8025:mailpit:8025; do
  doctor_key=${doctor_spec%%:*}; doctor_rest=${doctor_spec#*:}
  doctor_default=${doctor_rest%%:*}; doctor_rest=${doctor_rest#*:}
  doctor_service=${doctor_rest%%:*}; doctor_target=${doctor_rest#*:}
  if [ "$doctor_service" = mailpit ] && [ "$doctor_local" = false ]; then continue; fi
  doctor_port=$(value "$doctor_key"); doctor_port=${doctor_port:-$doctor_default}
  if ! printf '%s\n' "$doctor_port" | awk '/^[0-9]+$/ && $0 >= 1 && $0 <= 65535 { valid=1 } END { exit !valid }'; then fail "$doctor_key must be a TCP port from 1 to 65535."; continue; fi
  doctor_binding=$(docker compose port "$doctor_service" "$doctor_target" 2>/dev/null || true)
  if [ "${doctor_binding##*:}" = "$doctor_port" ]; then pass "$doctor_key is published by this Compose project."; continue; fi
  if [ "$doctor_local_daemon" = false ]; then
    warn "$doctor_key listener check requires access to the Docker host; remote contexts are not checked."; doctor_incomplete=1
  elif command -v lsof >/dev/null 2>&1; then
    if lsof -nP -iTCP:"$doctor_port" -sTCP:LISTEN -t >/dev/null 2>&1; then fail "$doctor_key is already in use by another listener. Choose an unused port in .env."; else pass "$doctor_key has no visible conflicting TCP listener."; fi
  elif command -v ss >/dev/null 2>&1; then
    if ss -H -ltn "sport = :$doctor_port" 2>/dev/null | awk 'NF { found=1 } END { exit !found }'; then fail "$doctor_key is already in use by another listener. Choose an unused port in .env."; else pass "$doctor_key has no visible conflicting TCP listener."; fi
  else
    warn "Install lsof or iproute2 (ss) to check $doctor_key before startup."; doctor_incomplete=1
  fi
done
unset doctor_environment doctor_url doctor_profiles doctor_smtp
# Inspect only containers selected by Compose; never start, stop, rebuild or reset.
doctor_database_ready=false
doctor_services='postgres backend frontend caddy database-init backend-migrate auth-migrate'
if [ "$doctor_local" = true ]; then doctor_services="$doctor_services mailpit"; fi
for doctor_service in $doctor_services; do
  doctor_id=$(docker compose ps --all --quiet "$doctor_service" 2>/dev/null || true)
  if [ -z "$doctor_id" ]; then fail "$doctor_service has not been created. Run ./setup.sh, then rerun doctor."; continue; fi
  if ! doctor_state=$(docker inspect --format '{{.State.Status}} {{.State.ExitCode}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$doctor_id" 2>/dev/null); then fail "$doctor_service state could not be read."; continue; fi
  doctor_expected_hash=$(docker compose config --hash "$doctor_service" 2>/dev/null | awk 'NF==2 {print $2}')
  doctor_running_hash=$(docker inspect --format '{{index .Config.Labels "com.docker.compose.config-hash"}}' "$doctor_id" 2>/dev/null || true)
  if [ -z "$doctor_expected_hash" ] || [ "$doctor_expected_hash" != "$doctor_running_hash" ]; then fail "$doctor_service does not match resolved Compose configuration. Run ./setup.sh to apply your intended configuration."; fi
  case "$doctor_service:$doctor_state" in
    database-init:exited\ 0\ *|backend-migrate:exited\ 0\ *|auth-migrate:exited\ 0\ *) pass "$doctor_service completed successfully." ;;
    postgres:running\ 0\ healthy) pass 'postgres is healthy.'; doctor_database_ready=true ;;
    backend:running\ 0\ healthy|frontend:running\ 0\ healthy|caddy:running\ 0\ none|caddy:running\ 0\ healthy|mailpit:running\ 0\ none|mailpit:running\ 0\ healthy) pass "$doctor_service is running with its expected health state." ;;
    *) fail "$doctor_service is missing, unhealthy, stopped, or its one-shot job failed. Inspect local logs privately; doctor does not repair it." ;;
  esac
done
if [ "$doctor_database_ready" = true ]; then
  doctor_version=$(awk '/^const SchemaVersion = [0-9]+$/ { print $4 }' go-b2b-starter/internal/db/postgres/runtime.go)
  doctor_auth=''
  for doctor_file in next_b2b_starter/auth-migrations/*.sql; do
    doctor_name=${doctor_file##*/}
    if ! printf '%s\n' "$doctor_name" | awk '/^[0-9]+_[a-z0-9_]+\.sql$/ {valid=1} END {exit !valid}'; then fail 'Auth migration filenames do not match the supported numbered SQL format.'; finish; fi
    if command -v sha256sum >/dev/null 2>&1; then doctor_hash=$(sha256sum "$doctor_file" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then doctor_hash=$(shasum -a 256 "$doctor_file" | awk '{print $1}')
    else fail 'Install sha256sum or shasum to verify applied auth migration checksums.'; finish; fi
    if ! printf '%s\n' "$doctor_hash" | awk '/^[a-f0-9]+$/ && length($0)==64 {valid=1} END {exit !valid}'; then fail 'Could not hash an auth migration.'; finish; fi
    doctor_auth="$doctor_auth{\"name\":\"$doctor_name\",\"checksum\":\"$doctor_hash\"},"
  done
  doctor_auth="[${doctor_auth%,}]"
  case "$doctor_version" in ''|*[!0-9]*) fail 'Cannot determine migration expectations from this checkout.' ;; *)
    if doctor_database=$(docker compose exec -T postgres sh -c 'export PGOPTIONS="-c default_transaction_read_only=on -c statement_timeout=5000"; exec psql --no-psqlrc -X -A -t -v ON_ERROR_STOP=1 -v expected_version="$1" -v expected_auth="$2" -U "$POSTGRES_USER" -d "$POSTGRES_DB"' doctor "$doctor_version" "$doctor_auth" < scripts/doctor-database.sql 2>/dev/null); then
      for doctor_check in business_migrations auth_migrations database_roles tenant_rls auth_isolation membership_roles; do
        if printf '%s\n' "$doctor_database" | awk -F= -v key="$doctor_check" '$1==key && $2=="ok" {found=1} END {exit !found}'; then pass "$doctor_check read-only database check passed."; else fail "$doctor_check check failed. Review migrations and role configuration; do not reset the database."; fi
      done
    else
      fail 'Database diagnostics could not complete. Check migrations/roles and PostgreSQL access; database errors and data are suppressed.'
    fi ;;
  esac
fi
finish
