import { getAuth } from "./configuration.ts";
import { getAuthDatabase } from "./database.ts";
import { isRole, permissionsForRole } from "./rbac.ts";
export async function verifyIdentity(headers: Headers) {
  const session = await getAuth().api.getSession({ headers, query: { disableCookieCache: true, disableRefresh: true } });
  if (!session || !session.user.emailVerified || new Date(session.session.expiresAt).getTime() <= Date.now()) return null;
  const orgId = session.session.activeOrganizationId;
  const result = orgId ? await getAuthDatabase().query<{ member_id: string; role: string; id: string; name: string; slug: string }>(
    `SELECT m.id AS member_id, m.role, o.id, o.name, o.slug FROM member m
     JOIN organization o ON o.id = m."organizationId" WHERE m."userId" = $1 AND o.id = $2`, [session.user.id, orgId]) : null;
  const membership = result?.rows[0];
  const validMembership = membership && isRole(membership.role) ? membership : null;
  return { session: session.session, user: session.user, membership: validMembership,
    permissions: validMembership ? permissionsForRole(validMembership.role) : [] };
}
export function bridgeIdentity(identity: NonNullable<Awaited<ReturnType<typeof verifyIdentity>>>) {
  const m = identity.membership;
  if (!m) throw new Error("Organization membership required");
  return { user_id: identity.user.id, email: identity.user.email, email_verified: identity.user.emailVerified,
    organization_id: m.id, member_id: m.member_id, roles: [m.role], permissions: identity.permissions,
    expires_at: new Date(identity.session.expiresAt).toISOString(), organization: { id: m.id, name: m.name, slug: m.slug },
    user: { name: identity.user.name } };
}
