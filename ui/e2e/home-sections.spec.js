import { test, expect } from '@playwright/test';

test.skip(!process.env.DESK_MUTATE || !process.env.DESK_LINK_WORKFLOW, 'set DESK_MUTATE=1 and DESK_LINK_WORKFLOW to a disposable workflow accepting repo, pr and issue variables');

// Waiting and in flight are sections of one home page at every width (JOB-130).
test('in flight stays a home section across widths, shortcuts and history', async ({ page, request, baseURL, isMobile }) => {
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const started = [];
  for (const [pr, issue] of [['990', 'JOB-130'], ['991', 'JOB-131']]) {
    const { run } = await call('run.start', { workflow: process.env.DESK_LINK_WORKFLOW, vars: { repo: 'gfx-labs/jobshop', pr, issue } });
    started.push(run.id);
  }
  const wide = { width: 1100, height: 800 };
  const narrow = { width: 412, height: 800 };
  const phoneNav = page.getByRole('navigation', { name: 'Mobile', exact: true });
  const inflight = page.getByRole('region', { name: 'In flight', exact: true });
  const waiting = page.getByRole('region', { name: 'Waiting', exact: true });
  const filter = page.getByRole('textbox', { name: 'Filter inflight', exact: true });
  const visible = async region => region.evaluate(node => {
    const box = node.getBoundingClientRect();
    const work = node.closest('.work').getBoundingClientRect();
    return box.height > 0 && box.top < work.bottom && box.bottom > work.top;
  });
  try {
    await page.setViewportSize(isMobile ? narrow : wide);
    await page.goto('./');
    // Both sections exist on the one home page at this width.
    await expect(waiting.locator('.tag')).toBeVisible();
    await expect(inflight.locator('.tag')).toBeVisible();
    await expect(page.getByRole('region', { name: 'Not moving', exact: true }).locator('.tag')).toBeVisible();
    await expect(inflight.getByRole('row').filter({ hasText: 'jobshop#991' })).toHaveCount(1);
    await filter.fill('jobshop#991');
    await expect(inflight.getByRole('row').filter({ hasText: 'jobshop#990' })).toHaveCount(0);

    // Resize both ways: same page, same filter, no page switch.
    for (const size of [narrow, wide, narrow]) {
      await page.setViewportSize(size);
      await expect(page).toHaveURL(/view=home|^[^?]*$/);
      await expect(filter).toHaveValue('jobshop#991');
      await expect(waiting.locator('.tag')).toBeVisible();
      await expect(inflight.locator('.tag')).toBeVisible();
      await expect(inflight.getByRole('row').filter({ hasText: 'jobshop#991' })).toHaveCount(1);
    }

    // The compact shortcut scrolls to the same home section and records it in the URL.
    await page.setViewportSize(narrow);
    await phoneNav.getByRole('button', { name: 'in flight', exact: true }).click();
    await expect(page).toHaveURL(/#inflight$/);
    await expect.poll(() => visible(inflight)).toBe(true);
    await expect(filter).toHaveValue('jobshop#991');
    await expect(waiting.locator('.tag')).toBeVisible();
    await phoneNav.getByRole('button', { name: /^waiting/ }).click();
    await expect(page).toHaveURL(/#waiting$/);
    await expect.poll(() => visible(waiting)).toBe(true);

    // Back and forward return to the same sections.
    await page.goBack();
    await expect(page).toHaveURL(/#inflight$/);
    await expect.poll(() => visible(inflight)).toBe(true);
    await page.goForward();
    await expect(page).toHaveURL(/#waiting$/);
    await expect.poll(() => visible(waiting)).toBe(true);

    // Direct navigation to the section lands on it at either width, and the
    // desktop width shows the very same section (it is not a mobile-only page).
    await page.goto('./?view=home#inflight');
    await expect.poll(() => visible(inflight)).toBe(true);
    await expect(inflight.getByRole('row').filter({ hasText: 'jobshop#991' })).toHaveCount(1);
    await page.setViewportSize(wide);
    await expect(inflight.getByRole('row').filter({ hasText: 'jobshop#991' })).toBeVisible();
    await expect(waiting).toBeVisible();

    // Another page and back keeps the in-flight filter and the section.
    await page.setViewportSize(narrow);
    await phoneNav.getByRole('button', { name: 'more', exact: true }).click();
    await page.getByRole('navigation', { name: 'Main', exact: true }).getByRole('button', { name: 'sys', exact: true }).click();
    await page.getByRole('navigation', { name: 'Main', exact: true }).getByRole('button', { name: 'home', exact: true }).click();
    await expect(inflight.getByRole('row').filter({ hasText: 'jobshop#991' })).toHaveCount(1);
    await expect(inflight.locator('.tag')).toBeVisible();
  } finally {
    for (const id of started) await call('run.cancel', { id });
  }
});
