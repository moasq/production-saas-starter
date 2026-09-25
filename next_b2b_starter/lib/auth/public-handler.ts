import { getAuth } from "./configuration.ts";
const allowed = new Set(["/magic-link/verify", "/sign-out", "/sign-in/magic-link"]);
export async function handleIdentityRequest(request: Request, forward: (request: Request) => Promise<Response> = (request) => getAuth().handler(request)): Promise<Response> {
  const path = new URL(request.url).pathname.replace(/^\/api\/identity/, "");
  if (!allowed.has(path)) return Response.json({ error: "Not found" }, { status: 404 });
  if (path === "/sign-in/magic-link" && request.method === "POST") {
    try {
      const body = await request.json();
      if (typeof body.email !== "string" || body.email.length > 254 || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(body.email.trim()) ||
          (body.name !== undefined && (typeof body.name !== "string" || body.name.trim().length < 2 || body.name.length > 100))) {
        return Response.json({ error: "A valid email and an optional name of 2–100 characters are required" }, { status: 400 });
      }
      // Keep the public SDK route within the same bounded contract as the signup UI.
      const input: Record<string, string> = { email: body.email.trim().toLowerCase() };
      if (body.name !== undefined) input.name = body.name.trim();
      for (const key of ["callbackURL", "newUserCallbackURL", "errorCallbackURL"]) {
        if (body[key] !== undefined) {
          if (typeof body[key] !== "string" || body[key].length > 2000) return Response.json({ error: "Invalid callback URL" }, { status: 400 });
          input[key] = body[key];
        }
      }
      const headers = new Headers(request.headers); headers.delete("content-length"); headers.set("content-type", "application/json");
      request = new Request(request.url, { method: request.method, headers, body: JSON.stringify(input) });
    } catch { return Response.json({ error: "Invalid sign-in request" }, { status: 400 }); }
  }
  return forward(request);
}
