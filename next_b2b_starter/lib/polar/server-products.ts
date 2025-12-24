/**
 * Server-side Products Fetching
 *
 * This module provides server-side data fetching for Polar products.
 * Use this in Server Components and Server Actions.
 * For Client Components, use the useProductsQuery hook instead.
 */

import { getMemberSession } from "@/lib/auth/stytch/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { getPolarClient } from "@/lib/polar/client";
import type { PolarPlan } from "./plans";

interface FetchProductsResult {
  success: boolean;
  products?: PolarPlan[];
  error?: string;
}

/**
 * Fetch products from Polar (Server-side)
 *
 * Fetches all active products from Polar with authentication and permission checks.
 * This function can only be called from Server Components or Server Actions.
 *
 * @returns Products array or error
 */
export async function fetchProducts(): Promise<FetchProductsResult> {
  // Authentication check
  const session = await getMemberSession();
  if (!session?.session_jwt) {
    console.info("[Polar Products] Unauthenticated request");
    return {
      success: false,
      error: "Authentication required.",
    };
  }

  const permissions = await getServerPermissions(session);
  if (!permissions.backendAvailable) {
    console.warn("[Polar Products] Backend unavailable", {
      error: permissions.backendError,
    });
    return {
      success: false,
      error: "Service temporarily unavailable",
    };
  }

  const profile = permissions.profile;
  if (!profile) {
    console.warn("[Polar Products] Profile unavailable");
    return {
      success: false,
      error: "Profile not available.",
    };
  }

  if (!permissions.canManageSubscriptions) {
    console.info("[Polar Products] Forbidden - insufficient permissions", {
      memberId: profile.member_id,
    });
    return {
      success: false,
      error: "You do not have access to manage subscriptions.",
    };
  }

  const client = getPolarClient();
  if (!client) {
    console.warn("[Polar Products] Polar client unavailable");
    return {
      success: false,
      error: "Billing service not configured.",
    };
  }

  try {
    console.info("[Polar Products] Fetching products");

    // Fetch all products from Polar
    const response = await client.products.list({
      // Only return active, non-archived products
      isArchived: false,
    });

    const products = response.result.items;

    console.info("[Polar Products] Products fetched successfully", {
      count: products.length,
      products: products.map((p) => ({
        id: p.id,
        name: p.name,
        pricesCount: p.prices?.length ?? 0,
      })),
    });

    // Transform products to PolarPlan format
    const transformedProducts: PolarPlan[] = products
      .reduce((acc, product) => {
        const price = product.prices?.[0];
        if (!price || !price.id) {
          console.warn("[Polar Products] Product has no usable price", {
            productId: product.id,
            name: product.name,
          });
          return acc;
        }

        const metadata = (product.metadata ?? {}) as Record<string, unknown>;

        const includedSeats =
          typeof metadata.included_seats === "number"
            ? metadata.included_seats
            : typeof metadata.max_seats === "number"
              ? metadata.max_seats
              : typeof metadata.seats === "number"
                ? metadata.seats
                : null;

        const includedInvoices =
          typeof metadata.included_invoices === "number"
            ? metadata.included_invoices
            : typeof metadata.invoice_limit === "number"
              ? metadata.invoice_limit
              : typeof metadata.invoices === "number"
                ? metadata.invoices
                : null;

        const benefits =
          product.benefits?.map((b) => b.description).filter(Boolean) ?? [];

        const planId = typeof metadata.plan_id === "string" ? metadata.plan_id : product.id;

        const plan: PolarPlan = {
          id: planId,
          name: product.name,
          description: product.description ?? null,
          price:
            price.amountType === "fixed" && price.priceAmount
              ? price.priceAmount / 100
              : 0,
          interval: (price.recurringInterval as "month" | "year") ?? "month",
          productId: product.id,
          priceId: price.id,
          includedSeats,
          includedInvoices,
          benefits,
          metadata,
        };

        acc.push(plan);
        return acc;
      }, [] as PolarPlan[])
      .sort((a, b) => a.price - b.price);

    return {
      success: true,
      products: transformedProducts,
    };
  } catch (error) {
    console.error("[Polar Products] Failed to fetch products", {
      error: error instanceof Error ? error.message : String(error),
      organizationId: profile.organization?.organization_id,
    });

    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to fetch products",
    };
  }
}
