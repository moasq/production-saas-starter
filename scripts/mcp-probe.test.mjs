import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { cpSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";
import { probeLocalTool, probeStdio } from "./mcp-probe.mjs";

const repository = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const fixtureCode = String.raw`
import { appendFileSync, writeFileSync } from "node:fs";
import { spawn } from "node:child_process";
import { createInterface } from "node:readline";
const mode = process.env.FIXTURE_MODE || "success";
writeFileSync("server.pid", String(process.pid));
const rpc = (message) => process.stdout.write(JSON.stringify({ jsonrpc: "2.0", ...message }) + "\n");
const tool = (name) => ({ name, description: "SECRET_DESCRIPTION", inputSchema: { type: "object" } });
process.stderr.write("SECRET_STDERR\n");
if (mode === "exit") process.exit(7);
if (mode === "noise") process.stdout.write("SECRET_INVALID_STDOUT\n");
if (mode === "stdout-flood") process.stdout.write("x".repeat(4096));
if (mode === "stderr-flood") process.stderr.write("SECRET".repeat(4096));
if (mode === "invalid-utf8") process.stdout.write(Buffer.from([0xff, 0x0a]));
if (mode === "tree") {
  const descendant = spawn(process.execPath, ["--input-type=module", "-e", "process.on('SIGTERM', () => {}); setInterval(() => {}, 1000)"], { stdio: "inherit" });
  writeFileSync("descendant.pid", String(descendant.pid));
  process.on("SIGTERM", () => {});
  descendant.on("exit", () => {});
  setInterval(() => {}, 1000);
}
let initialized = false;
createInterface({ input: process.stdin }).on("line", (line) => {
  const message = JSON.parse(line);
  appendFileSync("requests.jsonl", JSON.stringify(message) + "\n");
  if (["timeout", "tree"].includes(mode)) return;
  if (message.method === "initialize") {
    if (mode === "protocol-error") { rpc({ id: message.id, error: { code: -32000, message: "SECRET_SERVER_ERROR" } }); return; }
    if (mode === "wrong-id") { rpc({ id: message.id + 1, result: {} }); return; }
    if (mode === "malformed-envelope") { rpc({ id: message.id, result: {}, error: {} }); return; }
    if (mode === "server-requests") {
      rpc({ id: "ping-request", method: "ping" });
      rpc({ id: "private-roots", method: "roots/list" });
      rpc({ method: "notifications/message", params: { level: "info", data: "SECRET_NOTIFICATION" } });
    }
    const result = { protocolVersion: mode === "unsupported-version" ? "1999-01-01" : message.params.protocolVersion,
      capabilities: mode === "missing-tools" ? {} : { tools: {} }, serverInfo: { name: "fixture", version: "1.0.0" } };
    const encoded = JSON.stringify({ jsonrpc: "2.0", id: message.id, result }) + "\n";
    process.stdout.write(encoded.slice(0, 30));
    setTimeout(() => process.stdout.write(encoded.slice(30)), 5);
  } else if (message.method === "notifications/initialized") {
    initialized = true;
  } else if (message.method === "tools/list") {
    if (!initialized) process.exit(8);
    if (mode === "empty") { rpc({ id: message.id, result: { tools: [] } }); return; }
    if (mode === "bad-schema") { rpc({ id: message.id, result: { tools: [{ name: "alpha", inputSchema: [] }] } }); return; }
    if (mode === "unsafe-name") { rpc({ id: message.id, result: { tools: [tool("alpha\nSECRET_NAME")] } }); return; }
    if (mode === "duplicate-tool") { rpc({ id: message.id, result: { tools: [tool("alpha"), tool("alpha")] } }); return; }
    if (mode === "cycle") { rpc({ id: message.id, result: { tools: [], nextCursor: "repeat" } }); return; }
    if (mode === "pagination") {
      rpc({ id: message.id, result: message.params.cursor === undefined
        ? { tools: [tool("zeta")], nextCursor: "second" } : { tools: [tool("alpha")] } }); return;
    }
    rpc({ id: message.id, result: { tools: [tool("alpha")] } });
  } else if (message.method) process.exit(9);
});
`;

function fixture(t, mode = "success") {
  const base = mkdtempSync(join(tmpdir(), "starter MCP probe "));
  t.after(() => rmSync(base, { recursive: true, force: true }));
  const script = join(base, "server.mjs");
  writeFileSync(script, fixtureCode);
  return { base, script, options: { cwd: base, env: { ...process.env, FIXTURE_MODE: mode }, timeoutMs: 2000, shutdownGraceMs: 15 } };
}

function requests(base) {
  return readFileSync(join(base, "requests.jsonl"), "utf8").trim().split("\n").map((line) => JSON.parse(line));
}

async function isStopped(pid) {
  for (let attempt = 0; attempt < 50; attempt++) {
    try { process.kill(pid, 0); }
    catch (error) { if (error.code === "ESRCH") return true; throw error; }
    // Linux may briefly retain an orphan as a zombie until init reaps it.
    if (process.platform === "linux" && existsSync(`/proc/${pid}/stat`) && readFileSync(`/proc/${pid}/stat`, "utf8").split(") ")[1]?.startsWith("Z ")) return true;
    await new Promise((done) => setTimeout(done, 10));
  }
  return false;
}

test("transport handles split frames, initialization and paginated names without tool calls", async (t) => {
  const { base, script, options } = fixture(t, "pagination");
  assert.deepEqual(await probeStdio(process.execPath, [script], { ...options, allowedTools: ["zeta", "alpha"] }), ["alpha", "zeta"]);
  const sent = requests(base);
  assert.deepEqual(sent.map((item) => item.method), ["initialize", "notifications/initialized", "tools/list", "tools/list"]);
  assert.deepEqual(sent[0].params.capabilities, {});
  assert.deepEqual(sent[3].params, { cursor: "second" });
  assert.equal(await isStopped(Number(readFileSync(join(base, "server.pid")))), true);
});

test("responds to ping but refuses server-requested client capabilities", async (t) => {
  const { base, script, options } = fixture(t, "server-requests");
  assert.deepEqual(await probeStdio(process.execPath, [script], options), ["alpha"]);
  const sent = requests(base);
  assert.deepEqual(sent.find((item) => item.id === "ping-request").result, {});
  assert.equal(sent.find((item) => item.id === "private-roots").error.code, -32601);
  assert.equal(sent.some((item) => item.method === "tools/call"), false);
});

test("allowlists support exact inventory checks and selected-tool subsets", async (t) => {
  const empty = fixture(t, "empty");
  assert.deepEqual(await probeStdio(process.execPath, [empty.script], { ...empty.options, allowedTools: [] }), []);
  const nonempty = fixture(t);
  await assert.rejects(probeStdio(process.execPath, [nonempty.script], { ...nonempty.options, allowedTools: ["alpha", "missing"] }), /configured allowlist/);
  await assert.rejects(probeStdio(process.execPath, [nonempty.script], { ...nonempty.options, allowedTools: [] }), /configured allowlist/);
  const subset = fixture(t, "pagination");
  assert.deepEqual(await probeStdio(process.execPath, [subset.script], {
    ...subset.options, allowedTools: ["alpha"], allowlistMode: "subset",
  }), ["alpha", "zeta"]);
  await assert.rejects(probeStdio(process.execPath, [subset.script], {
    ...subset.options, allowedTools: ["missing"], allowlistMode: "subset",
  }), /configured allowlist/);
});

test("rejects protocol failures without returning raw server output", async (t) => {
  const cases = {
    "exit": /exited|stream closed/,
    "noise": /invalid JSON/,
    "invalid-utf8": /invalid JSON/,
    "protocol-error": /protocol error/,
    "wrong-id": /Unexpected MCP response/,
    "malformed-envelope": /Unexpected MCP response/,
    "unsupported-version": /initialization/,
    "missing-tools": /initialization/,
    "bad-schema": /tool definition/,
    "unsafe-name": /tool definition/,
    "duplicate-tool": /tool definition/,
    "cycle": /pagination/,
  };
  for (const [mode, expected] of Object.entries(cases)) {
    await t.test(mode, async (t) => {
      const { base, script, options } = fixture(t, mode);
      await assert.rejects(probeStdio(process.execPath, [script], options), (error) => {
        assert.match(error.message, expected);
        assert.doesNotMatch(error.message, /SECRET/);
        return true;
      });
      assert.equal(await isStopped(Number(readFileSync(join(base, "server.pid")))), true);
    });
  }
});

test("combined output budget covers stdout and stderr, including unterminated frames", async (t) => {
  for (const mode of ["stdout-flood", "stderr-flood"]) {
    const { base, script, options } = fixture(t, mode);
    await assert.rejects(probeStdio(process.execPath, [script], { ...options, maxBufferBytes: 1024 }), /byte limit/);
    assert.equal(await isStopped(Number(readFileSync(join(base, "server.pid")))), true);
  }
});

test("timeout kills an unresponsive server and its descendants", async (t) => {
  const { base, script, options } = fixture(t, "tree");
  const started = Date.now();
  await assert.rejects(probeStdio(process.execPath, [script], { ...options, timeoutMs: 750 }), /timed out/);
  assert.ok(Date.now() - started < 5000);
  for (const file of ["server.pid", "descendant.pid"]) {
    assert.equal(await isStopped(Number(readFileSync(join(base, file)))), true, `${file} must stop`);
  }
});

test("abort cleans up the child and invalid options never launch it", async (t) => {
  const { base, script, options } = fixture(t, "timeout");
  const controller = new AbortController();
  const abort = setTimeout(() => controller.abort(), 300);
  try { await assert.rejects(probeStdio(process.execPath, [script], { ...options, signal: controller.signal }), /interrupted/); }
  finally { clearTimeout(abort); }
  assert.equal(await isStopped(Number(readFileSync(join(base, "server.pid")))), true);
  await assert.rejects(probeStdio(process.execPath, [script], { timeoutMs: 120001 }), /Invalid MCP probe options/);
  await assert.rejects(probeStdio(process.execPath, [script], { allowedTools: ["alpha", "alpha"] }), /Invalid MCP probe options/);
  await assert.rejects(probeStdio(join(base, "missing-executable"), [], { shutdownGraceMs: 1 }), /Could not start/);
});

test("catalog probe uses enabled local launchers and rejects remote, disabled and unknown IDs", async (t) => {
  const { base } = fixture(t);
  mkdirSync(join(base, ".agents"));
  mkdirSync(join(base, "scripts"));
  cpSync(join(repository, ".agents/tools.json"), join(base, ".agents/tools.json"));
  for (const file of ["mcp-probe.mjs", "tool-catalog.mjs"]) cpSync(join(repository, "scripts", file), join(base, "scripts", file));
  writeFileSync(join(base, "scripts/mcp-launch.mjs"), fixtureCode);
  const catalog = JSON.parse(readFileSync(join(base, ".agents/tools.json"), "utf8"));
  catalog.servers.context7.tools = ["alpha"];
  writeFileSync(join(base, ".agents/tools.json"), JSON.stringify(catalog));
  assert.deepEqual(await probeLocalTool(base, "context7", { shutdownGraceMs: 15 }), ["alpha"]);
  assert.deepEqual(await probeLocalTool(base, "context7", {
    shutdownGraceMs: 15, env: { ...process.env, FIXTURE_MODE: "pagination" },
  }), ["alpha", "zeta"]);
  catalog.servers.context7.tools = ["missing"];
  writeFileSync(join(base, ".agents/tools.json"), JSON.stringify(catalog));
  await assert.rejects(probeLocalTool(base, "context7", { shutdownGraceMs: 15 }), /configured allowlist/);
  catalog.servers.context7.tools = ["alpha"];
  writeFileSync(join(base, ".agents/tools.json"), JSON.stringify(catalog));
  for (const id of ["better-auth", "magicui", "not-a-tool", "../context7"]) {
    await assert.rejects(probeLocalTool(base, id), /enabled local tool ID/);
  }
  const result = spawnSync(process.execPath, [join(base, "scripts/mcp-probe.mjs"), "context7"], { cwd: tmpdir(), encoding: "utf8", timeout: 5000 });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, "MCP transport verified; tools: alpha\n");
  assert.equal(result.stderr, "");
  const failure = spawnSync(process.execPath, [join(base, "scripts/mcp-probe.mjs"), "context7"], {
    env: { ...process.env, FIXTURE_MODE: "protocol-error" }, encoding: "utf8", timeout: 5000,
  });
  assert.equal(failure.status, 1);
  assert.equal(failure.stdout, "");
  assert.equal(failure.stderr, "MCP server returned a protocol error\n");
});
