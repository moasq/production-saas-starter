---
name: "go-backend"
description: "This skill applies when adding a Go API endpoint, changing SQLC queries or migrations, fixing tenant authorization, or testing backend behavior in this starter."
---

# Go backend workflow

Read `AGENTS.md` and `docs/ARCHITECTURE.md` first. Locate the nearest implementation
in `go-b2b-starter/internal/modules/` and follow its handler, service, and repository
boundary. Prefer an existing interface over a new abstraction with one caller.

1. State the actor, tenant, input, response shape, and failure behavior. Trace the
   route through `internal/modules/auth/` to the organization/account resolver.
   For identity or membership changes, also use `.agents/skills/auth-integration/SKILL.md`.
2. Keep HTTP decoding/status translation in handlers and application decisions in
   services. Pass request context to SQL/provider calls, bound outbound timeouts,
   and close resources. Start no detached goroutine without cancellation and an owner.
3. Scope reads and writes to the resolved tenant. Test another tenant's resource ID
   and a user without the required permission before accepting the happy path.
4. Edit queries in `go-b2b-starter/internal/db/postgres/sqlc/query/`, then run
   `make -C go-b2b-starter sqlc`. Inspect the generated diff. Run it again and verify
   the second generation produces no further changes. Add forward-only migrations
   for schema changes; test both a fresh database and preserved historical data.
5. Use a transaction for local changes that must commit together. External provider
   effects cannot share that transaction: define retry/reconciliation behavior and
   test partial failure. Return wrapped internal errors and safe public messages.
6. Add table-driven behavior tests at the changed boundary. Cover malformed input,
   provider/database failure, and duplicate delivery when the operation can repeat.
   Use HTTP fixtures for wire formats, not only mocks of internal interfaces.

From `go-b2b-starter/`, run `gofmt` on changed Go files, `go test -race ./...`, and
`go vet ./...` with the version in `go.mod`. Set `TEST_DATABASE_URL` only to a
disposable PostgreSQL instance with database-creation permission for migration tests;
an unset variable skips those tests. Docker-only alternatives are `make test` and
`make vet`; those do not claim race or migration coverage. Deployment changes also
need the setup/smoke procedure in `SETUP.md` and `scripts/smoke.sh`.

Return the endpoint/DTO changes needed by the frontend, authorization decisions,
test commands and outcomes, migration implications, and any live-provider checks
still missing. Do not label a fixture as a successful live provider operation.
