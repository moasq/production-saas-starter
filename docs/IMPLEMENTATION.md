# Minimal B2B starter

The revival targets developers deploying a small B2B SaaS. Keep Go, Next.js,
PostgreSQL, Stytch organization authentication, member management, and optional
Polar billing. Remove document upload, OCR, AI chat, RAG, embeddings, object
storage, and their infrastructure/dependencies from the application.

## Boundaries

- Go owns tenant authorization, members, persistence, and billing status.
- Next.js owns the interface and the existing session boundary.
- PostgreSQL is the only required stateful service. Authentication caches should
  be process-local; do not require Redis merely to cache provider metadata.
- Billing is opt-in. Missing billing credentials must not stop the core app.
- Missing authentication configuration shows setup instructions, never fake
  authenticated access. Production authentication must fail closed.
- Keep the modular monolith; remove unused abstractions and services instead of
  introducing a framework or maintaining multiple frontend/backend editions.

## Ownership and handoff

- Backend work owns Go except billing/Polar internals and Docker/deployment files:
  prune removed features, database/SQLC, auth/organizations, supported Go/modules,
  meaningful tests. Preserve existing database data and document migration limits.
- Frontend work owns Next.js except its Dockerfile: remove feature surfaces,
  simplify dashboard/navigation, optional billing, configuration state, dependency
  updates, lint/type/build checks. Read current subscription state from Go; keep billing optional.
- Integration work owns billing/Polar, deployment files, root setup/docs/CI,
  final review and real local container/browser verification.

## Completion checks

- Removed feature code, routes, provider secrets and package dependencies are gone.
- Core authorization, configuration, migrations and billing ownership/status have
  behavioral tests; frontend lint, type checking and build pass.
- Fresh container startup needs no local Go/Node and no AI/storage account.
- Production configuration uses runtime secrets and compatible architectures.
- A real local deployment and browser smoke check pass; provider-backed live
  payments/email remain explicitly unverified unless exercised with real accounts.
- Dependency scans distinguish actionable findings, unreachable findings and
  unavailable evidence. No obsolete production-ready or AI feature claims remain.

No deployment to an external host or changes to existing user data are required
for this implementation. Historical audit reports remain local evidence.
