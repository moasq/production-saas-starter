---
name: "service-connections"
description: "This skill applies when enabling an optional provider MCP, completing host authorization, verifying a service connection, or diagnosing Resend, PostHog, Linear or component-provider access in this starter."
---

# Service connection workflow

Read `AGENTS.md`, `.agents/tools.json` and `docs/AI_TOOLS.md`. Identify the requested
provider, intended account/workspace and operation before changing configuration.
Keep developer access separate from the application's auth, SMTP and Polar setup.

## Establish the actual state

1. Inspect callable host tools first. Prefer an already authorized host plugin
   when it serves the task; do not configure a second connection unnecessarily.
2. Run `node scripts/harness.mjs tools` to inspect repository selections. Treat an
   enabled entry as **configured** only. Do not report it as connected, authorized
   or tested from its presence in `.mcp.json` or a generated host adapter.
3. For a requested optional selection, edit `.agents/tools.json`, then run
   `node scripts/harness.mjs sync` and `node scripts/harness.mjs check`.
   Preserve catalog endpoints, including PostHog's `readonly=true` query and
   Linear's `/mcp/readonly` path. Do not silently broaden their permissions.
4. Complete any OAuth or account selection in the coding host's interactive flow.
   Keep credentials in the host's supported secret storage or private environment.
   Never put tokens in repository files, prompts, generated adapters or evidence.
5. After authorization, perform a narrow read appropriate to the task and confirm
   the intended workspace. Mark the connection **read-only verified** only for
   that operation. Report denied scopes, wrong accounts or unavailable tools
   precisely; retain successful earlier work while resolving the specific blocker.

## Keep verification bounded

Use `node scripts/mcp-probe.mjs <id>` for explicit local MCP initialization and
tool-list checks; it does not call tools or prove provider authorization. Use
`node scripts/harness.mjs check-mcp` only for Better Auth's public documentation
contract. Neither check establishes working application authentication, SMTP
delivery, analytics collection or billing.

Use public documentation queries without private source, customer records,
credentials or cookies. Read only data needed for the assigned provider task and
keep returned sensitive content out of committed reports. Record verification
time, provider/workspace, operation and outcome without secrets.

Treat resource creation, email sending, publication, billing and production
configuration changes as separate actions requiring the user's task authorization.
Use existing authorization when it covers the concrete action; do not ask again
merely because this skill was loaded. Configure or verify a connection without
creating test provider resources or sending messages as an incidental check.

Keep provider administration in the main authorized session. Reviewer roles use
documentation only, and local QA roles do not gain provider access from a test
assignment. Adding a developer MCP does not install a runtime SDK, enable
analytics, replace SMTP, switch Polar billing or migrate Better Auth.

Return configured, authorized and read-only verified states separately, with the
exact remaining user interaction or technical blocker. Do not call a connection
complete when only configuration generation or MCP initialization succeeded.
