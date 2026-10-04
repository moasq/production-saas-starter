// Exercise the built Next.js startup hook with synthetic configuration and local HTTP.
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { once } from "node:events";
import { createServer } from "node:net";
import { setTimeout as delay } from "node:timers/promises";
import { fileURLToPath } from "node:url";

const runtime = fileURLToPath(new URL("../.next/standalone/", import.meta.url));
const base = {
  PATH: process.env.PATH, NODE_ENV: "production", HOSTNAME: "127.0.0.1", NEXT_TELEMETRY_DISABLED: "1",
  APP_ENV: "PROD", APP_BASE_URL: "https://app.example.com",
  BETTER_AUTH_SECRET: "2ec34d0f74fe8a9c164fb05920abc285", AUTH_INTERNAL_SECRET: "9a429cbf2dd80e61423bd94aa7bfe501",
  PGHOST: "127.0.0.1", SMTP_HOST: "smtp.example.com", EMAIL_FROM: "Starter <login@example.com>",
};
const cases = [
  ["explicit local built server", { APP_ENV: "DEV", BETTER_AUTH_SECRET: "", AUTH_INTERNAL_SECRET: "", APP_BASE_URL: "" }, null],
  ["missing mode", { APP_ENV: "" }, /APP_ENV/],
  ["unknown mode", { APP_ENV: "production" }, /APP_ENV/],
  ["production HTTP", { APP_BASE_URL: "http://localhost:3000" }, /HTTPS/],
  ["production missing secret", { BETTER_AUTH_SECRET: "" }, /BETTER_AUTH_SECRET/],
  ["production placeholder", { BETTER_AUTH_SECRET: "change-me-before-production-12345678" }, /placeholder/],
  ["production capture mail", { SMTP_HOST: "mailpit" }, /Mailpit/],
  ["configured production server", {}, null],
];
for (const [name, changes, expectedError] of cases) {
  const reservation = createServer();
  reservation.listen(0, "127.0.0.1"); await once(reservation, "listening");
  const port = reservation.address().port;
  await new Promise(resolve => reservation.close(resolve));
  const env = { ...base, ...changes, PORT: String(port) };
  const child = spawn(process.execPath, ["server.js"], { cwd: runtime, env, stdio: ["ignore", "pipe", "pipe"] });
  const closed = once(child, "close");
  let output = "";
  child.stdout.on("data", data => { output += data; }); child.stderr.on("data", data => { output += data; });
  try {
    const deadline = Date.now() + 15000;
    if (expectedError) {
      while (child.exitCode === null && Date.now() < deadline) await delay(50);
      assert.notEqual(child.exitCode, null, `${name}: unsafe server stayed running`);
      assert.notEqual(child.exitCode, 0, `${name}: unsafe server reported successful exit`);
      assert.match(output, expectedError, `${name}: missing safe configuration error`);
      for (const value of [env.BETTER_AUTH_SECRET, env.AUTH_INTERNAL_SECRET]) {
        if (value) assert.ok(!output.includes(value), `${name}: secret leaked to logs`);
      }
    } else {
      let ready = false;
      while (!ready && child.exitCode === null && Date.now() < deadline) {
        try { ready = (await fetch(`http://127.0.0.1:${port}/api/health`, { signal: AbortSignal.timeout(500) })).ok; } catch { /* still starting */ }
        if (!ready) await delay(50);
      }
      assert.ok(ready, `${name}: server did not become ready`);
      if (env.APP_ENV === "DEV") {
        const login = await fetch(`http://127.0.0.1:${port}/auth`);
        assert.equal(login.status, 200);
        assert.match(await login.text(), /Connect authentication/);
      }
    }
    console.log(`PASS ${name}`);
  } finally {
    // Signal only this direct child, never a host process tree.
    if (child.exitCode === null) child.kill("SIGTERM");
    await closed;
  }
}
