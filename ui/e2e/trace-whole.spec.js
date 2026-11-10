import { test, expect } from '@playwright/test';
import { showDetails, showTimeline } from './trace-view.js';

// JOB-190: the trace always shows the whole graph from its top ticket. Opening
// it on a sub-item selects that item, brings it into view and keeps the rows.
// The live stream is mocked like Desk answers it (the graph from the root, the
// requested ticket inspected), so this runs against any Desk and writes nothing.
const at = (minutes) => new Date(Date.now() - minutes * 60_000).toISOString();
const ticket = (id, title, minutes, o = {}) => ({
  id,
  project: 'js',
  kind: 'step',
  title,
  status: 'open',
  priority: 0,
  fence: 1,
  revision: 1,
  created_at: at(minutes),
  updated_at: at(1),
  ...o,
});

// A root run with 40 steps, each created a minute apart: the last ones are far
// below the first rows of the timeline.
const count = 40;
const root = ticket('js-root', 'Root run', 100, { kind: 'run', status: 'in_progress' });
const steps = Array.from({ length: count }, (_, i) =>
  ticket(`js-step${i}`, `Step ${i}`, 99 - i, { parent_id: 'js-root' }),
);
const all = [root, ...steps];
const timeline = {
  root: 'js-root',
  requested: 'js-root',
  truncated: false,
  arrows: [],
  rows: all.map((t, i) => ({
    id: t.id,
    kind: t.kind,
    title: t.title,
    status: t.status,
    depth: i ? 1 : 0,
    parent: i ? 'js-root' : undefined,
    created_at: t.created_at,
    segments: [{ state: 'queued', start: t.created_at }],
    markers: [],
  })),
};
const graph = {
  root: 'js-root',
  requested: 'js-root',
  nodes: all,
  relations: steps.map((t) => ({ from: 'js-root', to: t.id, kind: 'parent' })),
  external: [],
  truncated: false,
};

// What /api/live sends for a trace: the whole graph whichever ticket was
// asked for, and that ticket (or ?inspect) as the inspected one.
async function open(page, query) {
  await page.route('**/api/live**', (route) => {
    const params = new URL(route.request().url()).searchParams;
    const wanted = params.get('inspect') || (params.get('trace') !== 'js-root' ? params.get('trace') : '');
    const inspected = all.find((t) => t.id === wanted);
    const body = {
      ticket: root,
      edges: [],
      comments: [],
      history: [],
      graph: { ...graph, requested: params.get('trace') },
      timeline: { ...timeline, requested: params.get('trace') },
      ...(inspected ? { inspected, inspection: { ticket: inspected, edges: [], comments: [], history: [] } } : {}),
    };
    return route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: `event: trace\ndata: ${JSON.stringify(body)}\n\n`,
    });
  });
  await page.goto('./?view=trace' + query);
}

// A timeline row is as wide as the chart and scrolls sideways: rows are only
// checked vertically.
const inView = async (inner, outer, both = true) => {
  const a = await inner.boundingBox();
  const b = await outer.boundingBox();
  expect(a.y).toBeGreaterThanOrEqual(b.y - 1);
  expect(a.y + a.height).toBeLessThanOrEqual(b.y + b.height + 1);
  if (!both) return;
  expect(a.x).toBeGreaterThanOrEqual(b.x - 1);
  expect(a.x + a.width).toBeLessThanOrEqual(b.x + b.width + 1);
};

test('opening the timeline on a sub-item shows every row with that one selected and in view', async ({ page }) => {
  await open(page, '&ref=js-step35');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl.locator('.tl-row')).toHaveCount(count + 1);
  await expect(tl).toContainText(`${count + 1} tickets`);
  const picked = tl.locator('.tl-row.tl-picked');
  await expect(picked).toHaveCount(1);
  await expect(picked).toContainText('Step 35');
  await expect(picked.getByRole('button', { name: 'Step 35' })).toHaveAttribute('aria-pressed', 'true');
  // The row sits low in a long list: it is scrolled into the chart.
  await inView(picked, tl.locator('.tl-scroll'), false);
  await showDetails(page);
  await expect(page.getByLabel('Ticket details').getByRole('heading', { level: 2 })).toContainText('Step 35');
  await expect(page.getByText('open as graph root')).toHaveCount(0);
  await expect(page.getByRole('button', { name: /graph root/ })).toHaveCount(0);
  await showTimeline(page);
  await inView(picked, tl.locator('.tl-scroll'), false);

  // Selecting another row keeps the same rows and does not rewrite the ref.
  await tl.getByRole('button', { name: 'Step 36' }).click();
  await expect(page).toHaveURL(/ref=js-step35/);
  await expect(page).toHaveURL(/inspect=js-step36/);
  await expect(tl.locator('.tl-row')).toHaveCount(count + 1);
  await expect(tl.locator('.tl-row.tl-picked')).toContainText('Step 36');
  await showDetails(page);
  await expect(page.getByLabel('Ticket details').getByRole('heading', { level: 2 })).toContainText('Step 36');
  await expect(page.getByText('open as graph root')).toHaveCount(0);
});

test('?ref=<root>&inspect=<sub-item> shows the same view as ?ref=<sub-item>', async ({ page }) => {
  await open(page, '&ref=js-root&inspect=js-step35');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl.locator('.tl-row')).toHaveCount(count + 1);
  await expect(tl.locator('.tl-row.tl-picked')).toContainText('Step 35');
  await inView(tl.locator('.tl-row.tl-picked'), tl.locator('.tl-scroll'), false);
  await showDetails(page);
  await expect(page.getByLabel('Ticket details').getByRole('heading', { level: 2 })).toContainText('Step 35');
});

test('selecting a row keeps every row on screen', async ({ page }) => {
  await open(page, '&ref=js-root');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl.locator('.tl-row')).toHaveCount(count + 1);
  await tl.getByRole('button', { name: 'Step 3', exact: true }).click();
  // At no point do the rows go: the count is asserted right after the click and
  // again once the selection has been drawn.
  await expect(tl.locator('.tl-row')).toHaveCount(count + 1);
  await expect(tl.locator('.tl-row.tl-picked')).toContainText('Step 3');
  await showDetails(page);
  await expect(page.getByLabel('Ticket details').getByRole('heading', { level: 2 })).toContainText('Step 3');
});
