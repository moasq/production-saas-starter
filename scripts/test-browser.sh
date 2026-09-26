#!/usr/bin/env sh
# Own a fresh synthetic stack; never target a running or restored installation.
set -eu
cd "$(dirname "$0")/.."
command -v docker >/dev/null
[ -d tests/browser/node_modules/@playwright/test ] || { echo 'Run npm ci --prefix tests/browser first.' >&2; exit 1; }
umask 077
browser_project="starter-browser-$(date +%s)-$$"
browser_setup="$browser_project-setup"
browser_http=${BROWSER_HTTP_PORT:-14250}
browser_https=${BROWSER_HTTPS_PORT:-18450}
browser_mail=${BROWSER_MAIL_PORT:-18050}
browser_setup_port=${BROWSER_SETUP_PORT:-14251}
browser_billing_fixture=${BROWSER_BILLING_FIXTURE:-false}
case "$browser_billing_fixture" in true|false) ;; *) echo 'BROWSER_BILLING_FIXTURE must be true or false.' >&2; exit 1;; esac
for browser_port in "$browser_http" "$browser_https" "$browser_mail" "$browser_setup_port"; do
  case "$browser_port" in ''|*[!0-9]*) echo 'Browser ports must be integers.' >&2; exit 1;; esac
  [ "$browser_port" -gt 1024 ] && [ "$browser_port" -lt 65536 ] || exit 1
done
browser_tmp=$(mktemp -d)
# Only this generated env is read. No developer credentials or billing keys.
cat > "$browser_tmp/env" <<ENV
POSTGRES_USER=starter
POSTGRES_DB=starter
POSTGRES_PASSWORD=browser-test-only-owner
APP_DATABASE_PASSWORD=browser-test-only-app
AUTH_DATABASE_PASSWORD=browser-test-only-auth
BETTER_AUTH_SECRET=browser-test-only-session-secret-000000000000
AUTH_INTERNAL_SECRET=browser-test-only-internal-secret-00000000000
APP_BASE_URL=http://localhost:$browser_http
SITE_ADDRESS=:80
BIND_ADDRESS=127.0.0.1
HTTP_PORT=$browser_http
HTTPS_PORT=$browser_https
APP_ENV=DEV
COMPOSE_PROFILES=local
MAILPIT_PORT=$browser_mail
SMTP_HOST=mailpit
SMTP_PORT=1025
BILLING_ENABLED=false
ENV
printf 'services: {}\n' > "$browser_tmp/browser.yaml"
if [ "$browser_billing_fixture" = true ]; then
  # UI-only billing responses are supplied by Playwright. This network cannot
  # reach Polar (or any external service), even if an interception is missed.
  cat >> "$browser_tmp/env" <<ENV
BILLING_ENABLED=true
POLAR_ENVIRONMENT=sandbox
POLAR_ACCESS_TOKEN=browser-fixture-not-a-real-token
POLAR_PRODUCT_ID=00000000-0000-4000-8000-000000000043
ENV
  cat > "$browser_tmp/browser.yaml" <<YAML
services:
  caddy:
    networks: [default, browser-edge]
  mailpit:
    networks: [default, browser-edge]
networks:
  default:
    internal: true
  browser-edge: {}
YAML
fi
# Clear Compose interpolation overrides for credentials/providers, even if callers
# have production settings exported in their shell. Keep Docker connection settings.
unset POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD APP_DATABASE_PASSWORD AUTH_DATABASE_PASSWORD BETTER_AUTH_SECRET AUTH_INTERNAL_SECRET APP_BASE_URL SITE_ADDRESS BIND_ADDRESS HTTP_PORT HTTPS_PORT APP_ENV COMPOSE_PROFILES MAILPIT_PORT SMTP_HOST SMTP_PORT SMTP_SECURE SMTP_USER SMTP_PASSWORD EMAIL_FROM BILLING_ENABLED POLAR_ENVIRONMENT POLAR_ACCESS_TOKEN POLAR_PRODUCT_ID
compose() { docker compose --env-file "$browser_tmp/env" -f compose.yaml -f "$browser_tmp/browser.yaml" -p "$browser_project" "$@"; }
cleanup() {
  docker rm -f "$browser_setup" >/dev/null 2>&1 || :
  compose down --volumes --remove-orphans --rmi local >/dev/null 2>&1 || :
  rm -rf "$browser_tmp"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
export BROWSER_TEST_RUN="$browser_project"
export BROWSER_BILLING_FIXTURE="$browser_billing_fixture"
export STARTER_URL="http://localhost:$browser_http" MAILPIT_URL="http://localhost:$browser_mail" SETUP_URL="http://localhost:$browser_setup_port"
node tests/browser/evidence.mjs prepare
compose up --build -d --wait --wait-timeout 240
browser_frontend=$(compose ps -q frontend)
BROWSER_IMAGE_ID=$(docker inspect --format '{{.Image}}' "$browser_frontend")
export BROWSER_IMAGE_ID
if [ "$browser_billing_fixture" = true ]; then
  # Copy only public action IDs/names, never the manifest's encryption material.
  BROWSER_BILLING_ACTIONS=$(docker exec "$browser_frontend" node -e 'const m=require("./.next/server/server-reference-manifest.json"); process.stdout.write(JSON.stringify(Object.fromEntries(Object.entries(m.node).filter(([,v])=>v.filename?.startsWith("lib/actions/billing/")).map(([id,v])=>[id,v.exportedName]))))')
  export BROWSER_BILLING_ACTIONS
fi
node tests/browser/evidence.mjs built
# A second instance of the same production image has no secrets/database/SMTP.
# This exercises the actual fail-closed setup screen, not a mocked component.
docker run -d --name "$browser_setup" -p "127.0.0.1:$browser_setup_port:3000" "$BROWSER_IMAGE_ID" >/dev/null
./scripts/smoke.sh
npm test --prefix tests/browser -- "$@"
