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
      const text = 'history turn ' + index + (index === 84 ? ' hidden journal tail' : '');
      await post(request, 'api/chat/send', { key: session.ticket.fields.key, text });
      await expect(page.locator('.transcript .markdown').last()).toContainText('echo: ' + text);
      await expect(page.locator('.transcript .markdown').last()).toBeInViewport();
    }
    await expect(page.locator('.chat-state')).toHaveText('idle');
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
    await page.goto('./?view=trace&layout=graph&ref=' + session.id);
    await page.getByRole('button', { name: 'Inspect ' + session.id, exact: true }).click();
    await loadBeginning(page.getByRole('region', { name: 'Read-only session' }).locator('.session-output'));
  } finally {
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});

test('consecutive tool calls sit tighter than separate messages', async ({ page, request, isMobile }) => {
  test.setTimeout(45_000);
  const name = 'tools-' + crypto.randomUUID();
  // "tool run" makes the fake agent reply with four tool calls, then an echo.
  const session = await post(request, 'api/chat/manage', { action: 'new', name, text: 'tool run one' });
  try {
    async function check(region) {
      await expect(region.locator('.markdown').last()).toContainText('echo: tool run two');
      const tools = region.locator('details.tool');
      await expect(tools).toHaveCount(8);
      const boxes = async () => Promise.all((await tools.all()).map(tool => tool.boundingBox()));
      const gaps = list => list.slice(1).map((box, i) => box.y - (list[i].y + list[i].height));
      const rows = await boxes();
      // Inside a run (tools 1-4, 5-8) rows touch; between a run and a message there is a gap.
      for (const [i, gap] of gaps(rows).entries()) {
        if (i === 3) continue;
        expect(Math.abs(gap), 'gap after tool ' + i).toBeLessThan(1);
      }
      const replyBox = await region.locator('.markdown').filter({ hasText: 'echo: tool run one' }).boundingBox();
      expect(replyBox.y - (rows[3].y + rows[3].height)).toBeGreaterThanOrEqual(7);
      expect(rows[0].y - (await region.locator('.operator').filter({ hasText: 'tool run one' }).boundingBox()).y).toBeGreaterThan(0);
      // Hit targets stay usable: the summary is at least one text line tall, taller on touch.
      const summary = tools.nth(5).locator('summary');
      const summaryBox = await summary.boundingBox();
      expect(summaryBox.height).toBeGreaterThanOrEqual(isMobile ? 26 : 20);
      await tools.nth(4).locator('summary').focus();
      await page.keyboard.press('Tab');
      await expect(summary).toBeFocused();
      await expect(summary).toHaveCSS('outline-style', 'solid');
      // Expanding remeasures the row: the next tool moves below the output, not over it.
      await summary.click();
      await expect(tools.nth(5)).toHaveAttribute('open', '');
      await expect(tools.nth(5).locator('pre')).toContainText('echo step 2');
      await expect.poll(async () => {
        const [output, next] = [await tools.nth(5).locator('pre').boundingBox(), await tools.nth(6).boundingBox()];
        return Math.round(next.y - (output.y + output.height));
      }).toBe(4);
      await summary.click();
      await expect(tools.nth(5)).not.toHaveAttribute('open', '');
    }
    await openChat(page, isMobile, name);
    // A second turn only after the first ends, or both share one turn.
    await expect(page.locator('.transcript .markdown').last()).toContainText('echo: tool run one');
    await post(request, 'api/chat/send', { key: session.ticket.fields.key, text: 'tool run two' });
    await check(page.locator('.transcript'));
    await page.goto('./?view=trace&layout=graph&ref=' + session.id);
    await page.getByRole('button', { name: 'Inspect ' + session.id, exact: true }).click();
    await check(page.getByRole('region', { name: 'Read-only session' }).locator('.session-output'));
  } finally {
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});
