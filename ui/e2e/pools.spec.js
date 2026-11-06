import { test, expect } from '@playwright/test';

// The Pools and triggers view at gasserver scale (JOB-81): about 40 pools
// in 3 instances, nearly all idle. The API is mocked, so this needs no
// live controller beyond the Desk that serves the UI.

// desk is the built-in app, the frame around every view (JOB-85).
const builtins = ['desk', 'attention', 'pools'];
const pathOf = (/** @type {string} */ url) => new URL(url).pathname.replace(/^.*?(?=\/api\/)/, '');

const row = (/** @type {string} */ name, /** @type {string[]} */ labels, /** @type {any} */ extra = {}) =>
  ({ name, project: 'js', labels, min: 0, max: 2, desired: 0, open: 0, working: 0, idle: 0, cordoned: 0, ...extra });

// 3 instances x 13 pools. sap has one working pool, uni has a quarantined
// one and a target override, luna is entirely idle.
const instances = ['sap', 'uni', 'luna'];
const suffixes = ['check', 'gate', 'dev', 'plan', 'review', 'sonnet-review', 'ci', 'merge', 'triage', 'docs', 'fix', 'release', 'audit'];
const poolList = instances.flatMap((inst) => suffixes.map((s) => {
  const labels = ['instance:' + inst, 'pack:github', ...(s.includes('review') ? ['review'] : [])];
  if (inst === 'sap' && s === 'dev') return row(inst + '-' + s, labels, { desired: 2, open: 2, working: 1, idle: 1 });
  if (inst === 'uni' && s === 'gate') return row(inst + '-' + s, labels, { target: 0 });
  return row(inst + '-' + s, labels);
}));

const triggerList = instances.flatMap((inst) => ['red', 'replan', 'stall'].map((s) => ({
  name: inst + '-' + s, holds: false, why: 'ready >= 1', labels: ['instance:' + inst, 'pack:github'],
  ...(inst === 'uni' && s === 'stall' ? { holds: true, stalls: ['issue-7'] } : {}),
})));

const poolsBody = {
  quarantined: { 'uni-dev': 'spawn: repo missing' },
  paused: {},
  host_max: 8,
  pools: [...poolList, row('uni-dev', ['instance:uni', 'pack:github'])],
  triggers: triggerList,
};

/** @param {import('@playwright/test').Page} page */
async function mock(page) {
  const plugins = await Promise.all(builtins.map(async (n) => {
    const r = await page.request.get('plugins/' + n + '/plugin.json');
    return { ...(await r.json()), source: 'builtin', url: 'plugins/' + n + '/' };
  }));
  await page.route('**/api/**', async (route) => {
    const path = pathOf(route.request().url());
    const json = (/** @type {any} */ b) => route.fulfill({ json: b });
    if (path === '/api/plugins') return json(plugins);
    if (path === '/api/pools') return json(poolsBody);
    if (path === '/api/attention') return json([]);
    if (path === '/api/push/key') return route.fulfill({ status: 404, json: { error: 'push off' } });
    if (path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    return json([]);
  });
}

const groupRow = (/** @type {import('@playwright/test').Page} */ page, /** @type {string} */ name) =>
  page.locator(`[data-testid=pool-group][data-group=${name}]`);

test('pools group by instance and fold idle pools', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=pools');
  await expect(page.getByTestId('scale')).toBeVisible();
  await expect(page.getByTestId('pool-group')).toHaveCount(3);

  // luna is all idle: one summary row, no pool rows.
  await expect(groupRow(page, 'luna').getByTestId('pool-row')).toHaveCount(0);
  await expect(groupRow(page, 'luna').getByTestId('idle-note')).toHaveText('13 idle');
  // sap shows its working pool and folds the other 12.
  await expect(groupRow(page, 'sap').getByTestId('pool-row')).toHaveCount(1);
  await expect(groupRow(page, 'sap').locator('[data-pool=sap-dev]')).toBeVisible();
  await expect(groupRow(page, 'sap').getByTestId('fold')).toHaveText('12 idle pools');
  // uni shows the quarantined pool and the target override, with the
  // quarantine reason folded into the row rather than a second table.
  await expect(groupRow(page, 'uni').locator('[data-pool=uni-dev]')).toContainText('quarantined');
  await expect(groupRow(page, 'uni').locator('[data-pool=uni-dev]')).toContainText('spawn: repo missing');
  await expect(groupRow(page, 'uni').locator('[data-pool=uni-gate]')).toContainText('target 0');
  await expect(groupRow(page, 'uni').getByTestId('pool-row')).toHaveCount(2);
  // The whole view is a handful of rows, not 39.
  expect(await page.locator('[data-testid=pool-row]').count()).toBeLessThan(6);
});

test('a group expands on click and the state is in the URL', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=pools');
  await groupRow(page, 'luna').locator('tr.grp').click();
  await expect(page).toHaveURL(/open=luna/);
  await expect(groupRow(page, 'luna').getByTestId('pool-row')).toHaveCount(13);
  await page.reload();
  await expect(groupRow(page, 'luna').getByTestId('pool-row')).toHaveCount(13);
  await groupRow(page, 'luna').locator('tr.grp').click();
  await expect(groupRow(page, 'luna').getByTestId('pool-row')).toHaveCount(0);
});

test('search narrows pools and triggers and round-trips in the URL', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=pools');
  await page.getByTestId('pool-search').fill('sonnet-review');
  await page.getByTestId('pool-search').press('Enter');
  await expect(page).toHaveURL(/q=sonnet-review/);
  await expect(page.getByTestId('pool-group')).toHaveCount(3);
  await groupRow(page, 'sap').locator('tr.grp').click();
  await expect(groupRow(page, 'sap').getByTestId('pool-row')).toHaveCount(1);
  await expect(page.getByTestId('trigger-row')).toHaveCount(0);
  await page.getByTestId('clear-filters').click();
  await expect(page).not.toHaveURL(/q=/);

  await page.goto('./?view=pools&q=uni-stall');
  await expect(page.getByTestId('pool-search')).toHaveValue('uni-stall');
  await expect(page.getByTestId('pool-row')).toHaveCount(0);
  await expect(page.locator('[data-trigger=uni-stall]')).toContainText('stalled');
});

test('label chips filter and combine', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=pools');
  const chips = page.getByTestId('label-chips');
  await chips.getByRole('button', { name: /^review / }).click();
  await expect(page).toHaveURL(/label=review/);
  // 6 review pools (2 per instance) and no trigger carries the label.
  await expect(page.getByTestId('pool-group')).toHaveCount(3);
  await expect(page.getByTestId('trigger-row')).toHaveCount(0);
  await chips.getByRole('button', { name: /^instance:sap / }).click();
  await expect(page).toHaveURL(/label=review%2Cinstance%3Asap|label=review,instance:sap/);
  await expect(page.getByTestId('pool-group')).toHaveCount(1);
  await groupRow(page, 'sap').locator('tr.grp').click();
  await expect(groupRow(page, 'sap').getByTestId('pool-row')).toHaveCount(2);
  await page.goto('./?view=pools&label=review,instance:sap');
  await expect(chips.getByRole('button', { name: /^review / })).toHaveClass(/on/);
});

test('group by project and none; triggers show stalled first', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=pools');
  // Stalled and firing triggers show, waiting ones fold per group.
  const uni = page.locator('[data-testid=trigger-group][data-group=uni]');
  await expect(uni.locator('[data-trigger=uni-stall]')).toContainText('stalled');
  await expect(uni.getByTestId('trigger-row')).toHaveCount(1);
  await expect(page.locator('[data-testid=trigger-group][data-group=sap]').getByTestId('trigger-row')).toHaveCount(0);

  await page.getByTestId('group-by').selectOption('project');
  await expect(page).toHaveURL(/group=project/);
  await expect(page.getByTestId('pool-group')).toHaveCount(1);
  await expect(groupRow(page, 'js').getByTestId('pool-row')).toHaveCount(3);

  await page.getByTestId('group-by').selectOption('none');
  await expect(page.locator('[data-testid=pool-group] tr.grp')).toHaveCount(0);
  await expect(page.getByTestId('pool-row')).toHaveCount(3);
});
