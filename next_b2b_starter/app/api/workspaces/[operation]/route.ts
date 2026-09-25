import { selectWorkspace, acceptInvitation, requestSignup, completeLogin } from "@/lib/auth/flows";
export async function POST(request: Request, context: { params: Promise<{ operation: string }> }) {
  try {
    const { operation } = await context.params;
    const body = await request.json();
    switch (operation) {
      case "select": await selectWorkspace(request.headers, body.organizationId); return Response.json({ success: true });
      case "accept-invitation": await acceptInvitation(request.headers, body.invitationId); return Response.json({ success: true });
      case "signup": return Response.json(await requestSignup(request.headers, body));
      case "complete-login": return Response.json({ destination: await completeLogin(request.headers, body.signupId) });
      default: return Response.json({ error: "Not found" }, { status: 404 });
    }
  } catch (error) { const code = error && typeof error === "object" && "statusCode" in error ? Number(error.statusCode) : 400;
    return Response.json({ error: "Could not complete request. Verify your details and session, or wait before retrying." }, { status: code >= 400 && code < 500 ? code : 400 }); }
}
