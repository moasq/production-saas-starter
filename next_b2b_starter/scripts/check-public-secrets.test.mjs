import { test } from "node:test";
import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync, chmodSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { execFileSync, spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { canary, checkArtifact, checkBuild, checkSource, checkSources } from "./check-public-secrets.mjs";

test("public credential references fail without printing their value", () => {
  for (const name of ["NEXT_PUBLIC_POLAR_API_SANDBOX_ACCESS_TOKEN", "NEXT_PUBLIC_AUTH_SECRET", "NEXT_PUBLIC_SMTP_PASSWORD", "NEXT_PUBLIC_API_KEY"]) {
    assert.throws(() => checkSource(`${name}=private-value`, "fixture.env"), error =>
      error.message.includes(name) && !error.message.includes("private-value"));
  }
  assert.doesNotThrow(() => checkSource("NEXT_PUBLIC_APP_URL=https://example.test\nPOLAR_ACCESS_TOKEN=runtime-only", "fixture.env"));
});

test("leaked synthetic secrets fail for bundles and image metadata", () => {
  for (const path of ["static/chunk.js", "server/app/index.html", "image configuration", "image build history"]) {
    assert.throws(() => checkArtifact(Buffer.from(`prefix ${canary} suffix`), path), error =>
      error.message.includes(path) && !error.message.includes(canary));
  }
  assert.doesNotThrow(() => checkArtifact(Buffer.from("process.env.POLAR_ACCESS_TOKEN"), "server.js"));
});

test("artifact verification refuses missing builds and catches prerendered leaks", t => {
  const dir = mkdtempSync(join(tmpdir(), "starter-secrets-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  assert.throws(() => checkBuild(dir));
  for (const path of [".next/static", ".next/server/app", "public"]) mkdirSync(join(dir, path), { recursive: true });
  writeFileSync(join(dir, ".next/BUILD_ID"), "test");
  writeFileSync(join(dir, ".next/static/chunk.js"), "safe");
  assert.doesNotThrow(() => checkBuild(dir));
  writeFileSync(join(dir, ".next/server/app/index.html"), canary);
  assert.throws(() => checkBuild(dir), /server\/app\/index.html/);
});

test("checked-in frontend and deployment inputs contain no public credential names", () => {
  assert.ok(checkSources() > 0);
});

test("source coverage includes hooks, proxy and new runtime directories", t => {
  const repo = mkdtempSync(join(tmpdir(), "starter-source-coverage-"));
  t.after(() => rmSync(repo, { recursive: true, force: true }));
  execFileSync("git", ["init", "--quiet", repo]);
  for (const relative of ["hooks/use-flow.ts", "proxy.ts", "src/new-feature/client.tsx"]) {
    const path = join(repo, "next_b2b_starter", relative);
    mkdirSync(join(path, ".."), { recursive: true });
    writeFileSync(path, "export const value = process.env.NEXT_PUBLIC_PRIVATE_KEY;");
    execFileSync("git", ["add", "."], { cwd: repo });
    assert.throws(() => checkSources(repo), error => error.message.includes(relative));
    writeFileSync(path, "export const value = process.env.NEXT_PUBLIC_APP_URL;");
    assert.doesNotThrow(() => checkSources(repo));
  }
});

test("failing build and Docker commands cannot print their private diagnostic payloads", { skip: process.platform === "win32" }, t => {
  const bin = mkdtempSync(join(tmpdir(), "starter-private-output-"));
  t.after(() => rmSync(bin, { recursive: true, force: true }));
  for (const tool of ["pnpm", "docker"]) {
    const path = join(bin, tool);
    writeFileSync(path, "#!/bin/sh\nprintf '%s\\n' 'SYNTHETIC_DIAGNOSTIC_SECRET'\nprintf '%s\\n' 'SYNTHETIC_DIAGNOSTIC_SECRET' >&2\nexit 1\n");
    chmodSync(path, 0o700);
  }
  const script = fileURLToPath(new URL("./check-public-secrets.mjs", import.meta.url));
  for (const args of [["--build"], ["--image", "synthetic-image"]]) {
    const result = spawnSync(process.execPath, [script, ...args], { encoding: "utf8", env: { ...process.env, PATH: `${bin}:${process.env.PATH}` } });
    assert.equal(result.status, 1);
    assert.doesNotMatch(result.stdout + result.stderr, /SYNTHETIC_DIAGNOSTIC_SECRET/);
    assert.match(result.stderr, /failed/i);
  }
});
