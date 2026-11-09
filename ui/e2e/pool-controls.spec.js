import { test, expect } from '@playwright/test';

// JOB-178: pool scale and session cordon controls on the system view. They
// change the isolated controller's pool, so only with DESK_MUTATE=1 against
// a root with a ticket pool (run-isolated.mjs defines `worker`).
test.skip(!process.env.DESK_MUTATE, 'set DESK_MUTATE=1 against an isolated controller');

const pool = process.env.DESK_SCALE_POOL || 'worker';

test.describe.configure({ mode: 'serial' });

test('scale a pool and cordon a session', async ({ page, request, baseURL }) => {
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  try {
    await page.goto('./?view=system');
    const row = page.getByRole('row').filter({ has: page.getByRole('cell', { name: pool, exact: true }) });
    await expect(row).toBeVisible();
    // The warm pool keeps one session open; its default is following ready work.
    await expect(row.getByRole('cell').nth(6)).toContainText('1 auto');
    await expect(row.getByRole('button', { name: 'auto' })).toBeDisabled();

    // Out-of-range values cannot be submitted.
    const input = row.getByRole('spinbutton', { name: 'Desired sessions for ' + pool });
    const set = row.getByRole('button', { name: 'set', exact: true });
    await expect(set).toBeDisabled();
    await input.fill('99');
    await expect(set).toBeDisabled();

    // Set a manual target by keyboard.
    await input.fill('2');
    await input.press('Enter');
    await expect(page.locator('.notice')).toHaveText(`${pool}: desired sessions set to 2`);
    await expect(row.getByRole('cell').nth(6)).toContainText('2 manual');
    await expect(row.getByRole('button', { name: 'auto' })).toBeEnabled();
    await expect(input).toHaveValue('');

    // Cordon a session of this pool and bring it back.
    const sessions = page.locator('.rows > p').filter({ has: page.getByRole('button', { name: /^cordon / }) });
    await expect(sessions.first()).toBeVisible();
    const target = sessions.first();
    const cordon = target.getByRole('button', { name: /^cordon / });
    const label = (await cordon.getAttribute('aria-label')).replace(/^cordon /, '');
    await cordon.click();
    await expect(page.locator('.notice')).toHaveText(`${label}: cordoned`);
    const again = page.locator('.rows > p').filter({ hasText: 'cordoned: manual' });
    await expect(again).toHaveCount(1);
    await expect(again.getByRole('button', { name: 'uncordon ' + label })).toBeVisible();
    await expect(row.getByRole('cell').nth(5)).toHaveText('1');
    await again.getByRole('button', { name: 'uncordon ' + label }).click();
    await expect(page.locator('.notice')).toHaveText(`${label}: back in rotation`);
    await expect(page.locator('.rows > p').filter({ hasText: 'cordoned: manual' })).toHaveCount(0);

    // Back to auto.
    await row.getByRole('button', { name: 'auto' }).click();
    await expect(page.locator('.notice')).toHaveText(`${pool}: scaling follows ready work`);
    await expect(row.getByRole('cell').nth(6)).toContainText('auto');

    // Nothing overflows the page, on the narrow project too.
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    expect(overflow).toBeLessThanOrEqual(0);
  } finally {
    await call('pool.scale', { pool, auto: true });
    const sessions = await call('session.list', {});
    for (const s of sessions) if (s.cordon === 'manual') await call('session.cordon', { id: s.id, uncordon: true });
  }
});

test('controls are hidden when Desk does not allow the methods', async ({ page }) => {
  await page.route('**/api/rpc', async (route) => {
    const response = await route.fetch();
    const list = await response.json();
    list.methods = list.methods.filter((m) => m !== 'pool.scale' && m !== 'session.cordon');
    await route.fulfill({ response, json: list });
  });
  await page.goto('./?view=system');
  const row = page.getByRole('row').filter({ has: page.getByRole('cell', { name: pool, exact: true }) });
  await expect(row).toBeVisible();
  await expect(page.getByRole('columnheader', { name: 'cordoned' })).toBeVisible();
  await expect(page.getByRole('columnheader', { name: 'set desired' })).toHaveCount(0);
  await expect(page.getByRole('spinbutton')).toHaveCount(0);
  await expect(page.getByRole('button', { name: /^(un)?cordon / })).toHaveCount(0);
  // Closing a session is not part of this grant and stays.
  await expect(page.getByRole('button', { name: 'close' }).first()).toBeVisible();
});
