"use client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useProductsQuery } from "@/lib/hooks/queries/use-products-query";
import { createCheckout } from "@/lib/actions/billing/create-checkout";
import { openBillingPortal } from "@/lib/actions/billing/open-portal";
import type { SubscriptionGateState } from "@/lib/polar/current-subscription";
import { checkoutUnavailableReason, subscriptionSummary } from "@/lib/polar/subscription-policy";
export function SubscriptionTab({ state, isLoading, error, onRefresh }: {
 state: SubscriptionGateState | null; isLoading: boolean; error: string | null; onRefresh: () => void;
}) {
 const canCheckout = !isLoading && !error && checkoutUnavailableReason(state) === null;
 const plans = useProductsQuery({ enabled: canCheckout });
 const summary = subscriptionSummary(error ? null : state);
 const [pending, setPending] = useState(false);
 const [actionError, setActionError] = useState<string | null>(null);
 async function start(productId?: string) {
  setPending(true); setActionError(null);
  try {
   const result = productId ? await createCheckout(productId) : await openBillingPortal();
   if (result.success) window.location.assign(result.data.url);
   else setActionError(result.error);
  } catch { setActionError("Could not connect to billing. Please retry."); }
  finally { setPending(false); }
 }
 if (isLoading) return <p>Loading billing…</p>;
 return <div className="space-y-6">
   {(error || state?.backendError || actionError) && <p role="alert" className="text-red-700">{error || state?.backendError || actionError}</p>}
   <div className="rounded-xl border p-6">
    <h3 className="text-lg font-semibold">{summary.title}</h3>
    <p className="mt-2 text-sm text-gray-600">{summary.description}</p>
    <p className="mt-2 text-sm text-gray-600">This starter supports one subscription plan. Plan switching is not available.</p>
    {!error && state?.backendAvailable && state.subscription?.currentPeriodEnd && <p className="mt-2 text-sm">Current period ends {new Date(state.subscription.currentPeriodEnd).toLocaleDateString()}</p>}
    <div className="mt-4 flex flex-wrap gap-3"><Button disabled={pending || Boolean(error) || !state?.isAuthenticated || state.reason === "BILLING_DISABLED" || state.reason === "INSUFFICIENT_PERMISSIONS"} onClick={() => start()}>Open billing portal</Button><Button variant="outline" disabled={pending} onClick={() => { setActionError(null); onRefresh(); }}>Refresh status</Button></div>
   </div>
   {canCheckout && <div className="grid gap-4 sm:grid-cols-2">
     {plans.isLoading && <p role="status">Loading the subscription plan…</p>}
     {plans.error && <p role="alert">{plans.error.message}</p>}
     {plans.data?.length === 0 && <p>No subscription plans are available.</p>}
     {plans.data?.map((plan) => <div key={plan.id} className="rounded-xl border p-6">
       <h3 className="font-semibold">{plan.name}</h3><p className="mt-2 text-sm text-gray-600">{plan.description}</p>
       <p className="my-4">{new Intl.NumberFormat(undefined, { style: "currency", currency: plan.currency }).format(plan.price)} / {plan.interval}</p>
       <Button disabled={pending} onClick={() => start(plan.productId)}>Subscribe</Button>
     </div>)}
   </div>}
 </div>;
}
