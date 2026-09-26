import "server-only";
import { getMemberSession } from "@/lib/auth/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { apiClient, unwrap } from "@/lib/api/api/client/api-client";
import { isPolarEnabled } from "./config";
import type { components } from "@/lib/api/generated/schema";
export type BillingStatus = components["schemas"]["BillingStatus"];
export interface SubscriptionGateState {
  isAuthenticated: boolean; isActive: boolean; reason?: string; status?: string | null;
  productId: string | null; planId: string | null;
  subscription: { id: string; status: string; productId: string; productName: string | null; currentPeriodEnd: string | null; cancelAtPeriodEnd: boolean } | null;
  backendAvailable: boolean; backendError?: string | null;
}
export async function resolveCurrentSubscription(): Promise<SubscriptionGateState> {
 const session = await getMemberSession();
 const empty: SubscriptionGateState = { isAuthenticated: Boolean(session), isActive: false, productId: null, planId: null, subscription: null, backendAvailable: true };
 if (!session?.cookie_header) return { ...empty, reason: "UNAUTHENTICATED" };
 if (!isPolarEnabled()) return { ...empty, reason: "BILLING_DISABLED" };
 const permissions = await getServerPermissions(session);
 if (!permissions.canManageSubscriptions) return { ...empty, reason: "INSUFFICIENT_PERMISSIONS" };
 try {
   const status = unwrap(await apiClient.GET("/subscriptions/status", { headers: { Cookie: session.cookie_header } }));
   return { ...empty, isActive: status.HasActiveSubscription, reason: status.Reason, status: status.SubscriptionStatus,
     productId: status.ProductID ?? null,
     subscription: status.SubscriptionID ? { id: status.SubscriptionID, status: status.SubscriptionStatus ?? "unknown", productId: status.ProductID ?? "", productName: null, currentPeriodEnd: status.CurrentPeriodEnd ?? null, cancelAtPeriodEnd: status.CancelAtPeriodEnd ?? false } : null };
 } catch { return { ...empty, backendAvailable: false, backendError: "Billing status is temporarily unavailable. Please retry.", reason: "BACKEND_UNAVAILABLE" }; }
}
