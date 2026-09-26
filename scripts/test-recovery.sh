#!/usr/bin/env sh
# Fresh synthetic projects only. Never accept a caller's database or .env.
set -eu
cd "$(dirname "$0")/.."
umask 077
recovery_tmp=$(mktemp -d)
recovery_id="starter-recovery-$(date +%s)-$$"
recovery_source="$recovery_id-source"
recovery_target="$recovery_id-restored"
recovery_http=${RECOVERY_HTTP_PORT:-14245}
recovery_https=${RECOVERY_HTTPS_PORT:-18445}
recovery_mail=${RECOVERY_MAIL_PORT:-18045}
recovery_target_http=${RECOVERY_RESTORE_HTTP_PORT:-14246}
recovery_target_https=${RECOVERY_RESTORE_HTTPS_PORT:-18446}
recovery_target_mail=${RECOVERY_RESTORE_MAIL_PORT:-18046}
for recovery_port in "$recovery_http" "$recovery_https" "$recovery_mail" "$recovery_target_http" "$recovery_target_https" "$recovery_target_mail"; do
  case "$recovery_port" in ''|*[!0-9]*) echo 'Recovery ports must be integers.' >&2; exit 1;; esac
  [ "$recovery_port" -gt 1024 ] && [ "$recovery_port" -lt 65536 ] || exit 1
done
# Clear ambient provider settings, while retaining Docker connection settings.
unset POSTGRES_USER POSTGRES_DB POSTGRES_PASSWORD APP_DATABASE_PASSWORD AUTH_DATABASE_PASSWORD BETTER_AUTH_SECRET AUTH_INTERNAL_SECRET APP_BASE_URL SITE_ADDRESS BIND_ADDRESS HTTP_PORT HTTPS_PORT APP_ENV COMPOSE_PROFILES MAILPIT_PORT SMTP_HOST SMTP_PORT SMTP_SECURE SMTP_USER SMTP_PASSWORD EMAIL_FROM BILLING_ENABLED POLAR_ENVIRONMENT POLAR_ACCESS_TOKEN POLAR_PRODUCT_ID
write_env() {
  cat <<ENV
POSTGRES_USER=starter
POSTGRES_DB=starter
POSTGRES_PASSWORD=recovery-test-only-owner
APP_DATABASE_PASSWORD=recovery-test-only-app
AUTH_DATABASE_PASSWORD=recovery-test-only-auth
BETTER_AUTH_SECRET=recovery-test-only-session-secret-0000000000
AUTH_INTERNAL_SECRET=recovery-test-only-internal-secret-000000000
APP_BASE_URL=http://localhost:$1
SITE_ADDRESS=:80
BIND_ADDRESS=127.0.0.1
HTTP_PORT=$1
HTTPS_PORT=$2
MAILPIT_PORT=$3
APP_ENV=DEV
COMPOSE_PROFILES=local
SMTP_HOST=mailpit
SMTP_PORT=1025
BILLING_ENABLED=false
ENV
}
write_env "$recovery_http" "$recovery_https" "$recovery_mail" > "$recovery_tmp/source.env"
write_env "$recovery_target_http" "$recovery_target_https" "$recovery_target_mail" > "$recovery_tmp/target.env"
# Build once; both disposable installations use these exact application images.
cat > "$recovery_tmp/images.yaml" <<YAML
services:
  backend:
    image: $recovery_id-backend:local
  backend-migrate:
    image: $recovery_id-backend:local
  frontend:
    image: $recovery_id-frontend:local
  auth-migrate:
    image: $recovery_id-frontend:local
YAML
source_compose() { docker compose --env-file "$recovery_tmp/source.env" -f compose.yaml -f "$recovery_tmp/images.yaml" -p "$recovery_source" "$@"; }
target_compose() { docker compose --env-file "$recovery_tmp/target.env" -f compose.yaml -f "$recovery_tmp/images.yaml" -p "$recovery_target" "$@"; }
cleanup() {
  source_compose down --volumes --remove-orphans >/dev/null 2>&1 || :
  target_compose down --volumes --remove-orphans >/dev/null 2>&1 || :
  docker image rm "$recovery_id-backend:local" "$recovery_id-frontend:local" >/dev/null 2>&1 || :
  rm -rf "$recovery_tmp"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
source_sql() { source_compose exec -T postgres psql -X -qAt -v ON_ERROR_STOP=1 -U starter -d starter "$@"; }
target_sql() { target_compose exec -T postgres psql -X -qAt -v ON_ERROR_STOP=1 -U starter -d starter "$@"; }
assert_status() {
  recovery_status=$(curl --silent --show-error --max-time 5 --output "$recovery_tmp/response" --write-out '%{http_code}' "http://localhost:$1$2")
  [ "$recovery_status" = "$3" ] || { echo "FAIL $2 expected $3, received $recovery_status" >&2; exit 1; }
}
source_compose build backend
source_compose build frontend
source_compose up --no-build -d --wait --wait-timeout 240
STARTER_URL="http://localhost:$recovery_http" ./scripts/smoke.sh
source_sql <<'SQL'
INSERT INTO auth."user" (id,name,email,"emailVerified") VALUES ('recovery-user','Recovery User','recovery@example.test',true);
INSERT INTO auth.organization (id,name,slug,"createdAt") VALUES ('recovery-alpha','Alpha','recovery-alpha',now()),('recovery-beta','Beta','recovery-beta',now());
INSERT INTO auth.member (id,"organizationId","userId",role,"createdAt") VALUES ('recovery-member','recovery-alpha','recovery-user','admin',now());
INSERT INTO organizations.organizations (name,slug,auth_org_id,polar_customer_external_id) VALUES ('Alpha','recovery-alpha','recovery-alpha','original-billing-alpha'),('Beta','recovery-beta','recovery-beta','original-billing-beta');
INSERT INTO organizations.accounts (organization_id,email,full_name,role,auth_user_id,auth_member_id) SELECT id,'recovery@example.test','Recovery User','admin','recovery-user','recovery-member' FROM organizations.organizations WHERE auth_org_id='recovery-alpha';
SQL
# Pause writers before the backup. The original database and volume remain intact.
source_compose stop caddy frontend backend >/dev/null
source_compose exec -T postgres pg_dump -U starter -d starter -Fc > "$recovery_tmp/database.dump"
source_compose exec -T postgres pg_dumpall -U starter --roles-only > "$recovery_tmp/roles.sql"
[ -s "$recovery_tmp/database.dump" ] && [ -s "$recovery_tmp/roles.sql" ]
source_compose up --no-build -d --wait --wait-timeout 120 >/dev/null

source_compose stop postgres >/dev/null
assert_status "$recovery_http" /livez 200
assert_status "$recovery_http" /readyz 503
source_compose up --no-build -d --wait --wait-timeout 60 postgres >/dev/null
assert_status "$recovery_http" /readyz 200
printf '%s\n' 'PASS: database outage returns readiness 503 while liveness remains 200; restart recovers.'

source_sql -c 'GRANT starter TO starter_app' >/dev/null
assert_status "$recovery_http" /readyz 503
assert_status "$recovery_http" /livez 200
source_sql -c 'REVOKE starter FROM starter_app' >/dev/null
source_sql -c 'ALTER TABLE organizations.accounts NO FORCE ROW LEVEL SECURITY' >/dev/null
assert_status "$recovery_http" /readyz 503
source_sql -c 'ALTER TABLE organizations.accounts FORCE ROW LEVEL SECURITY' >/dev/null
assert_status "$recovery_http" /readyz 200
printf '%s\n' 'PASS: live elevated-role and FORCE RLS drift are rejected and recovery is detected.'

# Deliberately claim version 10 against version 11 objects. The real immutable
# migration fails on an existing column, leaving the library's dirty marker.
# This is only fault injection on this generated database, never a repair recipe.
source_sql -c 'UPDATE public.schema_migrations SET version=10,dirty=false' >/dev/null
if source_compose run --rm --no-deps backend-migrate > "$recovery_tmp/migration.log" 2>&1; then
  echo 'FAIL: conflicting migration unexpectedly succeeded.' >&2; exit 1
fi
[ "$(source_sql -c 'SELECT version::text || chr(58) || dirty::text FROM public.schema_migrations')" = '11:true' ]
assert_status "$recovery_http" /readyz 503
assert_status "$recovery_http" /livez 200
if source_compose run --rm --no-deps backend-migrate > "$recovery_tmp/dirty.log" 2>&1; then
  echo 'FAIL: dirty migration was silently retried.' >&2; exit 1
fi
printf '%s\n' 'PASS: a real failed migration leaves a dirty ledger, blocks readiness and refuses automatic retry.'

# Restore only into a separate fresh volume. Standard roles are reprovisioned
# from private config rather than applying a role dump over the bootstrap owner.
target_compose up --no-build -d --wait --wait-timeout 60 postgres >/dev/null
target_compose run --rm --no-deps database-init >/dev/null
target_compose exec -T postgres pg_restore -U starter -d starter --clean --if-exists --exit-on-error < "$recovery_tmp/database.dump"
[ "$(target_sql -c 'SELECT count(*) FROM organizations.organizations')" = 2 ]
[ "$(target_sql -c 'SELECT count(*) FROM auth.member')" = 1 ]
[ "$(target_sql -c "SELECT tableowner FROM pg_tables WHERE schemaname='auth' AND tablename='user'")" = starter_auth ]
recovery_visible=$(target_sql <<'SQL'
SET ROLE starter_app;
SELECT count(*) FROM organizations.organizations;
SET app.tenant_id='recovery-alpha';
SELECT polar_customer_external_id FROM organizations.organizations;
SELECT count(*) FROM organizations.accounts;
SQL
)
[ "$recovery_visible" = "0
original-billing-alpha
1" ]
target_compose up --no-build -d --wait --wait-timeout 240 >/dev/null
STARTER_URL="http://localhost:$recovery_target_http" ./scripts/smoke.sh
# Both migration jobs must accept the restored ledgers; Go checks the clean business version.
recovery_version=$(awk '/^const SchemaVersion = [0-9]+$/ {print $4}' go-b2b-starter/internal/db/postgres/runtime.go)
[ "$(target_sql -c 'SELECT version::text || chr(58) || dirty::text FROM public.schema_migrations')" = "$recovery_version:false" ]
# Preserve the failed source for diagnosis until cleanup; restore never overwrote it.
[ "$(source_sql -c 'SELECT version::text || chr(58) || dirty::text FROM public.schema_migrations')" = '11:true' ]
printf '%s\n' 'PASS: separate restore preserves auth identities, membership, billing IDs, ownership and tenant RLS; restored app is ready.'
