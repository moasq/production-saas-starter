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

Generate SQLC with `make sqlc` after editing queries or schema input. Inspect the
generated diff; never edit `internal/db/postgres/sqlc/gen/` by hand.

| Check | Command | Prerequisites |
| --- | --- | --- |
| Formatting (read-only) | `make fmt-check` | Docker; pinned Go image from `go.mod` |
| Generated SQL matches source and two fresh runs agree | `make sqlc-check` | Docker; pinned SQLC image in `Makefile` |
| Race-enabled behavior tests | `make test-race` | Go version from `go.mod`, C compiler |
| Static analysis | `go vet ./...` or `make vet` | Host Go or Docker respectively |
| Required database behavior tests | `make test-db` | Host Go/C compiler and `TEST_DATABASE_URL` |

`sqlc-check` generates in a temporary directory and compares the full output tree;
it neither rewrites local changes nor requires a clean Git checkout. Format
reported authored Go files with `gofmt -w path/to/file.go`. Both checks also run in
the **Backend source contracts** CI workflow; race, vet and vulnerability checks
remain in the main backend CI job.

`TEST_DATABASE_URL` must identify a **disposable** PostgreSQL superuser instance;
tests create temporary databases and restricted roles. `make test-db` refuses a
missing URL, while plain `go test` skips those integration cases without one. The
Docker-only `make test` does not run the race detector or pass database credentials.
These tests use synthetic data and local HTTP fixtures, not live provider accounts.
