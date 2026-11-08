import { test, expect } from '@playwright/test';

// Learn with representative data. The live stream is mocked, so this runs
// against any Desk and changes nothing on the controller.
const day = 24 * 3600 * 1000;
const iso = (ago) => new Date(Date.now() - ago).toISOString();
const zero = '0001-01-01T00:00:00Z';
const metrics = (o = {}) => ({ from: iso(8 * day), to: iso(day), runs: 4, steps: 9, retries: 3, escalations: 1, findings: 2, recurrence: 2, ...o });
const learning = (id, o = {}) => ({ id, kind: 'failure', summary: 'problem ' + id, scope: 'repo:acme/widget', source: 'controller', status: 'open', created_at: iso(day), ...o });

const learnings = [
  learning(1, { summary: 'Step closed failed at the retry limit: go test ./... ran after every edit', fix: 'Run only the touched package while editing, the full suite once before push', status: 'applied', pr: 'https://github.com/acme/widget/pull/811', gate: 'not_helped: cited problem still recurs (memory edit #1)', ticket: 'js-123' }),
  learning(2, { kind: 'correction', summary: 'review resumed pr title not a conventional commit', fix: 'Use type(scope): lowercase summary', status: 'applied', pr: 'https://github.com/acme/widget/pull/811' }),
  learning(3, { kind: 'workaround', summary: 'vendorHash mismatch after an import with go.mod untouched', fix: 'Set lib.fakeHash, build goModules, paste the got value', status: 'applied', pr: 'https://github.com/acme/widget/pull/812', gate: 'helped: no steps since the merge', created_at: iso(20 * day) }),
  learning(4, { summary: 'Pushed to master directly', fix: 'Always open a PR from a branch', pr: 'https://github.com/acme/widget/pull/801' }),
  learning(5, { kind: 'correction', summary: 'Opened a PR without checking for an existing one', fix: 'gh pr list --state all --search <issue> first', status: 'rejected', pr: 'https://github.com/acme/widget/pull/790', created_at: iso(30 * day) }),
  learning(6, { summary: 'A very long problem title that goes on and on about a retry budget burned because the implement step kept running the whole suite and a migration collided with a parallel pull request and nobody noticed until review ' + 'x'.repeat(120), fix: 'Short fix' }),
  learning(7, { summary: 'Migration number collided with a parallel PR', created_at: iso(0.2 * day) }),
  learning(8, { kind: 'workaround', summary: 'python3 is not installed', fix: 'Use go or jq', status: 'applied', pr: 'memory:99', created_at: iso(40 * day) }),
  ...Array.from({ length: 14 }, (_, i) => learning(20 + i, { kind: 'workaround', summary: `Open workaround ${i + 1}: flaky tool needs a longer timeout`, fix: 'Retry once with 2x timeout' })),
];
const edits = [
  { id: 1, pr: 'https://github.com/acme/widget/pull/811', learnings: [1, 2, 404], pools: ['dev'], merged_at: iso(2 * day), baseline: metrics(), state: 'not_helped', after: metrics({ from: iso(2 * day), to: iso(0), steps: 5, recurrence: 3 }), reason: 'cited problem still recurs: 3 in 5 steps after, 2 in 9 before', evaluated_at: iso(0.1 * day), created_at: iso(2 * day) },
  { id: 2, pr: 'https://github.com/acme/widget/pull/812', learnings: [3], pools: [], merged_at: iso(19 * day), baseline: metrics({ recurrence: 1 }), state: 'helped', after: metrics({ steps: 0, runs: 0, retries: 0, escalations: 0, findings: 0, recurrence: 0 }), reason: 'no steps in the scoped pools since the merge', evaluated_at: iso(15 * day), created_at: iso(19 * day) },
  { id: 3, path: 'memory/pool/dev.md', before: 'old text\n'.repeat(50), existed: true, learnings: [], pools: ['dev'], merged_at: iso(0.05 * day), baseline: metrics(), state: 'pending', after: { from: zero, to: zero, runs: 0, steps: 0, retries: 0, escalations: 0, findings: 0, recurrence: 0 }, evaluated_at: zero, created_at: iso(0.05 * day) },
];

async function mockLive(page, data) {
  await page.route('**/api/live**', (route) => route.fulfill({
    status: 200,
    headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
    body: `event: improvements\ndata: ${JSON.stringify(data)}\n\n`,
  }));
}

test('learn leads with what needs a look and shows problem with attempted fix', async ({ page }) => {
  const errors = [];
  page.on('pageerror', (e) => errors.push(String(e)));
  await mockLive(page, { learnings, edits });
  await page.goto('./?view=improvements');
  await expect(page.locator('main .tag').first()).toContainText('/learn');
  await expect(page.getByLabel('Summary')).toContainText('not helped');
  const stage = page.getByRole('group', { name: 'Stage' });
  await expect(stage.getByRole('button', { name: /^not helped 1$/ })).toBeVisible();
  await expect(stage.getByRole('button', { name: /^helped 1$/ })).toBeVisible();

  // The first section is the one the gate judged, with both halves visible.
  const first = page.getByRole('region', { name: 'not helped' }).locator('.learn-row').first();
  await expect(first).toContainText('Step closed failed at the retry limit');
  await expect(first).toContainText('attempted fix Run only the touched package');
  await expect(first).toContainText('cited problem still recurs');
  await expect(first.getByRole('link', { name: 'acme/widget#811' })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/811');
  await expect(first).toContainText('+2 more');

  // Details hold the cited learnings, the measurements and what is missing.
  await first.getByText('details', { exact: true }).click();
  await expect(first).toContainText('problem #2');
  await expect(first).toContainText('#404, not in the loaded list');
  await expect(first.getByRole('row', { name: /same problem again 2 3/ })).toBeVisible();

  // A proposal is not called an attempt, and a missing fix says so.
  await expect(page.getByRole('region', { name: 'awaiting curator' }).locator('.learn-row', { hasText: 'Migration number collided' })).toContainText('suggested fix not recorded');
  await expect(page.getByRole('region', { name: 'fix proposed' }).locator('.learn-row').first()).toContainText('proposed fix');

  // History is collapsed but reachable.
  await expect(page.getByRole('region', { name: 'helped', exact: true }).locator('.learn-row')).toHaveCount(0);
  await page.getByRole('region', { name: 'helped', exact: true }).getByRole('button', { name: /^helped/, expanded: false }).click();
  await expect(page.getByRole('region', { name: 'helped', exact: true }).locator('.learn-row')).toHaveCount(1);
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
  expect(errors).toEqual([]);
});

test('learn filters, pages long sections and keeps long titles inside the viewport', async ({ page }) => {
  await mockLive(page, { learnings, edits });
  await page.goto('./?view=improvements');
  const awaiting = page.getByRole('region', { name: 'awaiting curator' });
  await expect(awaiting.locator('.learn-row')).toHaveCount(10);
  await awaiting.getByRole('button', { name: /^more/ }).click();
  await expect(awaiting.locator('.learn-row')).toHaveCount(16);

  const filter = page.getByLabel('Filter learn');
  await filter.fill('vendorhash');
  await expect(page.locator('.learn-row')).toHaveCount(1);
  await filter.fill('812');
  await expect(page.locator('.learn-row')).toHaveCount(1);
  await filter.fill('nothing matches this');
  await expect(page.getByText('Nothing matches the filter.')).toBeVisible();
  await filter.fill('');
  await page.getByRole('group', { name: 'Stage' }).getByRole('button', { name: /^rejected/ }).click();
  await expect(page.locator('.learn-row')).toHaveCount(1);
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
});

test('learn empty state', async ({ page }) => {
  await mockLive(page, { learnings: [], edits: [] });
  await page.goto('./?view=improvements');
  await expect(page.getByText('Nothing recorded yet.')).toBeVisible();
});
