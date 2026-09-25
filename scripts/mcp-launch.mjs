import { spawn } from "node:child_process";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { toolCatalog } from "./tool-catalog.mjs";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
try {
  const id = process.argv[2];
  const catalog = toolCatalog(root);
  const server = catalog.servers[id];
  if (process.argv.length !== 3 || !catalog.enabled.includes(id) || !server?.package) throw new Error("Expected an enabled local tool ID from .agents/tools.json");
  // All shell arguments are constrained by the catalog validator. Windows needs
  // cmd to resolve npx.cmd; the workspace path is a cwd option, never shell text.
  const child = spawn("npx", ["--yes", `${server.package}@${server.version}`, ...server.args], {
    cwd: resolve(root, server.cwd), stdio: "inherit", shell: process.platform === "win32",
  });
  child.on("error", () => { console.error("Could not start the pinned MCP tool. Install Node.js with npm/npx."); process.exitCode = 1; });
  child.on("exit", (code) => { process.exitCode = code ?? 1; });
  for (const signal of ["SIGINT", "SIGTERM"]) process.on(signal, () => child.kill(signal));
} catch (error) { console.error(error.message); process.exitCode = 1; }
