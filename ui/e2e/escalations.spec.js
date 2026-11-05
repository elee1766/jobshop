import { test, expect } from '@playwright/test';

// Escalation sessions (#379). The API is mocked, so this runs against any
// Desk build and needs no controller state.

const now = Date.now();
const iso = (/** @type {number} */ msAgo) => new Date(now - msAgo).toISOString();

const escalations = [
  { ticket: 'js-e1', title: 'Escalate run js-r1', cause: 'run', why: 'Run js-r1 (workflow pr-work) needs a human', held: 'js-s1', key: 'escalation:js-e1', session: 'js-sess1', open: true, since: iso(3 * 3600e3) },
  { ticket: 'js-e2', title: 'Wait past timeout', cause: 'wait', why: 'wait for review', held: 'js-e2', key: 'escalation:js-e2', session: 'js-sess2', open: true, since: iso(600e3) },
  { ticket: 'js-e0', title: 'Old hold', cause: 'label', why: 'needs the staging key', key: 'escalation:js-e0', session: 'js-sess0', open: false, resolution: 'done', since: iso(48 * 3600e3), closed: iso(24 * 3600e3) },
];

/** @param {import('@playwright/test').Page} page */
async function mock(page) {
  await page.route('**/api/**', async (route) => {
    // Strip any Desk prefix (DESK_PREFIX runs under /desk/).
    const path = new URL(route.request().url()).pathname.replace(/^.*?(?=\/api\/)/, '');
    /** @param {any} body */
    const json = (body) => route.fulfill({ json: body });
    if (path === '/api/escalations') return json(escalations);
    if (path === '/api/attention') return json([{ kind: 'ticket', id: 'js-e1', title: 'Escalate run js-r1', why: 'escalation', since: iso(3 * 3600e3), chat: 'escalation:js-e1' }]);
    if (path === '/api/chat') return json({ pool: 'operator', sessions: [
      { id: 'js-sess1', key: 'escalation:js-e1', title: 'Escalation js-e1', alive: true },
      { id: 'js-d1', key: 'desk:abc', title: 'x', alive: true },
    ] });
    if (path === '/api/sessions') return json([]);
    if (path.endsWith('/access')) return json({ write: true });
    if (path === '/api/session/js-sess1/stream') {
      const ev = { type: 'turn_start', text: '', context: 'You are the escalation chat for js-e1', time: iso(1000) };
      return route.fulfill({ contentType: 'text/event-stream', body: 'event: session\ndata: ' + JSON.stringify(ev) + '\n\nevent: end\ndata: {}\n\n' });
    }
    if (path.endsWith('/stream') || path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'event: end\ndata: {}\n\n' });
    return json([]);
  });
}

test('chat view lists escalation sessions apart from other chats', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=chat');
  const table = page.getByTestId('escalations');
  await expect(table).toBeVisible();
  const rows = table.getByTestId('escalation-row');
  await expect(rows).toHaveCount(3);
  // Open ones first, oldest first, closed last.
  await expect(rows.nth(0)).toContainText('Escalate run js-r1');
  await expect(rows.nth(0)).toContainText('run failed');
  await expect(rows.nth(0)).toContainText('open');
  await expect(rows.nth(1)).toContainText('Wait past timeout');
  await expect(rows.nth(1)).toContainText('wait timed out');
  await expect(rows.nth(2)).toContainText('closed · done');
  // The escalation chat is not repeated in the operator chat list.
  const list = page.getByTestId('chat-list');
  await expect(list.locator('a.item')).toHaveCount(1);
  await expect(list).not.toContainText('Escalation js-e1');
  expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)).toBeLessThanOrEqual(1);
});

test('clicking an escalation opens its session panel', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=chat');
  await page.getByTestId('escalation-row').nth(1).click();
  await expect(page.getByTestId('session-panel')).toBeVisible();
  await expect(page).toHaveURL(/session=js-sess2/);
  await page.getByTestId('session-panel-close').click();
  await expect(page.getByTestId('session-panel')).toHaveCount(0);
});

test('an escalation ticket in attention opens its chat, not the drawer', async ({ page }) => {
  await mock(page);
  await page.goto('./');
  await page.locator('[data-testid=attention-list] tr.click').first().click();
  await expect(page.getByTestId('session-panel')).toBeVisible();
  await expect(page).toHaveURL(/session=js-sess1/);
  await expect(page.getByTestId('drawer')).toHaveCount(0);
  // The pack prompt is context, not a visible operator message.
  await expect(page.locator('.line.turn.me')).toHaveCount(0);
  const ctx = page.getByTestId('chat-context');
  await expect(ctx).toBeVisible();
  await ctx.locator('summary').click();
  await expect(ctx).toContainText('escalation chat for js-e1');
});
