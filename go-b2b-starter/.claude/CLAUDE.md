# Backend guidance

Follow ../../AGENTS.md. Keep a small Go modular monolith, PostgreSQL/SQLC, Stytch
identity, optional Polar billing. Do not restore AI/document/RAG infrastructure.
Run go test -race ./... and go vet ./... after backend changes. Use a disposable
TEST_DATABASE_URL for migration integration checks. Provider failures must not
grant access. Never hand-edit SQLC output or silently erase user data.
