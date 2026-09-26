# Development

Run `./setup.sh` for the complete local environment. It includes self-hosted
Better Auth, the Go API, PostgreSQL and a local SMTP inbox. Use
`docker compose up --build -d --wait` after changes and
`docker compose logs -f backend frontend` for application logs.

For hot reload, use Go 1.27.1 and Node 24 LTS with pnpm 10.32.1. Use each app's
environment example, a local PostgreSQL connection and the same public URL and
internal bridge secret in both apps. The default database has no published port;
use a private local Compose override when running apps outside containers. Run
the schema jobs before starting `go run ./cmd/api` and `pnpm dev`. Never use the
migration owner's credentials for the running Go API.

The frontend container uses Node 26.9.0 (Current). CI checks the frontend on both
Node 24.21.0 LTS and Node 26.9.0; Node 24 LTS remains the recommended local
development version. The container installs the pinned pnpm version directly
because [Node no longer bundles Corepack from version 25](https://nodejs.org/download/release/v25.8.0/docs/api/corepack.html).
Node 26 is not yet LTS; consider [Node's release guidance](https://nodejs.org/en/about/previous-releases)
when selecting a production runtime.

## Verification

Use `./setup.sh --doctor` for read-only configuration, service, migration and
database-role diagnosis before debugging application code. Run its failure,
recovery and redaction tests with `node --test scripts/doctor.test.mjs`; these
use disposable fixtures and do not require Docker. The dedicated doctor CI
workflow also exercises a fresh Compose deployment and a stopped-service recovery.

```sh
cd go-b2b-starter
go test -race ./...
go vet ./...
cd ../next_b2b_starter
pnpm install --frozen-lockfile
pnpm verify
pnpm audit --prod --audit-level=high
```

Use backend `make test` / `make vet` when Go is not installed. Database tests
require `TEST_DATABASE_URL` pointing at a disposable database with permission to
create test databases. Never point these tests at production. CI also starts a
fresh Compose deployment, tests self-hosted authentication through captured email,
checks protected APIs, and verifies repeated setup preserves configuration.

Keep actual browser verification separate from API tests: use signup, the email
link, workspace selection, invitation acceptance, settings and logout in the UI.
Local SMTP capture is not external email-delivery verification. Polar payments
require their own provider-backed check.

## Database changes

Business SQL source and SQLC bindings live under
`go-b2b-starter/internal/db/postgres/sqlc`. Add forward migrations and regenerate
bindings with `make sqlc`; never edit generated files or applied migrations.
Better Auth has a separate versioned schema under the frontend. Migration jobs
run before applications become ready.

Every new tenant-owned business table needs RLS policies for both reads and
writes, appropriate tenant-bound references, and tests under `starter_app`.
Access it through the existing transaction-scoped tenant store. Test deliberately
unfiltered queries, forged IDs, absent context and connection reuse after
rollback. A query's explicit tenant predicate is useful but not the only boundary.

## Scope

Add product features as a small module under `internal/modules`. Better Auth owns
identity, sessions and membership; Go owns business authorization and data. The
frontend handles presentation and HttpOnly cookies. Developer agents and skills
can assist implementation without adding application AI features or runtime
infrastructure.
