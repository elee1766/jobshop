import { test, expect } from '@playwright/test';

// Wait edges in the trace graph (#1091), ported with JOB-150. The live stream
// is mocked, so this runs against any Desk and writes nothing.
const iso = (minutes) => new Date(Date.now() - minutes * 60_000).toISOString();
const ticket = (id, title, o = {}) => ({
  id,
  project: 'js',
  kind: 'task',
  title,
  status: 'open',
  priority: 0,
  fence: 0,
  revision: 1,
  created_at: iso(30),
  updated_at: iso(30),
  ...o,
});

const root = ticket('js-wait', 'parked on review');
const external = [
  { id: 'github.pr:acme/widget#7', type: 'github.pr', label: 'acme/widget#7', url: 'https://github.com/acme/widget/pull/7', at: iso(20) },
  { id: 'linear.state:JOB-9', type: 'linear.state', label: 'JOB-9 done', at: iso(10) },
];
const relations = [
  { from: 'js-wait', to: 'github.pr:acme/widget#7', kind: 'wait', type: 'github.pr', state: 'waiting', reason: 'review requested' },
  { from: 'js-wait', to: 'linear.state:JOB-9', kind: 'wait', type: 'linear.state', state: 'invalidated', reason: 'issue moved' },
];

test('wait edges are drawn with their state and external targets are dashed cards', async ({ page }) => {
  const trace = {
    ticket: root,
    edges: [],
    comments: [],
    history: [],
    graph: { root: 'js-wait', nodes: [root], relations, external, truncated: false },
  };
  await page.route('**/api/live**', (route) =>
    route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: `event: trace\ndata: ${JSON.stringify(trace)}\n\n`,
    }),
  );
  await page.goto('./?view=trace&ref=js-wait');
  const graph = page.getByRole('region', { name: 'Ticket dependency graph' });
  await expect(graph).toContainText('1 tickets · 2 external · 2 relationships');

  const cards = graph.locator('.graph-card.graph-external');
  await expect(cards).toHaveCount(2);
  await expect(cards.first()).toContainText('acme/widget#7');
  await expect(cards.first()).toContainText('external · github.pr');
  // A link only appears when the target has an http(s) URL.
  await expect(cards.first().getByRole('link')).toHaveAttribute('href', 'https://github.com/acme/widget/pull/7');
  await expect(cards.nth(1).getByRole('link')).toHaveCount(0);

  await expect(graph.locator('path.graph-wait[data-wait="waiting"]')).toHaveCount(1);
  await expect(graph.locator('path.graph-wait[data-wait="invalidated"]')).toHaveCount(1);
  await expect(graph.locator('text.graph-wait-label[data-wait="invalidated"]')).toHaveText('invalidated');
  await expect(graph.locator('path.graph-wait[data-wait="waiting"] > title')).toHaveText(
    'js-wait is parked on github.pr:acme/widget#7 (waiting: review requested)',
  );
});
