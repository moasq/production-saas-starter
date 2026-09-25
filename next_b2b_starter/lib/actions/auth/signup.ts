"use server";
import { headers } from "next/headers";
import { requestSignup } from "@/lib/auth/flows";
import { createActionError, createActionSuccess } from "@/lib/utils/server-action-helpers";
export async function signup(input: { email: string; name: string; organizationName: string }) {
  try { return createActionSuccess(await requestSignup(new Headers(await headers()), input)); }
  catch { return createActionError("Unable to send a signup link. Check your details or wait a few minutes before trying again."); }
}
