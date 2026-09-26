import { test, expect } from '@playwright/test';
import { randomUUID, createHash } from 'node:crypto';
import { readFile, writeFile } from 'node:fs/promises';
import { fixture, sourceHash } from './evidence.mjs';
const run = fixture();

async function capture(page, info, state) {
  await expect(page.locator('main')).toHaveCount(1);
  await expect(page.locator('h1')).toHaveCount(1);
  await expect(page.locator('h1')).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), `${state}: page must reflow`).toBe(true);
  expect(await page.locator('a button, button a, button button').count(), 'No nested interactive controls').toBe(0);
  const unnamedFields = await page.locator('input:not([type=hidden]), select, textarea').evaluateAll(elements => elements.filter(el => !el.labels?.length && !el.getAttribute('aria-label') && !el.getAttribute('aria-labelledby')).map(el => el.outerHTML));
  expect(unnamedFields, 'Inputs need associated accessible labels').toEqual([]);
  // The product has light tokens; dark OS preference must not invent a dark UI.
  expect(await page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(false);
  expect(await page.locator('body').evaluate(el => getComputedStyle(el).backgroundColor)).toBe('rgb(255, 255, 255)');
  expect(sourceHash(), 'Captures must match the built source').toBe(run.sourceHash);
  const path = info.outputPath(`${state}.png`);
  await page.screenshot({ path, fullPage: true, animations: 'disabled' });
  const metadata = { ...run, route: new URL(page.url()).pathname + new URL(page.url()).search,
    state, viewport: page.viewportSize(), requestedColorScheme: info.project.use.colorScheme,
    renderedTheme: 'light', browserVersion: page.context().browser().version(),
    capturedAt: new Date().toISOString(), screenshotSha256: createHash('sha256').update(await readFile(path)).digest('hex') };
  await writeFile(info.outputPath(`${state}.json`), JSON.stringify(metadata, null, 2));
  await info.attach(state, { path, contentType: 'image/png' });
}
async function headers(response) {
  expect(response?.ok()).toBe(true);
  const h = await response.allHeaders();
  expect(h['x-content-type-options']).toBe('nosniff');
  expect(h['x-frame-options']).toBe('DENY');
  expect(h['referrer-policy']).toBe('strict-origin-when-cross-origin');
  expect(h['permissions-policy']).toContain('camera=()');
  expect(h['content-security-policy']).toContain("frame-ancestors 'none'");
  expect(h['content-security-policy']).toContain("object-src 'none'");
  expect(h['content-security-policy']).not.toContain("'unsafe-eval'");
  expect(h['x-powered-by']).toBeUndefined();
}
async function tabTo(page, target) {
  for (let index = 0; index < 25; index++) {
    await page.keyboard.press('Tab');
    if (await target.evaluate(el => el === document.activeElement)) {
      const visible = await target.evaluate(el => {
        const box = el.getBoundingClientRect(); const style = getComputedStyle(el);
        return box.width > 0 && box.height > 0 && box.left >= 0 && box.right <= innerWidth + 1 && (style.outlineStyle !== 'none' || style.boxShadow !== 'none');
      });
      expect(visible, 'Keyboard focus is visible and on screen').toBe(true); return;
    }
  }
  throw new Error('Control was not reachable by keyboard');
}
async function magicLink(request, email) {
  let url;
  await expect.poll(async () => {
    const response = await request.get(`${run.mailpitURL}/api/v1/messages`);
    const list = await response.json();
    for (const entry of list.messages || []) {
      if (!(entry.To || []).some(recipient => recipient.Address === email)) continue;
      const message = await (await request.get(`${run.mailpitURL}/api/v1/message/${entry.ID}`)).json();
      url = (message.Text || message.HTML || '').match(/https?:\/\/[^\s<>"']+\/api\/identity\/magic-link\/verify[^\s<>"']*/)?.[0]?.replaceAll('&amp;', '&');
      if (url) return true;
    }
    return false;
  }, { message: 'Synthetic magic link delivered to local Mailpit', timeout: 20_000 }).toBe(true);
  expect(new URL(url).origin).toBe(run.baseURL);
  return url;
}

test('public routes, labels, keyboard, invalid link and fail-closed setup', async ({ page }, info) => {
  for (const [route, heading, state] of [
    ['/', 'Start with your team.', 'home'],
    ['/auth', 'Welcome back to Your App', 'sign-in'],
    ['/signup', 'Create your account', 'signup-empty'],
    ['/authenticate?error=invalid', 'Magic link is missing or invalid', 'invalid-link'],
    ['/invite', 'Join your workspace', 'missing-invitation'],
  ]) {
    await headers(await page.goto(route));
    await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible();
    await capture(page, info, state);
  }
  await page.goto('/signup');
  await expect(page.getByRole('button', { name: 'Continue' })).toBeDisabled();
  await tabTo(page, page.getByLabel('Full Name'));
  await page.keyboard.type('Browser Reviewer');
  await page.keyboard.press('Tab');
  await expect(page.getByLabel('Email', { exact: true })).toBeFocused();
  await page.keyboard.type('invalid-address');
  await expect(page.getByRole('button', { name: 'Continue' })).toBeDisabled();
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/\/auth/);
  await expect(page.getByLabel('Work email address')).toBeVisible();
  await expect.poll(async () => {
    try { return (await page.request.get(`${run.setupURL}/api/health`)).status(); } catch { return 0; }
  }).toBe(200);
  await headers(await page.goto(`${run.setupURL}/auth`));
  await expect(page.getByRole('heading', { name: 'Connect authentication' })).toBeVisible();
  await expect(page.getByRole('textbox')).toHaveCount(0);
  await capture(page, info, 'auth-unconfigured');
});

test('real signup, workspace, profile, team dialog and keyboard navigation', async ({ page, request }, info) => {
  test.skip(run.billingFixture, 'The separate real journey verifies billing-disabled behavior.');
  const email = `browser-${randomUUID().slice(0, 8)}@example.test`;
  await page.goto('/signup');
  await page.getByLabel('Full Name').fill('Browser Reviewer');
  await page.getByLabel('Email', { exact: true }).fill(email);
  await page.getByRole('button', { name: 'Continue' }).press('Enter');
  await expect(page.getByLabel('Organization Name')).toBeFocused();
  await page.getByLabel('Organization Name').fill('Browser Test Workspace');
  await capture(page, info, 'signup-workspace');
  // Hold only this browser's outgoing signup action to observe its real pending
  // state deterministically; releasing it still exercises the real server/SMTP.
  let release;
  const pending = new Promise(resolve => { release = resolve; });
  await page.route('**/signup', async route => {
    if (route.request().method() === 'POST') await pending;
    await route.continue();
  });
  try {
    await page.getByRole('button', { name: 'Create Account' }).press('Enter');
    await expect(page.getByRole('button', { name: 'Creating...' })).toBeDisabled();
    await capture(page, info, 'signup-pending');
  } finally { release(); }
  await expect(page.getByRole('heading', { name: 'Check your email' })).toBeVisible();
  await capture(page, info, 'signup-email-sent');
  await page.goto(await magicLink(request, email));
  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByRole('heading', { name: 'Welcome to your dashboard' })).toBeVisible();
  await capture(page, info, 'dashboard');
  if (info.project.use.viewport.width < 1024) {
    const trigger = page.getByRole('button', { name: 'Open sidebar' });
    await tabTo(page, trigger); await page.keyboard.press('Enter');
    const dialog = page.getByRole('dialog', { name: 'Workspace navigation' });
    await expect(dialog).toBeVisible();
    for (let index = 0; index < 5; index++) {
      await page.keyboard.press('Tab');
      expect(await dialog.evaluate(el => el.contains(document.activeElement)), 'Mobile navigation traps focus').toBe(true);
    }
    await page.keyboard.press('Escape'); await expect(dialog).toBeHidden(); await expect(trigger).toBeFocused();
    await trigger.press('Enter'); await dialog.getByRole('link', { name: 'Settings' }).press('Enter');
  } else {
    await page.getByRole('navigation', { name: 'Workspace', exact: true }).getByRole('link', { name: 'Settings' }).press('Enter');
  }
  await expect(page.getByRole('heading', { name: 'Workspace settings' })).toBeVisible();
  await expect(page.getByRole('button', { name: /Subscription & billing/ })).toHaveCount(0);
  await capture(page, info, 'settings-billing-disabled');
  await page.getByRole('button', { name: /Account & workspace/ }).press('Enter');
  await expect(page.getByLabel('Display name')).toBeVisible();
  await page.getByLabel('Display name').fill('Updated Browser Reviewer');
  await page.getByRole('button', { name: 'Save name' }).press('Enter');
  await expect(page.getByRole('status').filter({ hasText: 'Profile saved.' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Account menu for Updated Browser Reviewer' })).toBeVisible();
  await capture(page, info, 'profile-saved');
  await page.getByRole('button', { name: 'Back', exact: true }).press('Enter');
  await page.getByRole('button', { name: /Team access/ }).press('Enter');
  await expect(page.getByRole('heading', { name: 'Team roster' })).toBeVisible();
  await expect(page.getByRole('cell', { name: /Updated Browser Reviewer/ })).toBeVisible();
  await capture(page, info, 'team');
  const addMember = page.getByRole('button', { name: 'Add member' });
  await addMember.press('Enter');
  const dialog = page.getByRole('dialog', { name: 'Add a teammate' });
  await expect(dialog.getByLabel('Full name')).toBeFocused();
  await expect(dialog.getByLabel('Work email')).toBeVisible();
  await expect(dialog.getByLabel('Role', { exact: true })).toBeVisible();
  await capture(page, info, 'invite-dialog');
  await page.keyboard.press('Escape'); await expect(dialog).toBeHidden();
  await expect(addMember).toBeFocused();
});

test('billing UI fixtures: checkout, outage, cancellation and revoked access', async ({ page, request }, info) => {
  test.skip(!run.billingFixture, 'Run with BROWSER_BILLING_FIXTURE=true for provider-free billing UI checks.');
  for (const name of ['getSubscriptionStatus', 'getProducts', 'createCheckout', 'openBillingPortal']) {
    expect(Object.values(run.billingActions)).toContain(name);
  }
  const empty = { isAuthenticated: true, isActive: false, canStartCheckout: true,
    backendAvailable: true, productId: null, planId: null, subscription: null };
  const active = { ...empty, isActive: true, canStartCheckout: false,
    subscription: { id: 'fixture-subscription', status: 'active', productId: 'fixture-product',
      productName: null, currentPeriodEnd: '2099-01-01T00:00:00Z', cancelAtPeriodEnd: false } };
  let state = empty;
  const actions = [];
  // Only billing actions are mocked. Signup, session, membership and business
  // profile use the real disposable application. The Docker network is internal
  // so a missing interception cannot reach a provider with fixture credentials.
  await page.route('**/dashboard/settings*', async route => {
    const name = run.billingActions[route.request().headers()['next-action']];
    if (!name) return route.continue();
    actions.push(name);
    const statusResult = state.reason === 'INSUFFICIENT_PERMISSIONS'
      ? { success: false, error: 'You cannot view subscription details.' }
      : { success: true, data: state };
    const result = name === 'getSubscriptionStatus' ? statusResult
      : name === 'getProducts' ? { success: true, data: [{ id: 'fixture-product', productId: 'fixture-product',
        name: 'Workspace plan', description: 'Synthetic browser fixture', price: 12, currency: 'usd', interval: 'month' }] }
      : { success: false, error: 'Synthetic billing provider is unavailable.' };
    await route.fulfill({ status: 200, contentType: 'text/x-component',
      body: `0:{"a":"$@1","f":[],"b":""}\n1:${JSON.stringify(result)}\n` });
  });
  const email = `billing-ui-${randomUUID().slice(0, 8)}@example.test`;
  await page.goto('/signup');
  await page.getByLabel('Full Name').fill('Billing Reviewer');
  await page.getByLabel('Email', { exact: true }).fill(email);
  await page.getByRole('button', { name: 'Continue' }).press('Enter');
  await page.getByLabel('Organization Name').fill('Billing Fixture Workspace');
  await page.getByRole('button', { name: 'Create Account' }).press('Enter');
  await expect(page.getByRole('heading', { name: 'Check your email' })).toBeVisible();
  await page.goto(await magicLink(request, email));
  await expect(page).toHaveURL(/\/dashboard$/);
  await headers(await page.goto('/dashboard/settings?view=subscription'));
  await expect(page.getByRole('heading', { name: 'No active subscription', exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Subscribe', exact: true })).toBeVisible();
  await expect(page.getByText('This starter supports one subscription plan. Plan switching is not available.')).toBeVisible();
  await capture(page, info, 'billing-fixture-no-subscription');
  await tabTo(page, page.getByRole('button', { name: 'Open billing portal', exact: true }));
  await page.keyboard.press('Enter');
  await expect(page.getByRole('alert').filter({ hasText: 'Synthetic billing provider is unavailable.' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Open billing portal', exact: true })).toBeEnabled();
  await page.getByRole('button', { name: 'Subscribe', exact: true }).press('Enter');
  await expect(page.getByRole('button', { name: 'Subscribe', exact: true })).toBeEnabled();
  await capture(page, info, 'billing-fixture-checkout-error');

  async function refresh(next, title, captureName) {
    state = next;
    await page.getByRole('button', { name: 'Refresh status', exact: true }).press('Enter');
    await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Subscribe', exact: true })).toHaveCount(0);
    await capture(page, info, captureName);
  }
  await refresh({ ...empty, canStartCheckout: false, backendAvailable: false,
    backendError: 'Billing status is temporarily unavailable. Please retry.' },
  'Billing status unavailable', 'billing-fixture-outage-hides-cached-plan');
  await refresh(active, 'Active subscription', 'billing-fixture-active');
  await refresh({ ...active, subscription: { ...active.subscription, cancelAtPeriodEnd: true } },
    'Cancellation scheduled', 'billing-fixture-cancellation');
  await refresh({ ...empty, canStartCheckout: false, reason: 'INSUFFICIENT_PERMISSIONS' },
    'Billing status unavailable', 'billing-fixture-revoked');
  await expect(page.getByRole('alert').filter({ hasText: 'You cannot view subscription details.' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Open billing portal', exact: true })).toBeDisabled();
  expect(actions.filter(name => name === 'createCheckout')).toHaveLength(1);
  expect(actions.filter(name => name === 'openBillingPortal')).toHaveLength(1);
});
