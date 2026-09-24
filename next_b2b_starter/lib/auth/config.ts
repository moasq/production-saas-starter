import "server-only";
export function isAuthConfigured(): boolean {
  return Boolean(process.env.STYTCH_PROJECT_ID?.trim() && process.env.STYTCH_SECRET?.trim());
}
