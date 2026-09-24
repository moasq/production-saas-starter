# Working on this starter

Read README.md, docs/ARCHITECTURE.md and the relevant module before changing code.
The product is a minimal deployable B2B starter: organization auth, members,
profile and optional billing. Keep Go + Next.js + PostgreSQL. Do not add AI,
RAG, document processing, Redis, agents or a new backend framework by default.

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
