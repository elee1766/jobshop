import { test, expect } from '@playwright/test';
import { showDetails, showTimeline } from './trace-view.js';

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
        { at: at(44), kind: 'session', label: 'taken over by js-s-2' },
        { at: at(36), kind: 'pr', label: 'PR opened', url: 'https://github.com/acme/widget/pull/7', comment: 4 },
        { at: at(30), kind: 'event', label: 'label added: ci', actor: 'admin', seq: 12 },
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
      markers: [{ at: at(12), kind: 'comment', label: 'Looks fine, one nit', actor: 'alice', comment: 5 }],
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
  // The thread of each ticket: the full text behind a marker.
  const threads = {
    'js-rev': {
      comments: [{ id: 5, ticket_id: 'js-rev', author: 'alice', body: 'Looks fine, one nit\n\nPlease rename `x` to **count**.', created_at: at(12) }],
      history: [],
    },
    'js-dev': {
      comments: [{ id: 4, ticket_id: 'js-dev', author: 'js-s-2', body: 'PR https://github.com/acme/widget/pull/7 head abc1234', created_at: at(36) }],
      history: [{ seq: 12, ticket_id: 'js-dev', at: at(30), actor: 'admin', op: 'label_add', revision: 2, diff: { label: 'ci' } }],
    },
  };
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
      body: `event: trace\ndata: ${JSON.stringify(inspected ? { ...trace, inspected, inspection: { ...trace, ticket: inspected, ...(threads[inspected.id] || {}) } } : trace)}\n\n`,
    });
  });
  await page.goto('./?view=trace&ref=js-run' + extra);
}

test('a trace opens as the timeline and an old layout=graph URL still does', async ({ page }) => {
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl).toBeVisible();
  await expect(tl.getByRole('button', { name: 'graph' })).toHaveCount(0);
  await open(page, '&layout=graph');
  await expect(tl).toBeVisible();
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

test('collapsing a row hides what is nested under it and the pane lists every segment', async ({ page }) => {
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl.locator('.tl-row')).toHaveCount(4);
  await tl.getByRole('button', { name: 'Collapse js-rev' }).click();
  await expect(tl.locator('.tl-row')).toHaveCount(3);
  await tl.getByRole('button', { name: 'Expand js-rev' }).click();
  await expect(tl.locator('.tl-row')).toHaveCount(4);

  await tl.getByRole('button', { name: 'Implement #7' }).click();
  // The detail pane (JOB-185) replaces the chart on request (#1181).
  await showDetails(page);
  const detail = page.getByLabel('Ticket details', { exact: true });
  await expect(detail.locator('ol.events > li[data-kind="segment"]')).toHaveCount(5);
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
  // The detail pane says the same in text, and lists the turns.
  await tl.getByRole('button', { name: 'Implement #7' }).click();
  await showDetails(page);
  const detail = page.getByLabel('Ticket details', { exact: true });
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

  // The legend is gone; the accessible names above carry the states.
  await expect(tl.getByRole('list', { name: 'Segment states' })).toHaveCount(0);
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

test('the chart stays on top and a press shows its selection under it, tabs first (#1183)', async ({ page, isMobile }) => {
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(tl).toBeVisible();
  const lower = page.getByLabel('Ticket details', { exact: true });

  // Zoom, fold and pick a segment: the chart keeps its size and place.
  await tl.getByRole('button', { name: 'Zoom in' }).click();
  await tl.getByRole('button', { name: 'Collapse js-rev' }).click();
  const scroller = tl.locator('.tl-scroll');
  const before = await scroller.boundingBox();
  await tl.locator('.tl-row').filter({ hasText: 'Implement #7' }).locator('.tl-seg[data-state="retry"]').click();
  await expect(page).toHaveURL(/inspect=js-dev/);
  await expect(lower).toBeVisible();
  await expect(lower.getByRole('heading', { level: 2 })).toHaveCount(1);
  // The chosen segment is the one the events tab is on.
  await expect(lower.locator('ol.events > li.chosen')).toContainText('failed attempt');
  // Only the chosen panel is drawn, and the tabs come before it.
  await expect(lower.getByRole('tab', { selected: true })).toHaveText('events');
  await expect(lower.getByRole('tabpanel')).toHaveCount(1);
  const tabs = await lower.getByRole('tablist').boundingBox();
  const panel = await lower.getByRole('tabpanel').boundingBox();
  expect(tabs.y + tabs.height).toBeLessThanOrEqual(panel.y + 1);

  if (isMobile) {
    // A full-screen popup over the page, dismissed to the same chart.
    const popup = page.getByRole('dialog', { name: 'Selected item' });
    const size = await popup.boundingBox();
    expect(size.width).toBeGreaterThanOrEqual(page.viewportSize().width - 1);
    expect(size.height).toBeGreaterThanOrEqual(page.viewportSize().height - 1);
    await page.getByRole('button', { name: 'Close and return to the timeline', exact: true }).click();
    await expect(popup).toHaveCount(0);
  } else {
    const chart = await scroller.boundingBox();
    expect(chart.y + chart.height).toBeLessThanOrEqual(tabs.y + 1);
  }
  await expect(tl).toBeVisible();
  await expect(tl.locator('.tl-row')).toHaveCount(3);
  await expect(tl.getByRole('button', { name: 'Zoom out' })).toBeEnabled();
  await expect(tl.locator('.tl-seg[aria-pressed="true"]')).toHaveCount(1);
  await expect(tl.locator('.tl-row.tl-picked')).toContainText('Implement #7');
  const after = await scroller.boundingBox();
  expect(after.height).toBe(before.height);
});

test("the selected row's parent is shaded and named, not described in a block (#1183)", async ({ page }) => {
  await open(page, '&inspect=js-dev');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  const parent = tl.locator('.tl-row.tl-parent');
  await expect(parent).toHaveCount(1);
  await expect(parent).toContainText('pr-work for #7');
  await expect(parent.locator('.tl-name')).toHaveAttribute('title', /parent of the selected ticket/);
  await expect(tl.locator('.tl-row.tl-picked')).toContainText('Implement #7');
  await expect(page.locator('.ticket-parent')).toHaveCount(0);
  // Clearly weaker than the selection.
  const bg = (row) => row.evaluate((el) => getComputedStyle(el).backgroundColor);
  expect(await bg(parent)).not.toBe(await bg(tl.locator('.tl-row.tl-picked')));
});

test('a deep link to a detail tab opens that tab under the chart', async ({ page, isMobile }) => {
  await open(page, '&inspect=js-dev&tab=raw');
  await expect(page.getByRole('tab', { name: 'raw' })).toHaveAttribute('aria-selected', 'true');
  await showTimeline(page);
  await expect(page.getByRole('region', { name: 'Ticket timeline' })).toBeVisible();
  if (!isMobile) await expect(page.getByRole('tab', { name: 'raw' })).toBeVisible();
});

// #1182: comments and history events sit on the bars at their times. Their
// text is shown in the marker tab under the chart, not drawn in it.
test('markers open their comment or event in the marker tab, with author, time and full text', async ({ page }) => {
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await expect(page.getByRole('region', { name: 'Marker detail' })).toHaveCount(0);
  // No prose in the chart: a marker is a tick with a name, not text.
  await expect(tl.locator('.tl-marker')).toHaveCount(4);
  await expect(tl.locator('.tl-marker').first()).toHaveText('');

  const review = tl.locator('.tl-row').filter({ hasText: 'Review #7' });
  await review.getByRole('button', { name: /comment: Looks fine, one nit/ }).click();
  const panel = page.getByRole('region', { name: 'Marker detail' });
  await expect(page.getByRole('tab', { name: 'marker', selected: true })).toBeVisible();
  await expect(panel).toContainText('alice');
  await expect(panel.locator('time')).toHaveAttribute('datetime', /\d{4}-\d\d-\d\dT/);
  await expect(panel.locator('.markdown')).toContainText('Please rename');
  await expect(panel.locator('.markdown strong')).toHaveText('count');
  await showTimeline(page);
  await expect(review.locator('.tl-marker')).toHaveAttribute('aria-pressed', 'true');

  // A history event shows its label and the raw row; a PR marker its link.
  const dev = tl.locator('.tl-row').filter({ hasText: 'Implement #7' });
  await dev.getByRole('button', { name: /event: label added: ci/ }).click();
  await expect(panel).toContainText('label added: ci');
  await expect(panel.locator('pre')).toContainText('"label": "ci"');
  await showTimeline(page);
  await dev.getByRole('button', { name: /pr: PR opened/ }).click();
  await expect(panel.getByRole('link', { name: /acme\/widget PR #7/ })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/7');

  // Step through the row's markers from the panel, then close it.
  await expect(panel).toContainText('2 of 3');
  await panel.getByRole('button', { name: 'Next marker' }).click();
  await expect(panel).toContainText('3 of 3');
  await expect(panel.getByRole('button', { name: 'Next marker' })).toBeDisabled();
  await panel.getByRole('button', { name: 'Close marker detail' }).click();
  await expect(panel).toHaveCount(0);
  await expect(page.getByRole('tab', { name: 'events', selected: true })).toBeVisible();

  // Phone and desktop alike: the page does not overflow.
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(0);
});

test('a marker is reachable by keyboard', async ({ page }) => {
  await open(page);
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  const marker = tl.locator('.tl-row').filter({ hasText: 'Review #7' }).locator('.tl-marker');
  await marker.focus();
  await page.keyboard.press('Enter');
  await expect(page.getByRole('region', { name: 'Marker detail' })).toContainText('Please rename');
});

test('the trace has no bottom automation invocation line', async ({ page }) => {
  await open(page, '&inspect=js-dev');
  await expect(page.getByText('automation invocation')).toHaveCount(0);
});
