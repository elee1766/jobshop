import { test, expect } from '@playwright/test';

// Desk plugins (#517). Needs Desk started with the fixtures installed:
//   cp -r ui/e2e/plugins/* <root>/desk/plugins/
// and DESK_PLUGINS=1. With them installed, e2e/plugins/deploys replaces
// the built-in deploys view; desk.spec.js's deploys test then skips.
test.skip(!process.env.DESK_PLUGINS, 'set DESK_PLUGINS=1 with ui/e2e/plugins installed');

test('a user plugin view calls api.get and api.call; claim is refused', async ({ page }) => {
  await page.goto('./?view=e2e-test');
  await expect(page.locator('nav a[data-view=e2e-test]')).toHaveText(/E2E plugin/);
  await expect(page.getByTestId('e2e-get')).toHaveText(/^get ok/);
  await expect(page.getByTestId('e2e-call')).toHaveText('call ok version');
  await expect(page.getByTestId('e2e-post')).toHaveText(/claim refused .*not allowed/);
});

test('a plugin that throws shows an error card and Desk keeps working', async ({ page }) => {
  await page.goto('./?view=e2e-throws');
  await expect(page.getByTestId('plugin-error')).toContainText('e2e plugin import failure');
  await page.locator('nav a[data-view=work]').click();
  await expect(page.getByTestId('runs')).toBeVisible();
});

test('a user plugin with a built-in name replaces it', async ({ page }) => {
  await page.goto('./?view=deploys');
  await expect(page.getByTestId('user-deploys')).toBeVisible();
  await expect(page.locator('nav a[data-view=deploys]')).toHaveText(/Deploys \(user\)/);
});

test('the drawer shows a plugin section', async ({ page }) => {
  await page.goto('./?view=work');
  await expect(page.getByTestId('runs')).toBeVisible();
  const row = page.locator('[data-testid=runs] tr.click').first();
  test.skip((await row.count()) === 0, 'no runs on this controller');
  await row.click();
  await expect(page.getByTestId('e2e-drawer')).toHaveText(/^drawer for /);
});

test('plugin listing and RPC policy over HTTP', async ({ request }) => {
  const list = await (await request.get('api/plugins')).json();
  const names = list.map((/** @type {any} */ p) => p.name);
  expect(names).toEqual(expect.arrayContaining(['deploys', 'e2e-test', 'e2e-throws']));
  expect(list.find((/** @type {any} */ p) => p.name === 'deploys').overrides).toBe('builtin');
  const res = await request.post('api/rpc/claim', { data: {}, headers: { 'X-Jobshop-Desk': '1' } });
  expect(res.status()).toBe(403);
  expect((await res.json()).code).toBe('forbidden');
});
