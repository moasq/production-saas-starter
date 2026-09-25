import { createHash } from 'node:crypto';
import { readdir, readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { Pool } from 'pg';
const pool = new Pool({ ...(process.env.AUTH_DATABASE_URL ? { connectionString: process.env.AUTH_DATABASE_URL } : {}), options: '-c search_path=auth' });
const db = await pool.connect();
try {
  await db.query('BEGIN');
  await db.query("SELECT pg_advisory_xact_lock(hashtext('b2b-auth-migrations'))");
  if ((await db.query("SELECT current_schema() AS name")).rows[0].name !== 'auth') throw new Error('Create schema auth owned by the authentication role first');
  await db.query('CREATE TABLE IF NOT EXISTS starter_auth_migrations (name text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())');
  const dir = fileURLToPath(new URL('../auth-migrations/', import.meta.url));
  for (const name of (await readdir(dir)).filter((name) => name.endsWith('.sql')).sort()) {
    const sql = await readFile(`${dir}/${name}`, 'utf8');
    const checksum = createHash('sha256').update(sql).digest('hex');
    const prior = await db.query('SELECT checksum FROM starter_auth_migrations WHERE name=$1', [name]);
    if (prior.rows[0]) {
      if (prior.rows[0].checksum !== checksum) throw new Error(`Applied migration changed: ${name}`);
      continue;
    }
    await db.query(sql);
    await db.query('INSERT INTO starter_auth_migrations (name, checksum) VALUES ($1,$2)', [name,checksum]);
    console.log(`Applied ${name}`);
  }
  await db.query('COMMIT');
} catch (error) { await db.query('ROLLBACK'); throw error; }
finally { db.release(); await pool.end(); }
