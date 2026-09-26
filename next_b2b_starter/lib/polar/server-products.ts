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
  const selected = await client.products.get({ id: productId });
  const products: PolarPlan[] = [];
  for (const product of [selected]) {
    const price = product.prices[0];
    // The starter supports one fixed recurring price per product. More complex pricing belongs in your application.
    if (product.isArchived || !product.isRecurring || product.prices.length !== 1 || !price || price.amountType !== "fixed") continue;
    products.push({ id: product.id, productId: product.id, name: product.name, description: product.description ?? null,
      price: price.priceAmount / 100, currency: price.priceCurrency, interval: product.recurringInterval ?? "month" });
  }
  return { success: true, products };
 } catch { return { success: false, error: "Could not load billing plans. Check the billing configuration and retry." }; }
}
