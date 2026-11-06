import { test, expect } from '@playwright/test';

// Close on a session row and in the session pane (JOB-111): the confirm
// lists what the session holds (from a dry run), and OK closes it. The
// API is mocked.

const builtins = ['desk', 'nav', 'sessions', 'chat'];
const pathOf = (/** @type {string} */ url) => new URL(url).pathname.replace(/^.*?(?=\/api\/)/, '');
const sess = { id: 'js-s-c1abcd', alive: true, current: 'js-t1', model: 'claude-x', ticket: { pool: 'dev', status: 'open' } };

/** @param {import('@playwright/test').Page} page @param {any[]} calls */
async function mock(page, calls) {
  const plugins = await Promise.all(builtins.map(async (n) => {
    const r = await page.request.get('plugins/' + n + '/plugin.json');
    return { ...(await r.json()), source: 'builtin', url: 'plugins/' + n + '/' };
  }));
  await page.route('**/api/**', async (route) => {
    const path = pathOf(route.request().url());
    const json = (/** @type {any} */ b) => route.fulfill({ json: b });
    if (path === '/api/plugins') return json(plugins);
    if (path === '/api/sessions') return json([sess]);
    if (path === '/api/chat') return json({ pool: 'operator', sessions: [] });
    if (path.endsWith('/access')) return json({ write: false, reason: 'a worker' });
    if (path === '/api/rpc/session.close') {
      const body = route.request().postDataJSON();
      calls.push(body);
      const held = [{ id: 'js-t1', title: 'Fix the thing', kind: 'task', status: 'in_progress', pool: 'dev' }];
      return json(body.dry_run ? { id: body.id, dry_run: true, held } : { id: body.id, held, session: { id: body.id, status: 'closed' } });
    }
    if (path.endsWith('/stream')) return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    if (path.startsWith('/api/session/')) return json({ events: [], first: 1 });
    if (path === '/api/push/key') return route.fulfill({ status: 404, json: { error: 'push off' } });
    if (path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    return json([]);
  });
}

test('a session row closes after a confirm that lists its ticket', async ({ page }) => {
  /** @type {any[]} */
  const calls = [];
  await mock(page, calls);
  await page.goto('./?view=sessions');
  await page.getByTestId('session-close').click();
  const confirm = page.getByTestId('confirm');
  await expect(confirm).toContainText('js-t1 Fix the thing (handed back to dev)');
  expect(calls).toEqual([{ id: sess.id, dry_run: true }]);
  await confirm.getByRole('button', { name: 'Cancel' }).click();
  expect(calls).toHaveLength(1);

  await page.getByTestId('session-close').click();
  await page.getByTestId('confirm').getByRole('button', { name: 'OK' }).click();
  await expect(page.getByTestId('toasts')).toContainText('Session ' + sess.id + ' closed');
  expect(calls.at(-1)).toEqual({ id: sess.id, reason: 'closed in Desk' });
});

test('the session pane has Close', async ({ page }) => {
  /** @type {any[]} */
  const calls = [];
  await mock(page, calls);
  await page.goto('./?view=sessions&session=' + sess.id);
  await page.getByTestId('session-panel').getByTestId('chat-close-session').click();
  await page.getByTestId('confirm').getByRole('button', { name: 'OK' }).click();
  await expect(page.getByTestId('toasts')).toContainText('closed');
  expect(calls.map((c) => !!c.dry_run)).toEqual([true, false]);
});
