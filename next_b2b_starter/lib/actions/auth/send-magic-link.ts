"use server";
import { headers } from "next/headers";
import { requestLogin } from "@/lib/auth/flows";
import { createActionError, createActionSuccess } from "@/lib/utils/server-action-helpers";
export async function sendMagicLink(email: string, returnTo?: string) {
  try { return createActionSuccess(await requestLogin(new Headers(await headers()), email, returnTo)); }
  catch { return createActionError("Unable to send a link. Check the email address or wait a few minutes before trying again."); }
}
