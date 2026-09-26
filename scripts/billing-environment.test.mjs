import { test } from "node:test";
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

test("Compose passes the same explicit provider environment to both runtimes", () => {
  const directory = mkdtempSync(join(tmpdir(), "starter-billing-config-"));
  const path = join(directory, "test.env");
  // Config interpolation only: never contacts the daemon or starts containers.
  const env = { ...process.env };
  for (const name of Object.keys(env)) {
    if (/^(?:POSTGRES_|APP_DATABASE_|AUTH_DATABASE_|BETTER_AUTH_|AUTH_INTERNAL_|POLAR_|BILLING_|COMPOSE_)/.test(name)) delete env[name];
  }
  try {
    for (const value of [undefined, "", "sandbox", "production"]) {
      writeFileSync(path, ["POSTGRES_PASSWORD=test-only", "APP_DATABASE_PASSWORD=test-only",
        "AUTH_DATABASE_PASSWORD=test-only", "BETTER_AUTH_SECRET=test-only", "AUTH_INTERNAL_SECRET=test-only",
        "BILLING_ENABLED=true", "POLAR_ACCESS_TOKEN=test-only", "POLAR_PRODUCT_ID=test-only",
        ...(value === undefined ? [] : [`POLAR_ENVIRONMENT=${value}`]), ""].join("\n"), { mode: 0o600 });
      const config = JSON.parse(execFileSync("docker", ["compose", "--env-file", path, "-f", "compose.yaml", "config", "--format", "json"], {
        cwd: fileURLToPath(new URL("..", import.meta.url)), env, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
      }));
      for (const name of ["backend", "frontend"]) {
        assert.equal(config.services[name].environment.POLAR_ENVIRONMENT, value ?? "");
        assert.equal(config.services[name].environment.BILLING_ENABLED, "true");
      }
    }
  } finally { rmSync(directory, { recursive: true, force: true }); }
});
