"use server";
import { getMemberSession } from "@/lib/auth/stytch/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";
import { getPolarClient } from "@/lib/polar/client";
import { createActionError, createActionSuccess, type ActionResult } from "@/lib/utils/server-action-helpers";
export async function openBillingPortal(): Promise<ActionResult<{ url: string }>> {
 const session = await getMemberSession();
 if (!session) return createActionError("Authentication required.");
 const permissions = await getServerPermissions(session);
 const orgId = permissions.profile?.organization?.organization_id;
 if (!orgId || !permissions.canManageSubscriptions) return createActionError("Organization administrator access required.");
 try {
  const client = getPolarClient();
  if (!client) return createActionError("Billing is disabled.");
  const portal = await client.customerSessions.create({ externalCustomerId: orgId });
  return createActionSuccess({ url: portal.customerPortalUrl });
 } catch { return createActionError("Could not open the billing portal. Please retry."); }
}
