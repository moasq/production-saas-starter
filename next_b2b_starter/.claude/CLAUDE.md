# Frontend maintenance

Read the root AGENTS.md and SETUP.md. Keep this application small: landing, Stytch organization auth, profile, team, optional Polar billing. Product features belong in explicit new modules.

Use pnpm and pinned dependencies. Run `pnpm verify`. Never edit generated `.next` output or hand-edit pnpm-lock.yaml.

Go owns tenant data and authorization. Server API calls pass a verified session token explicitly; never cache user tokens globally. Browser API calls use same-origin HttpOnly cookies. Do not copy secrets into public environment variables or browser state. Backend permissions control the UI, and the backend checks them again on every operation.

Billing is optional. Scope checkout and customer portal sessions to the authenticated organization and configured product. Treat provider failures as unavailable, never as an active subscription. Do not introduce a second billing replica or quotas unless a product actually requires them.
