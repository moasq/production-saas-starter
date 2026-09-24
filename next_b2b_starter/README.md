# Web application

Next.js renders the landing page, sign-in flow, workspace, profile, team and optional billing settings. Go owns tenant data and permissions.

Use the root [setup guide](../SETUP.md) for Docker startup and provider configuration. For host development:

```sh
cp .env.example .env.local
pnpm install --frozen-lockfile
pnpm dev
```

Run `pnpm verify` before changing the application. It checks ESLint, TypeScript, regression tests and the production build.

Authentication uses the server Stytch SDK and HttpOnly cookies. Browser API calls stay on `/api`; server calls supply a verified bearer token to `API_BASE_URL_INTERNAL`. No browser SDK, public token or provider secrets are needed at build time. The Go API checks permissions on every protected operation.

Billing is disabled by default. When enabled, configure `POLAR_ENVIRONMENT`, `POLAR_ACCESS_TOKEN` and one fixed recurring `POLAR_PRODUCT_ID`. Checkout binds the authenticated workspace as the external customer ID. Go reads current Polar state and verifies checkout ownership. The hosted customer portal handles subscription changes. No webhook receiver or local billing replica is required.

Runtime packages are pinned. Next.js 16.3.6 and React 19.3.0 are the latest stable registry releases verified on September 25, 2026. Tailwind 3.4 and tailwind-merge 2.6 are a compatible pair. TypeScript 5.9 and ESLint 9 remain on the versions supported by Next's lint plugins: the React and accessibility plugins exclude ESLint 10, and the TypeScript parser currently requires TypeScript below 6.1.

The dashboard keeps sidebar state in React and formats dates with the platform API. There is no global UI store or query devtools dependency. Polar requests pin API version `2026-04`, matching the Go client and the installed SDK.
