#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
command -v docker >/dev/null 2>&1 || { echo "Install Docker with Compose first." >&2; exit 1; }
docker compose version >/dev/null
docker info >/dev/null 2>&1 || { echo "Start Docker first." >&2; exit 1; }
if [ ! -f .env ]; then
  umask 077
  starter_password=$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')
  [ "${#starter_password}" -eq 48 ] || { echo "Could not generate a database password; configure .env manually." >&2; exit 1; }
  sed "s/^POSTGRES_PASSWORD=$/POSTGRES_PASSWORD=$starter_password/" .env.example > .env
  echo "Created .env with a private database password. Add Stytch credentials to enable sign-in."
fi
docker compose config --quiet
if ! docker compose up --build -d --wait --wait-timeout 180; then
  echo "Startup failed. Inspect: docker compose logs --tail=80 backend frontend postgres" >&2
  exit 1
fi
echo "Starter is running. Default local URL: http://localhost:3000 (APP_BASE_URL in .env)."
