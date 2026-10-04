import test from "node:test";
import assert from "node:assert/strict";
import { applicationMode, readAuthConfiguration, validateServerConfiguration } from "../lib/auth/runtime-config.ts";

// Synthetic values; these tests make no network requests.
const valid = {
  APP_ENV: "PROD", APP_BASE_URL: "https://app.example.com",
  BETTER_AUTH_SECRET: "2ec34d0f74fe8a9c164fb05920abc285", AUTH_INTERNAL_SECRET: "9a429cbf2dd80e61423bd94aa7bfe501",
  PGHOST: "postgres", SMTP_HOST: "smtp.example.com", EMAIL_FROM: "Starter <login@example.com>",
};

test("only explicit application modes are accepted independently of NODE_ENV", () => {
  for (const APP_ENV of [undefined, "", "prod", "production", "PROD ", "staging"]) {
    assert.throws(() => validateServerConfiguration({ ...valid, APP_ENV, NODE_ENV: "production" }), /APP_ENV/);
  }
  assert.equal(applicationMode({ APP_ENV: "DEV", NODE_ENV: "production" }), "DEV");
  assert.doesNotThrow(() => validateServerConfiguration({ APP_ENV: "DEV", NODE_ENV: "production" }));
  assert.doesNotThrow(() => validateServerConfiguration({ ...valid, NODE_ENV: "development" }));
});

test("production rejects invalid or incomplete auth configuration before provider calls", () => {
  for (const key of ["APP_BASE_URL", "BETTER_AUTH_SECRET", "AUTH_INTERNAL_SECRET", "PGHOST", "SMTP_HOST", "EMAIL_FROM"]) {
    assert.throws(() => validateServerConfiguration({ ...valid, [key]: undefined }), new RegExp(key));
  }
  for (const APP_BASE_URL of ["http://app.example.com", "https://user:password@app.example.com", "https://app.example.com/path", "https://app.example.com?x=1", "https://app.example.com#x", "javascript:alert(1)", "https://app.example.com:99999", "https://app.example.com:0", "https://app.example.com?", "https://app.example.com/foo/..", "https://app.example.com\\path", "https://app.exa\tmple.com"]) {
    assert.throws(() => readAuthConfiguration({ ...valid, APP_BASE_URL }), /APP_BASE_URL/);
  }
  assert.throws(() => validateServerConfiguration({ ...valid, SMTP_HOST: "MAILPIT." }), /Mailpit/);
});

test("production rejects obvious secret placeholders without including values in errors", () => {
  for (const key of ["BETTER_AUTH_SECRET", "AUTH_INTERNAL_SECRET"]) {
    for (const value of ["short", "a".repeat(40), "é".repeat(40), "change-me-before-production-12345678", "test-only-session-secret-0000000000", "placeholder-secret-000000000000000", ` ${valid[key as keyof typeof valid]}`]) {
      assert.throws(() => readAuthConfiguration({ ...valid, [key]: value }), (error: unknown) => {
        assert.ok(error instanceof Error); assert.ok(error.message.includes(key)); assert.ok(!error.message.includes(value)); return true;
      });
    }
  }
  assert.throws(() => readAuthConfiguration({ ...valid, AUTH_INTERNAL_SECRET: valid.BETTER_AUTH_SECRET }), /independent/);
});

test("local built image uses real auth with explicit DEV and supports synthetic fixture secrets", () => {
  const config = readAuthConfiguration({ ...valid, APP_ENV: "DEV", NODE_ENV: "production", APP_BASE_URL: "http://localhost:3000", SMTP_HOST: "mailpit", BETTER_AUTH_SECRET: "test-only-session-secret-0000000000" });
  assert.equal(config.baseURL, "http://localhost:3000");
  assert.throws(() => readAuthConfiguration({ ...valid, APP_ENV: "DEV", APP_BASE_URL: undefined }), /APP_BASE_URL/);
  assert.equal(readAuthConfiguration({ ...valid, APP_BASE_URL: "https://APP.EXAMPLE.COM:443/" }).baseURL, "https://app.example.com");
  assert.doesNotThrow(() => readAuthConfiguration({ ...valid, PGHOST: undefined, AUTH_DATABASE_URL: "postgres://synthetic@postgres/auth" }));
});
