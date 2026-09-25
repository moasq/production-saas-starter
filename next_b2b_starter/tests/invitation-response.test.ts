import { test } from "node:test";
import assert from "node:assert/strict";
import { parseInvitationResponse } from "../lib/api/api/dto/invitation-response.ts";
test("member creation does not falsely claim an email was sent", () => {
 const result = parseInvitationResponse({ success: true, data: { member_id: "member-a", invite_sent: false } });
 assert.equal(result.memberId, "member-a"); assert.equal(result.inviteSent, false); assert.match(result.message!, /could not be sent/);
});
test("successful invitation unwraps the backend response", () => {
 assert.equal(parseInvitationResponse({ success: true, data: { member_id: "member-a", invite_sent: true } }).inviteSent, true);
 assert.throws(() => parseInvitationResponse({ success: false, message: "Forbidden" }), /Forbidden/);
});
