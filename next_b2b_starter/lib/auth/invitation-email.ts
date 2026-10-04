import { APIError } from "better-auth/api";
import { sendAuthEmail } from "./email.ts";
import { consumeInvitationLimit } from "./rate-limit.ts";

// Reserve before creating/resending the invitation. Better Auth intentionally
// swallows email callback errors, so the private bridge owns and awaits delivery.
export async function reserveInvitationDelivery(organizationId: string, email: string): Promise<void> {
  if (!await consumeInvitationLimit(organizationId, email)) {
    throw new APIError("TOO_MANY_REQUESTS", { message: "Too many invitation requests. Please wait before trying again." });
  }
}

export async function sendInvitationEmail(
  data: { id: string; email: string; organization: { id: string; name: string } },
  baseURL: string,
): Promise<void> {
  const url = new URL("/invite", baseURL);
  url.searchParams.set("id", data.id);
  await sendAuthEmail(data.email, `Invitation to ${data.organization.name}`,
    `You have been invited to ${data.organization.name}. Sign in with this email address and accept your invitation:\n\n${url}\n\nThis invitation expires in 48 hours.`);
}
