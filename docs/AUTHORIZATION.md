# Authorization policy

Application code is the permission authority. Better Auth stores identity,
sessions and each organization's current membership assignment. Go checks the
verified tenant and action permission before business operations and data access.
This replaces the old Stytch split between hardcoded standard roles and provider
custom-role policies; the starter does not query an external policy service.

## One declaration

`next_b2b_starter/lib/auth/rbac.ts` declares `rolePermissions`. The same declaration
produces the explicit grants returned by the private session bridge and the
Better Auth organization plugin's management grants. There is no second role map
in Go, browser storage, provider defaults or a settings dashboard.

| Membership role | Application grants | Behavior |
| --- | --- | --- |
| `admin` | `org:view`, `org:manage` | View the workspace; manage its name, members, invitations and optional billing |
| `manager` | `org:view` | View the workspace and edit their own profile |
| `member` | `org:view` | View the workspace and edit their own profile |

`org:manage` maps to organization update, member create/update/delete and invitation
create/cancel. Organization deletion is not granted. Self-profile editing is
scoped to the verified user and membership, not an organization management grant.
Only these three single role names are supported. Unknown, custom, combined and
case-variant names establish no membership authorization; PostgreSQL constraints
also reject unsupported role assignments. Go rejects unsupported roles and grants
from the bridge, including wildcards. A role label cannot restore a missing grant.

Dynamic access control and custom roles are not enabled. Adding them would require
an explicit contract change across assignment validation, auth-schema constraints,
bridge validation, the Go API schema and UI, with negative tenant/permission tests.
Do not add a provider policy alongside the existing application declaration.

## Policy changes and revocation

Change the declaration, run the RBAC and Go bridge tests, and deploy the changed
Next.js service. Policy is loaded when the process starts. Complete replacement
of old frontend instances before treating a restrictive policy rollout as
effective; mixed versions can return different grants. Go does not cache grants
or need a matching role-map release when existing permissions are removed.
Introducing a new action or role does require updating its Go contract explicitly.

Every protected Go request checks the signed cookie with the private bridge.
The bridge reads the live session and active organization membership; cookie
caching and session renewal are disabled. An administrator demotion, deleted
membership or revoked session therefore affects the next authorization check,
including requests using an already issued cookie. The private mutation bridge
also rechecks membership and management grants. There is no stale-data fallback
when the auth service or its database is unavailable.

The private bridge's `org:manage` check is essential even with plugin ACLs.
Better Auth can allow its configured creator role to bypass the role ACL for
role-assignment operations. Do not expose direct SDK membership mutations or
assume that a denied `roles.admin.authorize` call secures those operations by
itself. Public organization mutation routes remain blocked; supported mutations
pass through the application's current-membership and permission checks.

A request authorized before a concurrent revocation can already be in flight.
These checks do not cancel completed or previously authorized operations. Browser
controls may also remain visible until navigation refreshes the profile; their
cached visibility never authorizes an API call.

## Verification

- `pnpm --dir next_b2b_starter test` verifies current role behavior, unsupported
  roles, denied organization deletion and a restrictive policy edit that removes
  management from both bridge grants and the generated Better Auth role ACL.
  This tests the ACL object, not every SDK endpoint's creator-role exceptions.
- Go's `internal/platform/betterauth/authorization_test.go` mounts real auth and
  permission middleware against an HTTP bridge fixture. Reusing one cookie, it
  checks permission removal without changing the `admin` role, demotion, restored
  explicit permission, invalid roles/grants, service failure and session revocation.
  Denied requests must not reach the business mutation.
- `scripts/test-auth.mjs` exercises real self-hosted sessions and membership
  changes against a disposable Compose stack, including promotion followed by
  demotion on the same session and denial of self-promotion afterward. It also
  checks tenant boundaries, last-admin protection, member removal and logout.

Fixtures verify application behavior. Local Compose uses synthetic accounts and
Mailpit; neither check establishes external email delivery or live Polar behavior.

## Invitation delivery and retry

Invitations remain pending until the verified recipient accepts them. Both initial
send and resend await SMTP acceptance; this confirms acceptance by the configured
mail server, not delivery to the recipient's inbox. The bridge sends explicitly after
authorized Better Auth creation; SDK success alone is insufficient because its
organization email callback logs and swallows delivery errors. Public organization
mutation endpoints remain blocked, and no second callback sends duplicate mail. An initial SMTP failure returns
`invite_sent: false` with the saved invitation ID. A failed resend returns a dependency
error. No email mutation is retried automatically by the frontend.

The private bridge reserves a PostgreSQL cooldown shared by all
application replicas: one attempt per workspace and normalized recipient every
60 seconds, plus 20 attempts per workspace in each 10-minute window. Concurrent
resends cannot bypass the cooldown. Both the create and resend operations return
HTTP 429 with a wait message when limited. The reservation remains after SMTP
failure because a timeout may follow mail-server acceptance; an administrator can
retry after the cooldown. Resend keeps an unexpired pending invitation and renews
its 48-hour expiry. Expired invitations leave the pending list; invite the email
again to create a fresh invitation. Expired or canceled IDs cannot be resent or
accepted.

The disposable auth suite checks initial and repeat Mailpit captures, concurrent
resend denial, tenant and recipient checks, and accepted-link replay. With
`TEST_SESSION_EXPIRY=true TEST_INVITATION_FAILURES=true` and an explicit synthetic
`COMPOSE_PROJECT_NAME`, it also stops and restores that project's Mailpit service
to exercise saved-but-unsent invitations and failed/recovered resends. These are
local SMTP checks; separately verify a production provider and real inbox delivery
before launch. No external email account is used by this suite.
