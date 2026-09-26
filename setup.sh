#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
command -v docker >/dev/null 2>&1 || { echo "Install Docker with Compose first." >&2; exit 1; }
docker compose version >/dev/null
docker info >/dev/null 2>&1 || { echo "Start Docker first." >&2; exit 1; }
umask 077
if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created private .env for local sign-in and email capture."
fi
# Add only missing/blank values during an upgrade; never rotate existing secrets.
for starter_key in POSTGRES_PASSWORD APP_DATABASE_PASSWORD AUTH_DATABASE_PASSWORD BETTER_AUTH_SECRET AUTH_INTERNAL_SECRET; do
  if ! awk -F= -v key="$starter_key" '$1 == key && length($0) > length(key)+1 {found=1} END {exit !found}' .env; then
    starter_secret=$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')
    [ "${#starter_secret}" -eq 64 ] || { echo "Failed generating $starter_key." >&2; exit 1; }
    starter_tmp=$(mktemp .env.setup.XXXXXX)
    awk -F= -v key="$starter_key" '$1 != key' .env > "$starter_tmp"
    printf '%s=%s\n' "$starter_key" "$starter_secret" >> "$starter_tmp"
    mv "$starter_tmp" .env
  fi
done
chmod 600 .env
docker compose config --quiet
if ! docker compose up --build -d --wait --wait-timeout 240; then
  echo "Startup failed. Inspect: docker compose logs --tail=80 postgres database-init backend-migrate auth-migrate backend frontend" >&2
  echo "Existing PostgreSQL 17 volumes require the backup/restore procedure in docs/UPGRADING.md." >&2
  exit 1
fi
echo "Starter is running. Default local URL: http://localhost:3000 (APP_BASE_URL in .env)."
echo "Local email inbox: http://localhost:8025 (MAILPIT_PORT in .env). Production requires your SMTP provider."
