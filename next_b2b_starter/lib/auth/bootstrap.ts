import { getMemberSession } from "@/lib/auth/server";
import { getServerPermissions } from "@/lib/auth/server-permissions";

export async function authBootstrap() {
  const session = await getMemberSession();
  const permissions = await getServerPermissions(session);

  return {
    profile: permissions.profile,
    roles: permissions.roles,
    permissions: permissions.permissions,
  };
}
