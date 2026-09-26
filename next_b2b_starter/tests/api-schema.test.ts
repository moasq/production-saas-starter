import { test } from "node:test";
import assert from "node:assert/strict";
import { assertApiResponse } from "../scripts/api-contract.mjs";

test("mounted-response validation rejects stale wrappers and missing required fields", () => {
  const url = "http://localhost/api/auth/members";
  assert.doesNotThrow(() => assertApiResponse(url, "GET", 200, { success: true, data: { members: [], total: 0 } }));
  assert.throws(() => assertApiResponse(url, "GET", 200, { members: [], total: 0 }), /contract mismatch/);
  assert.throws(() => assertApiResponse(url, "GET", 200, { success: true, data: { members: [] } }), /contract mismatch/);
  assert.doesNotThrow(() => assertApiResponse(url, "GET", 403, ""));
  assert.throws(() => assertApiResponse(url, "GET", 200, ""), /contract mismatch/);
});

test("invitation contract distinguishes saved invitation from actual delivery", () => {
  const url = "http://localhost/api/auth/members";
  const body = { email: "fixture@example.test", name: "Fixture", role_slug: "member" };
  const response = { success: true, data: { member_id: "invitation:fixture", invite_sent: false } };
  assert.doesNotThrow(() => assertApiResponse(url, "POST", 200, response, body));
  assert.throws(() => assertApiResponse(url, "POST", 200, response, { ...body, role_slug: "owner" }), /contract mismatch/);
  assert.doesNotThrow(() => assertApiResponse(url, "POST", 403, { error: "Permission denied" }, { ...body, role_slug: "owner" }));
});
