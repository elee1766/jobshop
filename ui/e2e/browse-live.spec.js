import { test, expect } from '@playwright/test';

// The Linear and GitHub views against the real providers (JOB-163). It needs a
// Desk on a controller with a Linear identity for team JOB, a GitHub identity
// for gfx-labs/jobshop and the two held runs of docs/verify/desk-linear-github.md.
// Read only. Without that scope the whole file skips, so the suite still passes
// against a plain Desk.

test.beforeEach(async ({ request, baseURL }) => {
  const res = await request.post(new URL('api/rpc/browse.scope', baseURL).href, {
    data: {},
    headers: { 'x-jobshop-desk': '1' },
  });
  const scope = res.ok() ? await res.json() : { linear: [], github: [] };
  test.skip(
    !scope.linear.some((s) => s.team === 'JOB') || !scope.github.some((s) => s.repo === 'gfx-labs/jobshop'),
    'Desk has no live Linear and GitHub scope configured',
  );
});

test('linear lists real issues and opens one with its real links and run', async ({ page }) => {
  await page.goto('./?view=linear');
  const table = page.getByRole('table', { name: 'Issues' });
  await expect(table.locator('.browse-open').first()).toBeVisible();
  await page.goto('./?view=linear&item=JOB-163');
  const detail = page.getByRole('region', { name: 'Issue', exact: true });
  await expect(detail.getByRole('heading', { level: 2 }).first()).toContainText('native Linear and GitHub views');
  // A Jobshop run that names this issue links to its trace.
  const run = detail.getByRole('region', { name: 'Linked work' }).getByRole('link', { name: /native Linear and GitHub views/ });
  await expect(run).toBeVisible();
  await run.click();
  await expect(page).toHaveURL(/view=trace&ref=js-/);
  await expect(page).not.toHaveURL(/item=/);
});

test('github lists real prs and a pr shows checks and files', async ({ page }) => {
  await page.goto('./?view=github&list=all');
  await expect(page.getByRole('table', { name: /Pull requests/ }).locator('.browse-open').first()).toBeVisible();
  await page.goto('./?view=github&list=merged&item=gfx-labs%2Fjobshop%231110');
  const detail = page.getByRole('region', { name: 'Pull request', exact: true });
  await expect(detail.getByRole('heading', { level: 2 }).first()).toContainText('timeline view');
  await expect(detail.getByRole('tab', { name: /files 35/ })).toBeVisible();
  await expect(detail.getByRole('tab', { name: /checks 5/ })).toBeVisible();
  // The PR closes #1103, which is an issue of the same configured repo.
  await expect(detail.getByRole('region', { name: 'Linked work' }).getByRole('link', { name: /#1103/ }).first()).toBeVisible();
});

test('nothing the providers return carries a credential', async ({ page }) => {
  const bodies = [];
  page.on('response', async (res) => {
    if (res.url().includes('/api/rpc/browse.')) bodies.push(await res.text());
  });
  await page.goto('./?view=github&list=all&item=gfx-labs%2Fjobshop%231110');
  await expect(page.getByRole('region', { name: 'Pull request', exact: true }).getByRole('heading', { level: 2 }).first()).toBeVisible();
  await page.goto('./?view=linear&item=JOB-163');
  await expect(page.getByRole('region', { name: 'Issue', exact: true }).getByRole('heading', { level: 2 }).first()).toBeVisible();
  expect(bodies.length).toBeGreaterThan(2);
  expect(bodies.join('\n')).not.toMatch(/lin_api_|lin_oauth_|ghs_[A-Za-z0-9]|ghp_[A-Za-z0-9]|github_pat_|Bearer /);
});
