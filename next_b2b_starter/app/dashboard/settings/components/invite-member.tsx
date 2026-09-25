"use client";
import { useState, type FormEvent } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { InviteMemberRequest, InviteMemberResponse, MemberRole } from "@/lib/models/member.model";
export function InviteMember({ canInvite, onInvite }: { canInvite: boolean; onInvite: (request: InviteMemberRequest) => Promise<InviteMemberResponse | undefined> }) {
 const [email, setEmail] = useState(""); const [name, setName] = useState(""); const [role, setRole] = useState<MemberRole>("member");
 const [pending, setPending] = useState(false); const [error, setError] = useState<string | null>(null);
 async function submit(event: FormEvent) {
  event.preventDefault(); setPending(true); setError(null);
  try {
   const result = await onInvite({ email: email.trim(), name: name.trim(), role });
   if (!result) throw new Error("Organization context unavailable. Reload and try again.");
   if (result.inviteSent) toast.success("Invitation sent", { description: email });
   else toast.warning("Member added; email needs attention", { description: result.message });
  } catch (error) { setError(error instanceof Error ? error.message : "Unable to add member."); }
  finally { setPending(false); }
 }
 if (!canInvite) return null;
 return <form onSubmit={submit} className="space-y-4">
  <label className="block text-sm" htmlFor="member-name">Full name</label><Input id="member-name" required maxLength={100} value={name} onChange={(e) => setName(e.target.value)} disabled={pending} />
  <label className="block text-sm" htmlFor="member-email">Work email</label><Input id="member-email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} disabled={pending} />
  <label className="block text-sm" htmlFor="member-role">Role</label>
  <select id="member-role" className="h-10 w-full rounded-md border border-gray-300 bg-white px-3 text-sm" value={role} onChange={(e) => setRole(e.target.value as MemberRole)} disabled={pending}>
   <option value="member">Member</option><option value="manager">Manager</option><option value="admin">Admin — manages team and billing</option>
  </select>
  {error && <p role="alert" className="text-sm text-red-700">{error}</p>}
  <Button disabled={pending || !name.trim()} type="submit">{pending ? "Adding member…" : "Send invitation"}</Button>
 </form>;
}
