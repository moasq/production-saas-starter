import { redirect } from "next/navigation";
import { getMemberSession } from "@/lib/auth/server";
import { getAuth } from "@/lib/auth/configuration";
import WorkspaceContent from "./content";
export default async function Page() {
  const session = await getMemberSession();
  if (!session) redirect("/auth?returnTo=/workspaces");
  const organizations = await getAuth().api.listOrganizations({ headers: new Headers({ cookie: session.cookie_header }) });
  return <WorkspaceContent organizations={organizations.map(({ id, name }) => ({ id, name }))} email={session.user.email} />;
}
