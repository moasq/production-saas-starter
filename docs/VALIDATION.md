# Verification — 25 September 2026

Implementation is on `codex/minimal-b2b-starter`. No external deployment or live
provider account changes were made. The local Docker application is available at
http://localhost:3000. Authentication intentionally shows configuration instructions
until Stytch credentials and its RBAC policy are configured.

## Passed

- Go 1.27.1: full `go test -race ./...` and `go vet ./...`.
- Real PostgreSQL 17 integration: fresh schema, clean legacy ledgers 1–9,
  preserved organization/account data, repeated startup, dirty-ledger refusal.
  All 18 historical migration files match the original repository bytes.
- Backend behavior: provider-derived permissions, configuration from environment,
  cookie Origin checks, profile/member routes, invitation delivery reporting,
  delete/reinvite recovery, protected organization provider linkage.
- Billing behavior: disabled mode, cross-tenant and wrong-product rejection,
  expired/canceled state, provider errors, repeat checkout verification, and
  HTTP JSON fixtures using Polar's documented `external_customer_id` field.
- Frontend: ESLint, TypeScript, eight behavioral regressions, production build.
  The regressions cover request token isolation, mutation retry behavior,
  invitation responses, redirects, and no-content API responses.
- Fresh Compose build/start without provider credentials or host Go/Node;
  frontend/API health, public pages, and unauthenticated API rejection.
- Database outage returns API readiness 503; restoring PostgreSQL restores 200.
- API production mode starts correctly with TLS terminated by the private proxy.
- Frontend image built without provider credentials renders the sign-in form when
  runtime-only test configuration is supplied. No provider requests were submitted.
- Browser: landing, workspace setup state, and unauthenticated dashboard redirect;
  no browser errors observed on those paths. System font stack corrected after
  visual inspection. Authenticated pages are covered by code/contracts, not a
  live authenticated browser session.
- Go and frontend containers run as non-root users. Public provider secret build
  arguments are removed; local audit reports are ignored by Git.

## Dependency evidence

Retained production packages and compatible transitive dependencies were refreshed.
Go is 1.27.1; Docker Node is 24.21.0; Next.js is 16.3.6; React is 19.3.0;
Stytch Go is 18.1.0 and Node 14.2.0; Polar Node is 0.49.0. Base images use
version tags plus verified multi-architecture digests.

`pnpm audit --prod`: zero known vulnerabilities across all severities.
`govulncheck` 1.8.0: zero vulnerabilities in called symbols or imported packages.
It reports GO-2026-5932 in the unimported `golang.org/x/crypto/openpgp` package
inside a required module; there is no upstream fixed version. The finding has
not been suppressed. An initial run with the obsolete scanner 1.1.4 could not
parse Go 1.27; the final result uses the supported newer scanner.

Tailwind 3.4.19 plus tailwind-merge 2.6.1, TypeScript 5.9.3 and ESLint 9.39.4
remain intentionally compatible. ESLint 10 conflicts with the current Next.js
plugin peer range; Tailwind 4 and TypeScript 7 are separate migrations, not
requirements for removing the obsolete product features. PostgreSQL remains on
supported 17.11 to avoid an implicit database-major upgrade.

## Not verified

Live Stytch email/sign-in/invitation delivery, live Polar payment/portal lifecycle,
public DNS and certificate issuance, remote CI execution, and a restoration of an
actual user's production database. Existing pgvector databases require the
explicit upgrade guidance; legacy data is never automatically dropped.

CI now defines backend/race/migration/security, frontend, and fresh deployment
checks. These files have not been pushed or run by GitHub Actions in this turn.

## Primary references

- [Polar customer state](https://polar.sh/docs/api-reference/customers/state-external)
- [Polar checkout response](https://polar.sh/docs/api-reference/checkouts/get-session)
- [Stytch RBAC policy](https://stytch.com/docs/multi-tenant-auth/enterprise-ready/rbac/create-rbac-policy)
- [Go vulnerability](https://pkg.go.dev/vuln/GO-2026-5932)
- [Dependabot ecosystems](https://docs.github.com/en/code-security/reference/supply-chain-security/supported-ecosystems-and-repositories)
