import { test } from "node:test";
import assert from "node:assert/strict";
import { isRole, permissionsForRole, roles } from "../lib/auth/rbac.ts";

test("Better Auth management grants match the business permission boundary", () => {
  for (const role of ["admin", "manager", "member"] as const) {
    const manages = permissionsForRole(role).includes("org:manage");
    assert.equal(roles[role].authorize({ member: ["create", "update", "delete"] }).success, manages);
    assert.equal(roles[role].authorize({ invitation: ["create", "cancel"] }).success, manages);
    assert.equal(roles[role].authorize({ organization: ["update"] }).success, manages);
    assert.equal(permissionsForRole(role).includes("org:view"), true);
  }
});
test("unknown and combined roles never establish application authorization", () => {
  for (const role of ["owner", "admin,member", "ADMIN", "super-admin", ""]) {
    assert.equal(isRole(role), false);
    assert.deepEqual(permissionsForRole(role), []);
  }
  assert.equal(isRole(["admin"]), false);
});
