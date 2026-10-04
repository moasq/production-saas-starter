type Environment = Record<string, string | undefined>;

export function applicationMode(env: Environment = process.env): "DEV" | "PROD" {
  if (env.APP_ENV !== "DEV" && env.APP_ENV !== "PROD") {
    throw new Error("APP_ENV must be explicitly set to DEV or PROD");
  }
  return env.APP_ENV;
}

function authSecret(env: Environment, name: string, production: boolean): string {
  const value = env[name];
  if (!value || value.length < 32 || value.trim() !== value) {
    throw new Error(`${name} must contain at least 32 characters without surrounding whitespace`);
  }
  // Detect copied examples and repeated characters, not arbitrary secret entropy.
  if (production && (/(change[-_]?me|replace[-_]?me|placeholder|your[-_]secret|test[-_]only)/i.test(value) || new Set(value).size === 1)) {
    throw new Error(`${name} must not be a placeholder or repeated character in production`);
  }
  return value;
}

export function readAuthConfiguration(env: Environment = process.env) {
  const production = applicationMode(env) === "PROD";
  let baseURL: URL;
  try { baseURL = new URL(env.APP_BASE_URL || ""); } catch {
    throw new Error("APP_BASE_URL must be an explicit HTTP(S) origin");
  }
  if (!/^https?:\/\/[^/?#\\]+\/?$/.test(env.APP_BASE_URL!) || /\s/.test(env.APP_BASE_URL!) ||
      (baseURL.port !== "" && Number(baseURL.port) < 1) || baseURL.username || baseURL.password || baseURL.pathname !== "/" || baseURL.search || baseURL.hash) {
    throw new Error("APP_BASE_URL must be an HTTP(S) origin without credentials, path, query or fragment");
  }
  if (production && baseURL.protocol !== "https:") throw new Error("Production APP_BASE_URL must use HTTPS");
  const secret = authSecret(env, "BETTER_AUTH_SECRET", production);
  const bridgeSecret = authSecret(env, "AUTH_INTERNAL_SECRET", production);
  if (production && secret === bridgeSecret) throw new Error("BETTER_AUTH_SECRET and AUTH_INTERNAL_SECRET must be independent in production");
  if (!env.AUTH_DATABASE_URL?.trim() && !env.PGHOST?.trim()) throw new Error("AUTH_DATABASE_URL or PGHOST is required");
  if (!env.SMTP_HOST?.trim() || !env.EMAIL_FROM?.trim()) throw new Error("SMTP_HOST and EMAIL_FROM are required");
  if (production && env.SMTP_HOST.trim().toLowerCase().replace(/\.$/, "") === "mailpit") {
    throw new Error("Production SMTP_HOST must not use local Mailpit capture");
  }
  return { baseURL: baseURL.origin, secret };
}

// DEV may show the unconfigured setup screen. PROD must refuse to start before
// accepting requests. NODE_ENV controls Next optimization, not deployment policy.
export function validateServerConfiguration(env: Environment = process.env): void {
  if (applicationMode(env) === "PROD") readAuthConfiguration(env);
}
