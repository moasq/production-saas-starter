# AI-assisted development

The starter includes five coding roles and six focused skills. They guide work
on Go, Next.js, auth security, PR review, testing, and service connections without adding an app service or dependency.
Normal Docker setup requires none of these tools. The harness scripts use Node.js
22.18+ with its standard library; the project's Node 24 development runtime works.

## Start with the task

Open the repository in Codex or Claude Code and trust the project configuration
only after reviewing it. Ask for a bounded change, for example:

- “Use backend-builder to add a tenant-scoped endpoint and negative permission tests.”
- “Use frontend-builder to fix invitation delivery feedback and keyboard focus.”
- “Use auth-reviewer to inspect this session change without editing files.”
- “Use code-reviewer and the pr-review skill to review this PR before merging.”
- “Use quality-engineer to test signup and tenant switching with local synthetic users.”

Roles inherit the caller's model. Both reviewers declare a read-only Codex sandbox
default, explicit project MCP restrictions, and a Claude tool allowlist without shell or write tools. Live parent-session
permission overrides can supersede Codex defaults. Generated MCP restrictions cover
this catalog; inspect unrelated inherited host/plugin tools before delegating a
read-only review. Implementation roles use
the host's existing approval policy; the harness never disables it. Dispatch is
optional. A host without native roles can read the corresponding canonical brief
and follow it in the main conversation. No parallel workers start automatically.

| Source | Purpose | Host delivery |
| --- | --- | --- |
| `AGENTS.md` | Repository rules | Codex reads directly; `CLAUDE.md` imports it |
| `.agents/skills/*/SKILL.md` | Focused coding and integration procedures | Codex discovers directly; Claude receives generated copies |
| `.agents/agents/*.md` | Bounded role input/process/output | Generated `.codex/agents/*.toml` and `.claude/agents/*.md` |
| `.agents/tools.json` | Tool catalog, exact executable pins, enabled selection | Generates `.mcp.json`, `.codex/config.toml`, and `.cursor/mcp.json` |
| `.agents/sources.json` | Source revisions, hashes, adaptations | Validated by the harness script |

Generated adapter directories belong to this harness. Edit canonical sources,
then run from the repository root:

```sh
node scripts/harness.mjs sync
node scripts/harness.mjs check
node --test scripts/harness.test.mjs scripts/mcp-probe.test.mjs
```

`check` is offline and does not write files. CI runs the same structural and
behavior checks. Paths are relative to the script, so invoking it by its absolute
path from another directory works. No command modifies a global agent config,
installs packages, changes permissions, or starts the application. Tool startup and
the opt-in MCP probes can fetch pinned npm executables. See [Tools and integrations](AI_TOOLS.md)
for local developer tools, optional provider connections, and verification boundaries.

## Better Auth's own skills and MCP

The [official documentation MCP](https://better-auth.com/docs/ai-resources/mcp)
uses `https://mcp.better-auth.com/mcp`. Its reviewed tools are `search_docs` and
`get_doc`; both advertise read-only behavior. Codex restricts the server to those
tools, and each Claude role names them explicitly. The main Claude session retains
normal host tool approval; do not approve new tools just because the URL is familiar.
The remote service is not version-pinned and can change. To verify it explicitly:

```sh
node scripts/harness.mjs check-mcp
```

This initializes MCP, checks the tool inventory/annotations, and retrieves the
public documentation version index. It sends no repository content or credentials.
It does **not** prove the app can authenticate users. When querying documentation,
resolve the installed or requested Better Auth version via `get_doc('/llms.txt')`,
then search that version and fetch the returned document path unchanged. Use the
[official web index](https://better-auth.com/llms.txt) if MCP is unavailable.

The [official skills repository](https://github.com/better-auth/skills) is pinned
in `.agents/sources.json` to a full commit and SHA-256 hashes for `best-practices`
and `organization`. No license was declared at that revision, so their contents
are not redistributed here. Download the exact files for local reference with:

```sh
node scripts/harness.mjs fetch-skills
```

The command only downloads Markdown into the gitignored `.agents/cache/` directory.
It verifies both hashes before saving either file and runs no upstream commands.
The auth integration skill explains when to read them. Missing cache is allowed;
the pinned links remain usable. An altered cache causes `check` to fail. Do not
commit cached files or infer a license from Better Auth's separate runtime package.

To update, inspect the upstream commit and license, review both file diffs, update
their revision/URL/hash records, then rerun fetch, sync, check, and tests. Do not
silently follow `main` or execute `npx ...@latest` from a fetched example.

## Application auth boundary

The application uses self-hosted Better Auth; [issue #64](https://github.com/moasq/production-saas-starter/issues/64)
records its replacement of Stytch. Inspect the current checkout's manifest,
lockfile, routes, and architecture before choosing provider-specific advice.
The developer harness is optional and does not run application authentication.

Next.js serves Better Auth at `/api/identity`, with a
private `POST /internal/auth/session` bridge protected by `X-Internal-Auth-Secret`
and the original cookie. Go receives current session/membership permissions rather
than authorizing from cached JWT claims. Review that this endpoint is never exposed
by Caddy, rejects missing/wrong internal secrets, and rechecks revoked sessions and
membership. The database split gives `starter_auth` access to auth tables,
`starter_app` a non-owner/non-bypass role for business tables with forced RLS and
transaction-local `app.tenant_id`, and a separate one-shot migration owner. Verify
cross-tenant access and pooled-connection tenant reset against the actual database.
See [architecture](ARCHITECTURE.md), [upgrade steps](UPGRADING.md), and
[verification evidence and limits](VALIDATION.md). Local integration checks do
not establish external SMTP delivery, live billing, or a production restore.

Keep developer documentation MCP separate from
Better Auth's application MCP authentication plugin; this starter does not need that
plugin, agent authentication, or any AI product feature.

## Evidence and provenance

The authored skills and tool catalog adapt Agentic Ship's canonical-source ownership,
bounded handoffs, developer integrations, secret boundaries, and verification principles. The inspected
commit and paths are in `.agents/sources.json`. No Convex implementation or toolkit
engine was copied. Tool configuration provenance is also recorded in `.agents/tools.json`;
the upstream MIT notice is retained in `.agents/licenses/agentic-ship.txt`.
New harness files use this repository's MIT license.

Host schema references: [Codex subagents](https://developers.openai.com/codex/subagents/),
[Codex MCP](https://developers.openai.com/codex/mcp/),
[Claude subagents](https://code.claude.com/docs/en/sub-agents),
[Claude skills](https://code.claude.com/docs/en/skills), and
[Claude memory imports](https://code.claude.com/docs/en/memory).

Structural checks demonstrate adapter consistency, safe output paths, and pinned
downloads. They cannot establish model quality, reliable automatic skill selection,
or completed live auth/email/billing flows. Use the auth skill's behavior matrix and
the normal application checks for implementation work. Related work: issues
[#48](https://github.com/moasq/production-saas-starter/issues/48),
[#49](https://github.com/moasq/production-saas-starter/issues/49), and
[#50](https://github.com/moasq/production-saas-starter/issues/50).
