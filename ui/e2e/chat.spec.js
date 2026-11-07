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
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    await page.getByRole('button', { name: 'new', exact: true }).click();
    await page.getByRole('textbox', { name: 'name', exact: true }).fill(name);
    await page.getByRole('textbox', { name: 'name', exact: true }).press('Enter');
    await page.getByRole('textbox', { name: /^First message to new chat:/ }).fill('Desk browser check. Reply with PONG, the Markdown link [Jobshop issue](https://github.com/elee1766/jobshop/issues/625), and the bare URL https://linear.app/acme/issue/ENG-2. Do not put the links in code spans or code fences. Do not use tools.');
    const created = page.waitForResponse((response) => response.url().endsWith('/api/chat/manage') && response.request().method() === 'POST');
    await page.getByRole('textbox', { name: /^First message to new chat:/ }).press('Enter');
    const response = await created;
    expect(response.ok()).toBeTruthy();
    id = (await response.json()).id;
    await expect(page.getByRole('button', { name: 'Choose chat', exact: true })).toContainText(name);
    await expect(page.locator('.transcript .markdown')).toContainText('PONG', { timeout: 180_000 });
    await expect(page.locator('.transcript .markdown').getByRole('link', { name: 'Jobshop issue', exact: true })).toHaveAttribute('href', 'https://github.com/elee1766/jobshop/issues/625');
    await expect(page.locator('.transcript .markdown a[href="https://linear.app/acme/issue/ENG-2"]')).toBeVisible();
    const composer = page.getByRole('textbox', { name: 'to ' + name, exact: true });
    await composer.fill(`Reply with only this Markdown link: [This session](jobshop://session/${id}). Do not use tools.`);
    await composer.press('Enter');
    await expect(page.locator('.transcript .markdown').getByRole('link', { name: 'This session', exact: true })).toHaveAttribute('href', '?view=trace&ref=' + id, { timeout: 180_000 });
    await page.getByRole('button', { name: 'new', exact: true }).click();
    await page.getByRole('dialog', { name: 'New chat', exact: true }).getByRole('button', { name: 'cancel', exact: true }).click();
    await expect(page.getByRole('textbox', { name: 'to ' + name, exact: true })).toBeVisible();
    await expect(page.locator('.transcript .markdown')).toContainText('PONG');
    await page.reload();
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
    await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill(name);
    await page.getByRole('table', { name: 'Recent chats', exact: true }).getByRole('button', { name, exact: true }).click();
    await expect(page.locator('.transcript .markdown')).toContainText('PONG');
    await post('archive');
    await page.reload();
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
    await page.getByRole('button', { name: /^archived/ }).click();
    await page.getByRole('table', { name: 'Archived chats', exact: true }).getByRole('button', { name, exact: true }).click();
    await expect(page.locator('.transcript .markdown')).toContainText('PONG');
    const restoring = page.waitForResponse((response) => response.url().endsWith('/api/chat/manage') && response.request().method() === 'POST');
    await page.getByRole('button', { name: 'restore', exact: true }).click();
    expect((await restoring).ok()).toBeTruthy();
    const restored = await (await page.request.get('api/chat')).json();
    expect(restored.sessions.find((chat) => chat.id === id).state).not.toBe('archived');
  } finally {
    if (id) await post('archive');
  }
});
