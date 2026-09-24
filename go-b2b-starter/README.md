# Go API

The API owns organization authorization, member/profile operations and optional
Polar subscription status. PostgreSQL stores organization/account projections;
Stytch remains authoritative for identity and membership.

See [root setup](../SETUP.md), [development](../DEVELOPMENT.md) and
[architecture](../docs/ARCHITECTURE.md). Docker is the shortest path. For host
development use Go 1.27.1, copy `example.env` to `app.env`, configure PostgreSQL
and run `go run ./cmd/api`. The API can start with auth unconfigured, but protected
requests fail closed. Billing credentials are required only when explicitly enabled.

Migrations are embedded and run before serving. Generate query bindings with
`make sqlc`; verify behavior with `go test -race ./...` and `go vet ./...`.
Set `TEST_DATABASE_URL` to a disposable PostgreSQL instance with database-creation
permission to exercise migration integration tests. Do not point tests at customer
or production databases.
