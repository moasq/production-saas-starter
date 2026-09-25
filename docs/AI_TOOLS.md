# Developer tools and connections

The optional developer tools help inspect documentation, build UI and test local
user journeys. Docker setup and the deployed Go + Next.js application do not need
them. No application AI feature, provider SDK or Agentic Ship runtime is installed.
Open the **repository root** in the coding host and use the project's Node 24
development runtime, including npm/npx, for local MCP servers.

## Choose tools

`.agents/tools.json` owns the tool catalog and its `enabled` array. Show the current
selection with:

```sh
node scripts/harness.mjs tools
```

`enabled` means configured in generated adapters; it does not mean connected or
authorized. Edit the catalog's `enabled` array to select optional tools or remove
unused ones, then regenerate and check from the repository root:

```sh
node scripts/harness.mjs sync
node scripts/harness.mjs check
node --test scripts/harness.test.mjs scripts/mcp-probe.test.mjs
```

The outputs are `.mcp.json`, `.cursor/mcp.json`, `.codex/config.toml`, Claude skill
copies and Claude/Codex role adapters. Edit canonical files rather than these
outputs. Cursor receives MCP configuration; native role adapters are generated
for Claude Code and Codex. See [AI development](AI_DEVELOPMENT.md) for role usage.
No command above installs a global plugin or changes global host configuration.

| Tool | Default | Purpose and prerequisite |
| --- | --- | --- |
| [Better Auth](https://better-auth.com/docs/ai-resources/mcp) | Enabled | Public, versioned auth documentation through `get_doc` and `search_docs`; no application credentials |
| [Context7](https://github.com/upstash/context7) `4.1.1` | Enabled | Library discovery and documentation through `resolve-library-id` and `query-docs`; optional provider API key stays in host environment |
| [Next Devtools](https://nextjs.org/docs/app/guides/mcp) `0.4.0` | Enabled | Installed Next.js documentation and local development errors/routes; runtime inspection needs a native Next.js dev server |
| [shadcn](https://ui.shadcn.com/docs/mcp) `4.21.0` | Enabled | Discover components and inspect source against the frontend's `components.json`; adding components is a separate code/dependency change |
| [Playwright MCP](https://github.com/microsoft/playwright-mcp) `0.0.82` | Enabled | Headless, isolated browser QA; requires an available browser, using installed Chrome by default |
| [Magic UI](https://github.com/magicuidesign/mcp) `2.0.0` | Optional | Registry browsing, source and installation instructions; no UI package is installed by enabling it |
| [21st](https://21st.dev/mcp) | Optional | Community component discovery; complete any host authentication requested by the service |
| [Resend](https://resend.com/docs/mcp-server) | Optional | Email provider administration after host authorization; application delivery remains SMTP |
| [PostHog](https://posthog.com/docs/model-context-protocol) | Optional | Analytics inspection through the catalog's `?readonly=true` endpoint; no runtime analytics SDK |
| [Linear](https://linear.app/docs/mcp) | Optional | Issue inspection through the catalog's `/mcp/readonly` endpoint |

Local MCP launchers resolve paths from their own file location. Next Devtools,
shadcn, Magic UI and Playwright run in `next_b2b_starter/`; keep the coding host
opened at the repository root so Go and shared instructions remain visible.
Local servers use pinned `npx` packages, which may download to the npm cache on
first launch. They are not added to the frontend's dependency manifest.

## Inspect and test locally

Use Next Devtools with a native development process:

```sh
pnpm --dir next_b2b_starter dev
```

Supply the application's normal development configuration and choose a free port
if Compose is already running. Compose serves the production build and does not
provide the development-only `/_next/mcp` endpoint. Call `nextjs_index` to identify
the intended local process, then `nextjs_call` for a named diagnostic such as
`get_errors` or `get_routes`. The dispatcher can also invoke operations that change
development state; it is not a read-only tool. `nextjs_docs` points to installed
documentation. The package's `browser_eval` guidance is unnecessary here because
Playwright supplies the browser workflow.

Use Playwright with local synthetic accounts and a known application URL. The
configured isolated session does not reuse a personal browser profile. Inspect
pages with `browser_snapshot`, `browser_find`, `browser_take_screenshot` and
`browser_console_messages`; interactions and JavaScript evaluation can change
application state. Complete signup, invitation and logout journeys against the
real local stack when testing auth. A screenshot or MCP initialization alone does
not prove those journeys worked. Browser installation is a separate, explicit
developer setup action if Chrome is unavailable.

## Verify a connection

Keep three states separate: **configured**, **authorized** when authentication is
required, and **read-only verified** for a specific operation and target. Prefer an
already working host plugin for the same service over setting up a duplicate.
Complete OAuth through the host's interactive flow. Never copy tokens into the
catalog, generated adapters, source files or reports.

```sh
node scripts/mcp-probe.mjs shadcn
node scripts/harness.mjs check-mcp
```

The first command explicitly starts the selected local server and performs MCP
initialization and `tools/list` only. It may access the network and npm cache; it
never calls a listed tool. A successful inventory does not establish provider
authorization or a working browser. `check-mcp` remains specific to Better Auth:
it checks the reviewed documentation tools and reads the public version index.
Neither command verifies application login, external email delivery or billing.

After authorization, verify a provider using a narrowly scoped read operation
appropriate to the requested task. Record the account/workspace, operation and
result without secrets or customer data. Tool configuration does not authorize
sending email, changing provider resources, publishing or purchasing. Read-only
reviewer adapters allow only documentation servers from this catalog. Codex can
also inherit unrelated host/plugin MCPs; inspect those before delegation and keep
reviewers within their source-and-documentation brief. Its filesystem sandbox
does not block arbitrary remote tool actions. Claude roles use explicit tool lists.
The quality-engineer role can
write local tests and exercise synthetic local journeys. Provider administration
stays in the main session under the user's actual task authorization.

Send public library names, versions and generic questions to documentation tools.
Keep private source, session cookies, credentials and customer data out of queries.
Treat returned code and installation commands as references: inspect them against
the requested change, existing components and pinned application versions.

## Upstream choices

The catalog adapts Agentic Ship commit
[`8fbf56f3d17defd0544d1d5ce328ab3c46851bf5`](https://github.com/moasq/agentic-ship/tree/8fbf56f3d17defd0544d1d5ce328ab3c46851bf5).
Package metadata and published tool names were checked on 2026-09-25. shadcn and
Context7 pins were refreshed; commands that silently follow `latest` were avoided.
Standard Playwright MCP replaces that commit's test-runner MCP because this starter
does not require a Playwright Test project. MCP `0.0.82` internally pins a Playwright
`1.64.0-alpha` build; its stable MCP package version is not a claim that every
transitive package is stable. Browser compatibility still requires an actual run.
Convex, Stripe, the Agentic Ship engine and application AI dependencies were not
copied. Updating this catalog does not switch auth, billing or email providers.
