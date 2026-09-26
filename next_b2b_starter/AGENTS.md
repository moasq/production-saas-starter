# Next.js project

Read shared rules in `../AGENTS.md`, `../README.md`, this project's `README.md`
and `../docs/ARCHITECTURE.md`. This project owns presentation, the HttpOnly session
boundary, Better Auth and its private bridge; Go remains the authority for
business permissions and tenant data.

Use these project-local procedures and roles:

- `.agents/skills/next-frontend/SKILL.md` and `.agents/agents/frontend-builder.md`
  for routes, components, API consumers and accessibility.
- `.agents/skills/auth-integration/SKILL.md` and `.agents/agents/auth-reviewer.md`
  for Better Auth, sessions, memberships and a read-only trust-boundary review.
- `.agents/skills/frontend-tools/SKILL.md` and `.agents/agents/quality-engineer.md`
  for local diagnostics and synthetic browser journeys.

Run `pnpm verify` here for frontend behavior, lint, typecheck and build. Generated
business types come from the Go-owned contract; request schema/authorization
changes from the backend owner. For browser checks, use the disposable runner
at `../scripts/test-browser.sh`; preserve existing stacks and real accounts.
Local auth checks do not establish external email delivery or live Polar billing.

Run `node ../scripts/harness.mjs check` for tooling changes. Canonical frontend
and Better Auth guidance lives here. Root only coordinates cross-project work;
generated `.claude/` and `.codex/` files are adapters, not instruction sources.
