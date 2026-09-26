// Pure configuration parsing for server callers and tests. Do not import this
// module into client components; config.ts/client.ts enforce the server boundary.
export type PolarConfig = { enabled: false } | {
  enabled: true;
  server: "sandbox" | "production";
  baseURL: string;
  accessToken: string;
  productId: string;
};

export function billingEnabled(env: NodeJS.ProcessEnv = process.env): boolean {
  const value = env.BILLING_ENABLED?.trim() ?? "";
  if (value === "" || value === "false") return false;
  if (value === "true") return true;
  throw new Error("BILLING_ENABLED must be true or false");
}

export function loadPolarConfig(env: NodeJS.ProcessEnv = process.env): PolarConfig {
  if (!billingEnabled(env)) return { enabled: false };
  const server = env.POLAR_ENVIRONMENT?.trim();
  if (server !== "sandbox" && server !== "production") {
    throw new Error("POLAR_ENVIRONMENT must be sandbox or production");
  }
  const productId = env.POLAR_PRODUCT_ID?.trim() ?? "";
  const accessToken = env.POLAR_ACCESS_TOKEN?.trim() ?? "";
  if (!productId) throw new Error("POLAR_PRODUCT_ID is required when BILLING_ENABLED=true");
  if (!accessToken) throw new Error("POLAR_ACCESS_TOKEN is required when BILLING_ENABLED=true");
  return { enabled: true, server, productId, accessToken,
    baseURL: server === "sandbox" ? "https://sandbox-api.polar.sh" : "https://api.polar.sh" };
}
