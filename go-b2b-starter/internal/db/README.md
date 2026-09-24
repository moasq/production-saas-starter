# PostgreSQL

PostgreSQL is the only required state service. Fresh installs contain organizations and accounts. Billing reads authoritative Polar state on demand, so there is no local quota or subscription replica.

SQL lives in `postgres/sqlc/query/`. SQLC generates the typed repository layer in `postgres/sqlc/gen/`; regenerate with `make sqlc` and do not hand-edit generated files. Domain repository interfaces belong to each module, and implementations receive the SQLC store. Keep every tenant query scoped by organization ID.

Migrations are embedded in the API binary and run before it serves traffic. `golang-migrate` uses its PostgreSQL lock and `schema_migrations` ledger, refuses dirty migration state, and skips already applied versions. Add new migrations after version 10.

Version 10 is a non-destructive baseline. It works on a fresh plain PostgreSQL instance and on the former clean versions 1–9; existing organizations, accounts, documents and billing data remain in place. Old feature tables and storage objects are not deleted automatically. Back up an existing database and verify the upgrade on a restored copy before production use. The baseline corrects the two conflicting legacy role constraints and accepts the `manager` role. Its downgrade intentionally refuses to destroy tenant data; restore a verified backup if a rollback is necessary. Historical migrations remain byte-for-byte in `postgres/archive/migrations/`. A source adapter validates old ledger versions against this archive while starting and advancing at version 10; removed feature migrations never run during fresh setup or upgrade.

Run database behavior tests against a disposable PostgreSQL service:

```sh
TEST_DATABASE_URL='postgres://postgres:test@localhost:5432/postgres?sslmode=disable' go test ./internal/db/postgres -v
```

The test account needs CREATE DATABASE. Tests create temporary databases, exercise fresh installs, repeated startup, legacy tenant preservation, manager roles, and dirty-state refusal, then remove only their own databases. Without this variable integration tests explicitly skip.
