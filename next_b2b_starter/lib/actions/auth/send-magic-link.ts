"use server";

import {
  getStytchB2BClient,
  getOrganizationIdsForMemberSearch,
} from "@/lib/auth/stytch/server";
import {
  createActionSuccess,
  createActionError,
  type ActionResult,
} from "@/lib/utils/server-action-helpers";

/**
 * Send Magic Link Server Action
 *
 * Validates that an email belongs to an existing member before sending a magic link.
 * This prevents unknown users from receiving authentication emails.
 *
 * @param email - The email address to send the magic link to
 * @returns ActionResult with success message or error
 */
export async function sendMagicLink(
  email: string
): Promise<ActionResult<{ message: string }>> {
  try {
    // Validate input
    if (!email || typeof email !== "string") {
      return createActionError("Email address is required");
    }

    const client = getStytchB2BClient();
    const organizationIds = await getOrganizationIdsForMemberSearch();

    if (!organizationIds.length) {
      console.error(
        "[Magic Link] No organization IDs configured for member search."
      );
      return createActionError(
        "Unable to process request. Please try again later."
      );
    }

    // Search for members with this email across all organizations
    // This checks if the user exists in ANY organization
    const searchResult = await client.organizations.members.search({
      organization_ids: organizationIds,
      query: {
        operator: "AND",
        operands: [
          {
            filter_name: "member_emails",
            filter_value: [email.toLowerCase()],
          },
        ],
      },
    });

    // If no members found, reject without revealing this fact
    if (!searchResult.members || searchResult.members.length === 0) {
      // Return success to prevent user enumeration
      // But don't actually send an email
      console.info(
        "[Magic Link] No member found for email (not revealing to client):",
        email
      );
      return createActionSuccess({
        message:
          "If an account exists with that email, a magic link has been sent.",
      });
    }

    // Member exists - prepare login redirect URL
    const redirectUrl = process.env.NEXT_PUBLIC_APP_BASE_URL
      ? `${process.env.NEXT_PUBLIC_APP_BASE_URL}/authenticate`
      : "http://localhost:3000/authenticate";

    const memberOrganizationIds = Array.from(
      new Set(
        (searchResult.members ?? [])
          .map((member) => member.organization_id)
          .filter((orgId): orgId is string => Boolean(orgId))
      )
    );

    if (memberOrganizationIds.length === 0) {
      console.warn(
        "[Magic Link] Member search succeeded but no organization IDs were returned for email:",
        email
      );
      return createActionSuccess({
        message:
          "If an account exists with that email, a magic link has been sent.",
      });
    }

    if (memberOrganizationIds.length > 1) {
      console.warn(
        "[Magic Link] Email is associated with multiple organizations; issuing login link for all memberships.",
        {
          email,
          organizationIds: memberOrganizationIds,
        }
      );
    }

    // Send magic link for each organization the member belongs to
    await Promise.all(
      memberOrganizationIds.map((organizationId) =>
        client.magicLinks.email.loginOrSignup({
          email_address: email,
          organization_id: organizationId,
          login_redirect_url: redirectUrl,
        })
      )
    );

    console.info("[Magic Link] Successfully sent magic link to:", email);

    return createActionSuccess({
      message:
        "If an account exists with that email, a magic link has been sent.",
    });
  } catch (error: any) {
    console.error("[Magic Link] Error sending magic link:", error);

    // Return generic error to prevent user enumeration
    return createActionError(
      "Unable to process request. Please try again later.",
      process.env.NODE_ENV === "development" ? error.message : undefined
    );
  }
}
