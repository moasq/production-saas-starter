import { spawn } from "node:child_process";
import { realpathSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { toolCatalog } from "./tool-catalog.mjs";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const protocolVersions = ["2025-06-18", "2025-03-26", "2024-11-05"];
const toolName = /^[a-zA-Z0-9_.-]{1,128}$/;
const object = (value) => value !== null && typeof value === "object" && !Array.isArray(value);
const sleep = (milliseconds) => new Promise((done) => setTimeout(done, milliseconds));
class ProbeError extends Error {}

async function waitForClose(closed, milliseconds) {
  let timer;
  try { await Promise.race([closed, new Promise((done) => { timer = setTimeout(done, milliseconds); })]); }
  finally { clearTimeout(timer); }
}

async function terminate(child, closed, graceMs) {
  child.stdin.end();
  // A launcher can outlive (or exit before) its descendants. On POSIX, signal
  // the whole private process group even if the immediate child already exited.
  if (process.platform !== "win32") {
    const signal = async (name) => {
      if (!child.pid) return;
      for (let attempt = 0; ; attempt++) {
        try { process.kill(-child.pid, name); return; }
        catch (error) {
          if (error.code === "ESRCH") return;
          // Darwin can briefly report EPERM for an exiting group awaiting
          // reaping. Yield for that transition, but never hide persistent EPERM.
          if (error.code === "EPERM" && attempt < 5) { await sleep(20); continue; }
          throw new ProbeError("Could not stop the MCP process group");
        }
      }
    };
    await waitForClose(closed, graceMs);
    await signal("SIGTERM");
    await sleep(graceMs);
    await signal("SIGKILL");
  } else if (child.pid) {
    // Node has no Windows process-group kill. taskkill's /T includes children;
    // argument arrays avoid passing a configurable command through a shell.
    await new Promise((done) => {
      const killer = spawn("taskkill.exe", ["/PID", String(child.pid), "/T", "/F"], { stdio: "ignore", windowsHide: true });
      const timeout = setTimeout(() => { killer.kill(); child.kill(); done(); }, 2000);
      killer.once("error", () => { clearTimeout(timeout); child.kill(); done(); });
      killer.once("close", () => { clearTimeout(timeout); done(); });
    });
  }
  await waitForClose(closed, 1000);
  child.stdin.destroy();
  child.stdout.destroy();
  child.stderr.destroy();
}

// This probes protocol transport, not package safety. The executable still has
// the invoking user's filesystem, environment and network permissions. Only
// initialize, initialized and tools/list are initiated; no tool is ever called.
export async function probeStdio(command, args, options = {}) {
  const { cwd, env = process.env, signal, allowedTools, allowlistMode = "exact", timeoutMs = 120000,
    maxBufferBytes = 1024 * 1024, shutdownGraceMs = 250 } = options;
  if (typeof command !== "string" || !command || !Array.isArray(args) || args.some((arg) => typeof arg !== "string") ||
      !Number.isInteger(timeoutMs) || timeoutMs < 1 || timeoutMs > 120000 ||
      !Number.isInteger(maxBufferBytes) || maxBufferBytes < 1 || maxBufferBytes > 1024 * 1024 ||
      !Number.isInteger(shutdownGraceMs) || shutdownGraceMs < 1 || shutdownGraceMs > 1000 ||
      (signal !== undefined && !(signal instanceof AbortSignal)) ||
      !["exact", "subset"].includes(allowlistMode) ||
      (allowedTools !== undefined && (!Array.isArray(allowedTools) || allowedTools.some((name) => typeof name !== "string" || !toolName.test(name)) ||
        new Set(allowedTools).size !== allowedTools.length))) throw new ProbeError("Invalid MCP probe options");
  if (signal?.aborted) throw new ProbeError("MCP probe interrupted");

  let child;
  try {
    child = spawn(command, args, { cwd, env, stdio: ["pipe", "pipe", "pipe"],
      detached: process.platform !== "win32", windowsHide: true, shell: false });
  } catch { throw new ProbeError("Could not start the MCP process"); }
  const closed = new Promise((done) => child.once("close", done));
  let stopped = false;
  let failure;
  let failPromise;
  const failed = new Promise((_, reject) => { failPromise = reject; });
  let pending;
  let sequence = 0;
  let receivedBytes = 0;
  let buffer = Buffer.alloc(0);
  const fail = (message) => {
    if (failure || stopped) return;
    failure = new ProbeError(message);
    failPromise(failure);
    pending?.reject(failure);
  };
  const write = (message) => {
    if (failure || stopped) return;
    child.stdin.write(`${JSON.stringify(message)}\n`, (error) => { if (error) fail("MCP input stream closed unexpectedly"); });
  };
  const request = (method, params) => new Promise((done, reject) => {
    if (failure) { reject(failure); return; }
    const id = ++sequence;
    pending = { id, done, reject };
    write({ jsonrpc: "2.0", id, method, params });
  });
  function receive(message) {
    if (!object(message) || message.jsonrpc !== "2.0") throw new ProbeError("Invalid MCP JSON-RPC message");
    if (Object.hasOwn(message, "method")) {
      if (typeof message.method !== "string" || !message.method || Object.hasOwn(message, "result") || Object.hasOwn(message, "error") ||
          (message.params !== undefined && !object(message.params))) throw new ProbeError("Invalid MCP request or notification");
      if (Object.hasOwn(message, "id")) {
        if (typeof message.id !== "string" && !Number.isInteger(message.id)) throw new ProbeError("Invalid MCP request identifier");
        // Do not grant roots, sampling, elicitation or other client actions.
        write({ jsonrpc: "2.0", id: message.id, ...(message.method === "ping" ? { result: {} }
          : { error: { code: -32601, message: "Client method not supported" } }) });
      }
      return;
    }
    if (!pending || message.id !== pending.id || Object.hasOwn(message, "result") === Object.hasOwn(message, "error")) {
      throw new ProbeError("Unexpected MCP response");
    }
    if (Object.hasOwn(message, "error")) throw new ProbeError("MCP server returned a protocol error");
    if (!object(message.result)) throw new ProbeError("Invalid MCP result");
    const current = pending;
    pending = undefined;
    current.done(message.result);
  }
  function count(chunk) {
    receivedBytes += chunk.length;
    if (receivedBytes > maxBufferBytes) { fail("MCP output exceeded the byte limit"); return false; }
    return !failure && !stopped;
  }
  child.stdout.on("data", (chunk) => {
    if (!count(chunk)) return;
    buffer = Buffer.concat([buffer, chunk]);
    let newline;
    while (!failure && (newline = buffer.indexOf(10)) !== -1) {
      const line = buffer.subarray(0, newline);
      buffer = buffer.subarray(newline + 1);
      try { receive(JSON.parse(new TextDecoder("utf-8", { fatal: true }).decode(line))); }
      catch (error) { fail(error instanceof ProbeError ? error.message : "MCP stdout contained invalid JSON"); }
    }
  });
  // Drain stderr without retaining or printing package logs, which may be secret.
  child.stderr.on("data", count);
  child.once("error", () => fail("Could not start the MCP process"));
  child.once("exit", () => fail("MCP process exited before the probe completed"));
  child.stdout.once("end", () => fail("MCP output stream closed before the probe completed"));
  child.stdin.on("error", () => fail("MCP input stream closed unexpectedly"));
  child.stdout.on("error", () => fail("MCP output stream failed"));
  child.stderr.on("error", () => fail("MCP error stream failed"));
  const timer = setTimeout(() => fail("MCP probe timed out"), timeoutMs);
  const abort = () => fail("MCP probe interrupted");
  signal?.addEventListener("abort", abort, { once: true });

  try {
    const exchange = async () => {
      const initialized = await request("initialize", { protocolVersion: protocolVersions[0], capabilities: {},
        clientInfo: { name: "starter-transport-probe", version: "1.0.0" } });
      if (!protocolVersions.includes(initialized.protocolVersion) || !object(initialized.capabilities) || !object(initialized.capabilities.tools) ||
          !object(initialized.serverInfo) || typeof initialized.serverInfo.name !== "string" || typeof initialized.serverInfo.version !== "string") {
        throw new ProbeError("MCP initialization did not negotiate a supported tools capability");
      }
      write({ jsonrpc: "2.0", method: "notifications/initialized" });
      const names = new Set();
      const cursors = new Set();
      let cursor;
      do {
        const result = await request("tools/list", cursor === undefined ? {} : { cursor });
        if (!Array.isArray(result.tools)) throw new ProbeError("Invalid MCP tools list");
        for (const tool of result.tools) {
          if (!object(tool) || typeof tool.name !== "string" || !toolName.test(tool.name) || names.has(tool.name) ||
              !object(tool.inputSchema) || tool.inputSchema.type !== "object") throw new ProbeError("Invalid or duplicate MCP tool definition");
          names.add(tool.name);
        }
        cursor = result.nextCursor;
        if (cursor !== undefined) {
          if (typeof cursor !== "string" || !cursor || cursors.has(cursor) || cursors.size >= 100) throw new ProbeError("Invalid MCP tools pagination");
          cursors.add(cursor);
        }
      } while (cursor !== undefined);
      const listed = [...names].sort();
      if (allowedTools !== undefined && (allowedTools.some((name) => !names.has(name)) ||
          (allowlistMode === "exact" && listed.length !== allowedTools.length))) {
        throw new ProbeError("MCP tool inventory differs from the configured allowlist");
      }
      return listed;
    };
    return await Promise.race([exchange(), failed]);
  } finally {
    stopped = true;
    clearTimeout(timer);
    signal?.removeEventListener("abort", abort);
    await terminate(child, closed, shutdownGraceMs);
  }
}

export async function probeLocalTool(base, id, options = {}) {
  let catalog;
  try { catalog = toolCatalog(base); }
  catch { throw new ProbeError("Could not read a valid developer tool catalog"); }
  const server = catalog.servers[id];
  if (typeof id !== "string" || !catalog.enabled.includes(id) || !server?.package || server.url) {
    throw new ProbeError("Expected an enabled local tool ID from .agents/tools.json");
  }
  return probeStdio(process.execPath, [join(base, "scripts/mcp-launch.mjs"), id], {
    // The catalog restricts host-visible tools, not what the package may expose.
    // Check that the selected tools exist without enabling or calling the rest.
    ...options, cwd: base, allowedTools: server.tools, allowlistMode: "subset",
  });
}

if (process.argv[1] && realpathSync(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const controller = new AbortController();
  let interrupt;
  const stop = (signal) => { interrupt = signal; controller.abort(); };
  process.on("SIGINT", stop);
  process.on("SIGTERM", stop);
  try {
    if (process.argv.length !== 3) throw new ProbeError("Usage: node scripts/mcp-probe.mjs <enabled-local-tool-id>");
    const names = await probeLocalTool(root, process.argv[2], { signal: controller.signal });
    console.log(`MCP transport verified; tools: ${names.join(", ") || "(none)"}`);
  } catch (error) {
    console.error(error instanceof ProbeError ? error.message : "MCP transport probe failed");
    process.exitCode = interrupt === "SIGINT" ? 130 : interrupt === "SIGTERM" ? 143 : 1;
  } finally {
    process.removeListener("SIGINT", stop);
    process.removeListener("SIGTERM", stop);
  }
}
