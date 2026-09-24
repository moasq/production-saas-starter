#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
command -v docker >/dev/null 2>&1 || { echo "Install Docker with Compose first." >&2; exit 1; }
docker compose version >/dev/null
docker info >/dev/null 2>&1 || { echo "Start Docker first." >&2; exit 1; }
if [ ! -f .env ]; then
  command -v openssl >/dev/null 2>&1 || { echo "Install openssl or copy .env.example to .env and set POSTGRES_PASSWORD." >&2; exit 1; }
  umask 077
  starter_password=$(openssl rand -hex 24)
  sed "s/^POSTGRES_PASSWORD=$/POSTGRES_PASSWORD=$starter_password/" .env.example > .env
  echo "Created .env with a private database password. Add Stytch credentials to enable sign-in."
fi
docker compose config --quiet
docker compose up --build -d --wait --wait-timeout 180
echo "Starter is running. Default local URL: http://localhost:3000 (APP_BASE_URL in .env)."
