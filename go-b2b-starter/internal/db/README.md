# PostgreSQL and tenant isolation

Business SQL is in `postgres/sqlc/query/`; regenerate its bindings with `make sqlc`.
Every application query uses `tenant.Store.Run` with an already verified tenant
context. It begins a transaction, sets `app.tenant_id` with `set_config(..., true)`,
runs generated queries and commits. Rollback and pooled-connection reuse cannot
carry the tenant setting to the next operation. SQL must still include meaningful
organization predicates; FORCE RLS is an additional independent boundary.

Organizations and accounts enforce row security even for their table owner. The
API runs as a separate NOSUPERUSER NOBYPASSRLS role with no table ownership and no
auth-schema grants. An absent scope sees no tenant rows and cannot insert them.
The auth server owns only its separate auth schema; it does not query business data
in normal request handling.

Run the embedded migration command `api migrate` with privileged credentials before
starting the restricted API. The serving command validates clean schema version 11
and both forced row policies; it never executes migrations. Historical migrations
1–9 and baseline 10 remain unchanged. Migration 11 preserves legacy organization,
account and billing identifiers, adds neutral auth/billing mapping columns and
activates RLS. It snapshots every pre-cutover account's `full_name` in nullable
`legacy_full_name`. Current profile names can follow the shared Better Auth user
without erasing distinct original per-organization names; synchronization never
updates the snapshot and new accounts leave it empty. It never deletes customer data. Roll back with a coordinated,
verified pre-cutover backup, not a partial auth downgrade.

With a disposable superuser `TEST_DATABASE_URL`, `go test ./internal/db/postgres`
checks fresh and repeated migration, legacy data and billing linkage, two-tenant
reads/writes, unscoped access, rolled-back scope, pooled reuse and privileged-role
refusal. These integration tests explicitly skip when no test database is given.
