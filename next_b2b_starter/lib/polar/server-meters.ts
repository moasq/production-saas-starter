/**
 * Server-side Meters Fetching
 *
 * This module provides server-side data fetching for Polar meters and usage data.
 * Use this in Server Components and Server Actions.
 */

import { getMemberSession } from "@/lib/auth/stytch/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { listMeters } from "@/lib/polar/usage";
import { POLAR_METER_ID } from "@/lib/polar/config";

interface MeterInfo {
  id: string;
  name: string;
  aggregation: string | null;
  filter: Record<string, unknown> | null;
}

interface FetchMetersResult {
  success: boolean;
  data?: {
    meters: MeterInfo[];
    invoiceMeter: MeterInfo | null;
    organizationUsage: unknown | null;
    customerUsage: unknown | null;
  };
  error?: string;
}

/**
 * Fetch meters data from Polar (Server-side)
 *
 * Fetches all available meters and invoice meter details with authentication and permission checks.
 * This function can only be called from Server Components or Server Actions.
 *
 * @returns Meters data or error
 */
export async function fetchMeters(): Promise<FetchMetersResult> {
  // Authentication check
  const session = await getMemberSession();
  if (!session?.session_jwt) {
    console.info("[Polar Meters] Unauthenticated request");
    return {
      success: false,
      error: "Authentication required.",
    };
  }

  const permissions = await getServerPermissions(session);
  if (!permissions.backendAvailable) {
    console.warn("[Polar Meters] Backend unavailable", {
      error: permissions.backendError,
    });
    return {
      success: false,
      error: "Service temporarily unavailable",
    };
  }

  const profile = permissions.profile;
  if (!profile) {
    console.warn("[Polar Meters] Profile unavailable");
    return {
      success: false,
      error: "Profile not available.",
    };
  }

  if (!permissions.canManageSubscriptions) {
    console.info("[Polar Meters] Forbidden - insufficient permissions", {
      memberId: profile.member_id,
    });
    return {
      success: false,
      error: "You do not have access to manage subscriptions.",
    };
  }

  try {
    // Fetch all meters
    const meters = await listMeters();

    // Find the invoice meter details
    const invoiceMeter = meters.find((m) => m.id === POLAR_METER_ID);

    console.info("[Polar Meters] Meters data fetched successfully", {
      totalMeters: meters.length,
      invoiceMeterId: POLAR_METER_ID,
      hasInvoiceMeter: Boolean(invoiceMeter),
      organizationId: profile.organization?.organization_id,
    });

    // NOTE: Customer usage API not yet available in SDK
    // When available, we can fetch per-customer usage here

    return {
      success: true,
      data: {
        // All available meters
        meters: meters.map((m) => ({
          id: m.id,
          name: m.name,
          aggregation: m.aggregation?.func ?? null,
          filter: m.filter,
        })),

        // Invoice meter details
        invoiceMeter: invoiceMeter
          ? {
              id: invoiceMeter.id,
              name: invoiceMeter.name,
              aggregation: invoiceMeter.aggregation?.func ?? null,
              filter: invoiceMeter.filter,
            }
          : null,

        // Placeholders for when SDK supports customer usage endpoint
        organizationUsage: null,
        customerUsage: null,
      },
    };
  } catch (error) {
    console.error("[Polar Meters] Failed to fetch meters data", {
      error: error instanceof Error ? error.message : String(error),
      organizationId: profile.organization?.organization_id,
    });

    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to fetch meters data",
    };
  }
}
