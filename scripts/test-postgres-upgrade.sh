#!/usr/bin/env sh
# Disposable regression: the real Compose service must preserve old clusters,
# and a logical restore must retain tenant policies, ownership and data.
set -eu
cd "$(dirname "$0")/.."
umask 077
pg_test_dir=$(mktemp -d)
pg_test_id="starter-pg-upgrade-$(date +%s)-$$"
pg_test_old="$pg_test_id-old"
pg_test_new="$pg_test_id-new"
pg_test_container="$pg_test_id-source"
export POSTGRES_USER=starter POSTGRES_DB=starter POSTGRES_PASSWORD=test-only
export APP_DATABASE_PASSWORD=test-app AUTH_DATABASE_PASSWORD=test-auth
export BETTER_AUTH_SECRET=test-only AUTH_INTERNAL_SECRET=test-only
export COMPOSE_PROFILES=
# Explicit project names and an empty env file keep the running app untouched.
old_compose() { docker compose --env-file /dev/null -f compose.yaml -p "$pg_test_old" "$@"; }
new_compose() { docker compose --env-file /dev/null -f compose.yaml -p "$pg_test_new" "$@"; }
cleanup() {
  docker rm -f "$pg_test_container" >/dev/null 2>&1 || :
  old_compose down --volumes >/dev/null 2>&1 || :
  new_compose down --volumes >/dev/null 2>&1 || :
  docker volume rm "${pg_test_old}_postgres_data" >/dev/null 2>&1 || :
  rm -rf "$pg_test_dir"
}
trap cleanup EXIT
trap 'exit 1' HUP INT TERM
wait_source() {
  pg_test_attempt=0
  until docker exec "$pg_test_container" pg_isready -h 127.0.0.1 -U starter -d starter >/dev/null 2>&1; do
    pg_test_attempt=$((pg_test_attempt + 1))
    [ "$pg_test_attempt" -lt 60 ] || { docker logs "$pg_test_container" >&2; exit 1; }
    sleep 1
  done
}
docker volume create --label "com.docker.compose.project=$pg_test_old" \
  --label com.docker.compose.volume=postgres_data "${pg_test_old}_postgres_data" >/dev/null
docker run -d --name "$pg_test_container" \
  -e POSTGRES_USER -e POSTGRES_DB -e POSTGRES_PASSWORD \
  -v "${pg_test_old}_postgres_data:/var/lib/postgresql/data" \
  postgres:17.11-alpine >/dev/null
wait_source
docker exec -i "$pg_test_container" psql -X -v ON_ERROR_STOP=1 -U starter -d starter <<'SQL'
CREATE ROLE starter_auth LOGIN NOSUPERUSER NOBYPASSRLS;
CREATE ROLE starter_app LOGIN NOSUPERUSER NOBYPASSRLS;
CREATE SCHEMA auth AUTHORIZATION starter_auth;
CREATE TABLE auth.upgrade_identity (id text PRIMARY KEY);
ALTER TABLE auth.upgrade_identity OWNER TO starter_auth;
INSERT INTO auth.upgrade_identity VALUES ('preserved-user');
CREATE TABLE public.upgrade_tenant (tenant_id text PRIMARY KEY, value text);
INSERT INTO public.upgrade_tenant VALUES ('alpha', 'preserved-alpha'), ('beta', 'preserved-beta');
ALTER TABLE public.upgrade_tenant ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.upgrade_tenant FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_boundary ON public.upgrade_tenant
  USING (tenant_id = current_setting('app.tenant_id', true));
GRANT SELECT ON public.upgrade_tenant TO starter_app;
SQL
docker exec "$pg_test_container" pg_dump -U starter -d starter -Fc > "$pg_test_dir/database.dump"
docker stop "$pg_test_container" >/dev/null

# Reuse the legacy volume under the new mount, exactly as a normal upgrade does.
if old_compose up -d --wait --wait-timeout 30 postgres > "$pg_test_dir/refusal.log" 2>&1; then
  echo "FAIL: PostgreSQL 18 accepted the PostgreSQL 17 volume." >&2
  exit 1
fi
old_compose logs --no-color postgres > "$pg_test_dir/postgres.log"
grep -q 'upgrading the underlying database' "$pg_test_dir/postgres.log"
old_compose down >/dev/null
docker run --rm -v "${pg_test_old}_postgres_data:/source:ro" \
  --entrypoint sh postgres:17.11-alpine -ec \
  '[ "$(cat /source/PG_VERSION)" = 17 ] && [ ! -s /source/18/docker/PG_VERSION ]'
docker start "$pg_test_container" >/dev/null
wait_source
[ "$(docker exec "$pg_test_container" psql -XAt -U starter -d starter -c 'SELECT count(*) FROM public.upgrade_tenant')" = 2 ]
echo "PASS: PostgreSQL 17 refused safely; original cluster and rows remain readable."

new_compose up -d --wait --wait-timeout 60 postgres >/dev/null
new_compose run --rm --no-deps database-init >/dev/null
new_compose exec -T postgres pg_restore -U starter -d starter \
  --clean --if-exists --exit-on-error < "$pg_test_dir/database.dump"
[ "$(new_compose exec -T postgres psql -XAt -U starter -d starter -c 'SHOW data_directory')" = /var/lib/postgresql/18/docker ]
[ "$(new_compose exec -T postgres psql -XAt -U starter -d starter -c "SELECT tableowner FROM pg_tables WHERE schemaname='auth' AND tablename='upgrade_identity'")" = starter_auth ]
[ "$(new_compose exec -T postgres psql -XAt -U starter -d starter -c 'SELECT id FROM auth.upgrade_identity')" = preserved-user ]
# Exercise the restored restricted role, not merely the superuser's row count.
pg_test_visible=$(new_compose exec -T postgres psql -XAt -v ON_ERROR_STOP=1 -U starter -d starter <<'SQL'
SET ROLE starter_app;
SELECT count(*) FROM public.upgrade_tenant;
SET app.tenant_id = 'alpha';
SELECT value FROM public.upgrade_tenant;
SQL
)
[ "$pg_test_visible" = "SET
0
SET
preserved-alpha" ]
new_compose up -d --force-recreate --wait --wait-timeout 60 postgres >/dev/null
[ "$(new_compose exec -T postgres psql -XAt -U starter -d starter -c 'SELECT count(*) FROM public.upgrade_tenant')" = 2 ]
echo "PASS: fresh PostgreSQL 18 restored identities, ownership, tenant RLS and rows; data survives container recreation."
