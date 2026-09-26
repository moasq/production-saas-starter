# Architecture

```text
Browser -> Caddy -> Next.js (pages, Better Auth, HttpOnly session cookies)
                -> Go API (business authorization and tenant data)
Go -> private Next.js auth bridge (live session + active membership)
Next.js -> PostgreSQL auth schema (starter_auth role)
Go -> PostgreSQL organizations schema (starter_app role + row-level security)
Next.js + Go -> Polar (optional billing)
```

The starter remains a Go modular monolith with a Next.js frontend and one
PostgreSQL database. Better Auth runs inside Next.js; no external authentication
account, additional auth service, Redis, or application AI runtime is required.
SMTP delivers login links and invitations. Mailpit captures local development
email; replace it with a real SMTP provider in production.

The versioned [business API contract and generated client](decisions/0001-business-api.md)
define the Go/Next.js boundary. Go route tests and the disposable deployment suite
check the schema against mounted endpoints; frontend CI rejects generated drift.

[Request cancellation and timeouts](REQUEST_TIMEOUTS.md) keep one response-writer
owner and define the limits of cooperative deadlines, streams and background work.

## Authentication and authorization

Better Auth owns identities, sessions, organizations, memberships and roles.
Go resolves each protected request through the internal auth bridge using the
original signed session cookie. The bridge reads the current session and active
membership. It does not trust a client role, an organization header, or a
signature-only JWT. Cookie caching and sliding session renewal are disabled. Sessions expire after
eight hours, requiring a fresh login; private API calls cannot extend database
expiry without renewing a browser cookie. Missing, expired or revoked
sessions and removed memberships fail closed. An unavailable auth server denies
access rather than authorizing stale data.

Caddy returns 404 for `/internal/*`. The bridge additionally requires an internal
secret, an authenticated session, and authorization for the requested operation.
Only Caddy publishes application ports. `/api/identity/*` belongs to Better Auth;
existing business `/api/auth/*` routes belong to Go. Next.js server actions retain
the original cookie when calling Go. Never expose bridge credentials to browsers.

Roles are scoped to a membership, not a user globally. Admin can manage the
organization, team and billing; manager and member can view the workspace and
edit their own profile. The manager role deliberately has no implicit elevation.
The role-to-permission policy is defined in the auth module and tested. User
identity plus current membership plus an explicit action permission is required;
an admin in one workspace is not an admin in another.

## Data isolation

Business tables use PostgreSQL `ENABLE ROW LEVEL SECURITY` and `FORCE ROW LEVEL
SECURITY`. The Go runtime role cannot own protected tables or bypass RLS. A
transaction receives `app.tenant_id` from the verified organization, using
transaction-local configuration. Missing context sees no tenant rows; writes
must also satisfy the policy. The context ends on commit or rollback, including
when a connection returns to the pool.

This protects against accidental missing tenant predicates. It does not make an
arbitrary SQL injection or compromised application server harmless: application
code controls the database context. Parameterized SQL, verified membership,
least privilege and negative tests remain necessary. Database superusers and
migration operators are trusted administrative principals outside the tenant
boundary. There is no claim of physical database isolation between tenants.

`starter_auth` owns only the auth schema. `starter_app` has scoped business-table
permissions and no access to sessions. The owner account is supplied only to the
one-shot role provisioning and business migration jobs. Historical migrations
remain unchanged. Existing provider identifiers remain available for migration,
and existing Polar customer IDs remain attached to their original organizations.
See [Upgrading](UPGRADING.md) before importing an existing installation.

These choices follow [OWASP authorization guidance](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html),
[OWASP multi-tenant guidance](https://cheatsheetseries.owasp.org/cheatsheets/Multi_Tenant_Security_Cheat_Sheet.html),
and [PostgreSQL row-security behavior](https://www.postgresql.org/docs/17/ddl-rowsecurity.html).
They are engineering controls, not a compliance certification.

## Deployment and billing

Caddy terminates HTTPS. Credentials and public URLs are runtime configuration.
Startup waits for versioned migrations before application readiness. Dependency
locks and pinned images support reproducible builds. Local development mail is
an optional Compose profile, disabled for production.

Polar is optional. The core app does not require a subscription. Billing reads
current provider state and preserves the original external customer ID during
auth migration. Checkout verification checks tenant ownership and does not grant
access or maintain a local webhook replica. Provider availability is therefore
required for billing requests. Live payments and external email delivery need
separate deployment verification.

Add product features behind the same transaction-scoped tenant boundary, with
RLS on every new tenant-owned table and cross-tenant tests. No document, OCR, AI,
or RAG product domain is built in.
