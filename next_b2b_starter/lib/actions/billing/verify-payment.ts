"use server";
import { getMemberSession } from "@/lib/auth/server";
import { apiClient, unwrap } from "@/lib/api/api/client/api-client";
import { isPolarEnabled } from "@/lib/polar/config";
import type { BillingStatus } from "@/lib/polar/current-subscription";
import { createActionError, createActionSuccess, type ActionResult } from "@/lib/utils/server-action-helpers";
export async function verifyPayment(sessionId: string): Promise<ActionResult<BillingStatus>> {
 if (!isPolarEnabled()) return createActionError("Billing is disabled.");
 const session = await getMemberSession();
 if (!session?.cookie_header) return createActionError("Authentication required.");
 if (!sessionId || sessionId.length > 200) return createActionError("Invalid checkout session.");
 try {
  const status = unwrap(await apiClient.POST("/subscriptions/verify-payment", { body: { session_id: sessionId }, headers: { Cookie: session.cookie_header } }));
  return createActionSuccess(status);
 } catch { return createActionError("Your payment is not yet verified. Refresh billing after checkout completes."); }
}
