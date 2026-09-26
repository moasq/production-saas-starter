import { createHash } from "node:crypto";
import { lstatSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, isAbsolute, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { toolCatalog, mcpConfig, codexServers } from "./tool-catalog.mjs";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const endpoint = "https://mcp.better-auth.com/mcp";
const digest = (value) => createHash("sha256").update(value).digest("hex");

// Validate every path component before a read or write. Never follow an output
// symlink/junction or overwrite a hard-linked file outside this checkout.
export function safePath(base, relative) {
  if (isAbsolute(relative) || relative.includes("\\") || relative.split("/").some((s) => !s || s === "." || s === "..")) {
    throw new Error(`Unsafe harness path: ${relative}`);
  }
  let path = base;
  const parts = relative.split("/");
  for (const [index, part] of parts.entries()) {
    path = join(path, part);
    let stat;
    try { stat = lstatSync(path); } catch (error) { if (error.code === "ENOENT") continue; throw error; }
    if (stat.isSymbolicLink() || (!stat.isDirectory() && (!stat.isFile() || stat.nlink > 1))) {
      throw new Error(`Refusing linked or special harness path: ${relative}`);
    }
    if (index < parts.length - 1 && !stat.isDirectory()) throw new Error(`Not a directory: ${path}`);
  }
  return path;
}

function read(base, path) { return readFileSync(safePath(base, path), "utf8").replaceAll("\r\n", "\n"); }
function entries(base, path) { return readdirSync(safePath(base, path), { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name)); }

// Canonical frontmatter deliberately uses one JSON-quoted scalar per line (valid
// YAML). This avoids a YAML dependency and fails closed on unsupported syntax.
export function parseBrief(source, label) {
  const match = source.match(/^---\n([\s\S]+?)\n---\n([\s\S]+)$/);
  if (!match) throw new Error(`${label}: missing frontmatter or body`);
  const fields = {};
  for (const line of match[1].split("\n")) {
    const field = line.match(/^([a-z]+): (".*")$/);
    if (!field || field[1] in fields) throw new Error(`${label}: use unique JSON-quoted frontmatter scalars`);
    fields[field[1]] = JSON.parse(field[2]);
  }
  if (!/^[a-z][a-z0-9-]{2,49}$/.test(fields.name) || !fields.description?.trim()) throw new Error(`${label}: invalid name/description`);
  return { ...fields, body: match[2].trim() + "\n" };
}

// Source bodies use repository-relative paths; module entry points use their own cwd.
export function checkReferences(base, text, sourcePath, scope = "") {
  const paths = new Set();
  for (const [, path] of text.matchAll(/`((?:\.\.\/)?(?:(?:\.agents\/|go-b2b-starter\/|next_b2b_starter\/|docs\/|scripts\/)[^`\s]+|AGENTS\.md|README\.md|SETUP\.md))`/g)) {
    if (path.includes("/.agents/cache/") || path.startsWith(".agents/cache/")) continue;
    paths.add(relative(base, resolve(base, scope, path)).replaceAll("\\", "/"));
  }
  for (const [, href] of text.matchAll(/\[[^\]]*\]\(([^)\s]+)\)/g)) {
    if (/^(?:https?:|#)/.test(href)) continue;
    const target = href.split("#")[0];
    if (target) paths.add(relative(base, resolve(base, dirname(sourcePath), target)).replaceAll("\\", "/"));
  }
  for (const path of paths) lstatSync(safePath(base, path.replace(/\/$/, "")));
  return [...paths].sort();
}

// Validate command targets without executing tooling or application operations.
// This covers the documented local Node/shell scripts and package/Make targets;
// external CLI behavior still needs its actual verification evidence.
export function checkCommands(base, text, scope = "") {
  text = [...text.matchAll(/`+([^`]+)`+/g)].map((match) => match[1]).join("\n");
  for (const [, commandPath] of text.matchAll(/(?:node |sh |(?<![.\w/])\.\/)((?:\.\.\/)?scripts\/[a-zA-Z0-9./-]+\.(?:mjs|sh))/g)) {
    const path = relative(base, resolve(base, scope, commandPath)).replaceAll("\\", "/");
    lstatSync(safePath(base, path));
  }
  for (const [, command] of text.matchAll(/pnpm (?:--dir next_b2b_starter )?([a-z][a-z0-9:-]*)/g)) {
    if (["install", "audit"].includes(command)) continue;
    const pkg = JSON.parse(read(base, "next_b2b_starter/package.json"));
    if (!Object.hasOwn(pkg.scripts, command)) throw new Error(`Unknown documented pnpm command: ${command}`);
  }
  for (const [, command] of text.matchAll(/make (?:-C go-b2b-starter )?([a-z][a-z0-9-]*)/g)) {
    if (!new RegExp(`^${command}:`, "m").test(read(base, "go-b2b-starter/Makefile"))) throw new Error(`Unknown documented Make target: ${command}`);
  }
}

export const scopes = [
  { path: "", roles: ["code-reviewer", "orchestrator"], skills: ["dev-tools", "orchestration", "pr-review", "service-connections"], kinds: ["documentation", "provider"] },
  { path: "go-b2b-starter", roles: ["backend-builder"], skills: ["go-backend"], kinds: ["documentation"] },
  { path: "next_b2b_starter", roles: ["auth-reviewer", "frontend-builder", "quality-engineer"], skills: ["auth-integration", "frontend-tools", "next-frontend"], kinds: ["documentation", "development"] },
];
const scoped = (scope, path) => scope ? `${scope}/${path}` : path;

export function sources(base) {
  const data = JSON.parse(read(base, ".agents/sources.json"));
  if (data.schemaVersion !== 2 || data.mcp?.url !== endpoint || data.mcp.server !== "better-auth" ||
      JSON.stringify(data.mcp.allowedTools) !== JSON.stringify(["get_doc", "search_docs"]) ||
      !Array.isArray(data.authoredHere) || !Array.isArray(data.upstreamSkills)) throw new Error("Unexpected documentation MCP or provenance contract");
  const seen = new Set();
  for (const skill of data.upstreamSkills) {
    if (!/^[a-f0-9]{40}$/.test(skill.revision) || !/^[a-f0-9]{64}$/.test(skill.sha256) ||
        !["best-practices", "organization"].includes(skill.name) || seen.has(skill.name) ||
        skill.repository !== "https://github.com/better-auth/skills" ||
        skill.sourcePath !== `better-auth/${skill.name}/SKILL.md` ||
        skill.cachePath !== `next_b2b_starter/.agents/cache/better-auth/${skill.name}/SKILL.md` ||
        skill.url !== `https://raw.githubusercontent.com/better-auth/skills/${skill.revision}/${skill.sourcePath}` ||
        skill.redistributed !== false) throw new Error("Invalid pinned upstream skill");
    seen.add(skill.name);
    safePath(base, skill.cachePath);
  }
  if (seen.size !== 2) throw new Error("Both pinned Better Auth skills are required");
  return data;
}

export function adapterPlan(base) {
  const manifest = sources(base);
  const catalog = toolCatalog(base);
  if (catalog.servers["better-auth"]?.url !== endpoint ||
      JSON.stringify(catalog.servers["better-auth"].tools) !== JSON.stringify(manifest.mcp.allowedTools)) throw new Error("Unexpected Better Auth documentation contract");
  const generated = "# Generated by node scripts/harness.mjs sync. Edit canonical sources instead.\n";
  const outputs = new Map();
  const skills = [];
  const allNames = new Set();
  for (const scope of scopes) {
    const at = (path) => scoped(scope.path, path);
    const launcherPath = `${scope.path ? "../" : ""}scripts/mcp-launch.mjs`;
    const options = { launcherPath, allowedKinds: scope.kinds };
    const instructions = at("AGENTS.md");
    checkReferences(base, read(base, instructions), instructions, scope.path);
    checkCommands(base, read(base, instructions), scope.path);
    outputs.set(at("CLAUDE.md"), `<!-- Generated by node scripts/harness.mjs sync. -->\n@AGENTS.md\n`);
    if (scope.path) outputs.set(at(".claude/CLAUDE.md"), `<!-- Generated by node scripts/harness.mjs sync. -->\n@../AGENTS.md\n`);
    const mcp = JSON.stringify(mcpConfig(catalog, options), null, 2) + "\n";
    outputs.set(at(".mcp.json"), mcp);
    outputs.set(at(".cursor/mcp.json"), mcp);
    outputs.set(at(".codex/config.toml"), `${generated}\n${codexServers(catalog, false, options)}\n`);
    const roleNames = [];
    for (const item of entries(base, at(".agents/agents"))) {
      if (!item.isFile() || !item.name.endsWith(".md")) throw new Error(`Unexpected role entry: ${item.name}`);
      const path = at(`.agents/agents/${item.name}`);
      const source = read(base, path);
      checkReferences(base, source, path);
      checkCommands(base, source);
      const role = parseBrief(source, path);
      if (item.name !== `${role.name}.md` || !["write", "read-only"].includes(role.mode)) throw new Error(`Invalid role: ${path}`);
      roleNames.push(role.name);
      const readonly = role.mode === "read-only";
      // Role tools never include provider administration. Root coordinates; Go
      // uses source/docs; only Next.js implementation/QA gets development tools.
      const allowedKinds = !readonly && scope.path === "next_b2b_starter" ? ["documentation", "development"] : ["documentation"];
      const selected = catalog.enabled.filter((id) => allowedKinds.includes(catalog.servers[id].kind));
      const tools = ["Read", "Grep", "Glob", "WebFetch", ...(!readonly ? ["Edit", "Write", "Bash"] : []),
        ...selected.flatMap((id) => (catalog.servers[id].tools || ["*"]).map((name) => `mcp__${id}__${name}`))];
      outputs.set(at(`.claude/agents/${role.name}.md`), `---\nname: ${JSON.stringify(role.name)}\ndescription: ${JSON.stringify(role.description)}\nmodel: inherit\ncolor: ${role.color}\ntools: ${tools.join(", ")}\n---\n\n<!-- Generated from ${path}; do not edit. -->\n${role.body}`);
      outputs.set(at(`.codex/agents/${role.name}.toml`), `${generated}name = ${JSON.stringify(role.name)}\ndescription = ${JSON.stringify(role.description)}\n${readonly ? 'sandbox_mode = "read-only"\n' : ""}developer_instructions = ${JSON.stringify(role.body)}\n\n${codexServers(catalog, readonly, { launcherPath, allowedKinds })}\n`);
    }
    if (JSON.stringify(roleNames.sort()) !== JSON.stringify(scope.roles)) throw new Error(`Unexpected role ownership in ${scope.path || "root"}`);
    const skillNames = [];
    for (const item of entries(base, at(".agents/skills"))) {
      if (!item.isDirectory()) throw new Error(`Unexpected skill entry: ${item.name}`);
      const path = at(`.agents/skills/${item.name}/SKILL.md`);
      const source = read(base, path);
      const references = checkReferences(base, source, path);
      checkCommands(base, source);
      const skill = parseBrief(source, path);
      if (skill.name !== item.name || allNames.has(skill.name)) throw new Error(`Duplicate or mismatched skill: ${path}`);
      allNames.add(skill.name);
      skillNames.push(skill.name);
      skills.push(path);
      const provenance = manifest.authoredHere.find((entry) => entry.path === path);
      if (!provenance || provenance.name !== skill.name || provenance.source !== "authored-here" || provenance.license !== "MIT" ||
          !/^\d+\.\d+\.\d+$/.test(provenance.version) || !/^\d{4}-\d{2}-\d{2}$/.test(provenance.verifiedAt) ||
          !provenance.adaptations?.trim() || provenance.sha256 !== digest(source) ||
          JSON.stringify(provenance.references) !== JSON.stringify(references)) throw new Error(`Stale or missing skill provenance: ${path}`);
      outputs.set(at(`.claude/skills/${item.name}/SKILL.md`), source.replace(/\n---\n/, `\n---\n\n<!-- Generated from ${path}; do not edit. -->\n`));
    }
    if (JSON.stringify(skillNames.sort()) !== JSON.stringify(scope.skills)) throw new Error(`Unexpected skill ownership in ${scope.path || "root"}`);
  }
  if (JSON.stringify(skills.sort()) !== JSON.stringify(manifest.authoredHere.map((entry) => entry.path).sort())) throw new Error("Skill provenance inventory is incomplete");
  // Preflight ALL destinations before changing ANY files. Stale adapter files
  // are never removed implicitly: a scope migration must explicitly remove them.
  for (const path of outputs.keys()) safePath(base, path);
  for (const scope of scopes) for (const suffix of [".codex/agents", ".claude/agents", ".claude/skills"]) {
    const directory = scoped(scope.path, suffix);
    let existing;
    try { existing = entries(base, directory); } catch (error) { if (error.code === "ENOENT") continue; throw error; }
    for (const item of existing) {
      const path = `${directory}/${item.name}`;
      const expected = directory.endsWith("skills") ? `${path}/SKILL.md` : path;
      if (!outputs.has(expected)) throw new Error(`Unmanaged adapter: ${path}. Move its source into the matching project first.`);
      if (directory.endsWith("skills")) for (const file of entries(base, path)) {
        if (file.name !== "SKILL.md") throw new Error(`Unmanaged skill adapter: ${path}/${file.name}`);
      }
    }
  }
  return outputs;
}

export function sync(base, check = false) {
  const plan = adapterPlan(base);
  const drift = [];
  for (const [path, content] of plan) {
    let actual;
    try { actual = read(base, path); } catch (error) { if (error.code !== "ENOENT") throw error; }
    if (actual !== content) drift.push(path);
  }
  if (check && drift.length) throw new Error(`Generated adapter drift: ${drift.join(", ")}. Run node scripts/harness.mjs sync.`);
  if (!check) for (const path of drift) {
    const destination = safePath(base, path);
    mkdirSync(dirname(destination), { recursive: true });
    writeFileSync(destination, plan.get(path));
  }
  for (const skill of sources(base).upstreamSkills) {
    let cached;
    try { cached = readFileSync(safePath(base, skill.cachePath)); } catch (error) { if (error.code === "ENOENT") continue; throw error; }
    if (digest(cached) !== skill.sha256) throw new Error(`Modified upstream cache: ${skill.cachePath}`);
  }
  return plan.size;
}

export async function fetchSkills(base, fetcher = fetch) {
  const { upstreamSkills } = sources(base);
  // Download and verify both files before replacing any cache content.
  const contents = await Promise.all(upstreamSkills.map(async (skill) => {
    const response = await fetcher(skill.url, { redirect: "error", signal: AbortSignal.timeout(20000) });
    if (!response.ok) throw new Error(`Skill download failed (${response.status})`);
    const bytes = Buffer.from(await response.arrayBuffer());
    if (bytes.length > 1024 * 1024 || digest(bytes) !== skill.sha256) throw new Error(`Upstream skill digest mismatch: ${skill.name}`);
    return bytes;
  }));
  upstreamSkills.forEach((skill, index) => {
    const destination = safePath(base, skill.cachePath);
    mkdirSync(dirname(destination), { recursive: true });
    writeFileSync(destination, contents[index]);
  });
}

export async function checkMcp(base, fetcher = fetch) {
  const { mcp } = sources(base);
  let session;
  let id = 0;
  async function rpc(method, params) {
    const notification = method.startsWith("notifications/");
    const requestId = ++id;
    const response = await fetcher(mcp.url, { method: "POST", redirect: "error", signal: AbortSignal.timeout(20000),
      headers: { "Content-Type": "application/json", Accept: "application/json, text/event-stream", "MCP-Protocol-Version": "2025-03-26", ...(session ? { "Mcp-Session-Id": session } : {}) },
      body: JSON.stringify({ jsonrpc: "2.0", ...(!notification ? { id: requestId } : {}), method, params }) });
    if (!response.ok) throw new Error(`Documentation MCP failed (${response.status})`);
    session = response.headers.get("mcp-session-id") || session;
    if (notification) return {};
    const raw = await response.text();
    const payloads = response.headers.get("content-type")?.includes("text/event-stream")
      ? raw.split(/\r?\n/).filter((line) => line.startsWith("data: ")).map((line) => JSON.parse(line.slice(6))) : [JSON.parse(raw)];
    const result = payloads.find((payload) => payload.id === requestId);
    if (!result?.result || result.error) throw new Error(`Invalid documentation MCP result for ${method}`);
    return result.result;
  }
  await rpc("initialize", { protocolVersion: "2025-03-26", capabilities: {}, clientInfo: { name: "starter-harness-check", version: "1.0.0" } });
  await rpc("notifications/initialized", {});
  const result = await rpc("tools/list", {});
  const listed = result.tools?.map((tool) => tool.name).sort();
  if (JSON.stringify(listed) !== JSON.stringify(mcp.allowedTools) || result.tools.some((tool) => tool.annotations?.readOnlyHint !== true || tool.annotations?.destructiveHint !== false)) {
    throw new Error("Documentation MCP tools changed; review before use");
  }
  const document = await rpc("tools/call", { name: "get_doc", arguments: { path: "/llms.txt" } });
  if (document.isError || !document.content?.some((item) => item.type === "text" && item.text.includes("https://better-auth.com/"))) throw new Error("Documentation MCP did not return the version index");
  return listed;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const command = process.argv[2];
    if (process.argv.length !== 3) throw new Error("Usage: node scripts/harness.mjs sync|check|tools|fetch-skills|check-mcp");
    if (command === "sync" || command === "check") console.log(`${command}: ${sync(root, command === "check")} generated adapters verified`);
    else if (command === "fetch-skills") { await fetchSkills(root); console.log("Fetched two hash-verified official skills into the ignored local cache."); }
    else if (command === "check-mcp") console.log(`Read-only documentation MCP verified: ${(await checkMcp(root)).join(", ")}`);
    else if (command === "tools") {
      const catalog = toolCatalog(root);
      for (const [id, server] of Object.entries(catalog.servers)) console.log(`${catalog.enabled.includes(id) ? "enabled" : "optional"}\t${id}\t${server.version || server.auth}\t${server.purpose}`);
    }
    else throw new Error(`Unknown harness command: ${command}`);
  } catch (error) { console.error(error.message); process.exitCode = 1; }
}
