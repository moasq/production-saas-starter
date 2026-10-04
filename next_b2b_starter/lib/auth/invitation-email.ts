import { APIError } from "better-auth/api";
import { sendAuthEmail } from "./email.ts";
import { consumeInvitationLimit } from "./rate-limit.ts";

// Better Auth calls this for initial invitations and resends, including direct
// server API calls that do not pass through its public HTTP rate limiter.
export async function sendInvitationEmail(
  data: { id: string; email: string; organization: { id: string; name: string } },
  baseURL: string,
): Promise<void> {
  if (!await consumeInvitationLimit(data.organization.id, data.email)) {
    throw new APIError("TOO_MANY_REQUESTS", { message: "Too many invitation requests. Please wait before trying again." });
  }
  const url = new URL("/invite", baseURL);
  url.searchParams.set("id", data.id);
  await sendAuthEmail(data.email, `Invitation to ${data.organization.name}`,
    `You have been invited to ${data.organization.name}. Sign in with this email address and accept your invitation:\n\n${url}\n\nThis invitation expires in 48 hours.`);
}
