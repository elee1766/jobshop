import { test, expect } from '@playwright/test';

// JOB-126: time axis, compact cards and trace bars. Creates and closes
// tickets, so only on an explicitly opted-in disposable controller.
test.skip(!process.env.DESK_MUTATE, 'set DESK_MUTATE=1 against an isolated controller');

// Card left edges and relative tick labels against the axis scale, read from
// the rendered endpoint marks.
function alignment(page) {
  return page.evaluate(() => {
    const at = (selector) => Date.parse(document.querySelector(selector).dateTime);
    const left = (selector) => document.querySelector(selector).getBoundingClientRect().left;
    const start = at('.graph-stamp-start time');
    const end = at('.graph-stamp-box-end time');
    const startMark = left('.graph-mark-start');
    const pxPerMs = (left('.graph-mark-end') - startMark) / (end - start);
    const x = (t) => startMark + (t - start) * pxPerMs;
    return [
      ...[...document.querySelectorAll('.graph-card')].map(card => card.getBoundingClientRect().left - x(Date.parse(card.querySelector('.graph-meta time').dateTime))),
      ...[...document.querySelectorAll('.graph-stamp-tick')].map(tick => {
        const r = tick.getBoundingClientRect();
        return r.left + r.width / 2 - x(Date.parse(tick.querySelector('time').dateTime));
      }),
    ];
  });
}

test('trace axis aligns with cards, labels endpoints and keeps bars compact', async ({ page, request, baseURL }) => {
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const key = 'axis-' + crypto.randomUUID();
  const open = [];
  try {
    const root = await call('ticket.create', { title: key + ' root' });
    open.push(root.id);
    const child = await call('ticket.create', { title: key + ' child with a title long enough to wrap onto a second line', parent_id: root.id });
    open.push(child.id);
    const done = await call('ticket.create', { title: key + ' done', parent_id: root.id });
    await call('ticket.close', { id: done.id });

    await page.goto('./?view=trace&ref=' + root.id);
    const graph = page.getByRole('region', { name: 'Ticket dependency graph', exact: true });
    await expect(graph.locator('.graph-card')).toHaveCount(3);
    const axis = graph.getByRole('group', { name: 'Creation time axis' });
    await expect(axis.locator('.graph-stamp-start')).toContainText('start ');
    await expect(axis.locator('.graph-stamp-now')).toContainText('now ');

    // Alignment at fit, zoomed and panned, and zoomed out.
    const viewport = graph.getByRole('application', { name: 'Ticket graph navigation', exact: true });
    for (const keys of [[], ['+', '+', '+', 'ArrowRight', 'ArrowRight'], ['-', '-', '-', '-', '-', '-']]) {
      await viewport.focus();
      for (const k of keys) await viewport.press(k);
      for (const delta of await alignment(page)) expect(Math.abs(delta)).toBeLessThanOrEqual(1.5);
    }
    await viewport.press('0');

    // Exact time with zone: keyboard focus, Escape, and tap.
    const tip = page.getByRole('tooltip');
    await axis.locator('.graph-stamp-start').focus();
    await expect(tip).toHaveText(/\d{4}.*(UTC|GMT|[+-]\d)/);
    await page.keyboard.press('Escape');
    await expect(tip).toHaveCount(0);
    await axis.locator('.graph-stamp-now').click();
    await expect(tip).toBeVisible();
    await expect(axis.locator('.graph-stamp-now')).toHaveAttribute('aria-label', /^now .+, .+/);

    // Title first, one status marker with a full label, no visible status word.
    const card = graph.locator(`[data-ticket="${done.id}"]`);
    await expect(card.locator('.graph-select > .graph-title')).toHaveText(done.title);
    await expect(card.locator('.ticket-status')).toHaveText('C');
    await expect(card.locator('.ticket-status')).toHaveAttribute('aria-label', 'closed · done');
    for (const text of await graph.locator('.graph-card').allInnerTexts()) expect(text).not.toMatch(/\b(open|in_progress|closed)\b/);
    for (const box of await graph.locator('.graph-card').evaluateAll(cards => cards.map(c => c.scrollHeight - c.clientHeight))) expect(box).toBeLessThanOrEqual(1);
    await expect(page.locator('.ticket-overview > p').first()).not.toContainText('open');

    // Bars: same height, controls the same height, bounded input, no overflow.
    const bars = await page.evaluate(() => {
      const h = (el) => Math.round(el.getBoundingClientRect().height);
      const bar = document.querySelector('.trace-search-bar');
      return {
        search: h(bar), graph: h(document.querySelector('.graph-heading')),
        controls: [...bar.querySelectorAll('input, select, button')].map(h),
        input: document.querySelector('.trace-search-bar input').getBoundingClientRect().width,
        right: bar.querySelector('form').getBoundingClientRect().right,
        width: innerWidth, over: document.documentElement.scrollWidth - innerWidth,
      };
    });
    expect(bars.search).toBe(bars.graph);
    expect(new Set(bars.controls).size).toBe(1);
    expect(bars.input).toBeLessThanOrEqual(420);
    expect(bars.right).toBeLessThanOrEqual(bars.width);
    expect(bars.over).toBeLessThanOrEqual(1);
    // Search stays a plain form: no dropdown while typing.
    await page.getByRole('textbox', { name: 'Trace reference' }).fill(key);
    await expect(page.getByRole('listbox')).toHaveCount(0);
    await expect(page.getByRole('region', { name: 'Ticket search results' })).toHaveCount(0);

    // A completed trace ends at its last close, not now.
    await call('ticket.close', { id: child.id });
    await call('ticket.close', { id: root.id });
    open.length = 0;
    await page.goto('./?view=trace&ref=' + root.id);
    await expect(axis.locator('.graph-stamp-end')).toContainText('ended ');
    await expect(axis.locator('.graph-stamp-now')).toHaveCount(0);
    for (const delta of await alignment(page)) expect(Math.abs(delta)).toBeLessThanOrEqual(1.5);
  } finally {
    for (const id of open) await call('ticket.close', { id });
  }
});
