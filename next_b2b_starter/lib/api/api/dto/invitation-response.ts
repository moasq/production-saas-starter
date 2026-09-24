export interface InvitationResult { success: boolean; memberId?: string; inviteSent: boolean; message?: string; }
export function parseInvitationResponse(response: { success: boolean; data?: { member_id?: string; invite_sent?: boolean }; message?: string }): InvitationResult {
 if (!response.success || !response.data?.member_id) throw new Error(response.message || "The member could not be added.");
 return { success: true, memberId: response.data.member_id, inviteSent: response.data.invite_sent === true,
   message: response.data.invite_sent === true ? "Invitation sent." : "Member added, but the invitation email could not be sent. Use Resend invitation in the team list." };
}
