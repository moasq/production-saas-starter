export interface RequestOptions { headers?: Record<string, string>; skipAuth?: boolean; }
export interface ApiClientConfig { baseUrl: string; defaultHeaders: Record<string, string>; }
// No session tokens or cross-request authentication state live in this module.
export class ApiClient {
 private config: ApiClientConfig;
 constructor(config?: Partial<ApiClientConfig>) {
  this.config = { baseUrl: config?.baseUrl || (typeof window === "undefined" ? process.env.API_BASE_URL_INTERNAL || "http://localhost:8080/api" : "/api"), defaultHeaders: config?.defaultHeaders || {} };
 }
 getBaseUrl() { return this.config.baseUrl; }
 get<T>(path: string, options?: RequestOptions) { return this.request<T>(path, "GET", undefined, options); }
 post<T>(path: string, body?: unknown, options?: RequestOptions) { return this.request<T>(path, "POST", body, options); }
 put<T>(path: string, body?: unknown, options?: RequestOptions) { return this.request<T>(path, "PUT", body, options); }
 delete<T>(path: string, options?: RequestOptions) { return this.request<T>(path, "DELETE", undefined, options); }
 private async request<T>(path: string, method: string, body?: unknown, options?: RequestOptions): Promise<T> {
  const headers = new Headers({ ...this.config.defaultHeaders, ...options?.headers });
  if (body !== undefined) headers.set("Content-Type", "application/json");
  const send = () => fetch(`${this.config.baseUrl}${path}`, { method, headers, body: body === undefined ? undefined : JSON.stringify(body), credentials: "same-origin", cache: "no-store", signal: AbortSignal.timeout(15000) });
  let response = await send();
  if (response.status === 401 && !options?.skipAuth && !options?.headers?.Authorization && typeof window !== "undefined") {
    const refreshed = await fetch("/api/auth/session/refresh", { method: "POST", credentials: "same-origin" });
    if (refreshed.ok) response = await send();
  }
  const payload = response.status === 204 ? null : await response.json().catch(() => null);
  if (!response.ok) throw new Error(payload?.message || payload?.error || `Request failed (${response.status})`);
  return payload as T;
 }
}
export const apiClient = new ApiClient();
