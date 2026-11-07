import { test, expect } from '@playwright/test';

test('chat attachments stay local until sending and can be removed or dismissed', async ({ page, isMobile }) => {
  const writes = [];
  page.on('request', request => {
    if (request.method() === 'POST' && request.url().includes('/api/chat/')) writes.push(request.url());
  });
  await page.goto('./');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
  await page.getByRole('button', { name: 'new', exact: true }).click();
  await page.getByRole('textbox', { name: 'name', exact: true }).fill('attachment draft');
  await page.getByRole('textbox', { name: 'name', exact: true }).press('Enter');
  const attach = page.locator('.composer').getByRole('button', { name: 'attach', exact: true });
  const message = page.locator('#chat-message');
  const attachLeft = await attach.evaluate(button => {
    const text = document.createRange();
    text.selectNodeContents(button);
    return text.getBoundingClientRect().x;
  });
  const messageBox = await message.boundingBox();
  expect(Math.abs(attachLeft - messageBox.x)).toBeLessThanOrEqual(1);
  await page.locator('input[type=file]').setInputFiles({ name: 'notes.md', mimeType: 'text/markdown', buffer: Buffer.from('# Notes\nNot sent yet.') });
  await expect(page.getByRole('list', { name: 'Message attachments' })).toContainText('notes.md');
  await page.getByRole('button', { name: 'Remove attachment notes.md', exact: true }).click();
  await expect(page.getByRole('list', { name: 'Message attachments' })).toHaveCount(0);
  await page.getByRole('button', { name: 'paste', exact: true }).click();
  const dialog = page.getByRole('dialog', { name: 'Paste document', exact: true });
  await dialog.getByRole('textbox', { name: 'paste text to attach as a document', exact: true }).fill('A short block explicitly attached as a document.');
  await dialog.getByRole('button', { name: 'attach document', exact: true }).click();
  await expect(dialog).toBeHidden();
  await expect(page.getByRole('list', { name: 'Message attachments' })).toContainText('pasted-text-1.txt');
  await expect(page.getByRole('textbox', { name: /^First message to new chat:/ })).toHaveValue('');
  await page.getByRole('button', { name: 'Dismiss new chat', exact: true }).click();
  await expect(page.getByRole('list', { name: 'Message attachments' })).toHaveCount(0);
  expect(writes).toEqual([]);
});

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
      await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    } else if (width < 760) {
      await page.getByRole('button', { name: 'chats ▸', exact: true }).click();
    }
    await expect(page.getByRole('button', { name: 'new', exact: true })).toBeVisible();
    await page.getByRole('button', { name: 'hide', exact: true }).click();
    await expect(page.getByRole('button', { name: 'chats ▸', exact: true })).toBeVisible();
    await page.getByRole('button', { name: 'chats ▸', exact: true }).click();
    await expect(page.getByRole('button', { name: 'new', exact: true })).toBeVisible();
    expect(errors).toEqual([]);
  });
}

for (const view of ['trace', 'improvements', 'system']) {
  test(`${view} loads directly without browser errors`, async ({ page }) => {
    const errors = [];
    page.on('pageerror', (error) => errors.push(String(error)));
    await page.goto('./?view=' + view);
    await expect(page.locator('main .tag').first()).toContainText('/' + view);
    await expect(page.locator('.top nav .chosen')).toHaveText({ trace: 'trace', improvements: 'learn', system: 'sys' }[view]);
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

test('New chat floats below New without moving the chat list', async ({ page, isMobile }) => {
  await page.goto('./');
  await expect(page.locator('.status')).toContainText('sys ok');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
  const button = page.getByRole('button', { name: 'new', exact: true });
  const selector = page.getByRole('button', { name: 'Choose chat', exact: true });
  const before = await selector.boundingBox();
  await button.click();
  const popover = page.getByRole('dialog', { name: 'New chat', exact: true });
  await expect(popover).toBeVisible();
  await expect(popover.getByRole('textbox', { name: 'name', exact: true })).toBeFocused();
  const anchor = await button.boundingBox();
  const box = await popover.boundingBox();
  expect(box.height).toBeLessThanOrEqual(38);
  await expect(popover).toHaveCSS('font-size', await page.locator('.chat-tag').evaluate(node => getComputedStyle(node).fontSize));
  expect(box.y).toBeGreaterThanOrEqual(anchor.y + anchor.height);
  expect(box.y - anchor.y - anchor.height).toBeLessThanOrEqual(8);
  expect(box.x).toBeGreaterThanOrEqual(0);
  expect(box.x + box.width).toBeLessThanOrEqual(await page.evaluate(() => innerWidth));
  expect(await selector.boundingBox()).toEqual(before);
  await page.keyboard.press('Escape');
  await expect(popover).toBeHidden();
  await expect(button).toBeFocused();
  await button.click();
  await page.getByText('desk', { exact: true }).click();
  await expect(popover).toBeHidden();
});

test('waiting references open their ticket trace and support Back and direct links', async ({ page }) => {
  await page.goto('./');
  await expect(page.locator('.status')).toContainText('sys ok');
  const items = await (await page.request.get('api/attention')).json();
  const item = items.find(item => ['decision', 'ticket', 'run', 'question', 'wait', 'guard'].includes(item.kind));
  test.skip(!item, 'The live controller has no waiting ticket.');
  const waiting = page.getByRole('region', { name: 'Waiting', exact: true });
  const link = waiting.getByRole('link', { name: `Trace ${item.ref || item.id}`, exact: true });
  const href = await link.getAttribute('href');
  await link.click();
  await expect(page.getByRole('textbox', { name: 'Trace reference', exact: true })).toHaveValue(item.id);
  await expect(page.locator('.ticket-overview > h2')).toContainText(item.id);
  await page.goBack();
  await expect(waiting).toBeVisible();
  await page.goto(href);
  await expect(page.locator('.ticket-overview > h2')).toContainText(item.id);
});

test('named draft clearly creates on first send and can be dismissed without writes', async ({ page, isMobile }) => {
  const writes = [];
  page.on('request', request => {
    if (request.method() === 'POST' && request.url().includes('/api/chat/')) writes.push(request.url());
  });
  await page.goto('./');
  await expect(page.locator('.status')).toContainText('sys ok');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
  const selector = page.getByRole('button', { name: 'Choose chat', exact: true });
  const previous = await selector.textContent();
  await page.getByRole('button', { name: 'new', exact: true }).click();
  await page.getByRole('textbox', { name: 'name', exact: true }).fill('Draft only');
  await page.getByRole('textbox', { name: 'name', exact: true }).press('Enter');
  await expect(page.getByRole('region', { name: 'New chat draft', exact: true })).toContainText('Your first message will create this chat. Nothing has been sent yet.');
  await expect(selector).toContainText('new chat: Draft only');
  await expect(page.getByRole('textbox', { name: 'First message to new chat: Draft only', exact: true })).toBeFocused();
  await page.getByRole('textbox', { name: 'First message to new chat: Draft only', exact: true }).fill('Unsent message');
  await expect(page.getByRole('textbox', { name: 'First message to new chat: Draft only', exact: true })).toHaveValue('Unsent message');
  await page.getByRole('button', { name: 'Dismiss new chat', exact: true }).click();
  await expect(selector).toHaveText(previous);
  await expect(page.getByRole('region', { name: 'New chat draft', exact: true })).toHaveCount(0);
  expect(writes).toEqual([]);
});

test('archived chats form a vertical table with scrolling but no scrollbar chrome', async ({ page, isMobile }) => {
  await page.setViewportSize({ width: isMobile ? 412 : 1000, height: 400 });
  await page.goto('./');
  await expect(page.locator('.status')).toContainText('sys ok');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
  await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
  await page.getByRole('button', { name: /^archived/ }).click();
  const table = page.getByRole('table', { name: 'Archived chats', exact: true });
  await expect(table).toBeVisible();
  const chats = await (await page.request.get('api/chat')).json();
  await expect(table.getByRole('button')).toHaveCount(chats.sessions.filter(chat => chat.state === 'archived').length);
  const expected = chats.sessions.filter(chat => chat.state === 'archived').sort((a, b) => b.updated.localeCompare(a.updated) || a.title.localeCompare(b.title));
  expect(await table.getByRole('button').allTextContents()).toEqual(expected.map(chat => chat.title));
  if (expected.length) {
    const title = expected[0].title;
    await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill(title[0] + title.at(-1));
    await expect(table.getByRole('button', { name: title, exact: true })).toBeVisible();
    await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill(crypto.randomUUID());
    await expect(page.getByText('No matching chats.', { exact: true })).toBeVisible();
    await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill('');
  }
  await expect(page.getByRole('tablist', { name: 'Named chats' })).toHaveCount(0);
  const rows = table.locator('tbody tr');
  const boxes = await rows.evaluateAll(elements => elements.map(element => {
    const rect = element.getBoundingClientRect();
    return { x: rect.x, y: rect.y, bottom: rect.bottom };
  }));
  for (let i = 1; i < boxes.length; i++) {
    expect(boxes[i].x).toBe(boxes[0].x);
    expect(boxes[i].y).toBeGreaterThanOrEqual(boxes[i - 1].bottom);
  }
  const archive = page.locator('.chat-archive');
  expect(await archive.evaluate(element => element.scrollWidth - element.clientWidth)).toBeLessThanOrEqual(1);
  expect(await archive.evaluate(element => getComputedStyle(element).scrollbarWidth)).toBe('none');
  if (await archive.evaluate(element => element.scrollHeight > element.clientHeight)) {
    await archive.hover();
    await page.mouse.wheel(0, 300);
    await expect.poll(() => archive.evaluate(element => element.scrollTop)).toBeGreaterThan(0);
  }
  const first = table.getByRole('button').first();
  if (await first.count()) {
    await first.click();
    await expect(page.getByRole('button', { name: 'restore', exact: true })).toBeVisible();
  } else {
    await expect(page.getByText('No archived chats.', { exact: true })).toBeVisible();
  }
});
