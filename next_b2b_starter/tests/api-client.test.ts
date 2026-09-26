import { test } from "node:test";
import assert from "node:assert/strict";
import { ApiClient, unwrap } from "../lib/api/api/client/api-client.ts";

test("concurrent server requests keep their organization credentials isolated", async () => {
 const original = globalThis.fetch;
 const observed: (string | null)[] = [];
 globalThis.fetch = async (input) => {
  observed.push((input as Request).headers.get("Cookie"));
  await new Promise((resolve) => setTimeout(resolve, 5));
  return Response.json({ success: true });
 };
 try {
  const client = new ApiClient({ baseUrl: "https://example.test/api" });
  await Promise.all([
   client.GET("/auth/profile/me", { headers: { Cookie: "better-auth.session_token=organization-a" } }),
   client.GET("/auth/profile/me", { headers: { Cookie: "better-auth.session_token=organization-b" } }),
  ]);
  await client.GET("/auth/profile/me");
  assert.deepEqual(observed, ["better-auth.session_token=organization-a", "better-auth.session_token=organization-b", null]);
 } finally { globalThis.fetch = original; }
});

test("server mutations are not retried after a failed response", async () => {
 const original = globalThis.fetch;
 let calls = 0;
 globalThis.fetch = async () => { calls++; return Response.json({ message: "Unauthorized" }, { status: 401 }); };
 try {
  const client = new ApiClient({ baseUrl: "https://example.test/api" });
  await assert.rejects(client.POST("/auth/members", { body: { email: "test@example.test", name: "Test" }, headers: { Cookie: "better-auth.session_token=expired" } }).then(unwrap), /Unauthorized/);
  assert.equal(calls, 1);
 } finally { globalThis.fetch = original; }
});

test("successful no-content member deletion does not require a JSON body", async () => {
 const original = globalThis.fetch;
 globalThis.fetch = async () => new Response(null, { status: 204 });
 try {
  const client = new ApiClient({ baseUrl: "https://example.test/api" });
  assert.equal(unwrap(await client.DELETE("/auth/members/{member_id}", { params: { path: { member_id: "member-a" } } })), undefined);
 } finally { globalThis.fetch = original; }
});


test("server cookie mutations carry the public origin required by Go CSRF checks", async () => {
 const originalFetch = globalThis.fetch;
 const originalOrigin = process.env.APP_BASE_URL;
 let received: Headers | undefined;
 process.env.APP_BASE_URL = "https://workspace.example.test";
 globalThis.fetch = async (input) => { received = (input as Request).headers; return Response.json({ success: true }); };
 try {
   const client = new ApiClient({ baseUrl: "http://backend:8080/api" });
   await client.PUT("/auth/profile/me", { body: { name: "New name" }, headers: { Cookie: "better-auth.session_token=session-a" } });
   assert.equal(received?.get("Origin"), "https://workspace.example.test");
   assert.equal(received?.get("Cookie"), "better-auth.session_token=session-a");
 } finally {
   globalThis.fetch = originalFetch;
   if (originalOrigin === undefined) delete process.env.APP_BASE_URL;
   else process.env.APP_BASE_URL = originalOrigin;
 }
});

test("empty CORS denial and non-JSON proxy failures produce a safe request error", async () => {
 const originalFetch = globalThis.fetch;
 try {
   for (const status of [403, 502]) {
     globalThis.fetch = async () => new Response(status === 403 ? null : "private upstream details", { status });
     await assert.rejects(new ApiClient({ baseUrl: "https://example.test/api" }).GET("/auth/profile/me").then(unwrap), new RegExp(`Request failed \\(${status}\\)`));
   }
 } finally { globalThis.fetch = originalFetch; }
});
