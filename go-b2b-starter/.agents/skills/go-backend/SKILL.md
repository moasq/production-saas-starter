---
name: "go-backend"
description: "This skill applies when adding a Go API endpoint, changing SQLC queries or migrations, fixing tenant authorization, or testing backend behavior in this starter."
---

# Go backend workflow

Paths and commands below are repository-relative unless a working directory is stated.
Read `go-b2b-starter/AGENTS.md` and `docs/ARCHITECTURE.md` first. Locate the nearest implementation
in `go-b2b-starter/internal/modules/` and follow its handler, service, and repository
boundary. Prefer an existing interface over a new abstraction with one caller.

For public business API changes, update `go-b2b-starter/apicontract/openapi.json`
with the mounted implementation. Follow `docs/decisions/0001-business-api.md`:
run `pnpm --dir next_b2b_starter api:generate`, check the generated diff, and run
`pnpm --dir next_b2b_starter api:check`. Hand the tested schema and commit to the
frontend specialist; generated types do not replace Go authorization or validation.

1. State the actor, tenant, input, response shape, and failure behavior. Trace the
   route through `internal/modules/auth/` to the organization/account resolver.
   Require identity, current membership, action permission, and resource ownership
   before mutation. A request header, resource ID or role label is never authority.
   Finish a new endpoint with an HTTP contract test for its status/body and a
   denied request that proves the repository/provider was not called. Route
   registration alone is not completion. Organization mutation examples are in
   `internal/modules/organizations/authorization_contract_test.go`.
   For identity or membership changes, also use `next_b2b_starter/.agents/skills/auth-integration/SKILL.md`.
2. Keep HTTP decoding/status translation in handlers and application decisions in
   services. Pass request context to SQL/provider calls, bound outbound timeouts,
   and close resources. Start no detached goroutine without cancellation and an owner.
   Do not store request context in a service or replace it with `context.Background()`.
   An owner must cancel and join any goroutine it starts; test cancellation and error
   paths. Use `errors.Is`/`errors.As` for wrapped failures rather than matching strings.
3. Scope reads and writes to the resolved tenant. Test another tenant's resource ID
   and a user without the required permission before accepting the happy path.
4. Edit queries in `go-b2b-starter/internal/db/postgres/sqlc/query/`, then run
   `make -C go-b2b-starter sqlc`. Inspect the generated diff, then run `sqlc-check`
   from the same Makefile. It compares two clean generations with the working tree,
   including added/deleted output files, without rewriting it. Generated code in
   `internal/db/postgres/sqlc/gen/` may be read for debugging; change SQL/config and
   regenerate instead of hand-editing the output. Add forward-only migrations
   for schema changes; test both a fresh database and preserved historical data.
5. Use a transaction for local changes that must commit together. External provider
   effects cannot share that transaction: define retry/reconciliation behavior and
   test partial failure. Return wrapped internal errors and safe public messages.
   Use `tenant.Store.Run` for business data and transaction-local tenant context;
   use its provided generated queries throughout the transaction. Do not send mail
   or charge a provider from a transaction-retry callback. Before retrying a mutation,
   define its idempotency key/uniqueness rule and test duplicate delivery plus an
   ambiguous timeout. Never retry every error or add a queue without a product need.
6. Add table-driven behavior tests at the changed boundary. Cover malformed input,
   provider/database failure, and duplicate delivery when the operation can repeat.
   Use HTTP fixtures for wire formats, not only mocks of internal interfaces.

From `go-b2b-starter/`, run `make fmt-check`, `make sqlc-check`,
`go test -race ./...` (or `make test-race`), and `go vet ./...` with the version in
`go.mod`. Formatting and SQLC checks use pinned Docker tool images. The existing
backend CI runs race tests, vet and vulnerability checks; the backend-source job
enforces formatting and deterministic generation. A passing source check does
not replace behavior tests.

Set `TEST_DATABASE_URL` only to a disposable PostgreSQL instance with database-
creation permission, then run `make test-db`. This target refuses an unset URL;
plain `go test` skips database tests without one. Database tests cover migration
preservation, RLS, absent tenant context, rollback and pooled connection reuse.
Docker-only alternatives `make test` and `make vet` do not claim race or migration
coverage. Deployment changes also need the setup/smoke procedure in `SETUP.md`
and `scripts/smoke.sh`.

Use deterministic synchronization rather than sleeps in concurrent unit tests.
`testing/synctest` became generally available in Go 1.25; its Go 1.24 experimental
API and `GOEXPERIMENT=synctest` examples are obsolete for this starter. Prefer
`synctest.Test`/`synctest.Wait` only when their isolated goroutine/time model fits;
use HTTP/database fixtures for real I/O. See the
[Go 1.25 release notes](https://go.dev/doc/go1.25#testing-synctest).

Agentic Ship's [identity and ownership principles](https://github.com/moasq/agentic-ship/blob/8fbf56f3d17defd0544d1d5ce328ab3c46851bf5/.agents/skills/convex-structure/SKILL.md)
and [behavioral test guidance](https://github.com/moasq/agentic-ship/blob/8fbf56f3d17defd0544d1d5ce328ab3c46851bf5/.agents/skills/testing/SKILL.md)
informed this procedure. Its Convex APIs, runtime, queues and repair loop are not
part of this Go workflow.

Return the endpoint/DTO changes needed by the frontend, authorization decisions,
test commands and outcomes, migration implications, and any live-provider checks
still missing. Do not label a fixture as a successful live provider operation.
