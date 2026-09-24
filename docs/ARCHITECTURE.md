# Architecture

```text
Browser -> Caddy -> Next.js (pages, server actions, HTTP-only session cookies)
                -> Go API (organization authorization and application logic)
                       -> PostgreSQL (organizations and accounts)
Next.js + Go -> Stytch B2B (identity and organization membership)
Next.js + Go -> Polar (optional checkout, portal, current subscription state)
```

Go remains a modular monolith. Next.js manages the browser/session boundary and
calls the API with the verified session. Every protected API call resolves the
organization from that identity; a submitted organization or checkout ID is not
authorization. Member changes and billing management require organization
management permission.

The starter intentionally has one durable datastore. Provider metadata can use
bounded process-local caching, while identity and membership remain authoritative
at Stytch. PostgreSQL uses generated SQLC queries. Database migrations run before
readiness so an image cannot report ready against a missing schema.

Billing is a thin, optional integration. Polar's customer external ID is the
Stytch organization ID. Go reads customer state instead of maintaining a webhook
replica. Checkout verification validates ownership and makes no state changes.
This avoids replay-induced grants, stale local subscription flags, quota resets,
and a required background worker. The tradeoff is provider availability for
billing requests. The core app is not gated by a billing subscription.

Caddy is the deployment entry point. Only it publishes host ports. TLS terminates
there; backend configuration explicitly acknowledges proxy termination. Credentials
and the public application URL are runtime configuration. Locked packages and
pinned multi-architecture base images make builds reproducible.

No document/AI domain is built in. Add your B2B product's domain behind the same
organization boundary and give it behavioral authorization tests.
