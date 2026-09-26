import type { SubscriptionGateState } from "./current-subscription.ts";

// A false active flag alone can also mean an unavailable provider, disabled
// billing or revoked access. Only a successful authorized status read allows
// checkout. The server action re-reads this state before contacting Polar.
export function checkoutUnavailableReason(state: SubscriptionGateState | null): string | null {
  if (!state?.isAuthenticated || state.reason === "INSUFFICIENT_PERMISSIONS") {
    return "Organization administrator access required.";
  }
  if (state.reason === "BILLING_DISABLED") return "Billing is disabled.";
  if (state.isActive || state.subscription) {
    return "An existing subscription cannot be replaced through checkout. Open the billing portal for payments and cancellation.";
  }
  if (!state.backendAvailable || !state.canStartCheckout) {
    return "Could not verify your current subscription. Refresh status before starting checkout.";
  }
  return null;
}

export function subscriptionSummary(state: SubscriptionGateState | null): { title: string; description: string } {
  if (!state || !state.backendAvailable) {
    return { title: "Billing status unavailable", description: "Refresh status before starting checkout." };
  }
  if (state.reason === "BILLING_DISABLED") {
    return { title: "Billing is disabled", description: "Your workspace is available without a subscription." };
  }
  if (!state.isAuthenticated || state.reason === "INSUFFICIENT_PERMISSIONS") {
    return { title: "Billing access unavailable", description: "Organization administrator access is required." };
  }
  if (state.isActive) {
    return {
      title: state.subscription?.cancelAtPeriodEnd ? "Cancellation scheduled" : "Active subscription",
      description: state.subscription?.cancelAtPeriodEnd
        ? "Your subscription remains active until the current period ends. Open the billing portal to review cancellation."
        : "Open the billing portal for payments and cancellation.",
    };
  }
  if (!state.canStartCheckout) {
    return { title: "Billing status unavailable", description: "Refresh status before starting checkout." };
  }
  return { title: "No active subscription", description: "Subscribe to the configured plan. Your workspace remains available without a subscription." };
}
