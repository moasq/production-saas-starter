# Provider SDK contracts

Reviewed October 4, 2026 for issue [#33](https://github.com/moasq/production-saas-starter/issues/33).
The original issue named Stytch and a separate Polar Next.js adapter. Both were
removed during the lean starter migration. The current runtime uses Better Auth
inside Next.js, the Polar TypeScript SDK for three server operations, and a small
Go HTTP client for billing reads. No provider framework or webhook replica is added.

| Boundary | Selected contract | Evidence |
| --- | --- | --- |
| Better Auth | `better-auth` 1.7.7 and aligned core/adapters | [Release and migration notes](https://github.com/better-auth/better-auth/releases/tag/v1.7.7) |
| Polar writes/product read | `@polar-sh/sdk` 1.0.2, explicit `@polar-sh/sdk/2026-04` import | [Published package](https://registry.npmjs.org/@polar-sh/sdk/1.0.2), [maintained SDK](https://github.com/polarsource/polar/tree/main/sdk/typescript) |
| Polar Go reads | `Polar-Version: 2026-04` | `internal/platform/polar/client.go`, checkout/customer-state HTTP fixtures |
| Authorization | Live Better Auth membership and explicit Go permissions | [Authorization contract](AUTHORIZATION.md), auth/tenant suite |

These were current package releases at review time, not a promise that the pins
will remain latest. API version and package version are separate decisions.

## Better Auth 1.7.7 upgrade

The [critical upstream advisory](https://github.com/better-auth/better-auth/security/advisories/GHSA-965c-763c-88jm)
concerns Magic Link combined with social or Generic OAuth sign-in, database-backed
OAuth state, and the documented overlapping verification-storage configuration.
This starter has no social/Generic OAuth provider, hashes magic-link tokens and
does not configure global identifier hashing. The advisory does **not** establish
account takeover in this default configuration. Upgrade nevertheless before
extending authentication; the release also repairs active-organization refresh
and Kysely verification-consumption behavior.

Deploy all servers sharing verification storage together. Request new magic
links after the cutover; links issued before the upgrade no longer complete.
Deployments that added OAuth/SAML must also restart pending sign-in/account-linking
flows, and OAuth Proxy users must coordinate every participant. Custom verification
storage must account for the new purpose prefixes. The provider requires no
database schema or user/account migration. Keep existing users, memberships,
tenant IDs and billing customer mappings; do not delete verification tables or
rewrite applied migrations as an upgrade shortcut.

Follow [the upgrade guide](UPGRADING.md) for backups and isolated validation.
Rerun the self-hosted auth/tenant journeys after deployment. Local Mailpit captures
prove the local flow; they do not establish real external SMTP delivery. Do not
mix old/new auth nodes as a rollback strategy.

## Polar 1.x migration

The [old SDK repository](https://github.com/polarsource/polar-js) is archived.
SDK 1.x uses `createPolar`, versioned imports and snake_case wire fields instead
of the former `Polar` class/camelCase mapping. The starter keeps the `2026-04`
API in both applications rather than implicitly selecting the SDK's newer example
version. No Polar customer identifiers or database schemas change.

`next_b2b_starter/lib/polar/provider.ts` maps the three used operations into local
application values. Server actions still resolve membership/permissions through
Go before obtaining a billing client. Checkout and portal requests use the
verified workspace's external customer ID; they never send that tenant ID as the
merchant's `Polar-Organization` header. The server-only configuration entry point
keeps the token out of client components and billing remains disabled by default.

The new SDK returns JSON without runtime schema validation. The adapter validates
the fields consumed by the UI and requires an HTTPS hosted URL. Only one available
fixed recurring price, charged every single interval, is offered. Archived,
multi-price, metered and multi-interval products require an explicit product
extension, not an inaccurate price label.

Every SDK call has a ten-second timeout (the 1.x option is in **seconds**). SDK 1.x
performs one fetch per operation and the adapter adds no automatic retries.
Regression cases count attempts after 404, 429, 500, ambiguous network failure
and abort; a checkout timeout must not create a second request. This bounds a
request, not the entire server-action sequence. The Go read client separately
uses its existing ten-second timeout and request context.

## Verification and limits

`pnpm test` exercises the installed SDK with synthetic HTTP responses for product,
checkout and customer-session success, both selected environments, API/auth
headers, external identity binding, malformed successful responses, unsupported
pricing, failure handling and timeout propagation. No test contacts either Polar
environment. Existing Go fixtures exercise customer state, cancellation reversal,
expiry, checkout ownership and payment status. Frontend build/type checks verify
the server callers compile against the new adapter.

Live sandbox checkout, portal cancellation/reversal, payment and external SMTP
remain separate checks in [provider verification](PROVIDER_VERIFICATION.md).
There is no webhook endpoint in this starter, so webhook signature/event migration
is not claimed. Only the current [billing lifecycle](BILLING.md) is supported.
