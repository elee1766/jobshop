import { test, expect } from '@playwright/test';

// In-progress cards and ticket traces name the session working a step and
// open it read-only (JOB-54). The API is mocked.

const builtins = ['attention', 'chat', 'ticket'];
const pathOf = (/** @type {string} */ url) => new URL(url).pathname.replace(/^.*?(?=\/api\/)/, '');

const worker = { id: 'js-s-w1abcd', display_name: 'dev worker', pool: 'dev', model: 'claude-x', alive: true, busy: true, tool: 'go test' };
const step = { id: 'js-st1', title: 'implement', status: 'in_progress', holder: worker.id, fields: '{"step":"implement"}' };
const run1 = { id: 'js-r1', kind: 'run', title: 'Fix the thing', status: 'in_progress' };
const run2 = { id: 'js-r2', kind: 'run', title: 'Idle run', status: 'open' };

const stream = 'retry: 50\n\n'
  + 'id: 3\nevent: session\ndata: ' + JSON.stringify({ seq: 3, type: 'turn_start', text: 'work on JOB-1', time: new Date().toISOString() }) + '\n\n'
  + 'id: 4\nevent: session\ndata: ' + JSON.stringify({ seq: 4, type: 'text', text: 'running the tests now', time: new Date().toISOString() }) + '\n\n';

/** @param {import('@playwright/test').Page} page */
async function mock(page) {
  const plugins = await Promise.all(builtins.map(async (n) => {
    const r = await page.request.get('plugins/' + n + '/plugin.json');
    return { ...(await r.json()), source: 'builtin', url: 'plugins/' + n + '/' };
  }));
  await page.route('**/api/**', async (route) => {
    const path = pathOf(route.request().url());
    const json = (/** @type {any} */ b) => route.fulfill({ json: b });
    if (path === '/api/plugins') return json(plugins);
    if (path === '/api/attention') return json([]);
    if (path === '/api/work') {
      return json({
        runs: [run1, run2], batches: [],
        detail: { [run1.id]: { steps: [step] }, [run2.id]: { steps: [{ id: 'js-st2', title: 'plan', status: 'open' }] } },
        sessions: { [worker.id]: worker },
      });
    }
    if (path === '/api/chat') return json({ pool: 'operator', sessions: [] });
    if (path === '/api/sessions') return json([{ id: worker.id, alive: true, current: step.id, ticket: { pool: 'dev' } }]);
    // A live worker the admin could write to: read-only must still hide the composer.
    if (path.endsWith('/access')) return json({ write: true });
    if (path === '/api/ticket/' + run1.id) {
      return json({
        ticket: run1, comments: [], run: { run: run1, steps: [step] },
        trace: { id: run1.id, title: run1.title, status: 'in_progress', children: [{ ...step, session: worker }] },
        sessions: { [worker.id]: worker },
      });
    }
    if (path === '/api/session/' + worker.id + '/stream') return route.fulfill({ contentType: 'text/event-stream', body: stream });
    if (path.startsWith('/api/session/')) return json({ events: [], first: 1 });
    if (path === '/api/push/key') return route.fulfill({ status: 404, json: { error: 'push off' } });
    if (path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    return json([]);
  });
}

test('in-progress card names its session and opens it read-only', async ({ page }) => {
  await mock(page);
  await page.goto('./');
  const card = page.locator('[data-testid=progress-card][data-run=js-r1]');
  // The chip shows pool, model and tool; the long display name is only in
  // the tooltip so a task-title name cannot stretch the card.
  await expect(card.getByTestId('holder-pool').first()).toHaveText('dev');
  await expect(card.getByTestId('holder-name')).toHaveCount(0);
  await expect(card.getByTestId('holder').first()).toHaveAttribute('title', /dev worker/);
  await expect(card.getByTestId('holder-tool').first()).toHaveText('go test');
  await expect(page.locator('[data-testid=progress-card][data-run=js-r2]').getByTestId('no-holder')).toHaveText('waiting for a session');

  await card.getByTestId('holder').first().click();
  const panel = page.getByTestId('session-panel');
  await expect(panel.getByTestId('journal')).toContainText('running the tests now');
  await expect(panel.getByTestId('chat-readonly')).toContainText('Read-only');
  await expect(panel.getByTestId('chat-input')).toHaveCount(0);
  await expect(panel.getByTestId('chat-stop')).toHaveCount(0);
});

test('a run trace shows View session on its running step and opens it read-only', async ({ page }) => {
  await mock(page);
  await page.goto('./?ticket=js-r1');
  const node = page.getByTestId('trace-session');
  await expect(node).toContainText('View session');
  await expect(node).toContainText('dev');
  await expect(node.getByTestId('holder')).toHaveAttribute('title', /dev worker/);
  await node.getByTestId('holder').click();
  const panel = page.getByTestId('session-panel');
  await expect(panel.getByTestId('journal')).toContainText('running the tests now');
  await expect(panel.getByTestId('chat-readonly')).toBeVisible();
  await expect(panel.getByTestId('chat-input')).toHaveCount(0);
});
