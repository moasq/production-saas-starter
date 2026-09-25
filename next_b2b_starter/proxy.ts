import { NextResponse, type NextRequest } from "next/server";
import { getSessionCookie } from "better-auth/cookies";
export function proxy(request: NextRequest) {
  if (!getSessionCookie(request)) {
    const login = new URL("/auth", request.url); login.searchParams.set("returnTo", request.nextUrl.pathname);
    return NextResponse.redirect(login);
  }
  // Presence only avoids needless rendering; every protected page/API verifies the live session.
  return NextResponse.next();
}
export const config = { matcher: ["/dashboard/:path*"] };
