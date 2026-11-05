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
  await page.goto('/');
  await expect(page.locator('[data-testid=attention-empty], [data-testid=attention-list]')).toBeVisible();
  await expect(page.getByText('live', { exact: true })).toBeVisible();
  expect(errors).toEqual([]);
});

for (const [view, testid] of [['work', 'runs'], ['sessions', 'sessions'], ['pools', 'triggers'], ['deploys', 'deploys']]) {
  test(`${view} view renders from the URL and the nav`, async ({ page }) => {
    const errors = watchErrors(page);
    await page.goto('/?view=' + view);
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
  await page.goto('/?view=work');
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

test('session journal tail streams', async ({ page }) => {
  await page.goto('/?view=sessions');
  await expect(page.getByTestId('sessions')).toBeVisible();
  const row = page.locator('[data-testid=sessions] tr.click').first();
  test.skip((await row.count()) === 0, 'no sessions on this controller');
  await row.click();
  await expect(page).toHaveURL(/session=/);
  await expect(page.getByTestId('journal')).toBeVisible();
  await expect(page.locator('.chat .head')).toHaveText(/^(live|closed)/);
});

test('no horizontal page overflow', async ({ page }) => {
  for (const view of ['attention', 'work', 'sessions', 'pools', 'deploys']) {
    await page.goto('/?view=' + view);
    // The SSE stream stays open, so networkidle never comes.
    await expect(page.locator('main h1')).toBeVisible();
    await page.waitForTimeout(300);
    const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    expect(over, view).toBeLessThanOrEqual(1);
  }
});

// Artifacts (#508): list, read, ack and next. Needs at least two unread
// artifacts (`jobshop artifact deliver`); acks the first one.
test('artifacts list, read, ack and next', async ({ page }) => {
  const errors = watchErrors(page);
  await page.goto('/?view=artifacts');
  await expect(page.locator('[data-testid=artifacts-empty], [data-testid=artifact-list]')).toBeVisible();
  const rows = page.locator('[data-testid=artifact-list] .item');
  test.skip((await rows.count()) < 2, 'needs two unread artifacts');
  const first = await rows.nth(0).locator('.title').innerText();
  const second = await rows.nth(1).locator('.title').innerText();
  await expect(page.getByTestId('artifact-pane').locator('h2')).toHaveText(first);
  await page.keyboard.press('j');
  await expect(page.getByTestId('artifact-pane').locator('h2')).toHaveText(second);
  await page.keyboard.press('k');
  await expect(page.getByTestId('artifact-pane').locator('h2')).toHaveText(first);
  await page.keyboard.press('a');
  await expect(page.getByTestId('artifact-pane').locator('h2')).toHaveText(second);
  await expect(rows.filter({ hasText: first })).toHaveCount(0);
  expect(errors).toEqual([]);
});
