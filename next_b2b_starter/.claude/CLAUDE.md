# Frontend maintenance

Read the root AGENTS.md and SETUP.md. Keep this application small: landing, self-hosted Better Auth organization auth, profile, team, optional Polar billing. Product features belong in explicit new modules.

Use pnpm and pinned dependencies. Run `pnpm verify`. Never edit generated `.next` output or hand-edit pnpm-lock.yaml.

Go owns tenant data and authorization. Server API calls forward the current request cookie and origin explicitly; never cache session credentials globally. Go re-verifies current session and membership through the protected internal bridge. Browser API calls use same-origin HttpOnly cookies. Do not copy secrets into public environment variables or browser state. Canonical roles live in lib/auth/rbac.ts. The organization plugin applies that policy and the backend checks its returned permissions again on every operation. Keep auth database credentials separate from the business database role.

Billing is optional. Scope checkout and customer portal sessions to the authenticated organization and configured product. Treat provider failures as unavailable, never as an active subscription. Do not introduce a second billing replica or quotas unless a product actually requires them.
