"use server";
import { headers } from "next/headers";
import { completeLogin } from "@/lib/auth/flows";
import { createActionError, createActionSuccess } from "@/lib/utils/server-action-helpers";
export async function completeAuthentication(signupId?: string) {
  try { return createActionSuccess({ destination: await completeLogin(new Headers(await headers()), signupId) }); }
  catch { return createActionError("Could not finish signing in. The link may have expired. Request a new link or sign in to choose a workspace."); }
}
