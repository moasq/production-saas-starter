# Development

The easiest environment is `./setup.sh`. Use `docker compose up --build -d` after
source changes and `docker compose logs -f backend frontend` for logs.

For local hot reload, run PostgreSQL through Compose and use Go 1.27.1 and Node
24 LTS with pnpm 10.32.1 on your machine. Copy `go-b2b-starter/example.env` to
`go-b2b-starter/app.env`, and `next_b2b_starter/.env.example` to
`next_b2b_starter/.env.local`. Use a local PostgreSQL connection (the default
Compose database does not publish a host port; add a private local override).
Run `go run ./cmd/api` in the backend and `pnpm dev` in the frontend.

## Verification

```sh
cd go-b2b-starter
go test -race ./...
go vet ./...
cd ../next_b2b_starter
pnpm install --frozen-lockfile
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

Use `make test`/`make vet` in the backend when Go is not installed; they run the
pinned Go container. CI also builds and starts the entire Compose deployment from
a fresh database with billing disabled and without auth credentials, then checks
readiness and the public setup pages. Real Stytch emails and Polar payments need
a separate provider-backed verification run.

## Database changes

SQL source and generated bindings live in `internal/db/postgres/sqlc`. Add a
forward migration and update queries/schema, then run `make sqlc` from the backend.
Do not edit generated Go files by hand or rewrite applied migrations. Test a fresh
install and an upgrade before shipping schema changes. Migrations are embedded
into the API image and applied before it serves traffic.

## Scope

Add product features as a small module under `internal/modules` with handlers,
application logic and persistence queries. Prefer the existing conventions to a
new framework. Keep authentication/tenant authorization in Go, secrets and session
cookies on the server, and the interface in Next.js. Do not restore removed
infrastructure merely because a future feature might need it.
