"use server";

import {
  resolveCurrentSubscription,
  type SubscriptionGateState,
} from "@/lib/polar/current-subscription";
import {
  createActionError,
  createActionSuccess,
  type ActionResult,
} from "@/lib/utils/server-action-helpers";

/**
 * Get Subscription Status Server Action
 *
 * Fetches the current subscription status from Polar.
 * This replaces the /api/billing/status API route.
 *
 * @returns ActionResult with subscription state or error
 */
export async function getSubscriptionStatus(): Promise<
  ActionResult<SubscriptionGateState>
> {
  const state = await resolveCurrentSubscription();

  if (!state.isAuthenticated) {
    return createActionError(state.reason ?? "Authentication required.");
  }

  if (state.reason === "INSUFFICIENT_PERMISSIONS") {
    return createActionError("You cannot view subscription details.");
  }

  return createActionSuccess(state);
}
