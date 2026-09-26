# Business API ownership and generated client

Status: accepted. API contract version: 1.0.0.

Go remains the authority for business authorization, tenant data, and the current
billing snapshot. Next.js owns presentation, HTTP-only sessions, and the self-hosted
Better Auth identity service. The private bridge verifies current membership and
executes identity operations through Better Auth after Go's public permission
check. The bridge repeats its authorization checks; it is not a public alternate
business API. PostgreSQL owns durable data and tenant RLS.

The canonical public business contract is
[`go-b2b-starter/apicontract/openapi.json`](../../go-b2b-starter/apicontract/openapi.json), using
OpenAPI 3.1. It covers the mounted profile, member, organization and billing routes.
Identity, workspace-selection, private bridge and operational health endpoints
have separate owners and are intentionally outside this business client.

Browser requests use same-origin `/api` through Caddy to Go. Server components and
server actions use `API_BASE_URL_INTERNAL`, forwarding only the current request's
cookie. Mutations also forward `APP_BASE_URL` as Origin. No client singleton stores
session state, accepts a tenant override, or retries a mutation. Member lists are
the complete active-organization list; pagination and tenant query parameters are
not supported. Selecting a workspace goes through the authenticated Next.js flow.

The frontend uses pinned `openapi-typescript` to generate request/response/path
types and `openapi-fetch` to construct requests. Repositories map wire data to UI
models; they no longer guess alternative wrappers or assert caller-written DTOs.
DTO compatibility exports refer to the generated schemas. The existing response
field names, including billing's capitalized fields, are retained.

From `next_b2b_starter/`:

```sh
pnpm api:generate
pnpm api:check
pnpm verify
```

Change the schema and mounted implementation together. Generated files are
outputs; check mode must fail on drift without repairing it. Increment the
contract version for changes and document any breaking fields or operations.
Generated TypeScript catches invalid paths, missing bodies and unsupported role
values. It does not validate arbitrary network data at runtime or replace server
authorization and input validation.

The Go contract test enumerates the actual mounted module routes and proves
every documented operation requires authentication before accessing dependencies.
Frontend CI rejects generation drift and compiles negative type fixtures. The
deployment job validates real request/response bodies against the schema while
running the existing synthetic auth suite, including profile updates, member
invitations/resend/removal/roles and organization reads/updates. Reproduce after
installing frontend development dependencies on a **disposable** local stack:

```sh
API_CONTRACT_CHECK=true TEST_SESSION_EXPIRY=true node scripts/test-auth.mjs
```

Set `STARTER_URL`, `MAILPIT_URL`, and `COMPOSE_PROJECT_NAME` for a non-default test
stack. This suite creates synthetic users and changes their sessions; never run
it against customer data. Disabled-billing error behavior is covered; a successful
live payment verification still requires separately authorized provider testing.

No RPC framework, second backend, application AI feature, or client-side auth
token store is introduced. See the [architecture](../ARCHITECTURE.md) for the
membership and data-isolation boundaries.
