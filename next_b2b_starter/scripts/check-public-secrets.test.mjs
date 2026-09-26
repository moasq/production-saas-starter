import { test } from "node:test";
import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
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
