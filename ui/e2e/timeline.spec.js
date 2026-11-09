import { test, expect } from '@playwright/test';

// The trace timeline (#1103). The live stream is mocked, so this runs against
// any Desk and writes nothing.
const at = (minutes) => new Date(Date.now() - minutes * 60_000).toISOString();
const ticket = (id, title, o = {}) => ({
  id,
  project: 'js',
  kind: 'step',
  title,
  status: 'closed',
  priority: 0,
  fence: 1,
  revision: 1,
  created_at: at(60),
  updated_at: at(10),
  ...o,
});
const seg = (state, from, to, o = {}) => ({ state, start: at(from), ...(to === null ? {} : { end: at(to) }), ...o });

const run = ticket('js-run', 'pr-work for #7', { kind: 'run', status: 'in_progress', closed_at: undefined });
const dev = ticket('js-dev', 'Implement #7', { closed_at: at(20), resolution: 'done', created_at: at(60) });
const review = ticket('js-rev', 'Review #7', { status: 'open', created_at: at(60) });
const ask = ticket('js-ask', 'Question: which base?', { kind: 'task', status: 'open', created_at: at(30) });

const timeline = {
  root: 'js-run',
  truncated: false,
  rows: [
    { id: 'js-run', kind: 'run', title: run.title, status: 'in_progress', depth: 0, created_at: at(60), segments: [seg('working', 60, null)], markers: [] },
    {
      id: 'js-dev',
      kind: 'step',
      title: dev.title,
      status: 'closed',
      resolution: 'done',
      parent: 'js-run',
      depth: 1,
      created_at: at(60),
      started: at(55),
      closed_at: at(20),
      segments: [
        seg('queued', 60, 55),
        seg('working', 55, 45, { actor: 'js-s-1' }),
        seg('retry', 45, 44, { actor: 'js-s-1', detail: 'tests failed' }),
        seg('working', 44, 35, { actor: 'js-s-2' }),
        seg('waiting', 35, 20, {
          detail: 'CI on PR 7',
          ended: 'met',
          wait: [{ type: 'github.checks', label: 'acme/widget PR #7 checks: success', url: 'https://github.com/acme/widget/pull/7', state: 'met' }],
        }),
      ],
      markers: [
        { at: at(36), kind: 'pr', label: 'PR opened', url: 'https://github.com/acme/widget/pull/7' },
        { at: at(44), kind: 'session', label: 'taken over by js-s-2' },
      ],
    },
    {
      id: 'js-rev',
      kind: 'step',
      title: review.title,
      status: 'open',
      parent: 'js-run',
      depth: 1,
      created_at: at(60),
      started: at(19),
      segments: [seg('queued', 60, 19), seg('working', 19, 8, { actor: 'js-s-3' }), seg('human', 8, null, { detail: 'needs-human' }), ],
      markers: [{ at: at(12), kind: 'comment', label: 'Looks fine, one nit' }],
    },
    {
      id: 'js-ask',
      kind: 'task',
      title: ask.title,
      status: 'open',
      parent: 'js-rev',
      depth: 2,
      created_at: at(30),
      segments: [seg('deferred', 30, 15, { until: at(15) }), seg('queued', 15, null)],
      markers: [],
    },
  ],
  arrows: [
    { from: 'js-dev', to: 'js-rev', kind: 'blocks', state: 'conflict' },
    { from: 'js-dev', to: 'js-ask', kind: 'tracks' },
  ],
};

async function open(page, extra = '') {
  const trace = {
    ticket: run,
    edges: [],
    comments: [],
    history: [],
    graph: {
      root: 'js-run',
      nodes: [run, dev, review, ask],
      relations: [{ from: 'js-rev', to: 'js-dev', kind: 'blocks' }],
      external: [],
      truncated: false,
    },
    timeline,
  };
  // Like Desk, an inspected ticket comes back as "inspected".
  await page.route('**/api/live**', (route) => {
    const inspect = new URL(route.request().url()).searchParams.get('inspect');
    const inspected = [run, dev, review, ask].find((t) => t.id === inspect);
    return route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: `event: trace\ndata: ${JSON.stringify(inspected ? { ...trace, inspected, inspection: { ...trace, ticket: inspected } } : trace)}\n\n`,
    });
  });
  await page.goto('./?view=trace&ref=js-run' + extra);
}

test('the graph offers a timeline, the timeline a way back', async ({ page }) => {
  await open(page);
  const graph = page.getByRole('region', { name: 'Ticket dependency graph' });
  await graph.getByRole('button', { name: 'timeline' }).click();
  await expect(page).toHaveURL(/layout=timeline/);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl).toBeVisible();
  await tl.getByRole('button', { name: 'graph' }).click();
  await expect(page).not.toHaveURL(/layout=timeline/);
  await expect(graph).toBeVisible();
});

test('bars show state, retries, waits and conflicts without color alone', async ({ page }) => {
  await open(page, '&layout=timeline');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl).toContainText('4 tickets · 2 arrows');
  const dev = tl.locator('.tl-row').filter({ hasText: 'Implement #7' });
  await expect(dev.locator('.tl-seg')).toHaveCount(5);
  await expect(dev.locator('.tl-seg[data-state="retry"]')).toHaveAttribute('aria-label', /failed attempt · js-s-1 · tests failed/);
  await expect(dev.locator('.tl-seg[data-state="waiting"]')).toHaveAttribute('aria-label', /waiting on an event · CI on PR 7 · ended met/);
  // The blocked review row is outlined and its arrow is drawn red.
  await expect(tl.locator('.tl-row').filter({ hasText: 'Review #7' })).toHaveAttribute('data-conflict', 'true');
  await expect(tl.locator('path.tl-arrow[data-shape="conflict"]')).toHaveCount(1);
  await expect(tl.locator('path.tl-arrow[data-shape="tracks"]')).toHaveCount(1);
  // An open bar ends at the now line.
  await expect(tl.locator('.tl-nowline')).toHaveCount(1);
});

test('collapsing a row hides what is nested under it and the detail lists every segment', async ({ page }) => {
  await open(page, '&layout=timeline');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl.locator('.tl-row')).toHaveCount(4);
  await tl.getByRole('button', { name: 'Collapse js-rev' }).click();
  await expect(tl.locator('.tl-row')).toHaveCount(3);
  await tl.getByRole('button', { name: 'Expand js-rev' }).click();
  await expect(tl.locator('.tl-row')).toHaveCount(4);

  await tl.getByRole('button', { name: 'Implement #7' }).click();
  const detail = tl.getByLabel('Timeline of js-dev');
  await expect(detail.locator('.tl-segments > li')).toHaveCount(5);
  await expect(detail.getByRole('link', { name: /PR #7 checks: success/ })).toHaveAttribute(
    'href',
    'https://github.com/acme/widget/pull/7',
  );
});

test('zoom widens the chart and the page does not overflow', async ({ page }) => {
  await open(page, '&layout=timeline');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  const canvas = tl.locator('.tl-canvas');
  const before = (await canvas.boundingBox()).width;
  await tl.getByRole('button', { name: 'Zoom in' }).click();
  await expect.poll(async () => (await canvas.boundingBox()).width).toBeGreaterThan(before);
  await tl.getByRole('button', { name: 'fit' }).click();
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(0);
});
