# Billing environment contract

Billing is optional. `BILLING_ENABLED` accepts `true`, `false`, or an omitted/blank
value (disabled). Go and Next.js trim surrounding whitespace and reject any other
value. Disabled billing ignores unused Polar settings and makes no provider call.

When enabled, both applications require an explicit `POLAR_ENVIRONMENT` of
`sandbox` or `production`, plus nonblank `POLAR_ACCESS_TOKEN` and `POLAR_PRODUCT_ID`.
Compose supplies the same runtime values to both applications; `.env.example`
starts with billing disabled and sandbox selected. Direct deployments must set
the environment explicitly. An omitted environment is no longer silently sandbox.

| Deployment | Build mode | Polar environment |
| --- | --- | --- |
| Local testing | Any | `sandbox` |
| Production-built staging | `NODE_ENV=production` | `sandbox` |
| Real payments | `NODE_ENV=production` | `production` |

Build mode never selects the provider account. Both clients use fixed Polar API
origins for the selected environment; no arbitrary API URL is accepted from an
environment variable. Frontend SDK clients read the current validated runtime
configuration instead of retaining a client for an earlier environment/token.
Secrets are not public build variables.

Tokens and product IDs are opaque. Local validation cannot establish that they
belong to the same provider account/environment. A product read must succeed in
the selected environment before checkout can be created; a rejected token or
missing/incompatible product fails closed. There is no fallback to the other
environment. No webhook secret is configured because this starter reads Polar
state on demand and has no webhook consumer or replicated entitlement state.

Use [provider verification](PROVIDER_VERIFICATION.md) for the opt-in, read-only
sandbox product probe and payment/portal checklist. The probe refuses production.
Unit fixtures establish configuration and routing behavior, not live account
compatibility, payment completion, or production readiness. Switching environments
on an existing deployment also changes which provider owns saved customer IDs;
use separate databases for staging and production rather than switching a live
tenant database between accounts.

Go and Next.js execute the same synthetic configuration cases in
`go-b2b-starter/internal/platform/polar/testdata/config.json`.
Polar documents the explicit SDK server selector in its
[TypeScript SDK guide](https://docs.polar.sh/documentation/sdks/typescript-sdk).
