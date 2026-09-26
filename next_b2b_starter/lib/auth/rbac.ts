import { createAccessControl } from "better-auth/plugins/access";

export const roleNames = ["admin", "manager", "member"] as const;
export type Role = (typeof roleNames)[number];
export function isRole(value: unknown): value is Role { return typeof value === "string" && roleNames.includes(value as Role); }
export type ApplicationPermission = "org:view" | "org:manage";
export type RolePermissions = Readonly<Record<Role, readonly ApplicationPermission[]>>;

// Application code is the policy authority. Both the private Go bridge and Better
// Auth's membership operations derive their grants from this one declaration.
export const rolePermissions = {
  admin: ["org:view", "org:manage"],
  manager: ["org:view"],
  member: ["org:view"],
} as const satisfies RolePermissions;

export const accessControl = createAccessControl({
  organization: ["update", "delete"], member: ["create", "update", "delete"], invitation: ["create", "cancel"],
} as const);

export function createRolePolicy(policy: RolePermissions) {
  // Snapshot the declaration so its consumers cannot change policy independently.
  const grants = { admin: [...policy.admin], manager: [...policy.manager], member: [...policy.member] };
  const permissionsForRole = (role: unknown): ApplicationPermission[] => isRole(role) ? [...grants[role]] : [];
  const authRole = (role: Role) => accessControl.newRole(permissionsForRole(role).includes("org:manage")
    ? { organization: ["update"], member: ["create", "update", "delete"], invitation: ["create", "cancel"] }
    : { organization: [], member: [], invitation: [] });
  return { permissionsForRole, roles: { admin: authRole("admin"), manager: authRole("manager"), member: authRole("member") } };
}

export const { permissionsForRole, roles } = createRolePolicy(rolePermissions);
