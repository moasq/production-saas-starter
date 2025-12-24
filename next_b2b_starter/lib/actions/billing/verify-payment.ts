"use server";

import { getMemberSession } from "@/lib/auth/stytch/server";
import { apiClient } from "@/lib/api/api/client/api-client";
import {
  createActionError,
  createActionSuccess,
  type ActionResult
} from "@/lib/utils/server-action-helpers";

interface BillingStatus {
  organization_id: number;
  external_id?: string;
  has_active_subscription: boolean;
  can_process_invoices: boolean;
  invoice_count: number;
  reason: string;
  checked_at: string;
}

/**
 * Verify Payment Server Action
 *
 * Verifies a payment from a Polar checkout session by calling the Go backend.
 * Updates the organization's subscription status in the database.
 *
 * @param sessionId - The Polar checkout session ID
 */
export async function verifyPayment(
  sessionId: string
): Promise<ActionResult<BillingStatus>> {
  // Authenticate user
  const session = await getMemberSession();
  if (!session?.session_jwt) {
    console.info("[Billing] verify-payment attempted without authentication");
    return createActionError("Authentication required.");
  }

  // Validate session_id
  if (!sessionId || typeof sessionId !== "string") {
    return createActionError(
      "session_id is required.",
      "Invalid or missing session_id parameter"
    );
  }

  try {
    console.info("[Billing] Verifying payment from checkout session", {
      sessionId,
    });

    // Call Go backend to verify payment
    const billingStatus = await apiClient.post<BillingStatus>(
      "/subscriptions/verify-payment",
      { session_id: sessionId }
    );

    console.info("[Billing] Payment verification successful", {
      sessionId,
      hasActiveSubscription: billingStatus.has_active_subscription,
      reason: billingStatus.reason,
    });

    return createActionSuccess(billingStatus);
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : "Unknown error";
    console.error("[Billing] Payment verification failed", {
      sessionId,
      error: errorMessage,
    });

    // Check if it's a 404 (session not found)
    if (errorMessage.includes("404")) {
      return createActionError(
        "Checkout session not found.",
        errorMessage
      );
    }

    return createActionError(
      "Failed to verify payment.",
      errorMessage
    );
  }
}
