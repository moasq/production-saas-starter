"use server";

import { cookies } from "next/headers";
import { getStytchB2BClient } from "@/lib/auth/stytch/server";
import {
  SESSION_COOKIE_NAME,
  SESSION_JWT_COOKIE_NAME,
} from "@/lib/auth/constants";
import {
  getSessionDurationMinutes,
  getCookieConfig,
  getSecureCookieConfig,
} from "@/lib/auth/server-constants";
import {
  createActionError,
  createActionSuccess,
  type ActionResult,
} from "@/lib/utils/server-action-helpers";

export interface ConsumeMagicLinkResult {
  memberAuthenticated: boolean;
  intermediateSessionToken?: string;
  member?: {
    member_id: string;
    email_address: string;
    name: string;
  };
  organization?: {
    organization_id: string;
    organization_name: string;
  };
  mfaRequired?: unknown;
  primaryRequired?: unknown;
}

/**
 * Consume Magic Link Server Action
 *
 * Exchanges a magic link token for a session.
 * Sets session cookies on successful authentication.
 *
 * @param token - The magic link token from the URL
 * @param sessionDurationMinutes - Optional session duration (defaults to env config)
 * @returns ActionResult with authentication status
 */
export async function consumeMagicLink(
  token: string,
  sessionDurationMinutes?: number
): Promise<ActionResult<ConsumeMagicLinkResult>> {
  if (!token) {
    return createActionError("Magic link token is required.");
  }

  const duration = sessionDurationMinutes || getSessionDurationMinutes();

  try {
    const client = getStytchB2BClient();

    const result = await client.magicLinks.authenticate({
      magic_links_token: token,
      session_duration_minutes: duration,
    });

    if (!result.member_authenticated) {
      return createActionSuccess({
        memberAuthenticated: false,
        intermediateSessionToken: result.intermediate_session_token,
        member: result.member
          ? {
              member_id: result.member.member_id,
              email_address: result.member.email_address,
              name: result.member.name,
            }
          : undefined,
        organization: result.organization
          ? {
              organization_id: result.organization.organization_id,
              organization_name: result.organization.organization_name,
            }
          : undefined,
        mfaRequired: result.mfa_required ?? false,
        primaryRequired: result.primary_required ?? false,
      });
    }

    // Set session cookies
    const cookieStore = await cookies();
    const maxAgeSeconds = duration * 60;

    if (result.session_token) {
      cookieStore.set(SESSION_COOKIE_NAME, result.session_token, {
        ...getSecureCookieConfig(),
        maxAge: maxAgeSeconds,
      });
    }

    if (result.session_jwt) {
      cookieStore.set(SESSION_JWT_COOKIE_NAME, result.session_jwt, {
        ...getCookieConfig(),
        maxAge: maxAgeSeconds,
      });
    }

    return createActionSuccess({
      memberAuthenticated: true,
      member: result.member
        ? {
            member_id: result.member.member_id,
            email_address: result.member.email_address,
            name: result.member.name,
          }
        : undefined,
      organization: result.organization
        ? {
            organization_id: result.organization.organization_id,
            organization_name: result.organization.organization_name,
          }
        : undefined,
    });
  } catch (error: any) {
    const errorMessage =
      error?.error_message || error?.message || "Unable to verify magic link.";

    return createActionError(errorMessage);
  }
}
