import { test, expect } from '@playwright/test';

test('running ticket references remain clickable in desktop and compact layouts', async ({ page, request, baseURL, isMobile }) => {
  test.skip(!process.env.DESK_MUTATE || !process.env.DESK_LINK_WORKFLOW, 'set DESK_MUTATE=1 and DESK_LINK_WORKFLOW to a disposable workflow accepting repo, pr and issue variables');
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const { run } = await call('run.start', { workflow: process.env.DESK_LINK_WORKFLOW, vars: { repo: 'elee1766/jobshop', pr: '988', issue: 'JOB-116' } });
  try {
    await call('ticket.update', { id: run.id, meta: { linear: { workspace: 'gfxlabs' } } });
    await page.goto('./');
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'in flight', exact: true }).click();
    const row = page.getByRole('region', { name: 'In flight', exact: true }).getByRole('row').filter({ has: page.locator(`.ticket-label[title*="${run.id}"]`) });
    await expect(row.getByRole('link', { name: 'JOB-116', exact: true })).toHaveAttribute('href', 'https://linear.app/gfxlabs/issue/JOB-116');
    await expect(row.getByRole('link', { name: 'jobshop#988', exact: true })).toHaveAttribute('href', 'https://github.com/elee1766/jobshop/pull/988');
    await expect(row.getByRole('link', { name: 'jobshop#988', exact: true })).toBeVisible();
  } finally {
    await call('run.cancel', { id: run.id });
  }
});

test('issue-started run shows no PR, then the PR its step records, without a reload (JOB-131)', async ({ page, request, baseURL, isMobile }) => {
  test.skip(!process.env.DESK_MUTATE || !process.env.DESK_ISSUE_WORKFLOW, 'set DESK_MUTATE=1 and DESK_ISSUE_WORKFLOW to a disposable workflow with an issue variable and an agent step in a pool that spawns nothing');
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const { run, steps } = await call('run.start', { workflow: process.env.DESK_ISSUE_WORKFLOW, vars: { issue: '131' } });
  try {
    await page.goto('./');
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'in flight', exact: true }).click();
    const row = page.getByRole('region', { name: 'In flight', exact: true }).getByRole('row').filter({ has: page.locator(`.ticket-label[title*="${run.id}"]`) });
    await expect(row).toBeVisible();
    await expect(row.locator('a[href*="/pull/"]')).toHaveCount(0);
    if (!isMobile) await expect(row.locator('td.ref-col').nth(1)).toHaveText('–');
    // A mention is not the run's PR, the handoff line is.
    await call('comment.add', { id: steps[0].id, body: 'related: https://github.com/acme/widget/pull/5' });
    await call('comment.add', { id: steps[0].id, body: 'PR https://github.com/acme/widget/pull/131 head abc1234' });
    // The wide layout shows the PR column, the compact one links beneath the title.
    const link = row.locator('a[href="https://github.com/acme/widget/pull/131"]').locator('visible=true');
    await expect(link).toHaveCount(1);
    await expect(link).toHaveText('widget#131');
    await expect(link).toHaveClass(isMobile ? /ref-stack/ : /^$/);
    await expect(row.locator('a[href$="/pull/5"]')).toHaveCount(0);
    // Desktop: labeled issue and pr columns beside the title. Compact: the same
    // links stacked under it. Never both, and no horizontal overflow.
    const headers = page.getByRole('region', { name: 'In flight', exact: true }).getByRole('columnheader');
    if (isMobile) {
      await expect(headers.filter({ hasText: /^issue$/ })).toHaveCount(0);
      await expect(row.locator('td.ref-col').first()).toBeHidden();
    } else {
      await expect(headers.filter({ hasText: /^issue$/ })).toBeVisible();
      await expect(headers.filter({ hasText: /^pr$/ })).toBeVisible();
      await expect(row.locator('.ref-stack').first()).toBeHidden();
    }
    const over = await page.evaluate(() => { const s = document.querySelector('.table-scroll.refs'); return s.scrollWidth - s.clientWidth; });
    expect(over).toBeLessThanOrEqual(0);
  } finally {
    await call('run.cancel', { id: run.id });
  }
});
