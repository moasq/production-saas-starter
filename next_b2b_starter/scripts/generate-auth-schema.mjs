// Run against an empty auth schema; review and commit the generated SQL.
import { writeFile } from 'node:fs/promises';
import { getMigrations } from 'better-auth/db/migration';
import { createAuth } from '../lib/auth/configuration.ts';
import { getAuthDatabase } from '../lib/auth/database.ts';
const auth = createAuth();
const plan = await getMigrations(auth.options);
if (plan.unsafeChanges.length || plan.schemaProblems.length) throw new Error('Unsafe auth schema migration');
await writeFile(process.argv[2] || 'auth-migrations/0001_better_auth.sql', await plan.compileMigrations());
await getAuthDatabase().end();
