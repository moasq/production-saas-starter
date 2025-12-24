"use server";

import { fetchProducts } from "@/lib/polar/server-products";
import type { PolarPlan } from "@/lib/polar/plans";
import {
  createActionError,
  createActionSuccess,
  type ActionResult,
} from "@/lib/utils/server-action-helpers";

/**
 * Get Products Server Action
 *
 * Fetches all active products from Polar with authentication and permission checks.
 * This replaces the /api/billing/products API route.
 *
 * @returns ActionResult with products array or error
 */
export async function getProducts(): Promise<ActionResult<PolarPlan[]>> {
  const result = await fetchProducts();

  if (!result.success) {
    return createActionError(result.error ?? "Failed to fetch products");
  }

  return createActionSuccess(result.products ?? []);
}
