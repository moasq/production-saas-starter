# Authentication and tenant boundaries

Next.js owns self-hosted Better Auth users, sessions, organization memberships and
canonical admin/manager/member policy. Go sends the raw Cookie to
`POST /internal/auth/session`, authenticated with `X-Internal-Auth-Secret`.
The bridge reads live session and membership state on every request. There is no
cookie cache, bearer/JWT alternative or role-name fallback. Expiry, logout, member
removal and role changes take effect on the next request. The server-only bridge
secret must never be exposed through browser build variables or public ingress.

The bridge returns verified user/email, active organization, current membership,
explicit permissions and expiry. Go checks the organization binding and expiry,
then synchronizes organization/account mirrors using a transaction-local tenant
setting. The database FORCE RLS policies deny rows without matching scope; the
runtime role cannot own tables, bypass RLS, or run as superuser. A suspended local
business organization stays suspended regardless of identity-provider membership.

Members and profile changes are delegated to private bridge operations with the
original end-user Cookie. Both Go and the bridge enforce current org:manage for
member and organization changes. Payload organization IDs never choose the scope.
All cookie-authenticated writes require an Origin matching APP_BASE_URL.

Existing identity identifiers are retained as historical columns by migration 11.
The separate migration/import step creates Better Auth users and memberships;
users must verify a new magic link. Old cookies and outstanding identity links do
not authenticate. Existing Polar external customer IDs are preserved separately.
