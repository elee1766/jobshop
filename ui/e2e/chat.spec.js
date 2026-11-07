import { test, expect } from '@playwright/test';

test.skip(!process.env.DESK_CHAT, 'set DESK_CHAT=1 to start a real operator conversation');

test('named chat streams, survives reload, archives and restores', async ({ page, isMobile }) => {
  test.setTimeout(240_000);
  const name = 'desk-check-' + Date.now();
  let id;
  const post = async (action) => {
    const response = await page.request.post('api/chat/manage', { headers: { 'X-Jobshop-Desk': '1' }, data: { action, id } });
    expect(response.ok()).toBeTruthy();
  };
  try {
    await page.goto('./');
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'Chats', exact: true }).click();
    await page.getByRole('button', { name: 'New', exact: true }).click();
    await page.getByRole('textbox', { name: 'Name', exact: true }).fill(name);
    await page.getByRole('textbox', { name: /^to / }).fill('Desk browser check. Reply with PONG and do not use tools.');
    const created = page.waitForResponse((response) => response.url().endsWith('/api/chat/manage') && response.request().method() === 'POST');
    await page.getByRole('textbox', { name: /^to / }).press('Enter');
    const response = await created;
    expect(response.ok()).toBeTruthy();
    id = (await response.json()).id;
    await expect(page.getByRole('tab', { name, exact: true })).toBeVisible({ timeout: 60_000 });
    await expect(page.locator('.transcript .markdown')).toContainText('PONG', { timeout: 180_000 });
    await page.reload();
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'Chats', exact: true }).click();
    await page.getByRole('tab', { name, exact: true }).click();
    await expect(page.locator('.transcript .markdown')).toContainText('PONG');
    await post('archive');
    await page.reload();
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'Chats', exact: true }).click();
    await page.getByRole('button', { name: /^Archived/ }).click();
    await page.getByRole('table', { name: 'Archived chats', exact: true }).getByRole('button', { name, exact: true }).click();
    await expect(page.locator('.transcript .markdown')).toContainText('PONG');
    const restoring = page.waitForResponse((response) => response.url().endsWith('/api/chat/manage') && response.request().method() === 'POST');
    await page.getByRole('button', { name: 'Restore', exact: true }).click();
    expect((await restoring).ok()).toBeTruthy();
    const restored = await (await page.request.get('api/chat')).json();
    expect(restored.sessions.find((chat) => chat.id === id).state).not.toBe('archived');
  } finally {
    if (id) await post('archive');
  }
});
