import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { loadPolarConfig } from "../lib/polar/environment.ts";

const cases = JSON.parse(readFileSync(new URL("../../go-b2b-starter/internal/platform/polar/testdata/config.json", import.meta.url), "utf8")) as {
  name: string; env: NodeJS.ProcessEnv; error?: string; enabled?: boolean;
  baseURL?: string; accessToken?: string; productId?: string;
}[];

for (const fixture of cases) test(`shared Go/Next billing contract: ${fixture.name}`, () => {
  if (fixture.error) {
    assert.throws(() => loadPolarConfig(fixture.env), { message: fixture.error });
    return;
  }
  const config = loadPolarConfig(fixture.env);
  assert.equal(config.enabled, fixture.enabled);
  if (config.enabled) {
    assert.equal(config.baseURL, fixture.baseURL);
    assert.equal(config.productId, fixture.productId);
    assert.equal(config.accessToken, fixture.accessToken);
  } else assert.deepEqual(config, { enabled: false });
});
