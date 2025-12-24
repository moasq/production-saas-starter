/**
 * Server Action Helpers
 *
 * Utility functions and types for Next.js Server Actions.
 * Provides consistent patterns for authentication, permissions, and error handling.
 */

import { getMemberSession, requireMemberSession } from "@/lib/auth/stytch/server";
import { getServerPermissions, type ServerPermissions } from "@/lib/auth/server-permissions";
import type { B2BSessionsAuthenticateResponse } from "stytch";

/**
 * Standard result type for Server Actions
 * Ensures consistent return values across all actions
 */
export type ActionResult<T = void> =
  | { success: true; data: T }
  | { success: false; error: string; details?: string };

/**
 * Session result type for auth helpers
 */
export type SessionResult = {
  session: B2BSessionsAuthenticateResponse | null;
  permissions: ServerPermissions | null;
};

/**
 * Get the current member session (optional - returns null if not authenticated)
 * Use this when authentication is optional
 */
export async function getActionSession(): Promise<B2BSessionsAuthenticateResponse | null> {
  return await getMemberSession();
}

/**
 * Require authentication for a Server Action
 * Throws an error if user is not authenticated
 * Use this when authentication is mandatory
 */
export async function requireActionSession(): Promise<B2BSessionsAuthenticateResponse> {
  return await requireMemberSession();
}

/**
 * Get session and permissions together
 * Returns null for both if not authenticated
 */
export async function getActionSessionWithPermissions(): Promise<SessionResult> {
  const session = await getMemberSession();

  if (!session) {
    return { session: null, permissions: null };
  }

  const permissions = await getServerPermissions(session);

  return { session, permissions };
}

/**
 * Require session and permissions together
 * Throws error if not authenticated
 */
export async function requireActionSessionWithPermissions(): Promise<{
  session: B2BSessionsAuthenticateResponse;
  permissions: ServerPermissions;
}> {
  const session = await requireMemberSession();
  const permissions = await getServerPermissions(session);

  return { session, permissions };
}

/**
 * Create a standardized error result
 * Use this to return errors from Server Actions
 */
export function createActionError(
  error: string,
  details?: string
): ActionResult<never> {
  return {
    success: false,
    error,
    details,
  };
}

/**
 * Create a standardized success result
 * Use this to return success from Server Actions
 */
export function createActionSuccess<T>(data: T): ActionResult<T> {
  return {
    success: true,
    data,
  };
}

/**
 * Create a success result with no data
 */
export function createActionSuccessEmpty(): ActionResult<void> {
  return {
    success: true,
    data: undefined,
  };
}

/**
 * Wrap a Server Action with error handling
 * Catches errors and returns them as ActionResult
 */
export async function withErrorHandling<T>(
  fn: () => Promise<ActionResult<T>>
): Promise<ActionResult<T>> {
  try {
    return await fn();
  } catch (error: any) {
    console.error("[Server Action Error]", error);

    return createActionError(
      "An unexpected error occurred",
      process.env.NODE_ENV === "development" ? error.message : undefined
    );
  }
}

/**
 * Check if user has required permission
 * Returns error result if permission is missing
 */
export function requirePermission(
  permissions: ServerPermissions,
  check: (p: ServerPermissions) => boolean,
  errorMessage = "You do not have permission to perform this action"
): ActionResult<void> | null {
  if (!check(permissions)) {
    return createActionError(errorMessage, "Permission denied");
  }
  return null;
}
