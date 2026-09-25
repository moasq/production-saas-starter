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
export async function limitPublicEmail(headers: Headers, email: string): Promise<void> {
  const expected = new URL(process.env.APP_BASE_URL || "http://localhost:3000").origin;
  if (headers.get("origin") !== expected) throw new Error("Invalid origin");
  const ip = headers.get("x-forwarded-for")?.split(",")[0]?.trim() || "unknown";
  const [byEmail, byIP] = await Promise.all([
    consumeAuthLimit(`email:${email.toLowerCase()}`), consumeAuthLimit(`ip:${ip}`, 20),
  ]);
  if (!byEmail || !byIP) throw new Error("Too many requests. Please wait before trying again.");
}
