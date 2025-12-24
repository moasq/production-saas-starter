/**
 * Centralized Authentication Constants
 * Now supports both server and client contexts with proper separation
 */

// Static constants (safe for both contexts)
export const SESSION_COOKIE_NAME = "stytch_session";
export const SESSION_JWT_COOKIE_NAME = "stytch_session_jwt";
export const TOKEN_EXPIRY_GRACE_SECONDS = 60;

// Auth Routes (static, safe everywhere)
// Note: LOGOUT and MAGIC_LINK routes migrated to Server Actions (see lib/actions/auth/)
export const AUTH_ROUTES = {
  LOGIN: "/auth",
  CONSUME_MAGIC_LINK: "/api/auth/consume-magic-link",  // External Stytch callback (must remain)
  SESSION_REFRESH: "/api/auth/session/refresh",        // Token refresh endpoint (must remain)
  AUTHENTICATE_REDIRECT: "/authenticate",
  DASHBOARD: "/dashboard",
} as const;



