# Go project

Read the shared rules in `../AGENTS.md`, `../README.md` and
`../docs/ARCHITECTURE.md`. This project owns business endpoints, tenant
authorization, SQLC, forward migrations and the business API contract.

Use `.agents/skills/go-backend/SKILL.md` and
`.agents/agents/backend-builder.md` inside this directory. Go tests, database
checks and SQLC generation belong here. Follow the skill's repository-relative
commands; from this directory, `make test`, `make vet` and `make sqlc-check`
are the corresponding local targets. Database tests require a disposable database.

Identity, sessions and organization membership belong to Better Auth in Next.js.
For that boundary, consult
`../next_b2b_starter/.agents/skills/auth-integration/SKILL.md`; hand changes to
its owner rather than duplicating the auth policy here. Hand the tested Go-owned
API schema and permission/error contract to the frontend owner before parallel UI
work. Never hand-edit SQLC or TypeScript generated output.

Run `node ../scripts/harness.mjs check` for tooling changes. Canonical Go guidance
lives here; generated `.claude/` and `.codex/` files are adapters. Root guidance
coordinates shared work and does not replace these project instructions.
