import createClient, { type Client } from "openapi-fetch";
import type { paths } from "../../generated/schema.ts";

export interface ApiClientConfig { baseUrl: string; defaultHeaders: Record<string, string>; }

// Operations, path parameters, bodies and responses come from the Go-owned schema.
// Cookies are passed per request; this shared client never stores an actor/session.
export class ApiClient {
  readonly GET: Client<paths>["GET"];
  readonly POST: Client<paths>["POST"];
  readonly PUT: Client<paths>["PUT"];
  readonly DELETE: Client<paths>["DELETE"];
  constructor(config?: Partial<ApiClientConfig>) {
    const client = createClient<paths>({
      baseUrl: config?.baseUrl || (typeof window === "undefined" ? process.env.API_BASE_URL_INTERNAL || "http://localhost:8080/api" : "/api"),
      headers: config?.defaultHeaders,
      credentials: "same-origin", cache: "no-store",
      fetch: (request) => globalThis.fetch(request),
    });
    client.use({ onRequest({ request }) {
      if (typeof window === "undefined" && request.method !== "GET") {
        request.headers.set("Origin", new URL(process.env.APP_BASE_URL || "http://localhost:3000").origin);
      }
      return new Request(request, { signal: AbortSignal.any([request.signal, AbortSignal.timeout(15000)]) });
    }, onResponse({ response }) {
      // CORS/proxies can reject a request before a JSON handler runs.
      if (!response.ok && !response.headers.get("content-type")?.includes("application/json")) {
        return Response.json({ message: `Request failed (${response.status})` }, { status: response.status });
      }
    } });
    this.GET = client.GET; this.POST = client.POST; this.PUT = client.PUT; this.DELETE = client.DELETE;
  }
}

export function unwrap<T>(result: { data?: T; error?: unknown; response: Response }): T {
  if (!result.response.ok) {
    const error = result.error;
    const message = error && typeof error === "object"
      ? ("message" in error ? error.message : "error" in error ? error.error : undefined) : undefined;
    throw new Error(typeof message === "string" ? message : `Request failed (${result.response.status})`);
  }
  return result.data as T;
}
export const apiClient = new ApiClient();
