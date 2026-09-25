"use server";
import { getMemberSession } from "@/lib/auth/stytch/server";
import { apiClient } from "@/lib/api/api/client/api-client";
import { isPolarEnabled } from "@/lib/polar/config";
import type { BillingStatus } from "@/lib/polar/current-subscription";
import { createActionError, createActionSuccess, type ActionResult } from "@/lib/utils/server-action-helpers";
export async function verifyPayment(sessionId: string): Promise<ActionResult<BillingStatus>> {
 if (!isPolarEnabled()) return createActionError("Billing is disabled.");
 const session = await getMemberSession();
 if (!session?.session_jwt) return createActionError("Authentication required.");
 if (!sessionId || sessionId.length > 200) return createActionError("Invalid checkout session.");
 try {
  const status = await apiClient.post<BillingStatus>("/subscriptions/verify-payment", { session_id: sessionId }, { headers: { Authorization: `Bearer ${session.session_jwt}` } });
  return createActionSuccess(status);
 } catch { return createActionError("Your payment is not yet verified. Refresh billing after checkout completes."); }
}
