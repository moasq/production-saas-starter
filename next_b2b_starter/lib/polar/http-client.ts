import { HTTPClient, type HTTPClientOptions } from "@polar-sh/sdk/lib/http";

export function createPolarHttpClient(options?: HTTPClientOptions): HTTPClient {
  return new HTTPClient(options).addHook("beforeRequest", (request) => {
    // Keep this aligned with the Go client and the installed SDK's OpenAPI version.
    request.headers.set("Polar-Version", "2026-04");
  });
}
