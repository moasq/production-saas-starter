import "server-only";
export function isAuthConfigured(): boolean {
  return Boolean(process.env.BETTER_AUTH_SECRET && process.env.BETTER_AUTH_SECRET.length >= 32 &&
    (process.env.AUTH_DATABASE_URL || process.env.PGHOST) && process.env.SMTP_HOST && process.env.EMAIL_FROM);
}
