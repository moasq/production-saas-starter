import { test } from "node:test";
import assert from "node:assert/strict";
import { handleIdentityRequest } from "../lib/auth/public-handler.ts";
const url = "https://app.example.test/api/identity/sign-in/magic-link";
function request(body: unknown) { return new Request(url, { method: "POST", headers: { "content-type": "application/json", origin: "https://app.example.test" }, body: JSON.stringify(body) }); }

test("direct magic-link HTTP requests cannot bypass shared profile storage bounds", async () => {
  let forwarded = 0;
  for (const body of [
    { email: "person@example.test", name: "x".repeat(256) },
    { email: "x".repeat(245) + "@example.test" },
    { email: "person@example.test", name: ["admin"] },
  ]) {
    const response = await handleIdentityRequest(request(body), async () => { forwarded++; return Response.json({ status: true }); });
    assert.equal(response.status, 400);
  }
  assert.equal(forwarded, 0, "invalid input must never reach Better Auth or write a user");
});
test("public login accepts an omitted name and forwards only supported fields", async () => {
  let forwarded: Record<string, unknown> | undefined;
  const response = await handleIdentityRequest(request({ email: "Person@Example.test", callbackURL: "/authenticate", role: "admin", metadata: { organizationId: "other" } }), async (request) => {
    forwarded = await request.json(); return Response.json({ status: true });
  });
  assert.equal(response.status, 200);
  assert.deepEqual(forwarded, { email: "person@example.test", callbackURL: "/authenticate" });
});
test("raw membership mutation and token-bearing session endpoints are not public", async () => {
  for (const path of ["organization/update-member-role", "organization/remove-member", "get-session"]) {
    const response = await handleIdentityRequest(new Request(`https://app.example.test/api/identity/${path}`), async () => { throw new Error("Must not forward"); });
    assert.equal(response.status, 404);
  }
});
