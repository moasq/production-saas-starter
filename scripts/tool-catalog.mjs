import { readFileSync, lstatSync } from "node:fs";
import { join } from "node:path";

// This file only reads checked-in, non-secret tool configuration. It never logs in,
// installs a package, changes global host settings, or provisions an app service.
export function toolCatalog(base) {
  const path = join(base, ".agents/tools.json");
  for (const part of [join(base, ".agents"), path]) {
    const stat = lstatSync(part);
    if (stat.isSymbolicLink() || (stat.isFile() && stat.nlink > 1)) throw new Error("Linked tool catalog is not supported");
  }
  const catalog = JSON.parse(readFileSync(path, "utf8"));
  if (catalog.schemaVersion !== 1 || !/^[a-f0-9]{40}$/.test(catalog.source?.revision) ||
      catalog.source.repository !== "https://github.com/moasq/agentic-ship" ||
      !Array.isArray(catalog.enabled) || new Set(catalog.enabled).size !== catalog.enabled.length ||
      !catalog.servers || typeof catalog.servers !== "object" || Array.isArray(catalog.servers)) throw new Error("Invalid tool catalog");
  for (const [id, server] of Object.entries(catalog.servers)) {
    if (!/^[a-z0-9][a-z0-9-]+$/.test(id) || !["documentation", "development", "provider"].includes(server.kind) ||
        !server.purpose || !["none", "host-oauth"].includes(server.auth) || !validUrl(server.documentation)) throw new Error(`Invalid tool entry: ${id}`);
    if (server.url) {
      if (!validUrl(server.url) || server.package || server.version || server.args || server.cwd) throw new Error(`Invalid remote tool: ${id}`);
    } else if (!/^(@[a-z0-9-]+\/)?[a-z0-9-]+$/.test(server.package) || !/^\d+\.\d+\.\d+$/.test(server.version) ||
        !Array.isArray(server.args) || server.args.some((arg) => typeof arg !== "string" || !/^[a-zA-Z0-9@._:/=-]+$/.test(arg)) ||
        ![".", "next_b2b_starter"].includes(server.cwd) || server.auth !== "none") throw new Error(`Tool executable must use an exact pin and safe arguments: ${id}`);
    if ((server.kind === "documentation" || server.tools !== undefined) && (!Array.isArray(server.tools) || !server.tools.length ||
        server.tools.some((name) => typeof name !== "string" || !/^[a-zA-Z0-9_-]+$/.test(name)))) throw new Error(`Documentation tools need an explicit allowlist: ${id}`);
    const allowed = new Set(["purpose", "kind", "url", "package", "version", "args", "cwd", "documentation", "tools", "auth"]);
    if (Object.keys(server).some((key) => !allowed.has(key))) throw new Error(`Unsupported tool field: ${id}`);
  }
  for (const id of catalog.enabled) if (!Object.hasOwn(catalog.servers, id)) throw new Error(`Unknown enabled tool: ${id}`);
  return catalog;
}

function validUrl(value) {
  try { const url = new URL(value); return url.protocol === "https:" && !url.username && !url.password && (!url.search || url.search === "?readonly=true") && !url.hash; }
  catch { return false; }
}

export function mcpConfig(catalog, { launcherPath = "scripts/mcp-launch.mjs", allowedKinds = ["documentation", "development", "provider"] } = {}) {
  return { mcpServers: Object.fromEntries(catalog.enabled.filter((id) => allowedKinds.includes(catalog.servers[id].kind)).map((id) => {
    const server = catalog.servers[id];
    return [id, server.url ? { type: "http", url: server.url } : { command: "node", args: [launcherPath, id] }];
  })) };
}

export function codexName(id) { return id === "better-auth" ? id : `workspace-${id}`; }

export function codexServers(catalog, reviewer = false, { launcherPath = "scripts/mcp-launch.mjs", allowedKinds = reviewer ? ["documentation"] : ["documentation", "development", "provider"] } = {}) {
  return catalog.enabled.map((id) => {
    const server = catalog.servers[id];
    const lines = [`[mcp_servers.${codexName(id)}]`];
    if (!allowedKinds.includes(server.kind)) return [...lines, "enabled = false"].join("\n");
    if (server.url) lines.push(`url = ${JSON.stringify(server.url)}`);
    else lines.push('command = "node"', `args = ${JSON.stringify([launcherPath, id])}`, "startup_timeout_sec = 120");
    if (server.tools) lines.push(`enabled_tools = ${JSON.stringify(server.tools)}`);
    return lines.join("\n");
  }).join("\n\n");
}
