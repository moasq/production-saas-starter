import { Pool } from "pg";
let pool: Pool | undefined;
export function getAuthDatabase(): Pool {
  pool ??= new Pool({
    ...(process.env.AUTH_DATABASE_URL ? { connectionString: process.env.AUTH_DATABASE_URL } : {}),
    options: "-c search_path=auth", max: 10, connectionTimeoutMillis: 5000,
  });
  return pool;
}
