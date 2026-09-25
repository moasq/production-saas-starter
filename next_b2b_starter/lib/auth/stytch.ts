// Authentication URL builders for Stytch
// These functions handle URL construction with proper safety checks

export interface LoginUrlOptions {
  returnTo?: string;
}

export interface LogoutUrlOptions {
  returnTo?: string;
}

export function sanitizeReturnTo(value: string | undefined): string | undefined {
  if (!value) return undefined;
  const trimmed = value.trim();
  if (!trimmed.startsWith("/") || trimmed.startsWith("//") || /[\\\x00-\x1f]/.test(trimmed)) {
    return undefined;
  }
  return trimmed;
}

export function getBaseUrl(): string {
  if (typeof window !== "undefined") {
    return window.location.origin.replace(/\/$/, "");
  }

  const baseUrl =
    process.env.APP_BASE_URL ||
    null;

  const resolved = baseUrl || "http://localhost:3000";
  return resolved.replace(/\/$/, "");
}

function resolveRoute(value: string | undefined, fallback: string): URL {
  const base = getBaseUrl();
  const trimmed = value?.trim();

  if (!trimmed) {
    return new URL(fallback, `${base}/`);
  }

  if (/^https?:\/\//i.test(trimmed)) {
    return new URL(trimmed);
  }

  const path = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
  return new URL(path, `${base}/`);
}

export function buildLoginUrl(options?: LoginUrlOptions): string {
  const url = resolveRoute(undefined, "/auth");
  const returnTo = sanitizeReturnTo(options?.returnTo);

  if (returnTo) {
    url.searchParams.set("returnTo", returnTo);
  }

  return url.toString();
}

export function buildLogoutUrl(options?: LogoutUrlOptions): string {
  const url = resolveRoute(undefined, "/auth");
  const returnTo = sanitizeReturnTo(options?.returnTo);

  if (returnTo) {
    url.searchParams.set("returnTo", returnTo);
  }

  return url.toString();
}
