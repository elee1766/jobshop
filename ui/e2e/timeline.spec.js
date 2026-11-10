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

async function open(page, extra = '', { tl = timeline, sessions } = {}) {
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
    timeline: tl,
    ...(sessions ? { sessions } : {}),
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

test('a trace opens as the timeline, the graph is one click away and back', async ({ page }) => {
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl).toBeVisible();
  const graph = page.getByRole('region', { name: 'Ticket dependency graph' });
  await expect(graph).toHaveCount(0);
  await tl.getByRole('button', { name: 'graph' }).click();
  await expect(page).toHaveURL(/layout=graph/);
  await expect(graph).toBeVisible();
  await graph.getByRole('button', { name: 'timeline' }).click();
  await expect(page).not.toHaveURL(/layout=/);
  await expect(tl).toBeVisible();
});

test('layout=graph opens the graph and layout=timeline still opens the timeline', async ({ page }) => {
  await open(page, '&layout=graph');
  await expect(page.getByRole('region', { name: 'Ticket dependency graph' })).toBeVisible();
  await expect(page.getByRole('region', { name: 'Ticket timeline' })).toHaveCount(0);
  await open(page, '&layout=timeline');
  await expect(page.getByRole('region', { name: 'Ticket timeline' })).toBeVisible();
});

test('bars show state, retries, waits and conflicts without color alone', async ({ page }) => {
  await open(page);
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
  await open(page);
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
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  const canvas = tl.locator('.tl-canvas');
  const before = (await canvas.boundingBox()).width;
  await tl.getByRole('button', { name: 'Zoom in' }).click();
  await expect.poll(async () => (await canvas.boundingBox()).width).toBeGreaterThan(before);
  await tl.getByRole('button', { name: 'fit' }).click();
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(0);
});

// JOB-184: turns inside a claim, and the open claim's live state.
const turn = (from, to, o = {}) => ({ start: at(from), ...(to === null ? {} : { end: at(to) }), ...o });
const turnTimeline = {
  root: 'js-run',
  truncated: false,
  arrows: [],
  rows: [
    { id: 'js-run', kind: 'run', title: run.title, status: 'in_progress', depth: 0, created_at: at(120), segments: [seg('working', 120, null)], markers: [] },
    // A finished claim with two turns: active 14m in 2 turns, held 46m.
    {
      id: 'js-dev',
      kind: 'step',
      title: dev.title,
      status: 'closed',
      resolution: 'done',
      parent: 'js-run',
      depth: 1,
      created_at: at(120),
      closed_at: at(60),
      segments: [
        seg('working', 120, 60, {
          actor: 'js-s-1',
          turns: [turn(115, 105, { stop: 'end_turn' }), turn(90, 86, { stop: 'cancelled' })],
        }),
      ],
      markers: [],
    },
    // A claim with no turn rows: drawn as before, not as idle.
    { id: 'js-old', kind: 'step', title: 'Old claim', status: 'closed', parent: 'js-run', depth: 1, created_at: at(120), closed_at: at(70), segments: [seg('working', 120, 70, { actor: 'js-s-0' })], markers: [] },
    // Open claims, one per live state.
    { id: 'js-run-now', kind: 'step', title: 'Turn running', status: 'in_progress', parent: 'js-run', depth: 1, created_at: at(50), segments: [seg('working', 50, null, { actor: 'js-s-live', turns: [turn(40, 30, { stop: 'end_turn' }), turn(5, null)] })], markers: [] },
    { id: 'js-idle', kind: 'step', title: 'Held idle', status: 'in_progress', parent: 'js-run', depth: 1, created_at: at(50), segments: [seg('working', 50, null, { actor: 'js-s-idle', turns: [turn(40, 30, { stop: 'end_turn' })] })], markers: [] },
    { id: 'js-lost', kind: 'step', title: 'Host gone', status: 'in_progress', parent: 'js-run', depth: 1, created_at: at(50), segments: [seg('working', 50, null, { actor: 'js-s-lost', turns: [turn(45, 44, { stop: 'error' })] })], markers: [] },
  ],
};
const holders = {
  'js-s-live': { id: 'js-s-live', alive: true, busy: true, tool: 'go test ./...' },
  'js-s-idle': { id: 'js-s-idle', alive: true },
  'js-s-lost': { id: 'js-s-lost', host_lost: true },
};

test('a claim shows its turns, totals and a text form of each', async ({ page }) => {
  await open(page, '', { tl: turnTimeline, sessions: holders });
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  const dev = tl.locator('.tl-row').filter({ hasText: 'Implement #7' });
  const bar = dev.locator('.tl-seg');
  await expect(bar).toHaveAttribute('data-turns', '2');
  await expect(bar.locator('.tl-turn')).toHaveCount(2);
  await expect(bar).toHaveAttribute('aria-label', /active 14m in 2 turns · held 46m/);
  // The turn spans sit inside the bar.
  const outer = await bar.boundingBox();
  for (const span of await bar.locator('.tl-turn').all()) {
    const box = await span.boundingBox();
    expect(box.x).toBeGreaterThanOrEqual(outer.x - 1);
    expect(box.x + box.width).toBeLessThanOrEqual(outer.x + outer.width + 1);
  }
  // Held time is told from time in a turn by pattern: the bar and a turn differ.
  const looks = await bar.evaluate((el) => [getComputedStyle(el).backgroundImage, getComputedStyle(el.firstElementChild).backgroundImage]);
  expect(looks[0]).toContain('repeating-linear-gradient');
  expect(looks[1]).toBe('none');
  // A claim with no turn rows is the plain working bar.
  const old = tl.locator('.tl-row').filter({ hasText: 'Old claim' }).locator('.tl-seg');
  await expect(old).not.toHaveAttribute('data-turns', /.*/);
  await expect(old.locator('.tl-turn')).toHaveCount(0);
  expect(await old.evaluate((el) => getComputedStyle(el).backgroundImage)).toBe('none');
  await expect(old).not.toHaveAttribute('aria-label', /held/);
  // The detail under the chart says the same in text, and lists the turns.
  await tl.getByRole('button', { name: 'Implement #7' }).click();
  const detail = tl.getByLabel('Timeline of js-dev');
  await expect(detail).toContainText('active 14m in 2 turns · held 46m');
  await detail.locator('.tl-turns summary').click();
  await expect(detail.locator('.tl-turns li')).toHaveCount(2);
  await expect(detail.locator('.tl-turns li').nth(1)).toContainText('cancelled');
});

test('an open claim shows whether its session runs a turn, is idle or lost its host', async ({ page }) => {
  await open(page, '', { tl: turnTimeline, sessions: holders });
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  const bar = (title) => tl.locator('.tl-row').filter({ hasText: title }).locator('.tl-seg');

  const running = bar('Turn running');
  await expect(running).toHaveAttribute('data-live', 'running');
  await expect(running).toHaveAttribute('aria-label', /turn running: go test/);
  await expect(running.locator('.tl-turn[data-running]')).toHaveCount(1);
  expect(await running.evaluate((el) => getComputedStyle(el).animationName)).toBe('tl-breathe');
  // The ring is the cue that stays when motion is off.
  expect(await running.evaluate((el) => getComputedStyle(el).boxShadow)).not.toBe('none');

  const idle = bar('Held idle');
  await expect(idle).toHaveAttribute('data-live', 'idle');
  await expect(idle).toHaveAttribute('aria-label', /held, idle between turns/);
  expect(await idle.evaluate((el) => getComputedStyle(el).animationName)).toBe('none');

  const lost = bar('Host gone');
  await expect(lost).toHaveAttribute('data-live', 'lost');
  await expect(lost).toHaveAttribute('aria-label', /host not alive, nobody is working this ticket/);
  expect(await lost.evaluate((el) => getComputedStyle(el).animationName)).toBe('none');

  // Three different looks, not three colors of one: pattern or outline differ.
  const looks = await Promise.all(
    [running, idle, lost].map((b) => b.evaluate((el) => getComputedStyle(el).backgroundImage + '|' + getComputedStyle(el).boxShadow + '|' + getComputedStyle(el).borderTopWidth)),
  );
  expect(new Set(looks).size).toBe(3);

  // The legend names the two kinds of working time and the lost host.
  const legend = tl.getByRole('list', { name: 'Segment states' });
  await expect(legend).toContainText('in a turn');
  await expect(legend).toContainText('held, idle');
  await expect(legend).toContainText('host not alive');
});

test('with reduced motion the running bar is still, and keeps its ring', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await open(page, '', { tl: turnTimeline, sessions: holders });
  const running = page.locator('.tl-row').filter({ hasText: 'Turn running' }).locator('.tl-seg');
  await expect(running).toHaveAttribute('data-live', 'running');
  expect(await running.evaluate((el) => getComputedStyle(el).animationName)).toBe('none');
  expect(await running.evaluate((el) => getComputedStyle(el).boxShadow)).not.toBe('none');
});

test('an open claim of a holder Desk cannot describe is drawn as held, not as idle or lost', async ({ page }) => {
  await open(page, '', { tl: turnTimeline, sessions: {} });
  const idle = page.locator('.tl-row').filter({ hasText: 'Held idle' }).locator('.tl-seg');
  await expect(idle).toHaveAttribute('data-live', 'unknown');
  await expect(idle).not.toHaveAttribute('aria-label', /idle between turns|host not alive/);
});
