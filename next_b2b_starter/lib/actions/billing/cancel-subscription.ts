"use server";

import { getMemberSession } from "@/lib/auth/stytch/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { getActiveSubscription } from "@/lib/polar/subscription";
import { getPolarClient } from "@/lib/polar/client";
import type { SubscriptionCancel } from "@polar-sh/sdk/models/components/subscriptioncancel";
import type { CustomerCancellationReason } from "@polar-sh/sdk/models/components/customercancellationreason";
import {
  createActionError,
  createActionSuccess,
  type ActionResult
} from "@/lib/utils/server-action-helpers";

const ALLOWED_REASONS = new Set([
  "too_expensive",
  "missing_features",
  "switched_service",
  "unused",
  "customer_service",
  "low_quality",
  "too_complex",
  "other",
]);

interface CancelSubscriptionParams {
  cancelAtPeriodEnd?: boolean;
  reason?: string | null;
  comment?: string | null;
}

interface CancelSubscriptionData {
  success: boolean;
  cancelAtPeriodEnd: boolean;
  status: string | null;
  subscription: {
    id: string;
    status: string;
    cancelAtPeriodEnd: boolean;
    currentPeriodEnd: string | null;
  } | null;
}

/**
 * Cancel Subscription Server Action
 *
 * Updates a subscription with cancellation settings.
 * Can cancel immediately or at the end of the billing period.
 *
 * @param params - Optional cancellation parameters
 */
export async function cancelSubscription(
  params?: CancelSubscriptionParams
): Promise<ActionResult<CancelSubscriptionData>> {
  const client = getPolarClient();

  if (!client) {
    return createActionError(
      "Polar billing is not configured.",
      "Missing Polar client"
    );
  }

  // Authenticate user
  const session = await getMemberSession();
  if (!session?.session_jwt) {
    return createActionError("Authentication required.");
  }

  // Check permissions
  const permissions = await getServerPermissions(session);
  const profile = permissions.profile;

  if (!profile?.organization?.organization_id) {
    return createActionError(
      "Organization context required to manage subscriptions."
    );
  }

  if (!permissions.canManageSubscriptions) {
    console.info("[Polar] Subscription cancel forbidden - insufficient permissions", {
      memberId: profile.member_id,
    });
    return createActionError(
      "You do not have access to manage subscriptions.",
      "Missing subscription management permissions"
    );
  }

  // Parse parameters
  const cancelAtPeriodEnd =
    typeof params?.cancelAtPeriodEnd === "boolean" ? params.cancelAtPeriodEnd : true;

  const reason =
    typeof params?.reason === "string" && ALLOWED_REASONS.has(params.reason)
      ? params.reason
      : undefined;

  const comment =
    typeof params?.comment === "string" && params.comment.trim().length > 0
      ? params.comment.trim()
      : undefined;

  // Get active subscription
  const subscriptionResult = await getActiveSubscription({
    externalCustomerId: profile.organization.organization_id,
    customerEmail: profile.email,
    organizationId: profile.organization.organization_id,
  });

  const subscription = subscriptionResult.subscription;

  if (!subscription) {
    return createActionError(
      "No active subscription to update.",
      "User does not have an active subscription"
    );
  }

  // Prepare update payload
  const subscriptionUpdatePayload: SubscriptionCancel = {
    cancelAtPeriodEnd,
  };

  if (reason) {
    subscriptionUpdatePayload.customerCancellationReason = reason as CustomerCancellationReason;
  }

  if (comment) {
    subscriptionUpdatePayload.customerCancellationComment = comment;
  }

  // Update subscription
  try {
    await client.subscriptions.update({
      id: subscription.id,
      subscriptionUpdate: subscriptionUpdatePayload,
    });
  } catch (error) {
    console.error("[Polar] Failed to update subscription cancellation", error);
    return createActionError(
      "Failed to update subscription status. Please try again.",
      error instanceof Error ? error.message : "Unknown error"
    );
  }

  // Refresh subscription data
  const refreshed = await getActiveSubscription({
    externalCustomerId: profile.organization.organization_id,
    customerEmail: profile.email,
    organizationId: profile.organization.organization_id,
  });

  return createActionSuccess({
    success: true,
    cancelAtPeriodEnd,
    status: refreshed.status,
    subscription: refreshed.subscription
      ? {
          id: refreshed.subscription.id,
          status: refreshed.subscription.status,
          cancelAtPeriodEnd: refreshed.subscription.cancelAtPeriodEnd,
          currentPeriodEnd: refreshed.subscription.currentPeriodEnd?.toISOString() ?? null,
        }
      : null,
  });
}
