import { loadPolarConfig } from "../lib/polar/environment.ts";
import { execFileSync } from "node:child_process";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const project = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// This reports narrow operations, never overall deployment readiness. No mail,
// checkout, payment, customer or subscription is created by this command.
export async function checkProviders(env, selected = [], dependencies = {}) {
  const smtp = { provider: "smtp", operation: "connection_and_authentication", state: "not_configured", environment: "external", detail: "SMTP configuration is incomplete." };
  const polarEnvironment = env.POLAR_ENVIRONMENT?.trim();
  const polar = { provider: "polar", operation: "read_recurring_product", state: "not_configured", environment: ["sandbox", "production"].includes(polarEnvironment) ? polarEnvironment : "invalid", detail: "Optional billing is disabled or incomplete." };
  const checks = [smtp, polar];
  if (["mailpit", "localhost", "127.0.0.1", "::1"].includes(env.SMTP_HOST)) smtp.environment = "local";
  if (env.SMTP_HOST && env.EMAIL_FROM) {
    const port = Number(env.SMTP_PORT || "587");
    if (!Number.isInteger(port) || port < 1 || port > 65535 ||
        ![undefined, "", "true", "false"].includes(env.SMTP_SECURE) ||
        Boolean(env.SMTP_USER) !== Boolean(env.SMTP_PASSWORD)) {
      Object.assign(smtp, { state: "failed", detail: "Invalid SMTP port, TLS mode or incomplete credentials." });
    } else {
      Object.assign(smtp, { state: "configured", detail: "Connection not checked; email delivery not verified." });
      if (selected.includes("smtp")) {
        let transport;
        try {
          const createTransport = dependencies.createTransport || (await import("nodemailer")).default.createTransport;
          transport = createTransport({
            host: env.SMTP_HOST, port, secure: env.SMTP_SECURE === "true",
            requireTLS: Boolean(env.SMTP_USER) && env.SMTP_SECURE !== "true",
            ...(env.SMTP_USER ? { auth: { user: env.SMTP_USER, pass: env.SMTP_PASSWORD } } : {}),
            connectionTimeout: 10000, greetingTimeout: 10000, socketTimeout: 15000,
          });
          if (await transport.verify() !== true) throw new Error("SMTP verification did not succeed");
          Object.assign(smtp, { state: "verified", detail: "Connection/authentication succeeded; sender acceptance and delivery are NOT verified." });
        } catch {
          Object.assign(smtp, { state: "failed", detail: "SMTP connection/authentication failed. Inspect provider diagnostics privately." });
        } finally {
          try { transport?.close(); } catch {
            Object.assign(smtp, { state: "failed", detail: "SMTP transport cleanup failed; inspect locally." });
          }
        }
      }
    }
  }
  let polarConfig;
  try { polarConfig = loadPolarConfig(env); } catch {
    Object.assign(polar, { state: "failed", detail: "Invalid or incomplete billing configuration. Check BILLING_ENABLED, POLAR_ENVIRONMENT, POLAR_PRODUCT_ID and POLAR_ACCESS_TOKEN." });
  }
  if (polarConfig?.enabled) {
    if (!["sandbox", "production"].includes(polar.environment) || !uuid.test(polarConfig.productId)) {
      Object.assign(polar, { state: "failed", detail: "Invalid Polar environment or product ID format." });
    } else {
      Object.assign(polar, { state: "configured", detail: "Product access and payment lifecycle not verified." });
      if (selected.includes("polar")) {
        if (polar.environment !== "sandbox") {
          Object.assign(polar, { state: "failed", detail: "This check only permits Polar sandbox; production was not contacted." });
        } else {
          try {
            const response = await (dependencies.fetch || fetch)(`https://sandbox-api.polar.sh/v1/products/${polarConfig.productId}`, {
              headers: { Authorization: `Bearer ${polarConfig.accessToken}`, "Polar-Version": "2026-04" },
              redirect: "error", signal: AbortSignal.timeout(10000),
            });
            if (!response.ok) throw new Error("Provider rejected product read");
            const product = await response.json();
            if (product.id !== polarConfig.productId || product.is_archived !== false || product.is_recurring !== true) throw new Error("Expected active recurring product");
            Object.assign(polar, { state: "verified", detail: "Sandbox token can read the configured active recurring product; payments/portal are NOT verified." });
          } catch {
            Object.assign(polar, { state: "failed", detail: "Sandbox product read failed or returned an incompatible product. Inspect provider diagnostics privately." });
          }
        }
      }
    }
  }
  return {
    schemaVersion: 1, scope: "provider-connectivity-only", checks,
    remaining: ["external email delivery and link/invitation journeys", "sandbox checkout, ownership, portal and subscription lifecycle", "production deployment verification"],
    exitCode: checks.some(check => check.state === "failed") ? 1 :
      !selected.length || selected.some(name => checks.find(check => check.provider === name)?.state !== "verified") ? 2 : 0,
  };
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const flags = process.argv.slice(2);
  if (flags.some(flag => !["--smtp", "--polar-sandbox"].includes(flag))) {
    console.error("Usage: check-providers.mjs [--smtp] [--polar-sandbox]");
    process.exitCode = 2;
  } else {
    const selected = [...new Set(flags.map(flag => flag === "--smtp" ? "smtp" : "polar"))];
    const report = await checkProviders(process.env, selected);
    let revision = "unknown", workingTree = "unknown";
    try {
      const options = { cwd: project, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] };
      revision = execFileSync("git", ["rev-parse", "HEAD"], options).trim();
      workingTree = execFileSync("git", ["status", "--porcelain"], options).trim() ? "modified" : "clean";
    } catch { /* Container images may omit Git. */ }
    console.log(JSON.stringify({ revision, workingTree, checkedAt: new Date().toISOString(), ...report }, null, 2));
    process.exitCode = report.exitCode;
  }
}
