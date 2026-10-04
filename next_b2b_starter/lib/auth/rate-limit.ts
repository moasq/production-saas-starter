import { createHash } from "node:crypto";
import { getAuthDatabase } from "./database.ts";
export async function consumeAuthLimit(key: string, max = 5, window = 600): Promise<boolean> {
  const now = Date.now();
  const bucket = Math.floor(now / (window * 1000));
  const hashed = createHash("sha256").update(`${key}:${bucket}`).digest("hex");
  const result = await getAuthDatabase().query<{ count: number }>(
    `INSERT INTO auth_rate_limit (key, count, expires_at) VALUES ($1, 1, $2)
     ON CONFLICT (key) DO UPDATE SET count = auth_rate_limit.count + 1 RETURNING count`,
    [hashed, new Date((bucket + 1) * window * 1000)],
  );
  await getAuthDatabase().query("DELETE FROM auth_rate_limit WHERE key IN (SELECT key FROM auth_rate_limit WHERE expires_at < now() LIMIT 100)");
  return result.rows[0].count <= max;
}
// A stable key with a database-clock deadline is a true cooldown: requests on
// opposite sides of a fixed window must not both send mail. Failed attempts do
// not extend the deadline. Keep the reservation after an SMTP failure because a
// timeout can occur after the server accepted the message.
export async function consumeInvitationLimit(organizationId: string, email: string): Promise<boolean> {
  const key = createHash("sha256").update(`invitation:${organizationId}:${email.trim().toLowerCase()}`).digest("hex");
  const result = await getAuthDatabase().query(
    `INSERT INTO auth_rate_limit (key, count, expires_at) VALUES ($1, 1, now() + interval '60 seconds')
     ON CONFLICT (key) DO UPDATE SET count = 1, expires_at = EXCLUDED.expires_at
     WHERE auth_rate_limit.expires_at <= now() RETURNING key`, [key],
  );
  if (result.rowCount !== 1) return false;
  return consumeAuthLimit(`invitation-org:${organizationId}`, 20, 600);
}
export async function limitPublicEmail(headers: Headers, email: string): Promise<void> {
  const expected = new URL(process.env.APP_BASE_URL || "http://localhost:3000").origin;
  if (headers.get("origin") !== expected) throw new Error("Invalid origin");
  const ip = headers.get("x-forwarded-for")?.split(",")[0]?.trim() || "unknown";
  const [byEmail, byIP] = await Promise.all([
    consumeAuthLimit(`email:${email.toLowerCase()}`), consumeAuthLimit(`ip:${ip}`, 20),
  ]);
  if (!byEmail || !byIP) throw new Error("Too many requests. Please wait before trying again.");
}
