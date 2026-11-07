import { test, expect } from '@playwright/test';

test.skip(!process.env.DESK_MUTATE || !process.env.DESK_LINK_WORKFLOW, 'set DESK_MUTATE=1 and DESK_LINK_WORKFLOW to a disposable workflow accepting repo, pr and issue variables');

test('running ticket references remain clickable in desktop and compact layouts', async ({ page, request, baseURL, isMobile }) => {
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
