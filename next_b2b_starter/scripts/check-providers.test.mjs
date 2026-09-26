import { test } from "node:test";
import assert from "node:assert/strict";
import { checkProviders } from "./check-providers.mjs";

const productId = "00000000-0000-4000-8000-000000000001";
const config = {
  SMTP_HOST: "smtp.example.test", EMAIL_FROM: "sender@example.test", SMTP_PORT: "587",
  SMTP_USER: "private-user", SMTP_PASSWORD: "private-password", SMTP_SECURE: "false",
  BILLING_ENABLED: "true", POLAR_ENVIRONMENT: "sandbox", POLAR_ACCESS_TOKEN: "private-token", POLAR_PRODUCT_ID: productId,
};
const forbiddenNetwork = { fetch() { assert.fail("Unexpected network request"); }, createTransport() { assert.fail("Unexpected SMTP request"); } };

test("missing/configured credentials never imply verified and default performs no network requests", async () => {
  const missing = await checkProviders({}, [], forbiddenNetwork);
  assert.equal(missing.exitCode, 2);
  assert.deepEqual(missing.checks.map(c => c.state), ["not_configured", "not_configured"]);
  const configured = await checkProviders(config, [], forbiddenNetwork);
  assert.equal(configured.exitCode, 2);
  assert.deepEqual(configured.checks.map(c => c.state), ["configured", "configured"]);
});

test("Polar probe is GET-only, sandbox-only, timeout-bounded and rejects redirects", async () => {
  const report = await checkProviders(config, ["polar"], { async fetch(url, options) {
    assert.equal(url, `https://sandbox-api.polar.sh/v1/products/${productId}`);
    assert.equal(options.method || "GET", "GET");
    assert.equal(options.redirect, "error");
    assert.ok(options.signal instanceof AbortSignal);
    assert.equal(options.headers["Polar-Version"], "2026-04");
    return Response.json({ id: productId, is_archived: false, is_recurring: true });
  } });
  assert.equal(report.exitCode, 0);
  assert.equal(report.checks[1].state, "verified");
  assert.ok(report.remaining.length);
  const refused = await checkProviders({ ...config, POLAR_ENVIRONMENT: "production" }, ["polar"], forbiddenNetwork);
  assert.equal(refused.exitCode, 1);
  assert.equal(refused.checks[1].state, "failed");
});

test("provider errors and invalid products fail without leaking credentials or payloads", async () => {
  for (const fetcher of [
    async () => { throw new Error("private-token private-password customer@example.test"); },
    async () => Response.json({ detail: "private-token" }, { status: 403 }),
    async () => Response.json({ id: productId, is_archived: true, is_recurring: true }),
    async () => Response.json({ id: "wrong-product", is_archived: false, is_recurring: true }),
  ]) {
    const report = await checkProviders(config, ["polar"], { fetch: fetcher });
    assert.equal(report.exitCode, 1);
    assert.doesNotMatch(JSON.stringify(report), /private-token|private-password|customer@example/);
  }
});

test("SMTP verifies transport, closes it, uses TLS and does not send mail", async () => {
  let closed = false, verified = false;
  const report = await checkProviders(config, ["smtp"], { createTransport(options) {
    assert.equal(options.requireTLS, true);
    assert.equal(options.connectionTimeout, 10000);
    return { async verify() { verified = true; return true; }, close() { closed = true; }, sendMail() { assert.fail("Must not send email"); } };
  } });
  assert.ok(verified && closed);
  assert.equal(report.checks[0].state, "verified");
  assert.match(report.checks[0].detail, /delivery are NOT verified/);
  assert.doesNotMatch(JSON.stringify(report), /private-user|private-password|sender@example/);
});

test("SMTP failure closes transport and reports failure without raw server output", async () => {
  let closed = false;
  const report = await checkProviders(config, ["smtp"], { createTransport() {
    return { async verify() { throw new Error("private-password"); }, close() { closed = true; } };
  } });
  assert.ok(closed);
  assert.equal(report.exitCode, 1);
  assert.doesNotMatch(JSON.stringify(report), /private-password/);
});

test("missing requested providers remain incomplete and invalid configuration fails", async () => {
  assert.equal((await checkProviders({}, ["smtp", "polar"], forbiddenNetwork)).exitCode, 2);
  for (const values of [{ SMTP_PORT: "wrong" }, { SMTP_USER: "" }, { POLAR_PRODUCT_ID: "not-an-id" }, { POLAR_ENVIRONMENT: "unknown" }]) {
    assert.equal((await checkProviders({ ...config, ...values }, ["smtp", "polar"], {
      async fetch() { return Response.json({ id: productId, is_archived: false, is_recurring: true }); },
      createTransport() { return { async verify() { return true; }, close() {} }; },
    })).exitCode, 1);
  }
});

test("billing probe rejects ambiguous configuration without contacting a provider", async () => {
  for (const values of [{ BILLING_ENABLED: "1" }, { POLAR_ENVIRONMENT: "" }, { POLAR_ACCESS_TOKEN: " " }, { POLAR_PRODUCT_ID: " " }]) {
    const report = await checkProviders({ ...config, ...values }, ["polar"], forbiddenNetwork);
    assert.equal(report.exitCode, 1);
    assert.equal(report.checks[1].state, "failed");
    assert.doesNotMatch(JSON.stringify(report), /private-token|private-password/);
  }
});
