import { test, expect } from '@playwright/test';

// These checks use a live controller, including under the prefix proxy.
for (const width of [1000, 790, 700, 412, 360]) {
  test(`Desk layout and chat rail at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 });
    const errors = [];
    page.on('pageerror', (error) => errors.push(String(error)));
    page.on('console', (message) => { if (message.type() === 'error') errors.push(message.text()); });
    await page.goto('./');
    await expect(page.getByRole('region', { name: 'Waiting', exact: true })).toBeVisible();
    await expect(page.locator('.status')).toContainText('sys ok');
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
    if (width < 600) {
      await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'Chats', exact: true }).click();
    } else if (width < 760) {
      await page.getByRole('button', { name: 'Chats ▸', exact: true }).click();
    }
    await expect(page.getByRole('button', { name: 'New', exact: true })).toBeVisible();
    await page.getByRole('button', { name: 'Hide', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Chats ▸', exact: true })).toBeVisible();
    await page.getByRole('button', { name: 'Chats ▸', exact: true }).click();
    await expect(page.getByRole('button', { name: 'New', exact: true })).toBeVisible();
    expect(errors).toEqual([]);
  });
}

for (const view of ['trace', 'improvements', 'system']) {
  test(`${view} loads directly without browser errors`, async ({ page }) => {
    const errors = [];
    page.on('pageerror', (error) => errors.push(String(error)));
    await page.goto('./?view=' + view);
    await expect(page.locator('main .tag').first()).toContainText('/' + view);
    await expect(page.locator('.status')).toContainText('sys ok');
    expect(errors).toEqual([]);
  });
}

test('blocked browser storage does not prevent rendering', async ({ page }) => {
  await page.addInitScript(() => Object.defineProperty(window, 'localStorage', { get() { throw new Error('disabled'); } }));
  await page.goto('./');
  await expect(page.getByRole('region', { name: 'Waiting', exact: true })).toBeVisible();
  await expect(page.locator('.status')).toContainText('sys ok');
});
