import "server-only";
import { readAuthConfiguration } from "./runtime-config";
export function isAuthConfigured(): boolean {
  try { readAuthConfiguration(); return true; } catch { return false; }
}
