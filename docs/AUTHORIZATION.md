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

A request authorized before a concurrent revocation can already be in flight.
These checks do not cancel completed or previously authorized operations. Browser
controls may also remain visible until navigation refreshes the profile; their
cached visibility never authorizes an API call.

## Verification

- `pnpm --dir next_b2b_starter test` verifies current role behavior, unsupported
  roles, denied organization deletion and a restrictive policy edit that removes
  management from both bridge grants and Better Auth operation authorization.
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
