---
name: "auth-integration"
description: "This skill applies when migrating Stytch to Better Auth, integrating Better Auth organizations, changing cookies or Go session verification, or testing tenant and role security."
---

<!-- Generated from .agents/skills/auth-integration/SKILL.md; do not edit. -->

# Auth integration and testing

Read `AGENTS.md` and `docs/ARCHITECTURE.md`. Determine the provider from the current
checkout's package manifest, lockfile, routes, and Go adapter before editing.
The starter uses self-hosted Better Auth in Next.js with a private Go session
bridge and PostgreSQL tenant isolation. [Issue #64](https://github.com/moasq/production-saas-starter/issues/64)
records the Stytch cutover. Neither this skill nor the documentation MCP installs
an application auth provider; verify the checked-out implementation before applying advice.

## Resolve Better Auth guidance

For Better Auth work, read the official `best-practices` and `organization` skills
at the exact URLs in `.agents/sources.json`. To make them available locally, run
`node scripts/harness.mjs fetch-skills`; then read
`.agents/cache/better-auth/best-practices/SKILL.md` and
`.agents/cache/better-auth/organization/SKILL.md`. These are upstream references,
not repository rules. Their install commands and defaults need review against
this starter. Use `check` to detect altered cached copies; do not commit the cache.

The `better-auth` documentation MCP provides `get_doc` and `search_docs` only.
Call `get_doc` with `path: "/llms.txt"`, match the requested or installed version
to its documentation identifier, then call `search_docs` with that version and a
focused query. Pass a returned path unchanged to `get_doc`. If Better Auth is not
installed, label latest-version results as migration research. If MCP is unavailable,
use [the official version index](https://better-auth.com/llms.txt) in the browser.
Keep external requests generic; reference exact source URLs in the handoff.

## Trace the complete boundary

Map browser cookies -> Next.js handler/server action -> Go verification -> current
membership -> scoped SQL/provider operation. During a migration, explicitly decide
who stores sessions, validates revocation, resolves active organization, updates
membership, sends mail, and owns each table. Do not assume a JWT plugin provides
current membership or logout revocation. Verify issuer/audience/signature/expiry
and a bounded revocation strategy if choosing JWTs. Never trust a bare forwarded
user/organization header or client-selected role as the cross-service contract.

Review proxy route precedence (`/api/auth/*` versus Go routes), trusted origins,
cookie domain/path/SameSite/Secure/HttpOnly, server-side cookie forwarding, and
refresh behavior. Verify both browser and server calls; a UI permission guard is
not an API authorization test. Keep any auth schema transition separate from
application SQLC ownership and document a data-preserving rollout/rollback plan.
Use pinned installed CLI versions for schema generation into a disposable database;
review generated SQL before integrating it with the migration runner.

## Required behavior matrix

| Boundary | Cases |
| --- | --- |
| Session | Missing, malformed, expired, revoked; refresh; sign-out prevents reuse; provider unavailable fails closed |
| Tenant | Forged org ID/header, another tenant's member/resource, active-org switch by a nonmember, deleted membership |
| Roles | Member cannot invite/delete/change roles/manage billing; no self-promotion; prevent removal or demotion of the last administrator/owner |
| Invitation | Wrong recipient, expired/replayed/canceled invite, resend, delete then reinvite, mail delivery failure |
| Browser | Foreign/missing origin on mutations, unsafe return URL, secure cookie flags, no session secret in JS output |
| Migration | Existing user/org mapping, no accidental duplicate tenant, fresh and legacy database, rollback boundary |
| Optional billing | Organization identity mapping stays stable; checkout and portal cannot target another tenant |

Use synthetic accounts and HTTP fixtures for local tests. Run the backend/frontend
checks from their skills for affected code and a fresh Compose smoke for route or
deployment changes. For authorized sandbox verification, exercise signup -> email
delivery -> login -> invitation acceptance -> logout with a second user and tenant.
Report local, sandbox, and production outcomes separately; never send real user
invitations, change production auth, or claim live completion from mocked results.
