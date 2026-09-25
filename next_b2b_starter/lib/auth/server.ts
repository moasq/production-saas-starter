import "server-only";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { verifyIdentity } from "./identity";
import { buildLoginUrl } from "./urls";
import { isAuthConfigured } from "./config";
export async function getMemberSession() {
  const requestHeaders = await headers();
  if (!isAuthConfigured()) return null;
  const identity = await verifyIdentity(new Headers(requestHeaders));
  return identity ? { ...identity, cookie_header: requestHeaders.get("cookie") || "" } : null;
}
export type VerifiedSession = NonNullable<Awaited<ReturnType<typeof getMemberSession>>>;
export async function requireMemberSession(options?: { returnTo?: string }): Promise<VerifiedSession> {
  const session = await getMemberSession();
  if (!session) redirect(buildLoginUrl({ returnTo: options?.returnTo }));
  if (!session.membership) redirect("/workspaces");
  return session;
}
