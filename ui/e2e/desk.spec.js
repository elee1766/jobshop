import { test, expect } from '@playwright/test';

// Read-only checks against a live controller: every view renders without
// console errors or CSP violations, and URL state round-trips.

/** @param {import('@playwright/test').Page} page */
function watchErrors(page) {
  /** @type {string[]} */
  const errors = [];
  page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
  page.on('pageerror', (e) => errors.push(String(e)));
  return errors;
}

test('attention is first and shows a list or the empty state', async ({ page }) => {
  const errors = watchErrors(page);
  await page.goto('./');
  await expect(page.locator('[data-testid=attention-empty], [data-testid=attention-list]')).toBeVisible();
  await expect(page.getByText('live', { exact: true })).toBeVisible();
  expect(errors).toEqual([]);
});

const viewChecks = [['work', 'runs'], ['sessions', 'sessions'], ['pools', 'triggers']];
for (const [view, testid] of viewChecks) {
  test(`${view} view renders from the URL and the nav`, async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('./?view=' + view);
    // /api/pools evaluates every trigger and takes ~5s on gasserver (#520).
    await expect(page.getByTestId(testid)).toBeVisible({ timeout: 15_000 });
    await page.locator('nav a[data-view=attention]').click();
    await expect(page).toHaveURL(/\/$/);
    await page.goBack();
    await expect(page).toHaveURL(new RegExp('view=' + view));
    await expect(page.getByTestId(testid)).toBeVisible({ timeout: 15_000 });
    expect(errors).toEqual([]);
  });
}

test('run drawer opens from work, is linkable and closes', async ({ page }) => {
  await page.goto('./?view=work');
  await expect(page.getByTestId('runs')).toBeVisible();
  const row = page.locator('[data-testid=runs] tr.click').first();
  test.skip((await row.count()) === 0, 'no runs on this controller');
  await row.click();
  await expect(page.getByTestId('drawer')).toBeVisible();
  await expect(page).toHaveURL(/ticket=/);
  await expect(page.getByTestId('drawer').getByText('Steps')).toBeVisible();
  const url = page.url();
  await page.keyboard.press('Escape');
  await expect(page.getByTestId('drawer')).toHaveCount(0);
  await page.goto(url);
  await expect(page.getByTestId('drawer')).toBeVisible();
});

test('session chat panel opens from the sessions list and closes', async ({ page, isMobile }) => {
  const errors = watchErrors(page);
  await page.goto('./?view=sessions');
  await expect(page.getByTestId('sessions')).toBeVisible({ timeout: 15_000 });
  const row = page.locator('[data-testid=sessions] tr.click').first();
  test.skip((await row.count()) === 0, 'no sessions on this controller');
  const id = (await row.locator('td').first().locator('a').innerText()).trim();
  await row.locator('td').first().locator('a').click();
  const panel = page.getByTestId('session-panel');
  await expect(panel).toBeVisible();
  await expect(page).toHaveURL(new RegExp('session=' + id));
  await expect(page.getByTestId('journal')).toBeVisible();
  await expect(page.getByTestId('chat-status')).toHaveText(/^\s*(live|closed|connecting|reconnecting)/);
  const noOverflow = () => page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
  expect(await noOverflow()).toBeLessThanOrEqual(1);

  const box = /** @type {{ x: number, y: number, width: number, height: number }} */ (await panel.boundingBox());
  const vp = /** @type {{ width: number, height: number }} */ (page.viewportSize());
  if (isMobile) {
    // Full screen, page behind does not scroll, dismiss is reachable.
    expect(box.width).toBeGreaterThanOrEqual(vp.width - 1);
    expect(box.height).toBeGreaterThanOrEqual(vp.height - 1);
    expect(await page.evaluate(() => getComputedStyle(document.body).overflow)).toBe('hidden');
  } else {
    // Side panel: the list stays visible beside it.
    expect(box.width).toBeLessThan(vp.width / 2);
    await expect(page.getByTestId('sessions')).toBeVisible({ timeout: 15_000 });
  }

  // Reload keeps the panel.
  await page.reload();
  await expect(panel).toBeVisible();

  // Dismiss with the close button.
  await page.getByTestId('session-panel-close').click();
  await expect(panel).toHaveCount(0);
  await expect(page).not.toHaveURL(/session=/);
  expect(await page.evaluate(() => document.body.style.overflow)).toBe('');

  // Escape closes it too, and back reopens it.
  await row.locator('td').first().locator('a').click();
  await expect(panel).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(panel).toHaveCount(0);
  // Close stepped back, so forward reopens it and back dismisses it again.
  await page.goForward();
  await expect(panel).toBeVisible();
  await page.goBack();
  await expect(panel).toHaveCount(0);
  expect(errors).toEqual([]);
});

test('session panel at 360px is full screen without horizontal overflow', async ({ page }) => {
  await page.setViewportSize({ width: 360, height: 740 });
  await page.goto('./?view=chat');
  const link = page.locator('[data-testid=chat-list] a.item').first();
  test.skip((await link.count()) === 0, 'no chat sessions on this controller');
  await link.click();
  const panel = page.getByTestId('session-panel');
  await expect(panel).toBeVisible();
  const box = /** @type {{ width: number, height: number }} */ (await panel.boundingBox());
  expect(box.width).toBe(360);
  expect(box.height).toBe(740);
  expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)).toBeLessThanOrEqual(1);
  expect(await panel.evaluate((el) => el.scrollWidth - el.clientWidth)).toBeLessThanOrEqual(1);
  const close = page.getByTestId('session-panel-close');
  const cb = /** @type {{ width: number, height: number }} */ (await close.boundingBox());
  expect(cb.width).toBeGreaterThanOrEqual(40);
  await close.click();
  await expect(panel).toHaveCount(0);
});

test('a session id in a ticket drawer opens the panel above it', async ({ page }) => {
  const items = await (await page.request.get('api/work')).json();
  const run = items.runs?.[0];
  test.skip(!run, 'no runs on this controller');
  const t = await (await page.request.get('api/ticket/' + run.id)).json();
  const holder = t.ticket.holder;
  test.skip(!holder, 'run has no session holder');
  await page.goto('./?ticket=' + run.id);
  await page.getByTestId('drawer').getByRole('link', { name: holder }).first().click();
  await expect(page.getByTestId('session-panel')).toBeVisible();
  await expect(page.getByTestId('drawer')).toHaveCount(0);
});

test('no horizontal page overflow', async ({ page }) => {
  for (const view of ['attention', 'chat', 'work', 'activity', 'sessions', 'pools']) {
    await page.goto('./?view=' + view);
    // The SSE stream stays open, so networkidle never comes.
    await expect(page.locator('main h1')).toBeVisible();
    await page.waitForTimeout(300);
    const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    expect(over, view).toBeLessThanOrEqual(1);
  }
});
