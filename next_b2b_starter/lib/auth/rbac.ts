import { createAccessControl } from "better-auth/plugins/access";

export const roleNames = ["admin", "manager", "member"] as const;
export type Role = (typeof roleNames)[number];
export function isRole(value: unknown): value is Role { return typeof value === "string" && roleNames.includes(value as Role); }
export function permissionsForRole(role: string): string[] {
  if (role === "admin") return ["org:view", "org:manage"];
  if (role === "manager" || role === "member") return ["org:view"];
  return [];
}
export const accessControl = createAccessControl({
  organization: ["update", "delete"], member: ["create", "update", "delete"], invitation: ["create", "cancel"],
} as const);
export const roles = {
  admin: accessControl.newRole({ organization: ["update"], member: ["create", "update", "delete"], invitation: ["create", "cancel"] }),
  manager: accessControl.newRole({ organization: [], member: [], invitation: [] }),
  member: accessControl.newRole({ organization: [], member: [], invitation: [] }),
};
