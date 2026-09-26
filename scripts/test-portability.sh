#!/usr/bin/env sh
# Native Linux-container smoke test. Owns only a generated project and its images.
set -eu
cd "$(dirname "$0")/.."
case "$(docker info --format '{{.OSType}}/{{.Architecture}}')" in
  linux/x86_64|linux/amd64) portability_arch=amd64; portability_other=arm64; portability_machine=183 ;;
  linux/aarch64|linux/arm64) portability_arch=arm64; portability_other=amd64; portability_machine=62 ;;
  *) echo 'Portability checks require native Linux AMD64 or ARM64 containers.' >&2; exit 1 ;;
esac
[ "$#" -le 1 ] && [ "${1:-$portability_arch}" = "$portability_arch" ] || {
  echo "Expected native $portability_arch; emulation does not establish native runtime support." >&2; exit 1;
}
umask 077
portability_tmp=$(mktemp -d)
portability_project="starter-portability-$(date +%s)-$$"
portability_http=${PORTABILITY_HTTP_PORT:-14240}
portability_https=${PORTABILITY_HTTPS_PORT:-18440}
portability_mail=${PORTABILITY_MAIL_PORT:-18040}
for portability_port in "$portability_http" "$portability_https" "$portability_mail"; do
  case "$portability_port" in ''|*[!0-9]*) echo 'Portability ports must be integers.' >&2; exit 1;; esac
  [ "$portability_port" -gt 1024 ] && [ "$portability_port" -lt 65536 ] || exit 1
done
# Ignore ambient application/provider settings and the caller's .env. Retain
# Docker connection settings; never switch context or touch existing projects.
unset DOCKER_DEFAULT_PLATFORM POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD APP_DATABASE_PASSWORD AUTH_DATABASE_PASSWORD BETTER_AUTH_SECRET AUTH_INTERNAL_SECRET APP_BASE_URL SITE_ADDRESS BIND_ADDRESS HTTP_PORT HTTPS_PORT APP_ENV COMPOSE_PROFILES MAILPIT_PORT SMTP_HOST SMTP_PORT SMTP_SECURE SMTP_USER SMTP_PASSWORD EMAIL_FROM BILLING_ENABLED POLAR_ENVIRONMENT POLAR_ACCESS_TOKEN POLAR_PRODUCT_ID
cat > "$portability_tmp/test.env" <<ENV
POSTGRES_USER=starter
POSTGRES_DB=starter
POSTGRES_PASSWORD=portability-test-only-owner
APP_DATABASE_PASSWORD=portability-test-only-app
AUTH_DATABASE_PASSWORD=portability-test-only-auth
BETTER_AUTH_SECRET=portability-test-only-session-secret-0000000000
AUTH_INTERNAL_SECRET=portability-test-only-bridge-secret-00000000000
APP_BASE_URL=http://localhost:$portability_http
SITE_ADDRESS=:80
BIND_ADDRESS=127.0.0.1
HTTP_PORT=$portability_http
HTTPS_PORT=$portability_https
MAILPIT_PORT=$portability_mail
APP_ENV=DEV
COMPOSE_PROFILES=local
SMTP_HOST=mailpit
SMTP_PORT=1025
BILLING_ENABLED=false
ENV
cat > "$portability_tmp/images.yaml" <<YAML
services:
  backend:
    image: $portability_project-backend:local
  backend-migrate:
    image: $portability_project-backend:local
  frontend:
    image: $portability_project-frontend:local
  auth-migrate:
    image: $portability_project-frontend:local
YAML
portability_compose() { docker compose --env-file "$portability_tmp/test.env" -f compose.yaml -f "$portability_tmp/images.yaml" -p "$portability_project" "$@"; }
cleanup() {
  portability_compose down --volumes --remove-orphans >/dev/null 2>&1 || :
  docker image rm "$portability_project-backend:local" "$portability_project-frontend:local" >/dev/null 2>&1 || :
  rm -rf "$portability_tmp"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM

# Sequential builds keep the local memory requirement below simultaneous Go and
# Next.js builds. No registry credentials, provider secrets or image push required.
portability_compose build backend
portability_compose build frontend
for portability_service in backend frontend; do
  portability_image="$portability_project-$portability_service:local"
  [ "$(docker image inspect "$portability_image" --format '{{.Os}}/{{.Architecture}}')" = "linux/$portability_arch" ] || {
    echo "$portability_service image does not match the native target." >&2; exit 1;
  }
done
portability_compose up --no-build -d --wait --wait-timeout 240
STARTER_URL="http://localhost:$portability_http" ./scripts/smoke.sh
portability_compose exec -T frontend node -e "if (process.platform !== 'linux' || process.arch !== ('$portability_arch' === 'amd64' ? 'x64' : 'arm64')) process.exit(1)"
printf 'PASS: native linux/%s production images, migrations and local development Compose stack.\n' "$portability_arch"

# Cross-compilation is a different claim from runtime support. The scratch export
# performs no foreign-architecture RUN command and needs no QEMU registration.
docker build --platform "linux/$portability_other" --target binary --output "type=local,dest=$portability_tmp/cross" go-b2b-starter
portability_ident=$(od -An -tx1 -N6 "$portability_tmp/cross/api" | tr -d ' \n')
[ "$portability_ident" = '7f454c460201' ] || { echo 'Expected a little-endian 64-bit ELF binary.' >&2; exit 1; }
portability_elf_machine=$(od -An -tu2 -j18 -N2 "$portability_tmp/cross/api" | tr -d ' \n')
[ "$portability_elf_machine" = "$portability_machine" ] || { echo 'Cross-compiled Go executable has the wrong CPU architecture.' >&2; exit 1; }
printf 'PASS: linux/%s Go binary cross-compiled on %s and verified by its ELF machine header.\n' "$portability_other" "$portability_arch"
