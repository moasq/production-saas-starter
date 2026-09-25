"use server";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { getAuth } from "@/lib/auth/configuration";
import { requireOrigin } from "@/lib/auth/flows";
import { sanitizeReturnTo } from "@/lib/auth/urls";
export async function logout(returnTo?: string): Promise<never> {
  const requestHeaders = new Headers(await headers());
  requireOrigin(requestHeaders);
  // A failed revocation must not be reported as a successful logout.
  await getAuth().api.signOut({ headers: requestHeaders });
  redirect(sanitizeReturnTo(returnTo) || "/");
}
