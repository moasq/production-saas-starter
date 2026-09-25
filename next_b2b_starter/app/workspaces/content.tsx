"use client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { logout } from "@/lib/actions/auth/logout";
export default function WorkspaceContent({ organizations, email }: { organizations: { id: string; name: string }[]; email: string }) {
  const [name, setName] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  async function submit(path: string, body: object) {
    setPending(true); setError("");
    try {
      const response = await fetch(path, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });
      if (!response.ok) throw new Error("Could not open this workspace. Check your access and try again.");
      // Full navigation discards data cached for the previously selected organization.
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination -- New tenant identity must discard all browser application state.
      window.location.assign("/dashboard");
    } catch (error) { setError(error instanceof Error ? error.message : "Please retry"); setPending(false); }
  }
  return <main className="min-h-screen bg-gray-50 px-6 py-20"><div className="mx-auto max-w-md rounded-2xl border bg-white p-8 shadow-sm">
    <h1 className="text-2xl font-semibold">Your workspaces</h1><p className="mt-2 text-sm text-gray-600">Signed in as {email}</p>
    {error && <p role="alert" className="mt-4 text-sm text-red-600">{error}</p>}
    <div className="mt-6 space-y-3">{organizations.map((org) => <Button key={org.id} variant="outline" className="w-full justify-start" disabled={pending} onClick={() => submit("/api/workspaces/select", { organizationId: org.id })}>{org.name}</Button>)}</div>
    <form className="mt-8 space-y-3 border-t pt-6" onSubmit={(event) => { event.preventDefault(); void submit("/api/workspaces", { name }); }}>
      <label htmlFor="workspace-name" className="text-sm font-medium">Create a workspace</label>
      <Input id="workspace-name" placeholder="Acme Inc" value={name} onChange={(event) => setName(event.target.value)} minLength={2} maxLength={100} required disabled={pending} />
      <Button type="submit" disabled={pending || name.trim().length < 2} className="w-full">Create workspace</Button>
    </form><Button variant="ghost" className="mt-4 w-full" disabled={pending} onClick={() => logout("/")}>Log out</Button>
  </div></main>;
}
