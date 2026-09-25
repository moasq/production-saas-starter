import "server-only";
import { Polar } from "@polar-sh/sdk";
import { isPolarEnabled } from "./config";
import { createPolarHttpClient } from "./http-client";
let cachedClient: Polar | null = null;
export function getPolarClient(): Polar | null {
  if (!isPolarEnabled()) return null;
  if (!process.env.POLAR_ACCESS_TOKEN) throw new Error("BILLING_ENABLED requires POLAR_ACCESS_TOKEN");
  const server = process.env.POLAR_ENVIRONMENT || "sandbox";
  if (server !== "sandbox" && server !== "production") throw new Error("POLAR_ENVIRONMENT must be sandbox or production");
  cachedClient ??= new Polar({ accessToken: process.env.POLAR_ACCESS_TOKEN, server, httpClient: createPolarHttpClient() });
  return cachedClient;
}
