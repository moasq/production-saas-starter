import { test } from "node:test";
import assert from "node:assert/strict";
import { createRolePolicy, isRole, permissionsForRole, rolePermissions, roles } from "../lib/auth/rbac.ts";

test("Better Auth management grants match the business permission boundary", () => {
  for (const role of ["admin", "manager", "member"] as const) {
    const manages = permissionsForRole(role).includes("org:manage");
    assert.equal(roles[role].authorize({ member: ["create", "update", "delete"] }).success, manages);
    assert.equal(roles[role].authorize({ invitation: ["create", "cancel"] }).success, manages);
    assert.equal(roles[role].authorize({ organization: ["update"] }).success, manages);
    assert.equal(roles[role].authorize({ organization: ["delete"] }).success, false);
    assert.equal(permissionsForRole(role).includes("org:view"), true);
  }
});
test("unknown and combined roles never establish application authorization", () => {
  for (const role of ["owner", "admin,member", "ADMIN", "super-admin", "", null, undefined, ["admin"], {}]) {
    assert.equal(isRole(role), false);
    assert.deepEqual(permissionsForRole(role), []);
  }
  assert.equal(isRole(["admin"]), false);
});

test("a restrictive application policy removes bridge and Better Auth grants together", () => {
  const restricted = createRolePolicy({ ...rolePermissions, admin: ["org:view"] });
  assert.deepEqual(restricted.permissionsForRole("admin"), ["org:view"]);
  for (const permission of [
    { member: ["create"] }, { member: ["update"] }, { member: ["delete"] },
    { invitation: ["create"] }, { invitation: ["cancel"] }, { organization: ["update"] },
  ] as const) {
    assert.equal(restricted.roles.admin.authorize(permission).success, false);
    assert.equal(roles.admin.authorize(permission).success, true);
  }
});

test("a denied role and returned grant arrays cannot gain permissions implicitly", () => {
  const restricted = createRolePolicy({ ...rolePermissions, admin: [] });
  assert.deepEqual(restricted.permissionsForRole("admin"), []);
  assert.equal(restricted.roles.admin.authorize({ member: ["create"] }).success, false);
  const grants = permissionsForRole("member");
  grants.push("org:manage");
  assert.deepEqual(permissionsForRole("member"), ["org:view"]);
  assert.equal(roles.member.authorize({ member: ["create"] }).success, false);
});
