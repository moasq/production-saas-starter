import { test } from "node:test";
import assert from "node:assert/strict";
import { ApiClient } from "../lib/api/api/client/api-client.ts";
import { MemberRepository } from "../lib/api/api/repositories/member-repository.ts";

test("member removal accepts the backend's HTTP 204 contract", async () => {
 const original = globalThis.fetch;
 globalThis.fetch = async () => new Response(null, { status: 204 });
 try {
   const repo = new MemberRepository(new ApiClient({ baseUrl: "https://example.test/api" }));
   assert.equal(await repo.removeMember("member-a"), true);
 } finally { globalThis.fetch = original; }
});

test("resending invitations uses the authenticated member route and verifies delivery", async () => {
 const original = globalThis.fetch;
 let route = "";
 globalThis.fetch = async (url) => { route = (url as Request).url; return Response.json({ success: true, data: { invite_sent: true } }); };
 try {
   const repo = new MemberRepository(new ApiClient({ baseUrl: "https://example.test/api" }));
   assert.equal(await repo.resendInvitation("member-a"), true);
   assert.equal(route, "https://example.test/api/auth/members/member-a/resend-invitation");
   globalThis.fetch = async () => Response.json({ success: true, data: { invite_sent: false } });
   await assert.rejects(repo.resendInvitation("member-a"), /could not be sent/);
 } finally { globalThis.fetch = original; }
});


test("invitation cooldown is shown to the caller without retrying the mutation", async () => {
 const original = globalThis.fetch;
 let calls = 0;
 globalThis.fetch = async () => { calls++; return Response.json({ success: false, error: "Too many invitation requests. Please wait before trying again." }, { status: 429 }); };
 try {
   const repo = new MemberRepository(new ApiClient({ baseUrl: "https://example.test/api" }));
   await assert.rejects(repo.resendInvitation("invitation:pending"), /Please wait before trying again/);
   assert.equal(calls, 1);
 } finally { globalThis.fetch = original; }
});
