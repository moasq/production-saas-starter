import { NextResponse, type NextRequest } from "next/server";
import { SESSION_COOKIE_NAME, SESSION_JWT_COOKIE_NAME } from "@/lib/auth/constants";
export function proxy(request: NextRequest) {
 if (!request.cookies.get(SESSION_COOKIE_NAME) && !request.cookies.get(SESSION_JWT_COOKIE_NAME)) {
   const login = new URL("/auth", request.url);
   login.searchParams.set("returnTo", request.nextUrl.pathname);
   return NextResponse.redirect(login);
 }
 // Presence is an optimization only; the dashboard and backend verify the actual session.
 return NextResponse.next();
}
export const config = { matcher: ["/dashboard/:path*"] };
