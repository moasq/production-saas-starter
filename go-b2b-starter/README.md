# Go API

Go owns business APIs, tenant authorization and optional Polar status. The existing
Next.js server runs self-hosted Better Auth; each protected Go request verifies the
opaque session cookie and current membership through its private HTTP bridge.
There is no session/JWT cache or external identity account requirement.

Use the root Compose setup. For host development configure `example.env` as
`app.env`, including `AUTH_BRIDGE_URL`, `AUTH_INTERNAL_SECRET` (32+ characters) and
the restricted PostgreSQL runtime credentials. The bridge and Next.js must share
the same secret. Browser writes require the configured `APP_BASE_URL` origin.

Run `go run ./cmd/api migrate` once with migration credentials. Run the API with
`go run ./cmd/api` using a separate nonowner, NOSUPERUSER, NOBYPASSRLS database role.
Startup validates schema version and forced tenant row policies. Never give the
serving process migration credentials. Billing remains optional.

Generate SQLC with `make sqlc`. Run `go test -race ./...` and `go vet ./...`.
`TEST_DATABASE_URL` enables integration tests and must identify a disposable
PostgreSQL superuser instance; tests create temporary databases and restricted roles.
