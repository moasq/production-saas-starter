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

Runtime packages are pinned. Next.js 16.3.6 and React 19.3.0 are the latest stable registry releases verified on September 25, 2026. Tailwind 3.4 and tailwind-merge 2.6 are a compatible pair. TypeScript 5.9 and ESLint 9 remain on the versions supported by Next's lint plugins: the React and accessibility plugins exclude ESLint 10, and the TypeScript parser currently requires TypeScript below 6.1.

The dashboard keeps sidebar state in React and formats dates with the platform API. There is no global UI store or query devtools dependency. Polar requests pin API version `2026-04`, matching the Go client and the installed SDK.
