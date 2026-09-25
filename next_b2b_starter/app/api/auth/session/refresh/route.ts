import { NextResponse, type NextRequest } from "next/server";
import { cookies } from "next/headers";
import { getStytchB2BClient } from "@/lib/auth/stytch/server";
import { getBaseUrl } from "@/lib/auth/stytch";
import { SESSION_COOKIE_NAME, SESSION_JWT_COOKIE_NAME } from "@/lib/auth/constants";
import { getSessionDurationMinutes, getCookieConfig } from "@/lib/auth/server-constants";
export async function POST(request: NextRequest) {
 if (request.headers.get("origin") !== new URL(getBaseUrl()).origin) return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
 const store = await cookies();
 const sessionToken = store.get(SESSION_COOKIE_NAME)?.value;
 if (!sessionToken) return NextResponse.json({ error: "Authentication required" }, { status: 401 });
 try {
  const session = await getStytchB2BClient().sessions.authenticate({ session_token: sessionToken, session_duration_minutes: getSessionDurationMinutes() });
  const response = NextResponse.json({ success: true });
  const options = { ...getCookieConfig(), maxAge: getSessionDurationMinutes() * 60 };
  response.cookies.set(SESSION_COOKIE_NAME, session.session_token, options);
  response.cookies.set(SESSION_JWT_COOKIE_NAME, session.session_jwt, options);
  return response;
 } catch {
  const response = NextResponse.json({ error: "Session expired. Sign in again." }, { status: 401 });
  response.cookies.delete(SESSION_COOKIE_NAME); response.cookies.delete(SESSION_JWT_COOKIE_NAME);
  return response;
 }
}
