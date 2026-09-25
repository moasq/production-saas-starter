// Explicit, reviewed cutover only. No remote-provider calls and no writes to the business schema.
import { createHash } from 'node:crypto';
import { Pool } from 'pg';
if (!process.env.LEGACY_DATABASE_URL) throw new Error('LEGACY_DATABASE_URL is required for the reviewed source snapshot');
const apply = process.argv.includes('--apply');
const sourcePool = new Pool({ connectionString: process.env.LEGACY_DATABASE_URL, max: 1 });
const targetPool = new Pool({ ...(process.env.AUTH_DATABASE_URL ? { connectionString: process.env.AUTH_DATABASE_URL } : {}), options: '-c search_path=auth', max: 1 });
const source = await sourcePool.connect(), target = await targetPool.connect();
const roleMap = { owner: 'admin', admin: 'admin', manager: 'manager', reviewer: 'manager', approver: 'manager', member: 'member', employee: 'member' };
const identitySQL = 'SELECT current_database() AS database, inet_server_addr()::text AS server, inet_server_port() AS port';
try {
  await source.query('BEGIN ISOLATION LEVEL REPEATABLE READ READ ONLY');
  const sourceID = (await source.query(identitySQL)).rows[0], targetID = (await target.query(identitySQL)).rows[0];
  if (JSON.stringify(sourceID) !== JSON.stringify(targetID)) throw new Error('Source and destination must be schemas in the same PostgreSQL database');
  const sourceRole = (await source.query('SELECT rolsuper OR rolbypassrls AS allowed FROM pg_roles WHERE rolname=current_user')).rows[0];
  if (!sourceRole?.allowed) throw new Error('The reviewed source connection must be able to read all legacy rows despite tenant RLS');
  const organizations = (await source.query(`SELECT id, auth_org_id, polar_customer_external_id, slug, name, created_at FROM organizations.organizations WHERE status='active' ORDER BY id`)).rows;
  const accounts = (await source.query(`SELECT a.id, a.organization_id, a.email, a.full_name, a.role, a.stytch_member_id, a.created_at
    FROM organizations.accounts a JOIN organizations.organizations o ON o.id=a.organization_id
    WHERE a.status='active' AND o.status='active' ORDER BY a.id`)).rows;
  const pairs = new Set(), memberIDs = new Set();
  for (const org of organizations) {
    if (!org.auth_org_id || !org.polar_customer_external_id) throw new Error('Review required: organization identity or Polar mapping is missing');
    if (!accounts.some((a) => a.organization_id === org.id && roleMap[a.role] === 'admin')) throw new Error('Review required: an active organization has no active administrator');
  }
  for (const account of accounts) {
    account.email = String(account.email).trim().toLowerCase();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(account.email) || !roleMap[account.role]) throw new Error('Review required: unknown role or invalid email');
    const pair = `${account.organization_id}:${account.email}`;
    if (pairs.has(pair)) throw new Error('Review required: duplicate normalized email within one organization');
    pairs.add(pair);
    account.memberID = account.stytch_member_id || `legacy-member-${account.id}`;
    if (memberIDs.has(account.memberID)) throw new Error('Review required: duplicate legacy member identifiers');
    memberIDs.add(account.memberID);
  }
  await target.query('BEGIN');
  await target.query("SELECT pg_advisory_xact_lock(hashtext('b2b-auth-legacy-import'))");
  const markers = new Map((await target.query('SELECT source_kind, source_id, target_id FROM legacy_import')).rows.map((row) => [`${row.source_kind}:${row.source_id}`, row.target_id]));
  const summary = { mode: apply ? 'apply' : 'dry-run', organizations: 0, users: 0, memberships: 0, previouslyImported: markers.size, skippedInactive: true, legacyRowsChanged: 0 };
  const users = new Map((await target.query('SELECT id, lower(email) AS email FROM "user"')).rows.map((row) => [row.email,row.id]));
  for (const org of organizations) {
    if (markers.has(`organization:${org.id}`)) continue;
    const conflicts = await target.query('SELECT id FROM organization WHERE id=$1 OR slug=$2', [org.auth_org_id,org.slug]);
    if (conflicts.rowCount) throw new Error('Review required: an existing auth organization conflicts with the legacy import');
    summary.organizations++;
    if (apply) {
      await target.query('INSERT INTO organization (id,name,slug,"createdAt") VALUES ($1,$2,$3,$4)', [org.auth_org_id,org.name,org.slug,org.created_at]);
      await target.query("INSERT INTO legacy_import (source_kind,source_id,target_id) VALUES ('organization',$1,$2)", [org.id,org.auth_org_id]);
    }
  }
  for (const account of accounts) {
    if (markers.has(`account:${account.id}`)) continue;
    const org = organizations.find((org) => org.id === account.organization_id);
    let userID = users.get(account.email);
    if (!userID) {
      userID = `legacy-user-${createHash('sha256').update(account.email).digest('hex')}`;
      users.set(account.email,userID); summary.users++;
      if (apply) await target.query('INSERT INTO "user" (id,name,email,"emailVerified","createdAt","updatedAt") VALUES ($1,$2,$3,false,$4,$4)', [userID,account.full_name,account.email,account.created_at]);
    }
    const conflict = await target.query('SELECT id FROM member WHERE id=$1 OR ("organizationId"=$2 AND "userId"=$3)', [account.memberID,org.auth_org_id,userID]);
    if (conflict.rowCount) throw new Error('Review required: an existing membership conflicts with the legacy import');
    summary.memberships++;
    if (apply) {
      await target.query('INSERT INTO member (id,"organizationId","userId",role,"createdAt") VALUES ($1,$2,$3,$4,$5)', [account.memberID,org.auth_org_id,userID,roleMap[account.role],account.created_at]);
      await target.query("INSERT INTO legacy_import (source_kind,source_id,target_id) VALUES ('account',$1,$2)", [account.id,account.memberID]);
    }
  }
  await target.query(apply ? 'COMMIT' : 'ROLLBACK');
  await source.query('COMMIT');
  console.log(JSON.stringify(summary));
} catch (error) {
  await Promise.allSettled([source.query('ROLLBACK'),target.query('ROLLBACK')]);
  console.error(error instanceof Error ? error.message : 'Legacy import failed');
  process.exitCode = 1;
} finally { source.release(); target.release(); await Promise.all([sourcePool.end(),targetPool.end()]); }
