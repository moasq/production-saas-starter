import "server-only";
import { Polar } from "@polar-sh/sdk";
import { loadPolarConfig } from "./environment";
import { createPolarHttpClient } from "./http-client";
export function getPolarClient(): Polar | null {
  const config = loadPolarConfig();
  if (!config.enabled) return null;
  return new Polar({ accessToken: config.accessToken, server: config.server, httpClient: createPolarHttpClient() });
}
