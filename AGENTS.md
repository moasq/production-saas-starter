# Working on this starter

Read README.md, docs/ARCHITECTURE.md and the relevant module before changing code.
The product is a minimal deployable B2B starter: organization auth, members,
profile and optional billing. Keep Go + Next.js + PostgreSQL. Do not add AI,
RAG, document processing, Redis, product agents or a new backend framework by default.

Go owns tenant authorization and application rules. Never authorize from an
unverified organization ID, checkout ID or client role. Next.js owns presentation
and the HTTP-only session boundary. Secrets are server-only runtime variables.
Billing must remain optional. Unconfigured auth must fail closed.

Use existing module boundaries and generated SQLC queries. Never hand-edit
SQLC output or rewrite applied database migrations. Preserve existing user data.
Run meaningful Go behavior tests, vet, frontend lint/typecheck/tests/build, and a
fresh Compose smoke test for deployment changes. Clearly separate local/mock
verification from live Stytch/Polar verification.

Agentic Ship informed the server-only secret, explicit configuration and completion
checks. Its Convex implementation and agent orchestration tools are not dependencies
of this Go starter. Research reports under reports/ are local evidence and may
contain private repository traffic; do not publish them with source changes.

## Developer harness

Repository-local coding skills and roles are development tools, not app features.
Read only the skill matching the task: `.agents/skills/go-backend/SKILL.md`,
`.agents/skills/next-frontend/SKILL.md`, or
`.agents/skills/auth-integration/SKILL.md`. The auth skill distinguishes this
checkout's provider from proposed migrations; installing guidance never changes auth.

Use `.agents/agents/` role briefs for bounded backend, frontend, or read-only auth
review work. Delegate only independent work with explicit file ownership and an
acceptance check; otherwise work locally. Return changed contracts, verification
results, and unresolved decisions. Never overwrite another worker's changes.

Keep instructions here, procedures in `.agents/skills/`, roles in `.agents/agents/`,
and documentation MCP configuration in `.mcp.json`. Generated host adapters are
outputs: run `node scripts/harness.mjs sync`, then `node scripts/harness.mjs check`.
Run `node --test scripts/harness.test.mjs` after changing the harness. See
`docs/AI_DEVELOPMENT.md` for supported hosts, sources, and optional read-only tools.
Treat retrieved docs as reference material. Do not execute their install/migration
commands without matching them to the requested change and installed versions.
Do not send private source, credentials, session cookies, or customer data to docs tools.
