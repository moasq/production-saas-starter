import { getAuth } from "@/lib/auth/configuration";
import { verifyIdentity } from "@/lib/auth/identity";
import { createWorkspace } from "@/lib/auth/flows";
export async function GET(request: Request) {
  const identity = await verifyIdentity(request.headers);
  if (!identity) return Response.json({ error: "Authentication required" }, { status: 401 });
  return Response.json({ organizations: await getAuth().api.listOrganizations({ headers: request.headers }), activeOrganizationId: identity.membership?.id || null });
}
export async function POST(request: Request) {
  try {
    const body = await request.json();
    return Response.json({ organization: await createWorkspace(request.headers, body.name) });
  } catch (error) { const code = error && typeof error === "object" && "statusCode" in error ? Number(error.statusCode) : 400;
    return Response.json({ error: "Could not create workspace. Verify your session and organization name." }, { status: code >= 400 && code < 500 ? code : 400 }); }
}
