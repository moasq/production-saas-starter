# Optional billing

Enable with `BILLING_ENABLED=true`, a server-only `POLAR_ACCESS_TOKEN`, a configured `POLAR_PRODUCT_ID`, and an explicit
`POLAR_ENVIRONMENT=sandbox` or `production`. Both applications must use the same
provider environment. The frontend creates checkout and customer portal sessions
using the authenticated organization as the external customer ID.

The Go API resolves the caller's organization and reads Polar's customer state.
It does not maintain a second subscription database, process webhooks, meter usage,
or gate unrelated starter pages. An unavailable provider returns an error instead
of treating a customer as free or paid. Checkout verification checks ownership
and performs no writes, so repeated redirects cannot reset quotas.

This deliberately small integration requires Polar availability for billing reads.
For application-specific entitlements, add a separately tested authorization rule.
If your application needs offline billing reads or asynchronous fulfillment, add a
verified, durable, idempotent webhook inbox and reconciliation at that point.

API references: https://polar.sh/docs/api-reference/customers/state-external and
https://polar.sh/docs/api-reference/checkouts/get-session.
