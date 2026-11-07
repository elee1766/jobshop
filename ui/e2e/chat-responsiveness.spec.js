import { test, expect } from '@playwright/test';

test.skip(!process.env.DESK_MUTATE, 'requires an isolated controller and the ACP test agent');

async function post(request, path, data) {
  const response = await request.post(path, { headers: { 'X-Jobshop-Desk': '1' }, data });
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

async function openChat(page, isMobile, name) {
  await page.goto('./');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
  await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
  await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill(name);
  await page.getByRole('table', { name: 'Recent chats', exact: true }).getByRole('button', { name, exact: true }).click();
}

test('busy chat shows accepted messages immediately and streams working state', async ({ page, request, isMobile }) => {
  const name = 'delivery-' + crypto.randomUUID();
  const session = await post(request, 'api/chat/manage', { action: 'new', name, text: 'hello' });
  try {
    await openChat(page, isMobile, name);
    await expect(page.locator('.transcript .markdown').last()).toContainText('echo: hello');
    await post(request, 'api/chat/send', { key: session.ticket.fields.key, text: 'hold this turn' });
    await expect(page.locator('.transcript .markdown').last()).toContainText('echo: hold this turn');
    await expect.soft(page.locator('.chat-state')).toHaveText('working');
    const message = 'queued while busy ' + crypto.randomUUID();
    await page.locator('#chat-message').fill(message);
    let release;
    let acknowledged;
    const gate = new Promise(resolve => { release = resolve; });
    const persisted = new Promise(resolve => { acknowledged = resolve; });
    await page.route('**/api/chat/send', async route => {
      const response = await route.fetch();
      acknowledged();
      await gate;
      await route.fulfill({ response });
    }, { times: 1 });
    const accepted = page.waitForResponse(response => response.url().endsWith('/api/chat/send'));
    await page.locator('#chat-message').press('Enter');
    const pending = page.locator('.operator').filter({ hasText: message });
    try {
      await persisted;
      await expect(pending).toContainText('sending…');
    } finally {
      release();
    }
    expect((await accepted).ok()).toBeTruthy();
    await expect(pending).toContainText('queued');
    await page.getByRole('toolbar', { name: 'Chat actions' }).getByRole('button', { name: 'stop', exact: true }).click();
    await expect(page.locator('.transcript .markdown').last()).toContainText('echo: ' + message);
    await expect(page.locator('.operator').filter({ hasText: message })).toHaveCount(1);
    await expect(page.locator('.chat-state')).toHaveText('idle');
    await expect(page.locator('.message-delivery')).toHaveCount(0);
  } finally {
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});

test('chat and read-only inspector reach the beginning of journal history', async ({ page, request, isMobile }) => {
  test.setTimeout(90_000);
  const name = 'history-' + crypto.randomUUID();
  const first = 'first visible message ' + crypto.randomUUID();
  const session = await post(request, 'api/chat/manage', { action: 'new', name, text: first });
  try {
    await openChat(page, isMobile, name);
    await expect(page.locator('.transcript .markdown').last()).toContainText('echo: ' + first);
    for (let index = 0; index < 85; index++) {
      const text = 'history turn ' + index;
      await post(request, 'api/chat/send', { key: session.ticket.fields.key, text });
      await expect(page.locator('.transcript .markdown').last()).toContainText('echo: ' + text);
      await expect(page.locator('.transcript .markdown').last()).toBeInViewport();
    }
    async function loadBeginning(region) {
      await expect(region.locator('.markdown').last()).toContainText('echo: history turn 84');
      await expect(region.locator('.markdown').last()).toBeInViewport();
      let release;
      let fetched = false;
      const gate = new Promise(resolve => { release = resolve; });
      await page.route(/\/api\/session\/[^/]+\/events\?/, async route => {
        const response = await route.fetch();
        fetched = true;
        await gate;
        await route.fulfill({ response });
      }, { times: 1 });
      let anchor;
      try {
        await region.evaluate(element => { element.scrollTop = 0; });
        await expect.poll(() => fetched).toBe(true);
        anchor = await region.locator('.transcript-row').first().evaluate(element => ({ key: element.dataset.rowKey, index: Number(element.dataset.index), top: element.getBoundingClientRect().top }));
      } finally {
        release();
      }
      await expect(region.getByText('loading history…', { exact: true })).toHaveCount(0);
      const anchored = region.locator(`[data-row-key="${anchor.key}"]`);
      await expect.poll(async () => Number(await anchored.getAttribute('data-index'))).toBeGreaterThan(anchor.index);
      await expect.poll(async () => Math.abs((await anchored.boundingBox()).y - anchor.top)).toBeLessThan(4);
      await expect.poll(async () => {
        await region.evaluate(element => { element.scrollTop = 0; });
        const beginning = region.getByText('start of conversation', { exact: true });
        return await beginning.count() && await region.locator('.operator').filter({ hasText: first }).count();
      }).toBe(1);
      await expect(region.locator('.operator').filter({ hasText: first })).toHaveCount(1);
      await expect(region.getByRole('button', { name: 'load older', exact: true })).toHaveCount(0);
      expect(await region.locator('.transcript-row').count()).toBeLessThan(60);
      const prompt = region.locator('.raw-prompt').first();
      const rowKey = await prompt.evaluate(element => element.closest('.transcript-row').dataset.rowKey);
      const expanded = region.locator(`[data-row-key="${rowKey}"] .raw-prompt`);
      await prompt.locator('summary').click();
      await expect(expanded).toHaveAttribute('open', '');
      await region.evaluate(element => { element.scrollTop = element.scrollHeight; });
      await expect(expanded).toHaveCount(0);
      await expect(region.locator('.markdown').last()).toBeInViewport();
      await region.evaluate(element => { element.scrollTop = 0; });
      await expect(expanded).toHaveAttribute('open', '');
    }
    const viewport = page.viewportSize();
    await page.setViewportSize({ width: viewport.width, height: 30000 });
    await openChat(page, isMobile, name);
    await expect(page.locator('.transcript').getByText('start of conversation', { exact: true })).toHaveCount(1);
    await expect(page.locator('.operator').filter({ hasText: first })).toHaveCount(1);
    await page.setViewportSize(viewport);
    await openChat(page, isMobile, name);
    await loadBeginning(page.locator('.transcript'));
    await page.goto('./?view=trace&ref=' + session.id);
    await page.getByRole('button', { name: 'Inspect ' + session.id, exact: true }).click();
    await loadBeginning(page.getByRole('region', { name: 'Read-only session' }).locator('.session-output'));
  } finally {
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});
