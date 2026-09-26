import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';
export const root = fileURLToPath(new URL('../../', import.meta.url));
export const runFile = resolve(root, 'tests/browser/.run.json');
export function sourceHash() {
  const files = execFileSync('git', ['ls-files', '-z', '--cached', '--others', '--exclude-standard', '--', 'next_b2b_starter', 'go-b2b-starter', 'compose.yaml', 'Caddyfile', 'scripts/test-browser.sh', 'tests/browser'], { cwd: root }).toString().split('\0').filter(Boolean).sort();
  const hash = createHash('sha256');
  for (const file of [...new Set(files)]) hash.update(file).update('\0').update(readFileSync(resolve(root, file))).update('\0');
  return hash.digest('hex');
}
export function fixture() {
  const data = JSON.parse(readFileSync(runFile, 'utf8'));
  if (!process.env.BROWSER_TEST_RUN || process.env.BROWSER_TEST_RUN !== data.runId) throw new Error('Run ./scripts/test-browser.sh to create a disposable fixture.');
  for (const key of ['baseURL', 'mailpitURL', 'setupURL']) {
    const url = new URL(data[key]);
    if (url.protocol !== 'http:' || !['localhost', '127.0.0.1'].includes(url.hostname) || url.username || url.password) throw new Error('Browser checks require loopback HTTP fixtures.');
  }
  if (data.sourceHash !== sourceHash()) throw new Error('Source changed since the fixture build; rebuild before capturing evidence.');
  return data;
}
if (process.argv[2] === 'prepare') {
  writeFileSync(runFile, JSON.stringify({ runId: process.env.BROWSER_TEST_RUN,
    revision: execFileSync('git', ['rev-parse', 'HEAD'], { cwd: root }).toString().trim(),
    sourceHash: sourceHash(), builtAt: new Date().toISOString(),
    billingFixture: process.env.BROWSER_BILLING_FIXTURE === 'true',
    baseURL: process.env.STARTER_URL, mailpitURL: process.env.MAILPIT_URL, setupURL: process.env.SETUP_URL }, null, 2));
}
if (process.argv[2] === 'built') {
  const data = fixture();
  data.frontendImage = process.env.BROWSER_IMAGE_ID;
  if (data.billingFixture) data.billingActions = JSON.parse(process.env.BROWSER_BILLING_ACTIONS || '{}');
  writeFileSync(runFile, JSON.stringify(data, null, 2));
}
