import "server-only";
import { loadPolarConfig } from "./environment";
import { createBillingProvider } from "./provider";
export function getPolarClient(): ReturnType<typeof createBillingProvider> | null {
  const config = loadPolarConfig();
  if (!config.enabled) return null;
  return createBillingProvider(config);
}
