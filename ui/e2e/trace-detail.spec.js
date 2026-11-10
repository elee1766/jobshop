import { test, expect } from '@playwright/test';

// The trace detail pane (JOB-185): summary, now, tabs and grouped events.
// The live stream is mocked, so this runs against any Desk and writes nothing.
// One clock for the fixture: a segment's end and the history row that wrote it
// share a timestamp on the server.
const base = Date.now();
const at = (minutes) => new Date(base - minutes * 60_000).toISOString();
// The server stamps a PR marker with its comment's own time.
const prAt = at(36);
const ticket = (id, title, o = {}) => ({
  id,
  project: 'js',
  kind: 'step',
  title,
  status: 'closed',
  priority: 0,
  fence: 1,
  revision: 3,
  created_at: at(60),
  updated_at: at(10),
  ...o,
});
const seg = (state, from, to, o = {}) => ({ state, start: at(from), ...(to === null ? {} : { end: at(to) }), ...o });
const turn = (from, to, o = {}) => ({ start: at(from), ...(to === null ? {} : { end: at(to) }), ...o });
const entry = (seq, minutes, actor, op, diff = {}) => ({ seq, ticket_id: 'js-dev', at: at(minutes), actor, op, revision: seq, diff });

const run = ticket('js-run', 'pr-work for #7', {
  kind: 'run',
  status: 'in_progress',
  fields: { workflow: 'pr-work', state: 'running', vars: { issue: '7' }, inputs: { repo: 'acme/widget' } },
});
// A step that failed once, was claimed again, waited for CI and closed.
const dev = ticket('js-dev', 'Implement #7', {
  closed_at: at(20),
  resolution: 'done',
  pool: 'dev',
  parent_id: 'js-run',
  fields: {
    run: 'js-run',
    step: 'implement',
    retries: 3,
    failures: 1,
    command: 'go test ./...',
    output: 'ok\n'.repeat(40),
    structured_output: { pr: 'https://github.com/acme/widget/pull/7' },
  },
  body: 'Implement the widget.\n\nSee the issue.',
});
// An open step in a turn right now.
const live = ticket('js-live', 'Review #7', {
  status: 'in_progress',
  holder: 'js-s-3',
  pool: 'refinery',
  parent_id: 'js-run',
  created_at: at(30),
  fields: { run: 'js-run', step: 'review', retries: 2, failures: 0 },
});
const human = ticket('js-human', 'Question: which base?', { kind: 'task', status: 'open', parent_id: 'js-run', created_at: at(30), labels: ['question'], body: 'Which base branch?' });
const script = ticket('js-script', 'Merge', { status: 'open', created_at: at(20), fields: { run: 'js-run', step: 'merge', type: 'script', script: 'merge.star:go', retries: 1, failures: 0 }, parent_id: 'js-run' });
const batch = ticket('js-batch', 'Batch of issues', { kind: 'batch', status: 'open', created_at: at(50), fields: { key: 'b1', issues: ['JOB-1', 'JOB-2'], state: 'running' } });
const all = [run, dev, live, human, script, batch];

const timeline = {
  root: 'js-run',
  truncated: false,
  arrows: [],
  rows: [
    { id: 'js-run', kind: 'run', title: run.title, status: 'in_progress', depth: 0, created_at: at(60), segments: [seg('working', 60, null)], markers: [] },
    {
      id: 'js-dev',
      kind: 'step',
      title: dev.title,
      pool: 'dev',
      status: 'closed',
      resolution: 'done',
      parent: 'js-run',
      depth: 1,
      created_at: at(60),
      started: at(55),
      closed_at: at(20),
      segments: [
        seg('queued', 60, 55),
        seg('retry', 55, 45, { actor: 'js-s-1', detail: 'tests failed', turns: [turn(54, 50, { stop: 'end_turn' })] }),
        seg('queued', 45, 44),
        seg('working', 44, 30, { actor: 'js-s-2', turns: [turn(43, 35, { stop: 'end_turn' })] }),
        seg('waiting', 30, 20, {
          detail: 'CI on PR 7',
          ended: 'met',
          wait: [{ type: 'github.checks', label: 'acme/widget PR #7 checks: success', url: 'https://github.com/acme/widget/pull/7', state: 'met' }],
        }),
      ],
      markers: [
        { at: prAt, kind: 'pr', label: 'PR opened', url: 'https://github.com/acme/widget/pull/7' },
        { at: at(44), kind: 'session', label: 'taken over by js-s-2' },
        { at: at(40), kind: 'comment', label: 'Rebased on master' },
      ],
    },
    {
      id: 'js-live',
      kind: 'step',
      title: live.title,
      pool: 'refinery',
      status: 'in_progress',
      parent: 'js-run',
      depth: 1,
      created_at: at(30),
      started: at(8),
      segments: [seg('queued', 30, 8), seg('working', 8, null, { actor: 'js-s-3', turns: [turn(7, 5, { stop: 'end_turn' }), turn(2, null)] })],
      markers: [],
    },
    { id: 'js-human', kind: 'task', title: human.title, status: 'open', parent: 'js-run', depth: 1, created_at: at(30), segments: [seg('human', 30, null, { detail: 'question' })], markers: [] },
    { id: 'js-script', kind: 'step', title: script.title, status: 'open', parent: 'js-run', depth: 1, created_at: at(20), segments: [seg('working', 20, 10), seg('deferred', 10, null, { until: new Date(Date.now() + 600_000).toISOString() })], markers: [] },
    { id: 'js-batch', kind: 'batch', title: batch.title, status: 'open', depth: 0, created_at: at(50), partial: true, segments: [seg('unknown', 50, 40), seg('working', 40, null)], markers: [] },
  ],
};

const history = {
  'js-dev': [
    entry(1, 60, 'admin', 'create'),
    entry(2, 55, 'js-s-1', 'claim', { holder: 'js-s-1', fence: 1 }),
    entry(3, 45, 'js-s-1', 'release', { reason: 'tests failed', failures: 1 }),
    entry(4, 44, 'js-s-2', 'claim', { holder: 'js-s-2', fence: 2 }),
    entry(5, 30, 'js-s-2', 'update', { fields: { wait: { reason: 'CI on PR 7' } } }),
    entry(6, 20, 'js-s-2', 'close', { resolution: 'done', note: 'PR merged' }),
  ],
  'js-live': [entry(7, 30, 'admin', 'create'), entry(8, 8, 'js-s-3', 'claim', { holder: 'js-s-3' })],
};
const comments = {
  'js-dev': [
    { id: 1, ticket_id: 'js-dev', author: 'js-s-2', body: 'Rebased on master', created_at: at(40) },
    { id: 2, ticket_id: 'js-dev', author: 'js-s-2', body: 'PR https://github.com/acme/widget/pull/7 head abc1234', created_at: prAt },
  ],
};
const sessions = { 'js-s-3': { id: 'js-s-3', alive: true, busy: true, tool: 'go test ./...' } };

function body(root, inspect, extra = {}) {
  const base = {
    ticket: root,
    edges: [],
    comments: [],
    history: [],
    graph: {
      root: root.id,
      nodes: all,
      relations: [
        { from: 'js-run', to: 'js-dev', kind: 'parent' },
        { from: 'js-run', to: 'js-live', kind: 'parent' },
        { from: 'js-live', to: 'js-dev', kind: 'blocks' },
      ],
      external: [],
      truncated: false,
    },
    timeline: extra.timeline || timeline,
    sessions,
  };
  const picked = all.find((t) => t.id === inspect);
  if (!picked || picked.id === root.id) return base;
  const shown = { ...base, ticket: picked, history: history[picked.id] || [], comments: (extra.comments || comments)[picked.id] || [] };
  return { ...base, inspected: picked, inspection: shown };
}

async function open(page, id, extra = '', opts = {}) {
  await page.route('**/api/live**', (route) => {
    const inspect = new URL(route.request().url()).searchParams.get('inspect');
    return route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: `retry: 100\nevent: trace\ndata: ${JSON.stringify(opts.make ? opts.make(inspect) : body(run, inspect))}\n\n`,
    });
  });
  await page.goto(`./?view=trace&ref=js-run&inspect=${id}${extra}`);
}

const pane = (page) => page.getByLabel('Ticket details', { exact: true });
const facts = (page) => pane(page).getByLabel('Summary', { exact: true });
const tabs = (page) => pane(page).getByRole('tablist', { name: 'Ticket detail' });

test('a retried step: facts, time by state, attempts and the closing line', async ({ page }) => {
  await open(page, 'js-dev');
  const f = facts(page);
  await expect(f).toContainText('started');
  await expect(f.locator('div').filter({ hasText: /^ended/ })).toBeVisible();
  await expect(f.locator('div').filter({ hasText: /^took/ })).toContainText('40m');
  // Two claims, one failed, three allowed.
  await expect(f.locator('div').filter({ hasText: /^attempts/ })).toContainText('2 · failed 1 of 3 allowed');
  await expect(f.locator('div').filter({ hasText: /^outcome/ })).toContainText('done');
  await expect(f.getByRole('link', { name: 'acme/widget PR #7' }).first()).toHaveCount(0);
  const share = pane(page).getByLabel('Time by state');
  await expect(share).toContainText('queued 6m');
  await expect(share).toContainText('failed attempts 10m');
  await expect(share).toContainText('in a turn 8m');
  await expect(share).toContainText('held, idle 6m');
  await expect(share).toContainText('waiting on an event 10m');
  const ended = pane(page).getByLabel('How it ended');
  await expect(ended).toContainText('PR merged');
});

test('summary and now fit above the tabs without scrolling the pane at 1000 px', async ({ page, isMobile }) => {
  test.skip(isMobile, 'a fixed 1000 px viewport');
  await page.setViewportSize({ width: 1000, height: 800 });
  await open(page, 'js-dev');
  const header = await pane(page).locator('.ticket-overview').boundingBox();
  const tablist = await tabs(page).boundingBox();
  // From the title to the tab bar is less than one screen, so the summary
  // is read without scrolling inside the pane.
  expect(tablist.y - header.y).toBeLessThan(420);
  await expect(pane(page).getByLabel('Summary')).toBeVisible();
  await expect(pane(page).getByLabel('How it ended')).toBeVisible();
});

test('events are grouped one entry per span, newest first, with comments and the PR in place', async ({ page }) => {
  await open(page, 'js-dev');
  const events = pane(page).locator('ol.events > li');
  // close, waiting span, PR marker, session marker, comment, working, queued, retry, queued: no raw rows.
  await expect(events.first()).toContainText('closed done');
  await expect(events.first()).toContainText('PR merged');
  await expect(events.nth(1)).toContainText('waiting on an event');
  await expect(events.nth(1)).toContainText('ended met');
  await expect(events.nth(1).getByRole('link', { name: /PR #7 checks: success/ })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/7');
  await expect(events.nth(2)).toContainText('PR opened');
  // The PR comment shows once, as the marker.
  await expect(pane(page).locator('ol.events > li[data-kind="comment"]')).toHaveCount(1);
  await expect(pane(page).locator('ol.events > li[data-kind="comment"]')).toContainText('Rebased on master');
  // The failed claim says who, how it ended and why.
  const retry = pane(page).locator('ol.events > li[data-kind="segment"]').filter({ hasText: 'failed attempt' });
  await expect(retry).toContainText('js-s-1');
  await expect(retry).toContainText('ended: released (tests failed)');
  // The second claim lists its turns behind a disclosure.
  const second = pane(page).locator('ol.events > li').filter({ hasText: 'js-s-2' }).filter({ has: page.locator('.tl-turns') });
  await second.locator('.tl-turns summary').click();
  await expect(second.locator('.tl-turns li')).toHaveCount(1);
  // Watching a session opens it beside the pane (over it on a phone), so it comes last.
  await retry.getByRole('button', { name: /view session/ }).click();
  await expect(page.getByRole('region', { name: 'Read-only session' })).toBeVisible();
  await page.getByRole('region', { name: 'Read-only session' }).getByRole('button', { name: 'dismiss ×', exact: true }).click();
  // No section lists every raw history row by default.
  await expect(pane(page).locator('.ticket-history')).toHaveCount(0);
  await pane(page).getByLabel('every history row').check();
  await expect(pane(page).locator('.ticket-history li')).toHaveCount(6);
  await expect(pane(page).locator('ol.events')).toHaveCount(0);
});

test('tabs are kept in the URL, reachable by keyboard, and an unknown one opens events', async ({ page }) => {
  await open(page, 'js-dev');
  const bar = tabs(page);
  await expect(bar.getByRole('tab')).toHaveText(['events', 'input / result', 'relations 2', 'body', 'comments 2', 'raw']);
  await expect(bar.getByRole('tab', { name: 'events' })).toHaveAttribute('aria-selected', 'true');
  await bar.getByRole('tab', { name: 'events' }).focus();
  await page.keyboard.press('ArrowRight');
  await expect(page).toHaveURL(/tab=io/);
  await expect(bar.getByRole('tab', { name: 'input / result' })).toBeFocused();
  const panel = pane(page).getByRole('tabpanel');
  await expect(panel.getByLabel('input')).toContainText('implement');
  await expect(panel.getByLabel('result')).toContainText('go test ./...');
  // Long output is collapsed behind its first line.
  const output = panel.getByLabel('result').locator('details').first();
  await expect(output).not.toHaveAttribute('open');
  await output.locator('summary').click();
  await expect(output.locator('pre')).toContainText('ok');
  await bar.getByRole('tab', { name: 'input / result' }).focus();
  await page.keyboard.press('End');
  await expect(page).toHaveURL(/tab=raw/);
  await expect(panel.locator('pre')).toContainText('"step": "implement"');
  await page.reload();
  await expect(bar.getByRole('tab', { name: 'raw' })).toHaveAttribute('aria-selected', 'true');
  await page.goto('./?view=trace&ref=js-run&inspect=js-dev&tab=nonsense');
  await expect(bar.getByRole('tab', { name: 'events' })).toHaveAttribute('aria-selected', 'true');
  await bar.getByRole('tab', { name: 'body' }).click();
  await expect(pane(page).getByRole('tabpanel')).toContainText('Implement the widget.');
  await bar.getByRole('tab', { name: 'comments 2' }).click();
  await expect(pane(page).locator('.ticket-comment')).toHaveCount(2);
});

test('relations list each ticket with its status and how long it took', async ({ page }) => {
  await open(page, 'js-dev', '&tab=relations');
  const rel = pane(page).getByLabel('Ticket relationships', { exact: true });
  await expect(rel.getByRole('region', { name: 'parent' })).toContainText('pr-work for #7');
  await expect(rel.getByRole('region', { name: 'parent' })).toContainText('so far');
  // js-live is blocked by js-dev, so it is a ticket js-dev blocks.
  const blocked = rel.getByRole('region', { name: 'blocks', exact: true });
  await expect(blocked).toContainText('Review #7');
  await expect(blocked.locator('.ticket-status')).toHaveAttribute('aria-label', 'in_progress');
  await blocked.getByRole('button', { name: /Review #7/ }).click();
  await expect(page).toHaveURL(/inspect=js-live/);
});

test('an open step in a turn says what it is doing now, with the tool', async ({ page }) => {
  await open(page, 'js-live');
  const now = pane(page).getByLabel('Now', { exact: true });
  await expect(now).toContainText('turn running: go test ./...');
  const holder = facts(page).locator('div').filter({ hasText: /^holder/ });
  await expect(holder).toContainText('js-s-3');
  await expect(holder).toContainText('turn running');
  await expect(facts(page).locator('div').filter({ hasText: /^attempts/ })).toContainText('1 · failed 0 of 2 allowed');
  await expect(facts(page).locator('div').filter({ hasText: /^open for/ })).toBeVisible();
  await holder.getByRole('button', { name: 'js-s-3' }).click();
  await expect(page.getByRole('region', { name: 'Read-only session' })).toBeVisible();
});

test('a human hold, a deferral, a controller-run step and a batch each render', async ({ page }) => {
  await open(page, 'js-human');
  await expect(pane(page).getByLabel('Now', { exact: true })).toContainText('question');
  await pane(page).getByRole('button', { name: 'read the question' }).click();
  await expect(page).toHaveURL(/tab=body/);
  await expect(pane(page).getByRole('tabpanel')).toContainText('Which base branch?');

  await open(page, 'js-script');
  const now = pane(page).getByLabel('Now', { exact: true });
  await expect(now).toContainText('deferred');
  await expect(now).toContainText('until');
  // No session claimed it: no attempts fact and no holder.
  await expect(facts(page).locator('div').filter({ hasText: /^attempts/ })).toHaveCount(0);
  await expect(facts(page).locator('div').filter({ hasText: /^holder/ })).toHaveCount(0);
  await pane(page).getByRole('tab', { name: 'input / result' }).click();
  await expect(pane(page).getByRole('tabpanel').getByLabel('input')).toContainText('merge.star:go');

  await open(page, 'js-batch');
  await expect(pane(page).getByLabel('Now', { exact: true })).toContainText('running');
  // Partial history is said, and what cannot be computed is left out.
  await expect(pane(page).locator('.summary-note')).toContainText('not kept');
  await pane(page).getByRole('tab', { name: 'input / result' }).click();
  await expect(pane(page).getByRole('tabpanel').getByLabel('input')).toContainText('JOB-1, JOB-2');
});

test('a run shows its steps as progress and its inputs', async ({ page }) => {
  await open(page, 'js-run');
  await expect(facts(page).locator('div').filter({ hasText: /^steps/ })).toContainText('1 of');
  await expect(pane(page).getByLabel('Now', { exact: true })).toContainText('running');
  await pane(page).getByRole('tab', { name: 'input / result' }).click();
  await expect(pane(page).getByRole('tabpanel').getByLabel('input')).toContainText('acme/widget');
  await expect(pane(page).getByRole('tabpanel').getByLabel('input')).toContainText('pr-work');
});

test('choosing a bar on the chart opens the events tab on that entry', async ({ page }) => {
  await open(page, 'js-run', '&tab=raw');
  const tl = page.getByRole('region', { name: 'Ticket timeline' });
  await tl.locator('.tl-row').filter({ hasText: 'Implement #7' }).locator('.tl-seg[data-state="retry"]').click();
  await expect(page).toHaveURL(/inspect=js-dev/);
  await expect(page).not.toHaveURL(/tab=/);
  const chosen = pane(page).locator('ol.events > li.chosen');
  await expect(chosen).toHaveCount(1);
  await expect(chosen).toContainText('failed attempt');
  // The detail under the chart is the pane: no second description.
  await expect(page.getByLabel('Timeline of js-dev')).toHaveCount(0);
});

test('the graph layout uses the same pane', async ({ page }) => {
  await open(page, 'js-dev', '&layout=graph');
  await expect(page.getByRole('region', { name: 'Ticket dependency graph' })).toBeVisible();
  await expect(facts(page).locator('div').filter({ hasText: /^attempts/ })).toContainText('failed 1 of 3 allowed');
  await expect(tabs(page).getByRole('tab', { name: 'events' })).toBeVisible();
});

test('tabs do not overflow and no page scrolls sideways', async ({ page }) => {
  for (const width of [390, 700, 1000]) {
    await page.setViewportSize({ width, height: 800 });
    await open(page, 'js-dev');
    const bar = await tabs(page).boundingBox();
    const view = page.viewportSize().width;
    expect(bar.x + bar.width).toBeLessThanOrEqual(view + 1);
    for (const tab of await tabs(page).getByRole('tab').all()) {
      const box = await tab.boundingBox();
      expect(box.x + box.width).toBeLessThanOrEqual(view + 1);
      // A tap target, not a sliver.
      expect(box.height).toBeGreaterThanOrEqual(22);
    }
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
    expect(overflow).toBeLessThanOrEqual(0);
  }
});

test('a live update keeps the chosen tab and the scroll position', async ({ page }) => {
  let late = false;
  const make = (inspect) => {
    // Once the test flips it, the step has one more comment.
    const more = late ? { 'js-dev': [...comments['js-dev'], { id: 3, ticket_id: 'js-dev', author: 'x', body: 'late comment', created_at: at(1) }] } : comments;
    return body(run, inspect, { comments: more });
  };
  await open(page, 'js-dev', '&tab=comments', { make });
  await expect(tabs(page).getByRole('tab', { name: 'comments 2' })).toBeVisible();
  const scroller = page.locator('.work');
  await scroller.evaluate((el) => el.scrollBy(0, 120));
  const before = await scroller.evaluate((el) => el.scrollTop);
  late = true;
  await expect(tabs(page).getByRole('tab', { name: 'comments 3' })).toBeVisible();
  await expect(page).toHaveURL(/tab=comments/);
  await expect(tabs(page).getByRole('tab', { name: 'comments 3' })).toHaveAttribute('aria-selected', 'true');
  expect(await scroller.evaluate((el) => el.scrollTop)).toBe(before);
});
