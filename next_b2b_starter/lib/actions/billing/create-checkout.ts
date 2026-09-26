"use server";
import { getMemberSession } from "@/lib/auth/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { getBaseUrl } from "@/lib/auth/urls";
import { getPolarClient } from "@/lib/polar/client";
import { fetchProducts } from "@/lib/polar/server-products";
import { resolveCurrentSubscription } from "@/lib/polar/current-subscription";
import { checkoutUnavailableReason } from "@/lib/polar/subscription-policy";
import { createActionError, createActionSuccess, type ActionResult } from "@/lib/utils/server-action-helpers";
export async function createCheckout(productId: string): Promise<ActionResult<{ url: string }>> {
 const session = await getMemberSession();
 if (!session) return createActionError("Authentication required.");
 const permissions = await getServerPermissions(session);
 const orgId = permissions.profile?.organization?.organization_id;
 if (!orgId || !permissions.canManageSubscriptions) return createActionError("Organization administrator access required.");
 try {
  const client = getPolarClient();
  if (!client) return createActionError("Billing is disabled.");
  if (!process.env.POLAR_PRODUCT_ID || productId !== process.env.POLAR_PRODUCT_ID) return createActionError("Select an available plan.");
  const state = await resolveCurrentSubscription();
  const unavailableReason = checkoutUnavailableReason(state);
  if (unavailableReason) return createActionError(unavailableReason);
  const result = await fetchProducts();
  if (!result.success || !result.products?.some((p) => p.productId === productId)) return createActionError("Select an available plan.");
  const checkout = await client.checkouts.create({ products: [productId], externalCustomerId: orgId,
    customerEmail: permissions.profile!.email, customerName: permissions.profile!.name,
    successUrl: `${getBaseUrl()}/dashboard?checkout_id={CHECKOUT_ID}`, returnUrl: `${getBaseUrl()}/dashboard/settings?view=subscription` });
  return createActionSuccess({ url: checkout.url });
 } catch { return createActionError("Could not start checkout. Please retry."); }
}
