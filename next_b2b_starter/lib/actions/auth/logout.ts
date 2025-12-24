"use server";

import { redirect } from "next/navigation";
import { cookies } from "next/headers";
import { getStytchB2BClient } from "@/lib/auth/stytch/server";
import {
  SESSION_COOKIE_NAME,
  SESSION_JWT_COOKIE_NAME,
} from "@/lib/auth/constants";

/**
 * Logout Server Action
 *
 * Revokes the Stytch session and clears session cookies.
 * Redirects user to the specified path or home page.
 *
 * @param returnTo - Optional path to redirect to after logout (must start with /)
 */
export async function logout(returnTo?: string): Promise<never> {
  const cookieStore = await cookies();

  const sessionToken = cookieStore.get(SESSION_COOKIE_NAME)?.value;
  const sessionJwt = cookieStore.get(SESSION_JWT_COOKIE_NAME)?.value;

  // Revoke the session with Stytch if we have a token
  if (sessionToken || sessionJwt) {
    try {
      const client = getStytchB2BClient();

      if (sessionToken) {
        await client.sessions.revoke({ session_token: sessionToken });
        console.info("[Logout] Session revoked via session_token");
      } else if (sessionJwt) {
        await client.sessions.revoke({ session_jwt: sessionJwt });
        console.info("[Logout] Session revoked via session_jwt");
      }
    } catch (error) {
      // Silently fail - user is logging out anyway
      // Session might already be expired or invalid
      console.warn("[Logout] Failed to revoke session (continuing anyway):", error);
    }
  }

  // Clear session cookies
  cookieStore.delete(SESSION_COOKIE_NAME);
  cookieStore.delete(SESSION_JWT_COOKIE_NAME);

  console.info("[Logout] Session cookies cleared");

  // Validate returnTo path for security
  const redirectPath = resolveReturnTo(returnTo);

  // Redirect to the specified path or home
  redirect(redirectPath);
}

/**
 * Resolve and validate the returnTo parameter
 * Prevents open redirect vulnerabilities
 */
function resolveReturnTo(returnTo?: string): string {
  if (!returnTo) return "/";

  const trimmed = returnTo.trim();

  // Must start with / and must NOT start with // (which could be a protocol-relative URL)
  if (!trimmed.startsWith("/") || trimmed.startsWith("//")) {
    console.warn(
      "[Logout] Invalid returnTo path (using default):",
      trimmed
    );
    return "/";
  }

  return trimmed;
}
