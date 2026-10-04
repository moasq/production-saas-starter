# Web application

Next.js renders the landing page, sign-in flow, workspace, profile, team and optional billing settings. Go owns tenant data and permissions.

Use the root [setup guide](../SETUP.md) for Docker startup and provider configuration. For host development:

```sh
cp .env.example .env.local
pnpm install --frozen-lockfile
pnpm dev
```

Run `pnpm verify` before changing the application. It checks ESLint, TypeScript, regression tests and the production build.

Authentication uses self-hosted Better Auth 1.7.6, its organization and magic-link plugins, PostgreSQL, and SMTP. Browser API calls use HttpOnly cookies; server calls forward the current cookie to `API_BASE_URL_INTERNAL`. Go verifies live session and membership through the secret-protected internal auth bridge on every protected operation. No session JWT or cookie cache delays revocation. Sessions have a fixed eight-hour lifetime: database reads and internal bridge calls never extend expiry; sign in again with a magic link after expiry. Local SMTP uses Mailpit; production SMTP requires separate delivery verification.

Apply the committed auth schema with `node scripts/migrate-auth.mjs`. The optional legacy importer defaults to dry-run and requires a reviewed active-membership snapshot before `--apply`; see the root upgrade guide. Never put owner database credentials in the frontend runtime.

Billing is disabled by default. When enabled, configure `POLAR_ENVIRONMENT`, `POLAR_ACCESS_TOKEN` and one fixed recurring `POLAR_PRODUCT_ID`. Checkout binds the authenticated workspace as the external customer ID. Go reads current Polar state and verifies checkout ownership. The hosted customer portal handles payments and cancellation; this starter does not implement plan switching or paid-feature entitlements. No webhook receiver or local billing replica is required. See the [supported billing lifecycle and limits](../docs/BILLING.md).

Packages are pinned in `package.json` and `pnpm-lock.yaml`. Tailwind 3.4 and tailwind-merge 2.6 remain a compatible pair. TypeScript 5.9 is the tested compiler; the locked parser accepts TypeScript below 6.1. ESLint 9 is retained temporarily because the React and accessibility plugins exclude ESLint 10, but ESLint 9 reached upstream end of life on August 6, 2026. See the [browser contract, coordinated migration plan and lint follow-up](../docs/decisions/0002-frontend-tooling.md) before updating these tools.

The dashboard keeps sidebar state in React and formats dates with the platform API. There is no global UI store or query devtools dependency. Polar requests pin API version `2026-04`, matching the Go client and the installed SDK.
