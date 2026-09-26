# Upgrade an existing installation

This release replaces Stytch with self-hosted Better Auth. It is an authentication
cutover, not a transparent session migration. Rehearse it on a restored copy of
your database before changing an installation used by customers.

## Back up and rehearse

1. Keep the previous application images, configuration and verified database dump.
   Restore that dump into a separate Compose project; keep the original running.
2. Preserve your PostgreSQL major version and volume mapping. Old installations
   with pgvector objects still need a compatible image until those objects are
   deliberately migrated or removed. Never attach an existing volume to a new
   major version merely by changing the image tag.
3. Transfer settings into the current `.env.example`; retain the existing database
   owner password. Generate independent application/auth database passwords,
   `BETTER_AUTH_SECRET`, and `AUTH_INTERNAL_SECRET`. Set the intended public URL
   and SMTP settings. Local rehearsal can use Mailpit.
4. Run the one-shot schema jobs against the restored database. Historical
   migrations 1–10 are unchanged. New migrations add provider-neutral identifiers
   and row-level security while retaining old provider IDs and existing data.
   Dirty migration ledgers fail closed and need deliberate repair.

The previous lean release retired document, OCR, AI/RAG, storage, Redis and
webhook billing domains. Their legacy tables are not automatically dropped.
Fresh installations use ordinary PostgreSQL and create only current schemas.

## PostgreSQL 17 to 18

The default image is now PostgreSQL 18.6. Its data directory is
`/var/lib/postgresql/18/docker`, with the named volume mounted at
`/var/lib/postgresql`. The volume name remains `postgres_data`: when an existing
17 volume is mounted there, the official image detects its `PG_VERSION` file and
refuses startup instead of creating an empty replacement database. See the
[official image's storage layout](https://hub.docker.com/_/postgres#pgdata) and
[PostgreSQL's major-version upgrade guidance](https://www.postgresql.org/docs/18/upgrading.html).

Do not delete the volume, remove its `PG_VERSION` file, or change `PGDATA` to
bypass this check. An image update does not convert PostgreSQL data files. Use a
logical backup and restore into a **different, fresh volume**. Rehearse before
the maintenance window; the following procedure covers the starter's default
database and its three roles. Additional roles, extensions (including pgvector),
tablespaces or databases require their own reviewed restore plan and compatible
PostgreSQL 18 image. Do not continue past any restore error.

1. Keep the old checkout, Compose project name, images, private environment and
   volume available. In that checkout, stop the application writers while leaving
   PostgreSQL 17 running. Substitute your actual old project name. Keep the dump
   directory outside this repository; it contains private data and role hashes.

   ```sh
   umask 077
   mkdir -p "$HOME/starter-upgrade-backup"
   docker compose -p starter-old stop caddy frontend backend
   docker compose -p starter-old exec -T postgres sh -ec \
     'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' \
     > "$HOME/starter-upgrade-backup/database.dump"
   docker compose -p starter-old exec -T postgres sh -ec \
     'pg_dumpall -U "$POSTGRES_USER" --roles-only' \
     > "$HOME/starter-upgrade-backup/roles.sql"
   ```

   Check both commands succeeded and securely retain both files. A successful
   restore and behavioral checks below are the backup verification. Review
   `roles.sql`: the standard roles are the configured database owner,
   `starter_app` and `starter_auth`. Provision any additional required roles in
   the destination before restoring, preserving their privileges. Do not blindly
   apply a roles dump over the bootstrap owner or drop existing roles.

2. In the new checkout, prepare a private `.env` as described above, with the
   **same database owner name and database name**, and configured application and
   auth passwords. Use a never-used Compose project name such as `starter-pg18`
   and different local HTTP, HTTPS and Mailpit ports for rehearsal. Start only the
   fresh database, then provision its restricted roles; this does not run app
   schema migrations or start application writers.

   ```sh
   docker compose -p starter-pg18 up -d --wait postgres
   docker compose -p starter-pg18 run --rm --no-deps database-init
   docker compose -p starter-pg18 exec -T postgres sh -ec \
     'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --exit-on-error' \
     < "$HOME/starter-upgrade-backup/database.dump"
   ```

   `--clean` is only for the separate destination: it replaces any empty schemas
   created by role provisioning and preserves the dump's table owners, grants and
   RLS policies. Never direct this command at the source cluster. The old volume
   is not mounted in this project. The role provisioning job sets runtime role
   passwords from the new private configuration; the database dump does not
   restore role passwords.

3. Verify the restored organization/account counts, auth identities, migration
   ledgers, object owners, grants and RLS policies before starting the application.
   Run schema jobs and the auth cutover procedure below if needed. Then start the
   new project and check its application and tenancy behavior.

   ```sh
   docker compose -p starter-pg18 up --build -d --wait
   STARTER_URL=http://localhost:YOUR_REHEARSAL_HTTP_PORT ./scripts/smoke.sh
   ```

   Do not run `scripts/test-auth.mjs` against restored customer data: that script
   creates synthetic users and mutates sessions. Use approved rehearsal accounts
   for the real-user acceptance checks. Repeat the final backup/restore with
   writes paused for cutover, and retain the same chosen destination project name
   on subsequent deployments so Compose reuses its restored volume.

Before new writes, rollback is to the saved old checkout and its original
PostgreSQL 17 mount (`postgres_data:/var/lib/postgresql/data`), environment and
project name. The old volume remains intact. If already checked out at the new
release and startup refused, return to that saved checkout to start 17 and take
the backup. After new writes, reconcile them before rollback as described below.
Never use `docker compose down --volumes` on either retained installation.

For a disposable regression without customer data, run
`./scripts/test-postgres-upgrade.sh`. It uses unique projects, removes only its
own resources, and checks rejection of a real 17 volume, recovery of its original
rows, a fresh 18 restore with ownership/RLS, and persistence after recreation.

## Reconcile identities before importing

Stytch represents a member separately in every organization. Better Auth uses
one user with separate organization memberships. The importer combines normalized
email identities while retaining each organization's ID and membership boundary.
The original per-organization profile name is retained in the business
`legacy_full_name` snapshot when the shared Better Auth profile becomes current.
It does not copy sessions or trust old email verification; users must prove their
email again through a new magic link.

Local account rows are historical snapshots, not proof that a person is still
entitled to access. Before applying an import, compare the candidate accounts and
roles with a current Stytch organization/member export. Remove or deactivate stale
membership snapshots in the rehearsal database, resolve conflicting or missing
emails, and confirm at least one legitimate administrator per active workspace.
Review every elevated role. The importer must not be used to restore removed
users from an old backup.

The frontend includes `scripts/import-legacy-auth.mjs`. Run it with the auth
schema connection (`PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, `PGPASSWORD`) and a
separate `LEGACY_DATABASE_URL` owner connection for reading the legacy business
rows. It defaults to a dry run; `--apply` is required to write. Run it only in a
short-lived migration job or controlled development shell. Never add the owner
connection to the long-running frontend or backend environment. Check its output
and resolve rejected rows before proceeding.

Existing organizations keep their original external customer identity for Polar.
Do not recreate those customer identities or substitute a user's Better Auth ID.
Verify billing state for a known organization in rehearsal without taking new
payments. Unconfigured billing must remain disabled.

## Cut over and verify

Pause writes during the final backup/import window, repeat the reviewed migration
against the final snapshot, then start the new application. Stytch sessions and
outstanding login/invitation links are invalid after cutover. Tell users to request
new links; resend still-needed invitations. Do not run both identity authorities
concurrently against the same live membership data.

Verify two distinct tenants, invitations, users belonging to both organizations,
all three roles, logout, member removal, expired links, profile changes and the
last-administrator guard. Confirm the Go database role is neither superuser nor
BYPASSRLS and does not own protected tables. Also verify SMTP externally and test
production billing separately; local email capture and mocks do not prove those
services work.

## Roll back

Before accepting new writes, rollback can restore the previous images,
configuration and database backup together. After users have written data to the
new system, first export and reconcile those changes; restoring an old backup
blindly would lose them. Do not run an automatic destructive down migration or
re-enable old Stytch links against a partially migrated database. Keep backups
and the previous deployment until the new flows have been accepted.
