# Billing scope

Billing is optional. An organization administrator can subscribe to the single
fixed recurring product configured by `POLAR_PRODUCT_ID`, see its current status,
and open Polar's hosted customer portal for payments and cancellation. Workspace,
team and profile access do not require payment. See [setup](../SETUP.md) for the
server-only credentials and matching provider environment.

## Supported lifecycle

| State | Starter behavior |
| --- | --- |
| Billing disabled | Core workspace works; no billing controls or provider calls. |
| No active subscription, verified by Go | Offer checkout for the configured product only. |
| Active or trialing subscription | Display the current period and open the customer portal; do not start a replacement checkout. |
| Cancellation scheduled | Keep showing active until the period ends; display the cancellation and retain the portal link. |
| Cancellation reversed in Polar | Refresh status to read the provider's current cancellation flag. The starter does not perform this mutation. |
| Subscription ended | The next successful status read reflects the ended subscription; checkout can become available again. |
| Provider unavailable or access revoked | Do not interpret missing status as an empty subscription or allow checkout. |

Use **Refresh status** after returning from the portal. Billing reads come from
Polar through Go; the browser may retain its last status until refreshed. A
successful checkout redirect is checked for organization and product ownership
and followed by a new provider read. It does not itself grant application access.
The hosted [customer portal](https://polar.sh/docs/documentation/features/customer-portal)
and [cancellation API](https://polar.sh/docs/api-reference/customer-portal/subscriptions/cancel)
are provider features; their availability still depends on the account setup.

## Deliberate limits

This starter does not implement plan switching, upgrades, scheduled downgrades,
proration, seat or usage pricing, quotas, or paid-feature entitlements. There is no
webhook inbox or local subscription replica. Do not enable portal product changes
or change `POLAR_PRODUCT_ID` for existing subscribers as a substitute for a tested
product migration: only the configured product is recognized. Keep the provider
portal configured for payments and cancellation until extending this lifecycle.

The server rechecks status before checkout and the browser disables the pending
button, but concurrent tabs or separate administrators can still open multiple
unpaid checkouts before Polar reports an active subscription. This is not an
atomic reservation or a guarantee of exactly one subscription. For a product that
requires that guarantee, implement and test a provider-compatible idempotency and
reconciliation design before accepting payments. Do not automatically retry a
checkout mutation after an ambiguous network failure.

## Verification boundary

Frontend regressions cover unavailable, revoked, disabled, active and scheduled
cancellation states. Run `BROWSER_BILLING_FIXTURE=true ./scripts/test-browser.sh`
for production-rendered billing UI at 390px and 1440px with both OS color preferences.
This opt-in mode uses an isolated network without internet access and intercepts
only billing server actions in Playwright. It exercises real local signup and
sessions, but the subscription, checkout and portal responses are explicit UI
fixtures. The default browser run still covers the billing-disabled starter.
Go behavior tests cover current-provider reads, cancellation
reversal, expiry, ownership and repeated checkout verification. These are local
fixtures, not live billing evidence. Before enabling payments, separately verify
the configured product, checkout, portal cancellation, reversal and refreshed
status with synthetic customers in Polar sandbox. Live payments and external
email delivery remain separate deployment checks; see
[provider verification](PROVIDER_VERIFICATION.md).
