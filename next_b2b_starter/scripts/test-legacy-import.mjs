// Creates and drops its own empty database. Never point this at a hosted production owner.
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { Pool } from 'pg';
if (!process.env.AUTH_IMPORT_TEST_DATABASE_URL) throw new Error('Set AUTH_IMPORT_TEST_DATABASE_URL to a disposable PostgreSQL owner connection');
const admin = new Pool({ connectionString: process.env.AUTH_IMPORT_TEST_DATABASE_URL });
const name = `auth_import_test_${randomUUID().replaceAll('-','')}`;
const targetURL = new URL(process.env.AUTH_IMPORT_TEST_DATABASE_URL); targetURL.pathname = `/${name}`;
const env = { ...process.env, AUTH_DATABASE_URL: targetURL.href, LEGACY_DATABASE_URL: targetURL.href };
const cwd = fileURLToPath(new URL('../', import.meta.url));
let db;
function run(file, args = [], expected = 0) {
  const result = spawnSync(process.execPath, [`scripts/${file}`, ...args], { cwd, env, encoding: 'utf8', timeout: 20000 });
  assert.equal(result.status, expected, result.stderr || result.stdout);
  return result.stdout.trim();
}
try {
  await admin.query(`CREATE DATABASE ${name}`);
  db = new Pool({ connectionString: targetURL.href });
  await db.query('CREATE SCHEMA auth; CREATE SCHEMA organizations');
  run('migrate-auth.mjs'); run('migrate-auth.mjs');
  await db.query(`
    CREATE TABLE organizations.organizations (id integer PRIMARY KEY, auth_org_id text, polar_customer_external_id text, slug text, name text, status text, created_at timestamptz DEFAULT now());
    CREATE TABLE organizations.accounts (id integer PRIMARY KEY, organization_id integer, email text, full_name text, role text, status text, stytch_member_id text, created_at timestamptz DEFAULT now());
    INSERT INTO organizations.organizations (id,auth_org_id,polar_customer_external_id,slug,name,status) VALUES
      (1,'organization-test-legacy-a','organization-test-legacy-a','legacy-alpha','Alpha','active'),
      (2,'organization-test-legacy-b','organization-test-legacy-b','legacy-beta','Beta','active'),
      (3,'organization-test-legacy-c','organization-test-legacy-c','legacy-inactive','Inactive','suspended');
    INSERT INTO organizations.accounts (id,organization_id,email,full_name,role,status,stytch_member_id) VALUES
      (1,1,'Shared@Example.test','Original Shared','owner','active','member-test-shared-a'),
      (2,2,'shared@example.test','Other Display Name','member','active','member-test-shared-b'),
      (3,2,'admin@example.test','Beta Admin','admin','active','member-test-admin-b'),
      (4,1,'suspended@example.test','Suspended','admin','suspended','member-test-suspended'),
      (5,1,'inactive@example.test','Inactive','member','inactive','member-test-inactive'),
      (6,3,'hidden@example.test','Hidden','owner','active','member-test-hidden');
  `);
  const before = (await db.query('SELECT * FROM organizations.organizations ORDER BY id')).rows;
  const dry = JSON.parse(run('import-legacy-auth.mjs'));
  assert.equal(dry.mode,'dry-run'); assert.equal(dry.users,2); assert.equal(dry.memberships,3); assert.equal(dry.organizations,2);
  assert.equal((await db.query('SELECT count(*)::int AS count FROM auth."user"')).rows[0].count,0);
  assert.equal((await db.query('SELECT count(*)::int AS count FROM auth.organization')).rows[0].count,0);
  const applied = JSON.parse(run('import-legacy-auth.mjs',['--apply'])); assert.equal(applied.memberships,3);
  const users = (await db.query('SELECT id,email,name,"emailVerified" FROM auth."user" ORDER BY email')).rows;
  assert.equal(users.length,2); assert.ok(users.every((user) => user.emailVerified === false));
  assert.equal(users.find((user) => user.email === 'shared@example.test').name, 'Original Shared');
  const memberships = (await db.query('SELECT id,"organizationId","userId",role FROM auth.member ORDER BY id')).rows;
  assert.equal(memberships.length,3);
  const shared = memberships.filter((row) => row.id.startsWith('member-test-shared'));
  assert.equal(shared[0].userId,shared[1].userId); assert.deepEqual(shared.map((row) => row.role),['admin','member']);
  assert.deepEqual(shared.map((row) => row.organizationId),['organization-test-legacy-a','organization-test-legacy-b']);
  assert.deepEqual((await db.query('SELECT * FROM organizations.organizations ORDER BY id')).rows,before);
  const repeat = JSON.parse(run('import-legacy-auth.mjs',['--apply'])); assert.equal(repeat.memberships,0); assert.equal(repeat.users,0); assert.equal(repeat.organizations,0);
  await db.query("DELETE FROM auth.member WHERE id='member-test-shared-b'");
  run('import-legacy-auth.mjs',['--apply']);
  assert.equal((await db.query("SELECT count(*)::int AS count FROM auth.member WHERE id='member-test-shared-b'")).rows[0].count,0,'Repeating an import must not reinstate removed access');
  await db.query("INSERT INTO organizations.accounts (id,organization_id,email,full_name,role,status) VALUES (7,1,'unknown@example.test','Unknown','super-admin','active')");
  run('import-legacy-auth.mjs',[],1);
  await db.query("DELETE FROM organizations.accounts WHERE id=7; INSERT INTO organizations.accounts (id,organization_id,email,full_name,role,status) VALUES (8,1,'SHARED@example.test','Duplicate','member','active')");
  run('import-legacy-auth.mjs',[],1);
  console.log('Legacy import passed: dry-run, deduplication, role/ID preservation, excluded inactive rows, idempotence, revocation, and ambiguity refusal.');
} finally {
  if (db) await db.end();
  await admin.query(`DROP DATABASE IF EXISTS ${name} WITH (FORCE)`);
  await admin.end();
}
