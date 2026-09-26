// Integration tests against a disposable local Compose stack, with real auth/email.
import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { execFileSync } from 'node:child_process';
const base = process.env.STARTER_URL || 'http://localhost:3000';
const inbox = process.env.MAILPIT_URL || 'http://localhost:8025';
for (const url of [base, inbox]) {
  assert.ok(['localhost', '127.0.0.1'].includes(new URL(url).hostname), 'Use a disposable local stack only');
}
const contract = process.env.API_CONTRACT_CHECK === 'true'
  ? await import('../next_b2b_starter/scripts/api-contract.mjs') : undefined;
const run = randomUUID().slice(0, 8);
const check = (name) => console.log(`PASS ${name}`);
class Client {
  cookies = new Map();
  setCookies = [];
  async request(path, method = 'GET', body, extraHeaders = {}, retry = 0) {
    const response = await fetch(new URL(path, base), {
      method, redirect: 'manual', signal: AbortSignal.timeout(20000),
      headers: { origin: base, cookie: [...this.cookies].map(([k,v]) => `${k}=${v}`).join('; '),
        ...(body ? {'content-type': 'application/json'} : {}), ...extraHeaders },
      body: body ? JSON.stringify(body) : undefined,
    });
    if (response.status === 429 && retry < 2) {
      const seconds = Math.min(61, Math.max(1, Number(response.headers.get('retry-after') || response.headers.get('x-retry-after')) || 60));
      await response.text();
      console.log(`Rate limit enforced; retrying test request after ${seconds}s.`);
      await new Promise(resolve => setTimeout(resolve, seconds * 1000 + 100));
      return this.request(path, method, body, extraHeaders, retry + 1);
    }
    for (const cookie of response.headers.getSetCookie()) {
      this.setCookies.push(cookie);
      const [pair] = cookie.split(';');
      const index = pair.indexOf('=');
      if (index > 0) this.cookies.set(pair.slice(0, index), pair.slice(index + 1));
    }
    const text = await response.text();
    let data; try { data = JSON.parse(text); } catch { data = text; }
    contract?.assertApiResponse(new URL(path, base), method, response.status, data, body);
    return { status: response.status, data, location: response.headers.get('location') };
  }
  async ok(path, method = 'GET', body) {
    const result = await this.request(path, method, body);
    assert.ok(result.status >= 200 && result.status < 300, `${method} ${path}: ${result.status} ${JSON.stringify(result.data).slice(0,300)}`);
    return result.data;
  }
}
async function mail(email, contains) {
  for (let attempt = 0; attempt < 40; attempt++) {
    const list = await fetch(`${inbox}/api/v1/messages`).then(r => r.json());
    for (const item of list.messages || []) {
      if (!(item.To || []).some(x => x.Address === email)) continue;
      const message = await fetch(`${inbox}/api/v1/message/${item.ID}`).then(r => r.json());
      const text = message.Text || message.HTML || '';
      if (text.includes(contains)) return text;
    }
    await new Promise(resolve => setTimeout(resolve, 250));
  }
  throw new Error(`Expected local email was not delivered for ${email}`);
}
async function login(label) {
  const client = new Client();
  const email = `auth-${run}-${label}@example.test`;
  await client.ok('/api/identity/sign-in/magic-link', 'POST', {email, name: `Test ${label}`, callbackURL: `${base}/workspaces`});
  const message = await mail(email, '/api/identity/magic-link/verify');
  const url = message.match(/https?:\/\/[^\s<>"']+\/api\/identity\/magic-link\/verify[^\s<>"']*/)?.[0]?.replaceAll('&amp;', '&');
  assert.ok(url, 'Magic-link email contains verification URL');
  const result = await client.request(url);
  assert.ok([200,302,303,307].includes(result.status), `Magic verification ${result.status}`);
  await client.ok('/api/workspaces');
  assert.equal((await client.request('/api/identity/get-session')).status,404, 'Raw session token endpoint is not exposed');
  assert.ok(client.cookies.size > 0);
  assert.ok(client.setCookies.some(c => /session_token=/.test(c) && /HttpOnly/i.test(c) && /SameSite=Lax/i.test(c)), 'Session cookie is HttpOnly and SameSite=Lax');
  const replay = await new Client().request(url);
  assert.ok(replay.status >= 400 || (replay.location || '').includes('error'), 'Magic link cannot be reused');
  return {client, email};
}
async function invite(admin, user, role) {
  const result = await admin.ok('/api/auth/members', 'POST', {email: user.email, name: 'Invited test user', role_slug: role});
  assert.equal(result.data.invite_sent, true, 'SMTP delivery acknowledged');
  const pendingId = result.data.member_id;
  assert.ok(pendingId.startsWith('invitation:'));
  await mail(user.email, pendingId.slice(11));
  return {pendingId, invitationId: pendingId.slice(11)};
}
async function accept(user, invitationId) {
  await user.client.ok('/api/workspaces/accept-invitation', 'POST', {invitationId});
  return (await user.client.ok('/api/auth/profile/me')).data;
}
const invalidName = await new Client().request('/api/identity/sign-in/magic-link','POST',{email:`auth-${run}-invalid@example.test`,name:'X'.repeat(300)});
assert.equal(invalidName.status,400,'Public login cannot bypass profile name bounds');
check('public login input bounds prevent uneditable oversized profiles');
const ownerA = await login('owner-a');
const ownerB = await login('owner-b');
const member = await login('member');
check('real magic-link email, verified sessions and one-time tokens');
const orgA = (await ownerA.client.ok('/api/workspaces', 'POST', {name:`Alpha ${run}`})).organization;
const orgB = (await ownerB.client.ok('/api/workspaces', 'POST', {name:`Beta ${run}`})).organization;
assert.notEqual(orgA.id, orgB.id);
const profileA = (await ownerA.client.ok('/api/auth/profile/me')).data;
const profileB = (await ownerB.client.ok('/api/auth/profile/me')).data;
assert.equal(profileA.organization.organization_id, orgA.id);
assert.equal(profileB.organization.organization_id, orgB.id);
assert.deepEqual(profileA.roles, ['admin']);
assert.equal(profileA.email_verified, true);
assert.equal(profileA.email, ownerA.email);
assert.equal((await ownerA.client.ok('/api/organizations')).data.id, orgA.id);
await ownerA.client.ok('/api/organizations', 'PUT', {name:`Updated Alpha ${run}`});
assert.equal((await ownerA.client.ok('/api/organizations')).data.name, `Updated Alpha ${run}`);
assert.ok((await ownerA.client.ok('/api/auth/members')).data.members.some(m => m.member_id === profileA.member_id));
check('two workspaces created through real authenticated frontend API and mirrored in Go');
const forbiddenSelection = await ownerA.client.request('/api/workspaces/select', 'POST', {organizationId:orgB.id});
assert.ok([400,401,403,404].includes(forbiddenSelection.status));
const forged = await ownerA.client.request('/api/auth/profile/me', 'GET', undefined, {'x-organization-id':orgB.id, 'x-tenant-id':orgB.id, 'x-role':'admin'});
assert.equal(forged.data.data.organization.organization_id, orgA.id);
const crossMember = await ownerA.client.request(`/api/auth/members/${profileB.member_id}`, 'DELETE');
assert.ok([400,401,403,404].includes(crossMember.status), `Cross-tenant member deletion: ${crossMember.status}`);
assert.equal((await ownerB.client.ok('/api/auth/profile/me')).data.member_id, profileB.member_id);
check('forged tenant headers, workspace selection and cross-tenant member ID denied');
const invitationA = await invite(ownerA.client, member, 'member');
await ownerA.client.ok(`/api/auth/members/${encodeURIComponent(invitationA.pendingId)}/resend-invitation`, 'POST');
const wrongRecipient = await ownerB.client.request('/api/workspaces/accept-invitation', 'POST', {invitationId:invitationA.invitationId});
assert.ok([400,401,403,404].includes(wrongRecipient.status));
const memberA = await accept(member, invitationA.invitationId);
assert.deepEqual(memberA.roles, ['member']);
assert.equal((await member.client.request('/api/auth/members')).status, 403);
assert.equal((await member.client.request('/api/organizations', 'PUT', {name:'Unauthorized'})).status, 403);
await member.client.ok('/api/auth/profile/me', 'PUT', {name:'Updated member'});
assert.equal((await member.client.ok('/api/auth/profile/me')).data.name, 'Updated member');
check('invitation delivery/resend/acceptance, wrong-recipient denial and member permissions');
await ownerA.client.ok(`/api/auth/members/${memberA.member_id}`, 'PUT', {role:'manager'});
assert.deepEqual((await member.client.ok('/api/auth/profile/me')).data.roles, ['manager']);
assert.equal((await member.client.request('/api/auth/members', 'POST', {email:'nobody@example.test',name:'No',role_slug:'admin'})).status,403);
check('role changes take effect on existing session; manager cannot grant permissions');
const invitationB = await invite(ownerB.client, member, 'member');
await accept(member, invitationB.invitationId);
const memberships = await member.client.ok('/api/workspaces');
assert.equal(memberships.organizations.length,2);
await member.client.ok('/api/workspaces/select','POST',{organizationId:orgA.id});
assert.deepEqual((await member.client.ok('/api/auth/profile/me')).data.roles,['manager']);
await member.client.ok('/api/workspaces/select','POST',{organizationId:orgB.id});
assert.deepEqual((await member.client.ok('/api/auth/profile/me')).data.roles,['member']);
check('one identity belongs to two organizations with independent roles');
await member.client.ok('/api/workspaces/select','POST',{organizationId:orgA.id});
await ownerA.client.ok(`/api/auth/members/${memberA.member_id}`,'DELETE');
assert.ok([401,403].includes((await member.client.request('/api/auth/profile/me')).status));
assert.ok([400,401,403,404].includes((await member.client.request('/api/workspaces/select','POST',{organizationId:orgA.id})).status));
await member.client.ok('/api/workspaces/select','POST',{organizationId:orgB.id});
assert.equal((await member.client.ok('/api/auth/profile/me')).data.organization.organization_id,orgB.id);
const renewed = await invite(ownerA.client,member,'member');
await accept(member,renewed.invitationId);
check('removal immediately denies old membership, preserves other tenant and allows reinvitation');
const lastAdmin = await ownerA.client.request(`/api/auth/members/${profileA.member_id}`,'PUT',{role:'member'});
assert.ok([400,403,409].includes(lastAdmin.status),`Last admin demotion: ${lastAdmin.status}`);
const removeAdmin = await ownerA.client.request(`/api/auth/members/${profileA.member_id}`,'DELETE');
assert.ok([400,403,409].includes(removeAdmin.status),`Last admin removal: ${removeAdmin.status}`);
assert.equal((await ownerA.client.request('/api/identity/organization/update-member-role','POST',{memberId:profileA.member_id,role:'member'})).status,404);
check('last administrator and alternate mutation-path protections');
const reinvitedProfile = (await member.client.ok('/api/auth/profile/me')).data;
await ownerA.client.ok(`/api/auth/members/${reinvitedProfile.member_id}`,'PUT',{role:'admin'});
const concurrentDemotions = await Promise.all([
  ownerA.client.request(`/api/auth/members/${profileA.member_id}`,'PUT',{role:'member'}),
  member.client.request(`/api/auth/members/${reinvitedProfile.member_id}`,'PUT',{role:'member'}),
]);
assert.equal(concurrentDemotions.filter(x => x.status >= 200 && x.status < 300).length,1,'Exactly one concurrent admin demotion can succeed');
const ownerRole = (await ownerA.client.ok('/api/auth/profile/me')).data.roles[0];
if (ownerRole !== 'admin') await member.client.ok(`/api/auth/members/${profileA.member_id}`,'PUT',{role:'admin'});
check('concurrent demotions cannot leave the tenant without an administrator');
const csrf = await ownerA.client.request('/api/organizations','PUT',{name:'CSRF'},{origin:'https://attacker.example'});
assert.equal(csrf.status,403);
assert.equal((await ownerA.client.request('/internal/auth/session','POST',{})).status,404);
const billing = await ownerA.client.ok('/api/subscriptions/status');
assert.equal(billing.BillingEnabled, false, 'Billing is disabled');
assert.equal(billing.HasActiveSubscription, false);
assert.equal((await ownerA.client.request('/api/subscriptions/verify-payment', 'POST', {session_id:randomUUID()})).status, 503);
check('cross-origin writes denied, private bridge hidden, billing disabled');
const oldCookies = new Map(member.client.cookies);
await member.client.ok('/api/identity/sign-out','POST',{});
member.client.cookies=oldCookies;
assert.ok([401,403].includes((await member.client.request('/api/auth/profile/me')).status));
check('logout revokes the server session even when replaying the old cookie');
// Force expiry for this test-created user; this is a negative-path fixture, not an auth bypass.
if (process.env.TEST_SESSION_EXPIRY === 'true') {
  const sql = (query) => execFileSync('docker',['compose','exec','-T','postgres','sh','-c','psql -At -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"'], {input:query,encoding:'utf8',stdio:['pipe','pipe','pipe']}).trim();
  const filter = `"userId" IN (SELECT id FROM auth."user" WHERE email = '${ownerB.email}')`;
  sql(`UPDATE auth.session SET "expiresAt"=now()+interval '1 hour', "updatedAt"=now()-interval '7 hours' WHERE ${filter};`);
  const expiresBefore = sql(`SELECT "expiresAt" FROM auth.session WHERE ${filter};`);
  await ownerB.client.ok('/api/auth/profile/me');
  assert.equal(sql(`SELECT "expiresAt" FROM auth.session WHERE ${filter};`),expiresBefore,'Private verification must not silently extend session expiry');
  check('fixed session lifetime stays synchronized with the browser cookie');
  execFileSync('docker',['compose','exec','-T','postgres','sh','-c','psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"'], {
    input:`UPDATE auth.session SET "expiresAt" = now() - interval '1 second' WHERE "userId" IN (SELECT id FROM auth."user" WHERE email = '${ownerB.email}');`, stdio:['pipe','pipe','pipe'],
  });
  assert.ok([401,403].includes((await ownerB.client.request('/api/auth/profile/me')).status));
  check('expired sessions rejected by frontend authority and Go API');
}
contract?.assertApiCoverage();
console.log(`Self-hosted authentication integration passed (fixture ${run}).`);
