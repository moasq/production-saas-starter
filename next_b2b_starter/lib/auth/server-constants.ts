import "server-only";

// Server-only: Read session duration from environment
export function getSessionDurationMinutes(): number {
  return (
    Number(process.env.STYTCH_SESSION_DURATION_MINUTES ?? "480") ||
    480
  );
}

// Server-only: Cookie config builder
export function getCookieConfig() {
  const isSecure = process.env.APP_BASE_URL?.startsWith("https://") ?? false;

  return {
    httpOnly: true, // Prevent XSS attacks from accessing cookies
    sameSite: "lax" as const,
    secure: isSecure,
    path: "/",
  };
}
