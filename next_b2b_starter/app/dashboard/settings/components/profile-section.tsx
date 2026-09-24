"use client";
import { useState, type FormEvent } from "react";
import { UserProfile, MemberHelpers } from "@/lib/models/member.model";
import { useUpdateProfile } from "@/lib/hooks/mutations/use-update-profile";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
export function ProfileSection({ profile }: { profile: UserProfile }) {
 const [name, setName] = useState(profile.name ?? "");
 const mutation = useUpdateProfile();
 async function submit(event: FormEvent) { event.preventDefault(); await mutation.mutateAsync({ name: name.trim() }).catch(() => undefined); }
 return <div className="grid gap-6 lg:grid-cols-2">
  <section className="rounded-xl border p-6"><h3 className="text-lg font-semibold">Your profile</h3><p className="mt-2 text-sm text-gray-600">{profile.email}</p>
   <form onSubmit={submit} className="mt-6 space-y-4"><label className="block text-sm" htmlFor="display-name">Display name</label>
    <Input id="display-name" value={name} maxLength={100} required onChange={(e) => setName(e.target.value)} />
    <Button disabled={mutation.isPending || !name.trim()} type="submit">{mutation.isPending ? "Saving…" : "Save name"}</Button>
    {mutation.error && <p role="alert" className="text-sm text-red-700">{mutation.error.message}</p>}
    {mutation.isSuccess && <p role="status" className="text-sm text-green-700">Profile saved.</p>}
   </form>
  </section>
  <section className="rounded-xl border p-6"><h3 className="text-lg font-semibold">{profile.organizationName}</h3>
   <p className="mt-2 text-sm text-gray-600">{MemberHelpers.getRoleConfig(profile.role).label}</p>
   <p className="mt-6 break-all text-sm text-gray-500">Workspace ID: {profile.organizationId}</p>
  </section>
 </div>;
}
