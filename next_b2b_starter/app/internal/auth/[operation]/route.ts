import { timingSafeEqual } from "node:crypto";
import { getAuth } from "@/lib/auth/configuration";
import { getAuthDatabase } from "@/lib/auth/database";
import { verifyIdentity, bridgeIdentity } from "@/lib/auth/identity";
import { isRole } from "@/lib/auth/rbac";

export async function POST(request: Request, context: { params: Promise<{ operation: string }> }) {
  const expected = process.env.AUTH_INTERNAL_SECRET || "";
  const supplied = request.headers.get("x-internal-auth-secret") || "";
  const expectedBytes = Buffer.from(expected), suppliedBytes = Buffer.from(supplied);
  if (expectedBytes.length < 32 || suppliedBytes.length !== expectedBytes.length || !timingSafeEqual(expectedBytes, suppliedBytes)) {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }
  try {
    const headers = new Headers({ cookie: request.headers.get("cookie") || "" });
    const identity = await verifyIdentity(headers);
    if (!identity?.membership) return Response.json({ error: "Authenticated membership required" }, { status: 401 });
    const { operation } = await context.params;
    if (operation === "session") return Response.json(bridgeIdentity(identity));
    const body = await request.json().catch(() => ({}));
    const auth = getAuth();
    const organizationId = identity.membership.id;
    if (operation === "profile") {
      if (typeof body.name !== "string" || body.name.trim().length < 2 || body.name.length > 100) return Response.json({ error: "Name must contain 2–100 characters" }, { status: 400 });
      await auth.api.updateUser({ headers, body: { name: body.name.trim() } });
      return Response.json({ success: true });
    }
    if (!identity.permissions.includes("org:manage")) return Response.json({ error: "Organization administrator access required" }, { status: 403 });
    if (operation === "members-list") {
      const active = await getAuthDatabase().query(`SELECT m.id AS member_id, u.email, u.name, ARRAY[m.role] AS roles,
        'active' AS status, u."emailVerified" AS email_verified, m."createdAt" AS created_at, u."updatedAt" AS updated_at
        FROM member m JOIN "user" u ON u.id = m."userId" WHERE m."organizationId" = $1 ORDER BY m."createdAt", m.id`, [organizationId]);
      const pending = await getAuthDatabase().query(`SELECT 'invitation:' || id AS member_id, email, '' AS name, ARRAY[role] AS roles,
        'pending' AS status, false AS email_verified, "createdAt" AS created_at, "createdAt" AS updated_at
        FROM invitation WHERE "organizationId" = $1 AND status = 'pending' ORDER BY "createdAt", id`, [organizationId]);
      const members = [...active.rows, ...pending.rows];
      return Response.json({ success: true, data: { members, total: members.length } });
    }
    if (operation === "members-invite") {
      if (!isRole(body.role) || typeof body.email !== "string" || body.email.length > 254) return Response.json({ error: "Valid email and role required" }, { status: 400 });
      try {
        const result = await auth.api.createInvitation({ headers, body: { email: body.email.trim().toLowerCase(), role: body.role, organizationId, resend: true } });
        return Response.json({ success: true, data: { member_id: `invitation:${result.id}`, invite_sent: true } });
      } catch (error) {
        // Keep authorization/validation failures visible. A saved pending invitation only
        // represents partial success when delivery or another server dependency failed.
        const status = error && typeof error === "object" && "statusCode" in error ? Number(error.statusCode) : 500;
        if (status >= 400 && status < 500) throw error;
        const pending = await getAuthDatabase().query("SELECT id FROM invitation WHERE \"organizationId\" = $1 AND email = $2 AND status = 'pending' ORDER BY \"createdAt\" DESC LIMIT 1", [organizationId, String(body.email).trim().toLowerCase()]);
        if (pending.rows[0]) return Response.json({ success: true, data: { member_id: `invitation:${pending.rows[0].id}`, invite_sent: false } });
        throw error;
      }
    }
    if (operation === "organization-update") {
      if (typeof body.name !== "string" || body.name.trim().length < 2 || body.name.length > 100) return Response.json({ error: "Valid organization name required" }, { status: 400 });
      const organization = await auth.api.updateOrganization({ headers, body: { organizationId, data: { name: body.name.trim() } } });
      return Response.json({ success: true, data: organization });
    }
    const memberId = typeof body.member_id === "string" ? body.member_id : "";
    if (!memberId || memberId.length > 200) return Response.json({ error: "Member required" }, { status: 400 });
    if (memberId.startsWith("invitation:")) {
      const invitationId = memberId.slice("invitation:".length);
      const pending = await getAuthDatabase().query("SELECT id, email, role FROM invitation WHERE id = $1 AND \"organizationId\" = $2 AND status = 'pending'", [invitationId, organizationId]);
      const invitation = pending.rows[0];
      if (!invitation) return Response.json({ error: "Invitation not found" }, { status: 404 });
      if (operation === "members-remove") {
        await auth.api.cancelInvitation({ headers, body: { invitationId } });
        return Response.json({ success: true });
      }
      if (operation === "members-resend" && isRole(invitation.role)) {
        await auth.api.createInvitation({ headers, body: { email: invitation.email, role: invitation.role, organizationId, resend: true } });
        return Response.json({ success: true, data: { invite_sent: true } });
      }
    } else {
      if (operation === "members-remove") {
        await auth.api.removeMember({ headers, body: { memberIdOrEmail: memberId, organizationId } });
        return Response.json({ success: true });
      }
      if (operation === "members-role" && isRole(body.role)) {
        await auth.api.updateMemberRole({ headers, body: { memberId, role: body.role, organizationId } });
        return Response.json({ success: true });
      }
    }
    return Response.json({ error: "Operation not supported" }, { status: 400 });
  } catch (error) {
    const code = typeof error === "object" && error !== null && "statusCode" in error ? Number(error.statusCode) : 500;
    const constraintError = typeof error === "object" && error !== null && "code" in error && error.code === "23514";
    const status = constraintError ? 409 : code >= 400 && code < 500 ? code : 503;
    return Response.json({ success: false, error: status === 503 ? "Authentication service temporarily unavailable" : "Operation not permitted" }, { status });
  }
}
