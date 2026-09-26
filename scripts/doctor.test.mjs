import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, copyFileSync, rmSync, existsSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { resolve, join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const secret = 'DO_NOT_PRINT_SECRET_'.repeat(3);
const checks = ['business_migrations','auth_migrations','database_roles','tenant_rls','auth_isolation','membership_roles'];
function fixture(t, { configured = true } = {}) {
  const dir = mkdtempSync(join(tmpdir(), 'starter doctor '));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  for (const file of ['setup.sh','scripts/doctor.sh','scripts/doctor-database.sql','.env.example','go-b2b-starter/internal/db/postgres/runtime.go']) {
    mkdirSync(dirname(join(dir,file)), { recursive: true }); copyFileSync(join(root,file),join(dir,file));
  }
  mkdirSync(join(dir,'next_b2b_starter/auth-migrations'), { recursive: true });
  writeFileSync(join(dir,'next_b2b_starter/auth-migrations/0001_test.sql'),'SELECT 1;');
  if (configured) writeFileSync(join(dir,'.env'), readFileSync(join(dir,'.env.example'),'utf8').replace(/^(POSTGRES_PASSWORD|APP_DATABASE_PASSWORD|AUTH_DATABASE_PASSWORD|BETTER_AUTH_SECRET|AUTH_INTERNAL_SECRET)=$/gm, `$1=${secret}`), { mode: 0o600 });
  const bin = join(dir,'bin'); mkdirSync(bin);
  writeFileSync(join(bin,'docker'), `#!${process.execPath}\n` + String.raw`
const fs = require('node:fs'); const args=process.argv.slice(2);
fs.appendFileSync(process.env.TEST_DOCKER_LOG, JSON.stringify(args)+'\n');
const bad=() => { console.error(process.env.TEST_SECRET); process.exit(7); };
if (args[0] === 'info') {
 if (process.env.TEST_DAEMON_FAIL) bad(); process.exit(0);
}
if (args[0] === 'context') { console.log('unix:///test/docker.sock'); process.exit(0); }
if (args[0] === 'inspect') {
 const service=args.at(-1);
 if (args.some(x=>x.includes('.Config.Labels'))) { console.log((process.env.TEST_STALE_SERVICE===service ? 'b' : 'a').repeat(64)); process.exit(0); }
 if (process.env.TEST_FAILED_SERVICE === service) console.log('exited 7 none');
 else if (['database-init','backend-migrate','auth-migrate'].includes(service)) console.log('exited 0 none');
 else console.log('running 0 ' + (['postgres','backend','frontend'].includes(service) ? 'healthy' : 'none'));
 process.exit(0);
}
if (args[0] !== 'compose') bad();
if (args.includes('version')) process.exit(0);
if (args.includes('--environment')) {
 if (!fs.readFileSync(0,'utf8').includes('image: scratch')) bad();
 const env=Object.fromEntries(fs.readFileSync('.env','utf8').split('\n').filter(x=>/^[A-Z_]+=/.test(x)).map(x=>[x.slice(0,x.indexOf('=')),x.slice(x.indexOf('=')+1)]));
 for (const [k,v] of Object.entries({...env,...JSON.parse(process.env.TEST_ENV_OVERRIDES || '{}')})) console.log(k+'='+v);
 process.exit(0);
}
if (args.includes('config')) { if (process.env.TEST_CONFIG_FAIL) bad(); if(args.includes('--hash')) console.log(args.at(-1)+' '+'a'.repeat(64)); process.exit(0); }
if (args.includes('port')) {
 if (process.env.TEST_UNBOUND) process.exit(1);
 console.log('127.0.0.1:'+({'80':'3000','443':'443','8025':'8025'}[args.at(-1)])); process.exit(0);
}
if (args.includes('ps')) { if (!process.env.TEST_NOT_STARTED) console.log(args.at(-1)); process.exit(0); }
if (args.includes('exec')) {
 const sql=fs.readFileSync(0,'utf8');
 if (!args.some(x=>x.includes('default_transaction_read_only=on')) || !sql.includes('SELECT')) bad();
 if (process.env.TEST_DATABASE_FAIL) bad();
 for (const name of ['business_migrations','auth_migrations','database_roles','tenant_rls','auth_isolation','membership_roles']) console.log(name+'='+(process.env.TEST_BAD_CHECK===name ? 'failed' : 'ok'));
 process.exit(0);
}
if (args.includes('up')) { if (process.env.TEST_START_FAIL) bad(); process.exit(0); }
bad();
`, { mode: 0o755 });
  writeFileSync(join(bin,'uname'), '#!/bin/sh\nprintf "%s\\n" "${TEST_HOST:-Linux}"\n', { mode: 0o755 });
  writeFileSync(join(bin,'lsof'), '#!/bin/sh\n[ "${TEST_PORT_BUSY:-}" = 1 ]\n', { mode: 0o755 });
  const log=join(dir,'docker.log');
  const run=(args=['--doctor'], env={}) => spawnSync('sh',[join(dir,'setup.sh'),...args], { cwd:tmpdir(), env:{...process.env,PATH:`${bin}:${process.env.PATH}`,TEST_DOCKER_LOG:log,TEST_SECRET:secret,...env}, encoding:'utf8', timeout:20000 });
  const calls=()=>existsSync(log)?readFileSync(log,'utf8').trim().split('\n').filter(Boolean).map(JSON.parse):[];
  return {dir,run,calls};
}
function clean(result) { assert.equal(result.error,undefined); assert.ok(!(result.stdout+result.stderr).includes(secret), 'secret must never be displayed'); }
test('healthy reruns are read-only and accept ports owned by this project', t=>{
 const f=fixture(t); const before=readFileSync(join(f.dir,'.env'));
 for (let i=0;i<2;i++) { const r=f.run(); clean(r); assert.equal(r.status,0,r.stdout+r.stderr); for(const key of checks) assert.match(r.stdout,new RegExp(`PASS ${key}`)); }
 assert.deepEqual(readFileSync(join(f.dir,'.env')),before);
 assert.ok(f.calls().every(args=>!args.some(a=>['up','down','build','run','restart','rm','stop','pull'].includes(a))));
});
test('missing configuration remains absent and points to setup',t=>{
 const f=fixture(t,{configured:false});const r=f.run();clean(r);assert.equal(r.status,1);assert.match(r.stdout,/Missing .env/);assert.equal(existsSync(join(f.dir,'.env')),false);
});
test('daemon and invalid Compose errors fail without leaking raw output',t=>{
 for(const key of ['TEST_DAEMON_FAIL','TEST_CONFIG_FAIL','TEST_DATABASE_FAIL']){const f=fixture(t);const r=f.run(['--doctor'],{[key]:'1'});clean(r);assert.equal(r.status,1,r.stdout);}
});
test('configuration mistakes and optional billing are diagnosed by name',t=>{
 for(const env of [{BETTER_AUTH_SECRET:'short'},{SMTP_HOST:'mailpit',COMPOSE_PROFILES:''},{APP_BASE_URL:'https://user:password@example.com'},{APP_BASE_URL:'http://localhost:99999'},{BILLING_ENABLED:'true',POLAR_ACCESS_TOKEN:'',POLAR_PRODUCT_ID:''},{SMTP_SECURE:'wrong'},{HTTP_PORT:'not-a-port'},{APP_ENV:'production'}]){
  const f=fixture(t);const r=f.run(['--doctor'],{TEST_ENV_OVERRIDES:JSON.stringify(env)});clean(r);assert.equal(r.status,1,r.stdout);
 }
});
test('occupied ports, failed jobs and missing migrations cannot report ready',t=>{
 for(const env of [{TEST_UNBOUND:'1',TEST_PORT_BUSY:'1'},{TEST_FAILED_SERVICE:'auth-migrate'},{TEST_NOT_STARTED:'1'},{TEST_BAD_CHECK:'business_migrations'},{TEST_BAD_CHECK:'membership_roles'},{TEST_STALE_SERVICE:'frontend'}]){
  const f=fixture(t);const r=f.run(['--doctor'],env);clean(r);assert.equal(r.status,1,r.stdout);
 }
});
test('Docker/WSL route uses Linux tools and rejects a native Windows shell',t=>{
 const f=fixture(t);const wsl=f.run(['--doctor'],{TEST_HOST:'Linux',WSL_DISTRO_NAME:'Ubuntu'});clean(wsl);assert.equal(wsl.status,0,wsl.stdout);
 const windows=f.run(['--doctor'],{TEST_HOST:'MINGW64_NT-10.0'});clean(windows);assert.equal(windows.status,1);assert.match(windows.stdout,/WSL2/);
});
test('setup preserves failure exit then recovers without rotating generated secrets',t=>{
 const f=fixture(t,{configured:false});const failed=f.run([],{TEST_START_FAIL:'1'});assert.equal(failed.status,1);assert.match(failed.stderr,/Startup failed/);
 const before=readFileSync(join(f.dir,'.env'));const success=f.run([]);assert.equal(success.status,0,success.stdout+success.stderr);assert.deepEqual(readFileSync(join(f.dir,'.env')),before);
 const doctor=f.run();clean(doctor);assert.equal(doctor.status,0,doctor.stdout+doctor.stderr);
});
test('unknown setup arguments fail without running Docker or writing config',t=>{
 const f=fixture(t,{configured:false});const result=f.run(['--delete-data']);assert.equal(result.status,2);assert.equal(existsSync(join(f.dir,'.env')),false);assert.deepEqual(f.calls(),[]);
});
