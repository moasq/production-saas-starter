#!/usr/bin/env sh
# Runs only in the one-shot database-init container, using the migration owner.
set -eu
: "${APP_DATABASE_PASSWORD:?Missing application database password}"
: "${AUTH_DATABASE_PASSWORD:?Missing authentication database password}"
psql -v ON_ERROR_STOP=1 --no-psqlrc \
  --set=app_password="$APP_DATABASE_PASSWORD" \
  --set=auth_password="$AUTH_DATABASE_PASSWORD" \
  --set=database="$PGDATABASE" <<'SQL'
SELECT 'CREATE ROLE starter_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'starter_app') \gexec
SELECT 'CREATE ROLE starter_auth LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'starter_auth') \gexec
ALTER ROLE starter_app WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS PASSWORD :'app_password';
ALTER ROLE starter_auth WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS PASSWORD :'auth_password';
REVOKE ALL ON DATABASE :"database" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"database" TO starter_app, starter_auth;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION starter_auth;
REVOKE ALL ON SCHEMA auth FROM PUBLIC;
SQL
