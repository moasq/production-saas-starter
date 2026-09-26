---
name: "dev-tools"
description: "Route and configure the repository's optional developer tools, check generated host adapters and distinguish configuration from verified connections."
---

# Developer tool routing

Read `AGENTS.md`, `.agents/tools.json` and `docs/AI_TOOLS.md`. The catalog owns
exact pins and selections; each project's adapters expose its allowed subset.
Root coordinates tool configuration. Go gets documentation tools; Next.js also
gets local development tools. Provider administration stays in the authorized
main session and is never inherited by implementation or reviewer roles.

Run these from the repository root, or invoke the scripts by absolute path:

```sh
node scripts/harness.mjs tools
node scripts/harness.mjs sync
node scripts/harness.mjs check
node --test scripts/harness.test.mjs scripts/mcp-probe.test.mjs
```

Edit the canonical catalog for requested selections; never edit generated adapters
or global settings. Prefer an existing working host integration. A local MCP may
fetch its pinned executable on first launch. `node scripts/mcp-probe.mjs <id>`
initializes and lists a selected local tool only; it does not verify app behavior.

Route Go work to `go-b2b-starter/.agents/skills/go-backend/SKILL.md`, frontend
runtime/component/browser work to
`next_b2b_starter/.agents/skills/frontend-tools/SKILL.md`, Better Auth guidance to
`next_b2b_starter/.agents/skills/auth-integration/SKILL.md`, and requested provider
connections to `.agents/skills/service-connections/SKILL.md`.

Reviewer adapters expose documentation tools only. Inspect unrelated inherited
host/plugin capabilities before delegation. Keep private source, credentials,
cookies and customer records out of documentation queries. Report configuration,
initialization and actual local/provider verification as separate outcomes.
