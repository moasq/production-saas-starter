import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { chmodSync, cpSync, existsSync, linkSync, lstatSync, mkdirSync, mkdtempSync, readFileSync, realpathSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { delimiter, dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import { test } from "node:test";
import { adapterPlan, checkCommands, checkMcp, checkReferences, fetchSkills, safePath, sources, sync } from "./harness.mjs";
import { toolCatalog, mcpConfig } from "./tool-catalog.mjs";

const repository = resolve(dirname(fileURLToPath(import.meta.url)), "..");
function fixture(t) {
  const root = mkdtempSync(join(tmpdir(), "starter harness "));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  for (const scope of ["", "go-b2b-starter/", "next_b2b_starter/"]) {
    cpSync(join(repository, `${scope}.agents`), join(root, `${scope}.agents`), { recursive: true, filter: (path) => !path.split(/[\\/]/).includes("cache") });
    cpSync(join(repository, `${scope}AGENTS.md`), join(root, `${scope}AGENTS.md`));
  }
  const manifest = JSON.parse(readFileSync(join(repository, ".agents/sources.json"), "utf8"));
  // Copy only harness contracts and their referenced files, never installed apps.
  const files = new Set([".mcp.json", "README.md", "SETUP.md", "docs/AI_DEVELOPMENT.md", "next_b2b_starter/README.md", "next_b2b_starter/package.json", "go-b2b-starter/Makefile",
    "scripts/harness.mjs", "scripts/tool-catalog.mjs", "scripts/mcp-launch.mjs", "scripts/mcp-probe.mjs", "scripts/harness.test.mjs", "scripts/mcp-probe.test.mjs", "scripts/test-browser.sh",
    ...manifest.authoredHere.flatMap((entry) => entry.references)]);
  for (const path of files) {
    const source = join(repository, path);
    if (lstatSync(source).isFile()) { mkdirSync(dirname(join(root, path)), { recursive: true }); cpSync(source, join(root, path)); }
    else mkdirSync(join(root, path), { recursive: true });
  }
  return root;
}

test("adapters are portable, deterministic, and check runs from another cwd", (t) => {
  const root = fixture(t);
  const count = adapterPlan(root).size;
  assert.equal(sync(root), count);
  assert.equal(sync(root, true), count);
  const result = spawnSync(process.execPath, [join(root, "scripts/harness.mjs"), "check"], { cwd: tmpdir(), encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
  for (const relative of ["CLAUDE.md", "go-b2b-starter/CLAUDE.md", "next_b2b_starter/CLAUDE.md", "go-b2b-starter/.claude/CLAUDE.md", "next_b2b_starter/.claude/CLAUDE.md"]) {
    const path = join(root, relative);
    const imported = readFileSync(path, "utf8").match(/^@(.+)$/m)[1];
    const scope = relative.startsWith("go-") ? "go-b2b-starter" : relative.startsWith("next_") ? "next_b2b_starter" : "";
    assert.equal(resolve(dirname(path), imported), join(root, scope, "AGENTS.md"));
  }
  const config = readFileSync(join(root, ".codex/config.toml"), "utf8");
  assert.match(config, /enabled_tools = \["get_doc","search_docs"\]/);
  assert.doesNotMatch(config, /approval_policy|danger-full-access|mcp-remote|\/Users\//);
  for (const name of ["auth-reviewer", "code-reviewer"]) {
    const scope = name === "auth-reviewer" ? "next_b2b_starter/" : "";
    const reviewer = readFileSync(join(root, `${scope}.codex/agents/${name}.toml`), "utf8");
    assert.match(reviewer, /sandbox_mode = "read-only"/);
    const claude = readFileSync(join(root, `${scope}.claude/agents/${name}.md`), "utf8");
    const tools = claude.match(/^tools: (.*)$/m)[1];
    assert.doesNotMatch(tools, /Bash|Write|Edit|Agent/);
    assert.match(tools, /mcp__better-auth__get_doc/);
  }
});

test("check detects drift without repairing it; sync repairs from canonical source", (t) => {
  const root = fixture(t);
  sync(root);
  const path = join(root, ".codex/config.toml");
  writeFileSync(path, "drift\n");
  assert.throws(() => sync(root, true), /adapter drift/);
  assert.equal(readFileSync(path, "utf8"), "drift\n");
  sync(root);
  assert.equal(sync(root, true), adapterPlan(root).size);
});

test("root coordinates while each project discovers only its canonical domain skills and roles", (t) => {
  const root = fixture(t);
  sync(root);
  const domains = [
    ["go-b2b-starter", "go-backend", "backend-builder"],
    ["next_b2b_starter", "next-frontend", "frontend-builder"],
    ["next_b2b_starter", "auth-integration", "auth-reviewer"],
  ];
  for (const [scope, skill, role] of domains) {
    assert.equal(existsSync(join(root, `.agents/skills/${skill}/SKILL.md`)), false);
    assert.equal(existsSync(join(root, `.claude/skills/${skill}/SKILL.md`)), false);
    assert.equal(existsSync(join(root, `.codex/agents/${role}.toml`)), false);
    assert.equal(existsSync(join(root, `${scope}/.agents/skills/${skill}/SKILL.md`)), true);
    assert.equal(existsSync(join(root, `${scope}/.codex/agents/${role}.toml`)), true);
    const result = spawnSync(process.execPath, ["../scripts/harness.mjs", "check"], { cwd: join(root, scope), encoding: "utf8", timeout: 5000 });
    assert.equal(result.status, 0, result.stderr);
  }
  assert.equal(existsSync(join(root, ".agents/skills/orchestration/SKILL.md")), true);
  const goMcp = JSON.parse(readFileSync(join(root, "go-b2b-starter/.mcp.json"), "utf8"));
  assert.deepEqual(goMcp.mcpServers.context7.args, ["../scripts/mcp-launch.mjs", "context7"]);
  assert.equal(goMcp.mcpServers.shadcn, undefined);
});

test("skill source edits require matching versioned provenance before any adapter writes", (t) => {
  const root = fixture(t);
  const path = "next_b2b_starter/.agents/skills/next-frontend/SKILL.md";
  writeFileSync(join(root, path), readFileSync(join(root, path), "utf8") + "\nA reviewed local adaptation.\n");
  assert.throws(() => sync(root), /Stale or missing skill provenance/);
  assert.equal(existsSync(join(root, "CLAUDE.md")), false);
  const manifest = sources(root);
  const entry = manifest.authoredHere.find((item) => item.path === path);
  entry.sha256 = createHash("sha256").update(readFileSync(join(root, path), "utf8").replaceAll("\r\n", "\n")).digest("hex");
  entry.version = "2.0.1";
  writeFileSync(join(root, ".agents/sources.json"), JSON.stringify(manifest));
  sync(root);
  assert.match(readFileSync(join(root, "next_b2b_starter/.claude/skills/next-frontend/SKILL.md"), "utf8"), /A reviewed local adaptation/);
});

test("broken local links and unknown documented commands fail without executing them", (t) => {
  const root = fixture(t);
  assert.throws(() => checkReferences(root, "[missing](../missing.md)", "docs/AI_DEVELOPMENT.md"), /ENOENT/);
  assert.throws(() => checkCommands(root, "`pnpm missing:task`"), /Unknown documented pnpm/);
  assert.throws(() => checkCommands(root, "`make -C go-b2b-starter absent-target`"), /Unknown documented Make/);
  assert.throws(() => checkCommands(root, "`node scripts/missing.mjs`"), /ENOENT/);
  assert.doesNotThrow(() => checkCommands(root, "A frontend owner can make a bounded change."));
  assert.doesNotThrow(() => checkCommands(root, "`node ../scripts/harness.mjs check`", "next_b2b_starter"));
});

test("scoped adapter preflight rejects nested links and leftover legacy roles before any write", (t) => {
  const root = fixture(t);
  const outside = mkdtempSync(join(tmpdir(), "harness outside "));
  t.after(() => rmSync(outside, { recursive: true, force: true }));
  symlinkSync(outside, join(root, "next_b2b_starter/.codex"), "junction");
  assert.throws(() => sync(root), /Refusing linked/);
  assert.equal(existsSync(join(root, "CLAUDE.md")), false);
  rmSync(join(root, "next_b2b_starter/.codex"));
  mkdirSync(join(root, "go-b2b-starter/.claude/agents"), { recursive: true });
  writeFileSync(join(root, "go-b2b-starter/.claude/agents/legacy.md"), "obsolete instruction");
  assert.throws(() => sync(root), /Unmanaged adapter/);
  assert.equal(existsSync(join(root, "CLAUDE.md")), false);
});

test("invalid canonical references and orphaned adapters fail before writes", (t) => {
  const root = fixture(t);
  const role = join(root, "go-b2b-starter/.agents/agents/backend-builder.md");
  const source = readFileSync(role, "utf8");
  writeFileSync(role, source + "\nRead `.agents/skills/missing/SKILL.md`.\n");
  assert.throws(() => sync(root), /ENOENT/);
  assert.equal(existsSync(join(root, "CLAUDE.md")), false);
  writeFileSync(role, source);
  sync(root);
  writeFileSync(join(root, ".codex/agents/obsolete.toml"), "old");
  assert.throws(() => sync(root, true), /Unmanaged adapter/);
});

test("generation refuses directory links and hard-linked outputs", (t) => {
  const root = fixture(t);
  const outside = mkdtempSync(join(tmpdir(), "starter outside "));
  t.after(() => rmSync(outside, { recursive: true, force: true }));
  symlinkSync(outside, join(root, ".codex"), "junction");
  assert.throws(() => sync(root), /Refusing linked/);
  assert.equal(existsSync(join(root, "CLAUDE.md")), false);
  assert.equal(existsSync(join(outside, "config.toml")), false);
  rmSync(join(root, ".codex"));
  writeFileSync(join(outside, "sentinel"), "do not replace");
  linkSync(join(outside, "sentinel"), join(root, "CLAUDE.md"));
  assert.throws(() => sync(root), /Refusing linked/);
  assert.equal(readFileSync(join(outside, "sentinel"), "utf8"), "do not replace");
  assert.throws(() => safePath(root, "../outside"), /Unsafe harness path/);
  assert.throws(() => safePath(root, "C:\\outside"), /Unsafe harness path/);
});

test("downloads verify both hashes before writing and check rejects altered cache", async (t) => {
  const root = fixture(t);
  const manifest = sources(root);
  const content = Buffer.from("Synthetic Markdown fixture\n");
  for (const skill of manifest.upstreamSkills) skill.sha256 = createHash("sha256").update(content).digest("hex");
  writeFileSync(join(root, ".agents/sources.json"), JSON.stringify(manifest));
  let calls = 0;
  await assert.rejects(fetchSkills(root, async () => new Response(++calls === 1 ? content : "bad digest")), /digest mismatch/);
  assert.equal(existsSync(join(root, "next_b2b_starter/.agents/cache")), false);
  await fetchSkills(root, async (url, options) => {
    assert.match(url, /raw\.githubusercontent\.com\/better-auth\/skills\/[a-f0-9]{40}\//);
    assert.equal(options.redirect, "error");
    return new Response(content);
  });
  sync(root);
  const cache = join(root, manifest.upstreamSkills[0].cachePath);
  writeFileSync(cache, "modified");
  assert.throws(() => sync(root, true), /Modified upstream cache/);
});

test("unapproved endpoints and unpinned skills are rejected", (t) => {
  const root = fixture(t);
  const manifest = sources(root);
  manifest.upstreamSkills[0].revision = "main";
  writeFileSync(join(root, ".agents/sources.json"), JSON.stringify(manifest));
  assert.throws(() => sources(root), /Invalid pinned/);
  cpSync(join(repository, ".agents/sources.json"), join(root, ".agents/sources.json"));
  writeFileSync(join(root, ".mcp.json"), JSON.stringify({ mcpServers: { unreviewed: { url: "https://example.com" } } }));
  assert.throws(() => sync(root, true), /adapter drift/);
});

test("tool selection produces matching host configs and keeps reviewers away from mutation tools", (t) => {
  const root = fixture(t);
  const catalog = toolCatalog(root);
  assert.equal(catalog.enabled.includes("resend"), false);
  catalog.enabled.push("resend", "linear");
  writeFileSync(join(root, ".agents/tools.json"), JSON.stringify(catalog));
  sync(root);
  const mcp = JSON.parse(readFileSync(join(root, ".mcp.json"), "utf8"));
  assert.deepEqual(mcp, mcpConfig(catalog, { allowedKinds: ["documentation", "provider"] }));
  assert.deepEqual(mcp, JSON.parse(readFileSync(join(root, ".cursor/mcp.json"), "utf8")));
  assert.equal(mcp.mcpServers.linear.url, "https://mcp.linear.app/mcp/readonly");
  assert.equal(mcp.mcpServers.shadcn, undefined);
  const frontendMcp = JSON.parse(readFileSync(join(root, "next_b2b_starter/.mcp.json"), "utf8"));
  assert.deepEqual(frontendMcp.mcpServers.shadcn.args, ["../scripts/mcp-launch.mjs", "shadcn"]);
  assert.equal(frontendMcp.mcpServers.resend, undefined);
  for (const name of ["code-reviewer", "auth-reviewer", "backend-builder"]) {
    const scope = name === "auth-reviewer" ? "next_b2b_starter/" : name === "backend-builder" ? "go-b2b-starter/" : "";
    const codex = readFileSync(join(root, `${scope}.codex/agents/${name}.toml`), "utf8");
    assert.match(codex, /\[mcp_servers.workspace-playwright\]\nenabled = false/);
    assert.match(codex, /\[mcp_servers.workspace-next-devtools\]\nenabled = false/);
    assert.match(codex, /\[mcp_servers.workspace-resend\]\nenabled = false/);
    assert.match(codex, /enabled_tools = \["resolve-library-id","query-docs"\]/);
    const claude = readFileSync(join(root, `${scope}.claude/agents/${name}.md`), "utf8").match(/^tools: (.*)$/m)[1];
    assert.doesNotMatch(claude, /mcp__(playwright|resend|linear|next-devtools)__/);
  }
  const frontend = readFileSync(join(root, "next_b2b_starter/.codex/agents/frontend-builder.toml"), "utf8");
  assert.match(frontend, /\[mcp_servers.workspace-resend\]\nenabled = false/);
  assert.match(frontend, /\[mcp_servers.workspace-playwright\]\nenabled = true\ncommand = "node"/);
});

test("Codex field inheritance enables Next tools and keeps root, Go and reviewer restrictions", (t) => {
  const root = fixture(t);
  const catalog = toolCatalog(root);
  catalog.enabled.push("resend");
  writeFileSync(join(root, ".agents/tools.json"), JSON.stringify(catalog));
  sync(root);
  // Read actual generated tables and apply Codex's field-wise parent/child merge.
  // An omitted child boolean must inherit the parent value, never silently reset.
  const layer = (path) => Object.fromEntries([...readFileSync(join(root, path), "utf8").matchAll(/\[mcp_servers\.([^\]]+)\]\n([^[]*)/g)].map(([, name, body]) => {
    const fields = Object.fromEntries([...body.matchAll(/^(enabled|command|url) = (.+)$/gm)].map(([, key, value]) => [key, JSON.parse(value)]));
    return [name, fields];
  }));
  const merge = (...layers) => {
    const result = {};
    for (const source of layers) for (const [name, fields] of Object.entries(source)) result[name] = { ...result[name], ...fields };
    return result;
  };
  const rootLayer = layer(".codex/config.toml");
  const nextLayer = layer("next_b2b_starter/.codex/config.toml");
  const goLayer = layer("go-b2b-starter/.codex/config.toml");
  const nextState = merge(rootLayer, nextLayer);
  for (const id of ["workspace-next-devtools", "workspace-shadcn", "workspace-playwright"]) {
    assert.equal(rootLayer[id].enabled, false);
    assert.equal(merge(rootLayer, goLayer)[id].enabled, false);
    assert.equal(nextState[id].enabled, true);
    assert.equal(merge(nextState, layer("next_b2b_starter/.codex/agents/frontend-builder.toml"))[id].enabled, true);
    assert.equal(merge(nextState, layer("next_b2b_starter/.codex/agents/auth-reviewer.toml"))[id].enabled, false);
    assert.equal(merge(nextState, layer(".codex/agents/code-reviewer.toml"))[id].enabled, false);
  }
  assert.equal(rootLayer["workspace-resend"].enabled, true);
  assert.equal(nextState["workspace-resend"].enabled, false);
  assert.equal(merge(rootLayer, goLayer)["workspace-resend"].enabled, false);
  for (const path of [".codex/config.toml", "go-b2b-starter/.codex/config.toml", "next_b2b_starter/.codex/config.toml", ".codex/agents/orchestrator.toml", ".codex/agents/code-reviewer.toml", "go-b2b-starter/.codex/agents/backend-builder.toml", "next_b2b_starter/.codex/agents/auth-reviewer.toml", "next_b2b_starter/.codex/agents/frontend-builder.toml", "next_b2b_starter/.codex/agents/quality-engineer.toml"]) {
    for (const fields of Object.values(layer(path))) {
      assert.equal(typeof fields.enabled, "boolean");
      assert.equal(typeof (fields.command || fields.url), "string", `Missing standalone transport in ${path}`);
    }
    if (path.includes("/agents/")) assert.equal(merge(rootLayer, layer(path))["workspace-resend"].enabled, false);
  }
});

test("tool catalog rejects floating pins, shell metacharacters, credentials and unknown selections before writes", (t) => {
  const root = fixture(t);
  const original = toolCatalog(root);
  for (const mutate of [
    (c) => { c.servers.shadcn.version = "latest"; },
    (c) => { c.servers.shadcn.args.push("x;unexpected"); },
    (c) => { c.servers.shadcn.cwd = "../outside"; },
    (c) => { c.servers.resend.headers = { Authorization: "fixture-only" }; },
    (c) => { c.servers.resend.url = "https://user:fixture@example.com/mcp"; },
    (c) => { c.servers.resend.url += "?key=fixture"; },
    (c) => { c.enabled.push("missing"); },
    (c) => { c.enabled.push(c.enabled[0]); },
  ]) {
    const catalog = structuredClone(original); mutate(catalog);
    writeFileSync(join(root, ".agents/tools.json"), JSON.stringify(catalog));
    assert.throws(() => sync(root));
    assert.equal(existsSync(join(root, "CLAUDE.md")), false);
  }
});

test("local MCP launcher resolves the frontend from its own path and propagates exit status", (t) => {
  const root = fixture(t);
  cpSync(join(repository, "scripts/mcp-launch.mjs"), join(root, "scripts/mcp-launch.mjs"));
  const bin = join(root, "test bin"); mkdirSync(bin);
  const reporter = `console.log(JSON.stringify({ cwd: process.cwd(), args: process.argv.slice(2) })); process.exit(17);\n`;
  if (process.platform === "win32") {
    writeFileSync(join(bin, "report.mjs"), reporter);
    writeFileSync(join(bin, "npx.cmd"), '@node "%~dp0report.mjs" %*\r\n');
  } else {
    writeFileSync(join(bin, "npx"), '#!/usr/bin/env node\n' + reporter);
    chmodSync(join(bin, "npx"), 0o755);
  }
  const pathKey = Object.keys(process.env).find((key) => key.toLowerCase() === "path") || "PATH";
  const run = (id) => spawnSync(process.execPath, [join(root, "scripts/mcp-launch.mjs"), id], {
    cwd: tmpdir(), encoding: "utf8", env: { ...process.env, [pathKey]: `${bin}${delimiter}${process.env[pathKey]}` }, timeout: 5000,
  });
  const result = run("shadcn");
  assert.equal(result.status, 17, result.stderr);
  assert.deepEqual(JSON.parse(result.stdout), { cwd: realpathSync(join(root, "next_b2b_starter")), args: ["--yes", `shadcn@${toolCatalog(root).servers.shadcn.version}`, "mcp"] });
  const disabled = run("magicui");
  assert.equal(disabled.status, 1);
  assert.equal(disabled.stdout, "");
  assert.match(disabled.stderr, /enabled local tool/);
  sync(root);
  for (const [scope, id] of [["go-b2b-starter", "context7"], ["next_b2b_starter", "shadcn"]]) {
    const config = JSON.parse(readFileSync(join(root, scope, ".mcp.json"), "utf8"));
    const result = spawnSync(process.execPath, config.mcpServers[id].args, {
      cwd: join(root, scope), encoding: "utf8", env: { ...process.env, [pathKey]: `${bin}${delimiter}${process.env[pathKey]}` }, timeout: 5000,
    });
    assert.equal(result.status, 17, result.stderr);
    assert.equal(JSON.parse(result.stdout).cwd, realpathSync(join(root, toolCatalog(root).servers[id].cwd)));
  }
});

function mcpFixture({ extraTool = false, documentError = false } = {}) {
  const methods = [];
  return { methods, fetcher: async (url, options) => {
    assert.equal(url, "https://mcp.better-auth.com/mcp");
    const request = JSON.parse(options.body);
    methods.push(request.method);
    if (methods.length > 1) assert.equal(options.headers["Mcp-Session-Id"], "test-session");
    let result;
    if (request.method === "initialize") result = { protocolVersion: "2025-03-26", capabilities: { tools: {} }, serverInfo: { name: "test", version: "1" } };
    else if (request.method === "notifications/initialized") { assert.equal(request.id, undefined); return new Response(null, { status: 202 }); }
    else if (request.method === "tools/list") result = { tools: [...["get_doc", "search_docs"], ...(extraTool ? ["mutate"] : [])].map((name) => ({ name, annotations: { readOnlyHint: true, destructiveHint: false } })) };
    else { assert.equal(request.params.name, "get_doc"); assert.deepEqual(request.params.arguments, { path: "/llms.txt" }); result = { isError: documentError, content: [{ type: "text", text: "https://better-auth.com/llms.txt" }] }; }
    return Response.json({ jsonrpc: "2.0", id: request.id, result }, { headers: { "mcp-session-id": "test-session" } });
  } };
}

test("MCP probe completes initialization and only calls the public documentation reader", async (t) => {
  const root = fixture(t);
  const server = mcpFixture();
  assert.deepEqual(await checkMcp(root, server.fetcher), ["get_doc", "search_docs"]);
  assert.deepEqual(server.methods, ["initialize", "notifications/initialized", "tools/list", "tools/call"]);
});

test("MCP probe fails on changed capabilities and document errors", async (t) => {
  const root = fixture(t);
  const changed = mcpFixture({ extraTool: true });
  await assert.rejects(checkMcp(root, changed.fetcher), /tools changed/);
  assert.equal(changed.methods.includes("tools/call"), false);
  await assert.rejects(checkMcp(root, mcpFixture({ documentError: true }).fetcher), /did not return/);
});
