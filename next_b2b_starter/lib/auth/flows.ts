import { APIError } from "better-auth/api";
import { randomUUID } from "node:crypto";
import { getAuth } from "./configuration.ts";
import { getAuthDatabase } from "./database.ts";
import { limitPublicEmail } from "./rate-limit.ts";
import { verifyIdentity } from "./identity.ts";
import { sanitizeReturnTo } from "./urls.ts";
export const emailMessage = "If this email address can receive mail, a sign-in link has been sent. Check your inbox.";
export function requireOrigin(headers: Headers): void {
  if (headers.get("origin") !== new URL(process.env.APP_BASE_URL || "http://localhost:3000").origin) throw new APIError("FORBIDDEN", { message: "Invalid origin" });
}
export function validateEmail(value: unknown): string {
  if (typeof value !== "string" || value.length > 254 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim())) throw new Error("A valid email address is required");
  return value.trim().toLowerCase();
}
function nameValue(value: unknown): string {
  if (typeof value !== "string" || value.trim().length < 2 || value.length > 100) throw new Error("Names must contain 2–100 characters");
  return value.trim();
}
export async function requestLogin(headers: Headers, rawEmail: unknown, returnTo?: string) {
  const email = validateEmail(rawEmail);
  await limitPublicEmail(headers, email);
  const path = sanitizeReturnTo(returnTo) || "/authenticate";
  await getAuth().api.signInMagicLink({ headers, body: { email, callbackURL: path, newUserCallbackURL: path, errorCallbackURL: "/authenticate?error=invalid_link" } });
  return { message: emailMessage };
}
export async function requestSignup(headers: Headers, input: { email: unknown; name: unknown; organizationName: unknown }) {
  const email = validateEmail(input.email), name = nameValue(input.name), organizationName = nameValue(input.organizationName);
  await limitPublicEmail(headers, email);
  const id = randomUUID();
  await getAuthDatabase().query("DELETE FROM signup_intent WHERE id IN (SELECT id FROM signup_intent WHERE expires_at<now() LIMIT 100)");
  await getAuthDatabase().query(`INSERT INTO signup_intent (id, email, name, organization_name, expires_at) VALUES ($1,$2,$3,$4,now()+interval '30 minutes')`, [id,email,name,organizationName]);
  const callbackURL = `/authenticate?signup=${id}`;
  await getAuth().api.signInMagicLink({ headers, body: { email, name, callbackURL, newUserCallbackURL: callbackURL, errorCallbackURL: "/authenticate?error=invalid_link" } });
  return { message: emailMessage };
}
export async function createWorkspace(headers: Headers, rawName: unknown, slug?: string) {
  requireOrigin(headers);
  const identity = await verifyIdentity(headers);
  if (!identity) throw new APIError("UNAUTHORIZED", { message: "Sign in with a verified email first" });
  const name = nameValue(rawName);
  const organization = await getAuth().api.createOrganization({ headers, body: { name, slug: slug || `workspace-${randomUUID()}` } });
  if (!organization) throw new Error("Could not create workspace");
  await getAuth().api.setActiveOrganization({ headers, body: { organizationId: organization.id } });
  return organization;
}
export async function selectWorkspace(headers: Headers, organizationId: unknown) {
  requireOrigin(headers);
  if (typeof organizationId !== "string" || !organizationId || organizationId.length > 200) throw new Error("Organization required");
  const identity = await verifyIdentity(headers);
  if (!identity) throw new APIError("UNAUTHORIZED", { message: "Sign in first" });
  const membership = await getAuthDatabase().query(`SELECT id FROM member WHERE "userId"=$1 AND "organizationId"=$2 AND role IN ('admin','manager','member')`, [identity.user.id, organizationId]);
  // Better Auth clears the active organization on a denied selector. Reject first to keep
  // a forged target from disrupting the user's current valid workspace, then let BA recheck.
  if (!membership.rowCount) throw new APIError("FORBIDDEN", { message: "Organization membership required" });
  return getAuth().api.setActiveOrganization({ headers, body: { organizationId } });
}
export async function acceptInvitation(headers: Headers, invitationId: unknown) {
  requireOrigin(headers);
  if (typeof invitationId !== "string" || !invitationId || invitationId.length > 200) throw new Error("Invitation required");
  if (!await verifyIdentity(headers)) throw new APIError("UNAUTHORIZED", { message: "Sign in with the invited email first" });
  const result = await getAuth().api.acceptInvitation({ headers, body: { invitationId } });
  await getAuth().api.setActiveOrganization({ headers, body: { organizationId: result.invitation.organizationId } });
  return result;
}
export async function completeLogin(headers: Headers, signupId?: string): Promise<string> {
  requireOrigin(headers);
  const identity = await verifyIdentity(headers);
  if (!identity) throw new Error("This link has expired or was already used. Request a new sign-in link.");
  if (signupId) {
    const db = getAuthDatabase();
    const claim = randomUUID();
    const found = await db.query(`UPDATE signup_intent SET claim_token=$3, claimed_at=now()
      WHERE id=$1 AND email=$2 AND expires_at>now()
      AND (organization_id IS NOT NULL OR claim_token IS NULL OR claimed_at<now()-interval '2 minutes') RETURNING *`, [signupId, identity.user.email.toLowerCase(), claim]);
    const intent = found.rows[0];
    if (!intent) throw new Error("Signup expired or is already completing. Sign in to choose your workspace.");
    try {
      let organizationId = intent.organization_id;
      if (!organizationId) {
        const slug = `signup-${signupId}`;
        const existing = await db.query(`SELECT o.id FROM organization o JOIN member m ON m."organizationId"=o.id WHERE o.slug=$1 AND m."userId"=$2`, [slug, identity.user.id]);
        organizationId = existing.rows[0]?.id || (await createWorkspace(headers, intent.organization_name, slug)).id;
        await db.query("UPDATE signup_intent SET organization_id=$2 WHERE id=$1 AND claim_token=$3", [signupId, organizationId, claim]);
      }
      await selectWorkspace(headers, organizationId);
      return "/dashboard";
    } finally {
      await db.query("UPDATE signup_intent SET claim_token=NULL, claimed_at=NULL WHERE id=$1 AND claim_token=$2", [signupId, claim]);
    }
  }

  if (identity.membership) return "/dashboard";
  const organizations = await getAuth().api.listOrganizations({ headers });
  if (organizations.length === 1) {
    await selectWorkspace(headers, organizations[0].id);
    return "/dashboard";
  }
  return "/workspaces";
}
