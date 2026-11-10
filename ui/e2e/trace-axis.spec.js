import { test, expect } from '@playwright/test';
import { showDetails } from './trace-view.js';

// JOB-126: trace bars. Creates and closes tickets, so only on an explicitly
// opted-in disposable controller.
test.skip(!process.env.DESK_MUTATE, 'set DESK_MUTATE=1 against an isolated controller');

test('trace bars share a height, keep bounded input and do not overflow', async ({ page, request, baseURL }) => {
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
    const child = await call('ticket.create', { title: key + ' child', parent_id: root.id });
    open.push(child.id);

    await page.goto('./?view=trace&ref=' + root.id);
    const timeline = page.getByRole('region', { name: 'Ticket timeline', exact: true });
    await expect(timeline.locator('.tl-row')).toHaveCount(2);

    // Bars: same height, controls the same height, bounded input, no overflow.
    const bars = await page.evaluate(() => {
      const h = (el) => Math.round(el.getBoundingClientRect().height);
      const bar = document.querySelector('.trace-search-bar');
      return {
        search: h(bar), heading: h(document.querySelector('.graph-heading')),
        controls: [...bar.querySelectorAll('input, select, button')].map(h),
        input: document.querySelector('.trace-search-bar input').getBoundingClientRect().width,
        right: bar.querySelector('form').getBoundingClientRect().right,
        width: innerWidth, over: document.documentElement.scrollWidth - innerWidth,
      };
    });
    expect(bars.search).toBe(bars.heading);
    expect(new Set(bars.controls).size).toBe(1);
    expect(bars.input).toBeLessThanOrEqual(420);
    expect(bars.right).toBeLessThanOrEqual(bars.width);
    expect(bars.over).toBeLessThanOrEqual(1);
    await showDetails(page);
    await expect(page.locator('.ticket-overview > p').first()).not.toContainText('open');
    // Search stays a plain form: no dropdown while typing.
    await page.getByRole('textbox', { name: 'Trace reference' }).fill(key);
    await expect(page.getByRole('listbox')).toHaveCount(0);
    await expect(page.getByRole('region', { name: 'Ticket search results' })).toHaveCount(0);
  } finally {
    for (const id of open) await call('ticket.close', { id });
  }
});
