import { test } from "node:test";
import assert from "node:assert/strict";
import { ApiClient } from "../lib/api/api/client/api-client.ts";

test("concurrent server requests keep their organization credentials isolated", async () => {
 const original = globalThis.fetch;
 const observed: (string | null)[] = [];
 globalThis.fetch = async (_input, init) => {
  observed.push(new Headers(init?.headers).get("Authorization"));
  await new Promise((resolve) => setTimeout(resolve, 5));
  return Response.json({ success: true });
 };
 try {
  const client = new ApiClient({ baseUrl: "https://example.test/api" });
  await Promise.all([
   client.get("/profile", { headers: { Authorization: "Bearer organization-a" } }),
   client.get("/profile", { headers: { Authorization: "Bearer organization-b" } }),
  ]);
  await client.get("/public", { skipAuth: true });
  assert.deepEqual(observed, ["Bearer organization-a", "Bearer organization-b", null]);
 } finally { globalThis.fetch = original; }
});

test("server mutations are not retried after a failed response", async () => {
 const original = globalThis.fetch;
 let calls = 0;
 globalThis.fetch = async () => { calls++; return Response.json({ message: "Unauthorized" }, { status: 401 }); };
 try {
  const client = new ApiClient({ baseUrl: "https://example.test/api" });
  await assert.rejects(client.post("/members", { name: "Test" }, { headers: { Authorization: "Bearer expired" } }), /Unauthorized/);
  assert.equal(calls, 1);
 } finally { globalThis.fetch = original; }
});

test("successful no-content member deletion does not require a JSON body", async () => {
 const original = globalThis.fetch;
 globalThis.fetch = async () => new Response(null, { status: 204 });
 try {
  const client = new ApiClient({ baseUrl: "https://example.test/api" });
  assert.equal(await client.delete("/auth/members/member-a"), null);
 } finally { globalThis.fetch = original; }
});
