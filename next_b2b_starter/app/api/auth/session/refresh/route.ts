import { NextResponse } from "next/server";
import { cookies } from "next/headers";

import { getStytchB2BClient } from "@/lib/auth/stytch/server";
import {
  SESSION_COOKIE_NAME,
  SESSION_JWT_COOKIE_NAME,
} from "@/lib/auth/constants";
import {
  getSessionDurationMinutes,
  getCookieConfig,
} from "@/lib/auth/server-constants";
import { isTokenExpired } from "@/lib/auth/token-utils";

export async function POST() {
  try {
    const cookieStore = await cookies();

    // First, check if we already have a valid JWT
    const existingJwt = cookieStore.get(SESSION_JWT_COOKIE_NAME)?.value ?? null;
    if (existingJwt && !isTokenExpired(existingJwt)) {
      return NextResponse.json({ sessionJwt: existingJwt });
    }

    // Try to get session token to exchange for JWT
    const sessionToken = cookieStore.get(SESSION_COOKIE_NAME)?.value ?? null;

    if (!sessionToken) {
      return NextResponse.json(
        { sessionJwt: null, error: "session_not_found" },
        { status: 401 }
      );
    }

    const client = getStytchB2BClient();

    try {
      const response = await client.sessions.authenticate({
        session_token: sessionToken,
        session_duration_minutes: getSessionDurationMinutes(),
      });

      const sessionJwt = (response as any)?.session_jwt ?? null;

      if (!sessionJwt) {
        return NextResponse.json(
          { sessionJwt: null, error: "session_missing_jwt" },
          { status: 401 }
        );
      }

      // Validate the new JWT before returning it
      if (isTokenExpired(sessionJwt)) {
        return NextResponse.json(
          { sessionJwt: null, error: "session_jwt_expired" },
          { status: 401 }
        );
      }

      const res = NextResponse.json({ sessionJwt });
      const maxAgeSeconds = getSessionDurationMinutes() * 60;

      res.cookies.set(SESSION_JWT_COOKIE_NAME, sessionJwt, {
        ...getCookieConfig(),
        maxAge: maxAgeSeconds,
      });

      return res;
    } catch {
      // Clear invalid session cookies
      const response = NextResponse.json(
        {
          sessionJwt: null,
          error: "session_invalid"
        },
        { status: 401 }
      );

      // Clear the invalid cookies
      response.cookies.delete(SESSION_COOKIE_NAME);
      response.cookies.delete(SESSION_JWT_COOKIE_NAME);

      return response;
    }
  } catch {
    return NextResponse.json(
      { sessionJwt: null, error: "refresh_failed" },
      { status: 500 }
    );
  }
}
