import "server-only";
import { getMemberSession } from "@/lib/auth/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { getPolarClient } from "./client";
import { loadPolarConfig } from "./environment";
import type { PolarPlan } from "./plans";
export async function fetchProducts(): Promise<{ success: boolean; products?: PolarPlan[]; error?: string }> {
 const session = await getMemberSession();
 if (!session) return { success: false, error: "Authentication required." };
 if (!(await getServerPermissions(session)).canManageSubscriptions) return { success: false, error: "Organization administrator access required." };
 try {
  const client = getPolarClient();
  if (!client) return { success: false, error: "Billing is disabled." };
  const config = loadPolarConfig();
  if (!config.enabled) return { success: false, error: "Billing is disabled." };
  const productId = config.productId;
  const plan = await client.getPlan(productId);
  const products: PolarPlan[] = plan ? [plan] : [];
  return { success: true, products };
 } catch { return { success: false, error: "Could not load billing plans. Check the billing configuration and retry." }; }
}
