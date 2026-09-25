"use client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { sendMagicLink } from "@/lib/actions/auth/send-magic-link";
import { logout } from "@/lib/actions/auth/logout";
export default function InviteContent({ invitationId, signedInEmail }: { invitationId: string; signedInEmail: string | null }) {
  const [email, setEmail] = useState(""); const [message, setMessage] = useState(""); const [pending, setPending] = useState(false);
  const returnTo = `/invite?id=${encodeURIComponent(invitationId)}`;
  async function accept() {
    setPending(true);
    try {
      const response = await fetch("/api/workspaces/accept-invitation", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ invitationId }) });
      if (!response.ok) throw new Error("This invitation has expired, was already accepted, or belongs to another email address. Ask your administrator to resend it.");
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination -- New tenant identity must discard all browser application state.
      window.location.assign("/dashboard");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Please retry"); setPending(false); }
  }
  return <main className="min-h-screen bg-gray-50 px-6 py-20"><div className="mx-auto max-w-md rounded-2xl border bg-white p-8 shadow-sm">
    <h1 className="text-2xl font-semibold">Join your workspace</h1>
    {message && <p role="status" className="mt-4 text-sm text-gray-600">{message}</p>}
    {!invitationId ? <p className="mt-4">The invitation link is missing. Ask your administrator for a new invitation.</p> : signedInEmail ? <>
      <p className="my-5 text-sm text-gray-600">Accept this invitation as {signedInEmail}.</p>
      <Button className="w-full" disabled={pending} onClick={accept}>Accept invitation</Button>
      <Button variant="ghost" className="mt-3 w-full" disabled={pending} onClick={() => logout(returnTo)}>Use another email</Button>
    </> : <form className="mt-6 space-y-4" onSubmit={async (event) => {
      event.preventDefault(); setPending(true);
      const result = await sendMagicLink(email, returnTo);
      setMessage(result.success ? "Check your inbox to verify the invited email, then accept the invitation." : result.error || "Please retry"); setPending(false);
    }}>
      <p className="text-sm text-gray-600">Verify the email address that received the invitation.</p>
      <Input type="email" aria-label="Invited email address" placeholder="you@company.com" required value={email} onChange={(event) => setEmail(event.target.value)} disabled={pending} />
      <Button className="w-full" type="submit" disabled={pending}>Send sign-in link</Button>
    </form>}
  </div></main>;
}
