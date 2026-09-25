# Optional billing

Enable with `BILLING_ENABLED=true`, a server-only `POLAR_ACCESS_TOKEN`, a configured `POLAR_PRODUCT_ID`, and an explicit
`POLAR_ENVIRONMENT=sandbox` or `production`. Both applications must use the same
provider environment. The frontend creates checkout and customer portal sessions
using the authenticated organization as the external customer ID. Existing companies
retain their original ID in `polar_customer_external_id`; Go reads that immutable
mapping under tenant RLS instead of deriving a new ID during auth migration.
Leave `BILLING_ENABLED` unset or `false` to run without any Polar configuration.

The Go API resolves the caller's organization and reads Polar's customer state.
It does not maintain a second subscription database, process webhooks, meter usage,
or gate unrelated starter pages. An unavailable provider returns an error instead
of treating a customer as free or paid. Checkout verification checks ownership
and performs no writes, so repeated redirects cannot reset quotas.
Only the configured product counts as this application's subscription. Checkout
ownership comes from Polar's `external_customer_id`; there is no client-supplied
tenant override or legacy nested-customer fallback. Missing or malformed provider
fields return an error. A documented missing customer returns an empty status.

This deliberately small integration requires Polar availability for billing reads.
For application-specific entitlements, add a separately tested authorization rule.
If your application needs offline billing reads or asynchronous fulfillment, add a
verified, durable, idempotent webhook inbox and reconciliation at that point.

The HTTP client pins `Polar-Version: 2026-04`, verified against Polar's
[current version declaration](https://github.com/polarsource/polar/blob/main/server/polar/version.py).
Review the provider's release notes and re-run the HTTP contract tests before
changing it. Tests use local HTTP fixtures; they do not prove live payments.

Contracts: [customer state](https://polar.sh/docs/api-reference/2026-04/customers/get-customer-state-by-external-id),
[checkout](https://polar.sh/docs/api-reference/2026-04/checkouts/get-checkout-session),
[API versioning](https://polar.sh/docs/api-reference/2026-04/versioning).
