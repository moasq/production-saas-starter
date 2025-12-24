import type { Subscription } from "@polar-sh/sdk/models/components/subscription";
import type { Meter } from "@polar-sh/sdk/models/components/meter";

import { getPolarClient } from "@/lib/polar/client";
import { POLAR_METER_ID } from "@/lib/polar/config";

export interface MeterUsageSummary {
  meterId: string;
  customerId: string;
  included: number;
  used: number;
  remaining: number;
  periodStart: Date;
  periodEnd: Date;
}

/**
 * Get invoice usage using Polar Meters Quantities API
 *
 * Finds the meter by name "invoice.processed" with count aggregation,
 * then fetches usage via meters.quantities() API.
 */
export async function getInvoiceUsage(
  subscription: Subscription
): Promise<MeterUsageSummary | null> {
  const client = getPolarClient();
  if (!client) {
    return null;
  }

  try {
    const start = subscription.currentPeriodStart;
    const end = subscription.currentPeriodEnd ?? new Date();

    // Get external customer ID from subscription (note: it's 'externalId' in the SDK)
    const externalCustomerId = subscription.customer?.externalId ?? null;

    // 1. List all meters to find the invoice meter
    const meters = await listMeters();


    // 2. Helper function to recursively search for matching filter clause
    const hasMatchingFilter = (clauses: any[]): boolean => {
      for (const clause of clauses) {
        // Check if this is a direct filter clause with property field
        if ('property' in clause) {
          const matches = (
            clause.property === "name" &&
            clause.value === "invoice.processed" &&
            (clause.operator === "eq" || clause.operator === "equals")
          );

          if (matches) {
            return true;
          }
        }

        // Check if this is a nested filter with its own clauses array
        if ('clauses' in clause && Array.isArray(clause.clauses)) {
          if (hasMatchingFilter(clause.clauses)) {
            return true;
          }
        }
      }
      return false;
    };

    // 3. Find meter that tracks "invoice.processed" events with count aggregation
    // The event name is in the filter clauses (potentially nested), not the meter name
    const invoiceMeter = meters.find((m) => {
      // Skip archived meters - we only want active meters
      if (m.archivedAt) {
        return false;
      }

      // Check if meter has count aggregation
      if (m.aggregation?.func !== "count") {
        return false;
      }

      // Check if filter exists
      const filter = m.filter;
      if (!filter || !filter.clauses) {
        return false;
      }

      // Recursively search for matching filter clause
      return hasMatchingFilter(filter.clauses);
    });

    // 4. Use only the found meter - no hardcoded fallbacks
    if (!invoiceMeter) {
      return null;
    }

    const meterId = invoiceMeter.id;

    // 5. Fetch quantities with customer filtering
    const response = await client.meters.quantities({
      id: meterId,
      startTimestamp: start,
      endTimestamp: end,
      interval: "month",
      customerId: subscription.customerId,
    });

    // 5. Extract total usage and get included invoices from product metadata
    const used = response.total;

    // Get included invoices from product metadata
    const productMetadata = subscription.product?.metadata ?? {};
    const included =
      typeof productMetadata.included_invoices === "number"
        ? productMetadata.included_invoices
        : typeof productMetadata.invoice_limit === "number"
          ? productMetadata.invoice_limit
          : typeof productMetadata.invoices === "number"
            ? productMetadata.invoices
            : 1000; // Default fallback if not in metadata

    const remaining = Math.max(0, included - used);

    return {
      meterId,
      customerId: subscription.customerId,
      included,
      used,
      remaining,
      periodStart: start,
      periodEnd: end,
    };
  } catch {
    return null;
  }
}

/**
 * List all available meters
 *
 * Returns all meters configured in Polar (e.g., "Invoice Processing")
 * Use this to discover available meters and their configurations.
 */
export async function listMeters(): Promise<Meter[]> {
  const client = getPolarClient();
  if (!client) {
    return [];
  }

  try {
    const response = await client.meters.list({});
    return response.result.items;
  } catch {
    return [];
  }
}

/**
 * Get customer usage for a specific meter
 *
 * NOTE: The SDK doesn't have a `customers()` endpoint yet.
 * This function is a placeholder for when that endpoint is available.
 * Currently returns null.
 */
export async function getCustomerMeterUsage(_meterId: string) {
  // Not yet implemented - SDK lacks meters.customers() endpoint
  return null;
}
