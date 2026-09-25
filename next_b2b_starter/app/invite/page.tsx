import { getMemberSession } from "@/lib/auth/server";
import InviteContent from "./content";
export default async function Page({ searchParams }: { searchParams: Promise<{ id?: string }> }) {
  const { id } = await searchParams;
  const session = await getMemberSession();
  return <InviteContent invitationId={id || ""} signedInEmail={session?.user.email || null} />;
}
