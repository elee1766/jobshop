import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

// The Linear and GitHub views (JOB-163). Read only: the live stream and the
// browse.* calls are mocked with what Linear and GitHub really answered (see
// pkg/browse/testdata and e2e/fixtures), so this runs against any Desk and
// writes nothing. e2e/browse-live.spec.js reads the real providers.

const dir = join(dirname(fileURLToPath(import.meta.url)), 'fixtures');
const load = (name) => JSON.parse(readFileSync(join(dir, name + '.json'), 'utf8'));
const linearList = load('linear-list');
const linearShow = load('linear-show');
const githubList = load('github-list');
const githubShow = load('github-show');

const scope = (extra = {}) => ({
  revision: 1,
  linear: [{ team: 'JOB', workspace: 'gfx-labs' }],
  github: [{ repo: 'elee1766/jobshop' }],
  writes: { enabled: false, reason: 'Provider writes are not connected yet.' },
  limits: { page_size: 50, max_page_size: 100, thread_page: 25 },
  ...extra,
});

const stream = (events) =>
  events.map(([kind, data]) => `event: ${kind}\ndata: ${JSON.stringify(data)}\n\n`).join('');

// mock answers the stream and the four browse methods. `calls` records each
// browse call so a test can see what the page asked for.
async function mock(page, over = {}) {
  const calls = [];
  const live = over.live ?? [['browse', scope()]];
  await page.route('**/api/live**', (route) =>
    route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: stream(live),
    }),
  );
  await page.route('**/api/rpc/browse.*', async (route) => {
    const method = new URL(route.request().url()).pathname.split('/').pop();
    const body = route.request().postDataJSON() ?? {};
    calls.push({ method, body });
    const answer = over[method] ? await over[method](body) : null;
    const fallback = {
      'browse.list': body.provider === 'github' ? githubList : linearList,
      'browse.show': body.provider === 'github' ? githubShow : linearShow,
      'browse.catalog': {
        states: [],
        projects: [{ id: '11111111-1111-1111-1111-111111111111', name: 'desk' }],
        users: [{ id: '22222222-2222-2222-2222-222222222222', name: 'eddy' }],
        labels: [],
        status: { state: 'fresh' },
      },
    }[method];
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(answer ?? fallback) });
  });
  return calls;
}

const nav = (page) => page.getByRole('navigation', { name: 'Main' });
async function open(page, isMobile, name) {
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  await nav(page).getByRole('button', { name, exact: true }).click();
}

test('linear and github are tabs of their own, not iframes', async ({ page, isMobile }) => {
  await mock(page);
  await page.goto('./');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  const names = await nav(page).getByRole('button').allTextContents();
  expect(names.slice(0, 6)).toEqual(['home', 'trace', 'learn', 'sys', 'linear', 'github']);
  await nav(page).getByRole('button', { name: 'linear', exact: true }).click();
  await expect(page).toHaveURL(/view=linear/);
  await expect(page.getByRole('table', { name: 'Issues' })).toBeVisible();
  await expect(page.locator('iframe')).toHaveCount(0);
  await expect(nav(page).getByRole('button', { name: 'linear', exact: true })).toHaveAttribute('aria-current', 'page');
});

test('linear list: groups, status letters, ids and owners from the provider', async ({ page, isMobile }) => {
  const calls = await mock(page);
  await page.goto('./?view=linear');
  const table = page.getByRole('table', { name: 'Issues' });
  await expect(table.getByRole('row', { name: /JOB-165/ })).toBeVisible();
  const first = linearList.items[0];
  await expect(table.getByRole('button', { name: first.title })).toBeVisible();
  // The one-letter status is named in full.
  await expect(table.getByRole('img', { name: /started/ }).first()).toBeVisible();
  // Groups follow the provider's state types.
  await expect(table.getByRole('button', { name: /active/ })).toBeVisible();
  const list = calls.find((c) => c.method === 'browse.list');
  expect(list.body).toMatchObject({ provider: 'linear', scope: 'JOB', view: 'open' });
  // A group folds.
  const group = table.getByRole('button', { name: /^▾ active/ });
  await group.click();
  await expect(table.getByRole('row', { name: new RegExp(first.ref) })).toHaveCount(0);
  await table.getByRole('button', { name: /^▸ active/ }).click();
  await expect(table.getByRole('row', { name: new RegExp(first.ref) })).toBeVisible();
  void isMobile;
});

test('the loaded list is searched as you type and the provider on Enter', async ({ page }) => {
  const calls = await mock(page);
  await page.goto('./?view=linear');
  const table = page.getByRole('table', { name: 'Issues' });
  await expect(table).toBeVisible();
  const box = page.getByRole('searchbox', { name: 'Search issues' }).or(page.getByLabel('Search issues'));
  await box.fill('baseline');
  await expect(table.getByRole('row', { name: /JOB-165/ })).toBeVisible();
  await expect(table.getByRole('row', { name: /JOB-175/ })).toHaveCount(0);
  await expect(page.getByText(/loaded match\. Enter searches Linear/)).toBeVisible();
  const before = calls.filter((c) => c.method === 'browse.list').length;
  await box.press('Enter');
  await expect(page).toHaveURL(/q=baseline/);
  await expect.poll(() => calls.filter((c) => c.method === 'browse.list').length).toBeGreaterThan(before);
  expect(calls.at(-1).body.query ?? calls.filter((c) => c.method === 'browse.list').at(-1).body.query).toBe('baseline');
});

test('load more asks for the next page and does not repeat rows', async ({ page }) => {
  const second = {
    ...linearList,
    items: [{ ...linearList.items[0], ref: 'JOB-1', id: 'JOB-1', title: 'An older issue' }, linearList.items[0]],
    next: '',
    has_more: false,
  };
  const calls = await mock(page, { 'browse.list': (body) => (body.cursor ? second : linearList) });
  await page.goto('./?view=linear');
  const table = page.getByRole('table', { name: 'Issues' });
  await expect(table).toBeVisible();
  await page.getByRole('button', { name: 'load more' }).click();
  await expect(table.getByRole('button', { name: 'An older issue' })).toBeVisible();
  await expect(table.getByRole('button', { name: linearList.items[0].title })).toHaveCount(1);
  await expect(page.getByRole('button', { name: 'load more' })).toHaveCount(0);
  expect(calls.filter((c) => c.method === 'browse.list').at(-1).body.cursor).toBe(linearList.next);
});

test('an issue opens in place: properties, linked work, threads, replies and disabled writes', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=linear');
  await page.getByRole('table', { name: 'Issues' }).getByRole('button', { name: /JOB-165|Refactor baseline/ }).first().click();
  await expect(page).toHaveURL(/item=JOB-/);
  const detail = page.getByRole('region', { name: 'Issue', exact: true });
  await expect(detail.getByRole('heading', { level: 2 }).first()).toContainText('Migrate Desk to React');
  // The properties are controls that would write, and they are disabled.
  const props = detail.getByRole('group', { name: /Properties/ });
  await expect(props.getByRole('button', { name: /▾/ }).first()).toHaveAttribute('aria-disabled', 'true');
  await expect(detail.getByText(/read only\. Provider writes are not connected/)).toBeVisible();
  await expect(detail.getByLabel(/New comment \(writes are not connected\)/)).toBeDisabled();
  // Real links only: the two attached PRs.
  const linked = detail.getByRole('region', { name: 'Linked work' });
  await expect(linked.getByRole('link', { name: /elee1766\/jobshop#1097/ }).first()).toBeVisible();
  // A reply sits under its parent and names who it answers.
  const reply = detail.locator('.browse-replies .browse-message').first();
  await expect(reply).toContainText(/replying to/);
  await expect(reply.getByRole('button', { name: 'reply (disabled)' })).toHaveAttribute('aria-disabled', 'true');
  // A thread folds and says how many replies it hides.
  await detail.locator('.browse-fold').first().click();
  await expect(detail.getByRole('button', { name: /show 1 replies/ })).toBeVisible();
  // The provider link opens a new tab and keeps no referrer or opener.
  const out = detail.getByRole('link', { name: /open in linear/ });
  await expect(out).toHaveAttribute('target', '_blank');
  await expect(out).toHaveAttribute('rel', /noopener/);
});

test('a linked PR in a configured repo opens in the github view, and a run opens its trace', async ({ page }) => {
  const withRun = {
    ...linearShow,
    linked: [...linearShow.linked, { kind: 'run', ref: 'js-run1', title: 'Desk migration run', state: 'running', relation: 'run' }],
  };
  await mock(page, { 'browse.show': (body) => (body.provider === 'linear' ? withRun : githubShow) });
  await page.goto('./?view=linear&item=JOB-150');
  const linked = page.getByRole('region', { name: 'Linked work' });
  await expect(linked.getByRole('link', { name: 'Desk migration run' })).toBeVisible();
  await linked.getByRole('link', { name: /elee1766\/jobshop#1097/ }).first().click();
  await expect(page).toHaveURL(/view=github/);
  await expect(page).toHaveURL(/item=elee1766%2Fjobshop%231097/);
  await expect(page.getByRole('region', { name: 'Pull request' }).getByRole('heading', { level: 2 }).first()).toContainText('migrate Desk');
  await page.goBack();
  await expect(page).toHaveURL(/view=linear/);
  await page.getByRole('region', { name: 'Linked work' }).getByRole('link', { name: 'Desk migration run' }).click();
  await expect(page).toHaveURL(/view=trace&ref=js-run1/);
});

test('github pr: header, tabs with counts, checks and files from the provider', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=github');
  const table = page.getByRole('table', { name: /Pull requests/ });
  await expect(table).toBeVisible();
  await expect(table.getByRole('columnheader', { name: 'ci' })).toBeVisible();
  await page.goto('./?view=github&item=elee1766%2Fjobshop%231097');
  const detail = page.getByRole('region', { name: 'Pull request', exact: true });
  await expect(detail.getByRole('heading', { level: 2 }).first()).toContainText('(JOB-150)');
  await expect(detail.getByText(/merged · elee-agent \(app\)/)).toBeVisible();
  await expect(detail.getByText(/\+7418/)).toBeVisible();
  const tabs = detail.getByRole('tablist');
  await expect(tabs.getByRole('tab', { name: /files 75/ })).toBeVisible();
  await expect(tabs.getByRole('tab', { name: /checks 5/ })).toBeVisible();
  await tabs.getByRole('tab', { name: /checks/ }).click();
  await expect(page).toHaveURL(/tab=checks/);
  const checks = detail.getByRole('table').first();
  await expect(checks.getByRole('cell', { name: 'go', exact: true })).toBeVisible();
  await expect(checks.getByRole('img', { name: 'passed' }).first()).toBeVisible();
  await tabs.getByRole('tab', { name: /files/ }).click();
  await expect(detail.getByRole('cell', { name: /^\.github|^ui|^pkg|^docs|^packages|^nix|^cmd/ }).first()).toBeVisible();
  await expect(detail.getByRole('row')).toHaveCount(76); // the header and every changed file
  await expect(detail.getByText(/files shown/)).toHaveCount(0); // nothing is cut
  await tabs.getByRole('tab', { name: 'discussion' }).click();
  await expect(detail.getByRole('region', { name: 'Linked work' }).getByRole('link', { name: /JOB-150/ })).toBeVisible();
});

test('a stale list says so and keeps its rows; a rate limit says when to retry', async ({ page }) => {
  const stale = {
    ...linearList,
    status: { state: 'stale', code: 'rate_limited', message: 'Linear rate limit reached.', fetched_at: new Date(Date.now() - 120_000).toISOString(), retry_at: new Date(Date.now() + 90_000).toISOString() },
  };
  await mock(page, { 'browse.list': () => stale });
  await page.goto('./?view=linear');
  const status = page.getByRole('status').filter({ hasText: 'rate limited' });
  await expect(status).toContainText('Showing an older copy.');
  await expect(status).toContainText(/Try again in \d+(s|m)/);
  await expect(page.getByRole('table', { name: 'Issues' }).getByRole('row', { name: /JOB-165/ })).toBeVisible();
  await expect(status.getByRole('button', { name: 'retry' })).toBeVisible();
});

test('auth, unavailable and empty states show a reason and no stale rows', async ({ page }) => {
  await mock(page, {
    'browse.list': () => ({ items: [], has_more: false, status: { state: 'error', code: 'auth', message: 'Linear rejected the service identity\u2019s credentials.' } }),
  });
  await page.goto('./?view=linear');
  const status = page.getByRole('status').filter({ hasText: 'auth' });
  await expect(status).toContainText('rejected the service identity');
  await expect(page.getByRole('table', { name: 'Issues' })).toHaveCount(0);
  // The provider's own message is shown, never a credential.
  await expect(page.locator('body')).not.toContainText(/lin_api_|ghs_|Bearer /);
});

test('an empty result says what to try', async ({ page }) => {
  await mock(page, { 'browse.list': () => ({ items: [], has_more: false, status: { state: 'fresh' } }) });
  await page.goto('./?view=linear&list=closed');
  await expect(page.getByText(/No issues match\. Try another list or clear the search\./)).toBeVisible();
});

test('a team or repo that is not configured is explained, not guessed', async ({ page }) => {
  await mock(page, { live: [['browse', scope({ linear: [], github: [] })]] });
  await page.goto('./?view=linear');
  await expect(page.getByText(/No Linear team is configured for Desk/)).toBeVisible();
  await page.goto('./?view=github');
  await expect(page.getByText(/No GitHub repository is configured for Desk/)).toBeVisible();
});

test('a scope outside the configuration falls back to a configured one', async ({ page }) => {
  const calls = await mock(page);
  await page.goto('./?view=linear&scope=OTHER');
  await expect(page.getByRole('table', { name: 'Issues' })).toBeVisible();
  expect(calls.find((c) => c.method === 'browse.list').body.scope).toBe('JOB');
});

test('a new revision on the live stream reads the list again without a loading flash', async ({ page }) => {
  let n = 0;
  const calls = await mock(page, {
    'browse.list': () => {
      n++;
      return n === 1 ? linearList : { ...linearList, items: [{ ...linearList.items[0], title: 'Changed at the provider' }, ...linearList.items.slice(1)] };
    },
  });
  // Two snapshots: the second moves the revision.
  await page.unroute('**/api/live**');
  await page.route('**/api/live**', (route) =>
    route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: stream([['browse', scope()]]),
    }),
  );
  await page.goto('./?view=linear');
  await expect(page.getByRole('table', { name: 'Issues' })).toBeVisible();
  const before = calls.filter((c) => c.method === 'browse.list').length;
  // Deliver a newer snapshot to the open page through its own event source
  // path: re-route and reload the stream by reconnecting.
  await page.unroute('**/api/live**');
  await page.route('**/api/live**', (route) =>
    route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: stream([['browse', scope({ revision: 2 })]]),
    }),
  );
  await page.evaluate(() => window.dispatchEvent(new Event('online')));
  await expect(page.getByRole('table', { name: 'Issues' }).getByRole('button', { name: 'Changed at the provider' })).toBeVisible();
  expect(calls.filter((c) => c.method === 'browse.list').length).toBeGreaterThan(before);
  await expect(page.getByText('loading', { exact: true })).toHaveCount(0);
});

test('attach to chat adds a provider reference that is not a ticket id', async ({ page, isMobile }) => {
  test.skip(isMobile, 'the phone shows chats on their own tab');
  await mock(page);
  await page.goto('./?view=linear');
  const table = page.getByRole('table', { name: 'Issues' });
  await expect(table).toBeVisible();
  await table.getByRole('row', { name: /JOB-165/ }).getByRole('button', { name: /Actions for JOB-165/ }).click();
  const menu = page.getByRole('menu', { name: 'Actions for JOB-165' });
  await expect(menu.getByRole('menuitem', { name: 'open in trace' })).toHaveCount(0);
  await expect(menu.getByRole('menuitem', { name: 'copy reference' })).toBeVisible();
  const out = menu.getByRole('menuitem', { name: /open in Linear/ });
  await expect(out).toHaveAttribute('href', /^https:\/\/linear\.app\//);
  await expect(menu.getByText('reference', { exact: true })).toBeVisible();
});

test('mobile shows one pane at a time with a way back', async ({ page, isMobile }) => {
  test.skip(!isMobile, 'phone layout only');
  await mock(page);
  await page.goto('./?view=github');
  const list = page.getByRole('region', { name: /GitHub/ });
  await expect(list.getByRole('table')).toBeVisible();
  await expect(page.getByRole('region', { name: 'Pull request' })).toBeHidden();
  await page.goto('./?view=github&item=elee1766%2Fjobshop%231097');
  await expect(page.getByRole('region', { name: 'Pull request' })).toBeVisible();
  await expect(list).toBeHidden();
  const body = await page.evaluate(() => ({ scroll: document.documentElement.scrollWidth, view: window.innerWidth }));
  expect(body.scroll, 'no horizontal overflow').toBeLessThanOrEqual(body.view);
  await page.getByRole('button', { name: 'Back to the list' }).click();
  await expect(list).toBeVisible();
  await expect(page).not.toHaveURL(/item=/);
});

test('the state is in the URL and survives a reload, Back and Forward', async ({ page }) => {
  await mock(page);
  await page.goto('./?view=github&kind=all&list=merged&group=checks&item=elee1766%2Fjobshop%231097&tab=files');
  await expect(page.getByRole('tab', { name: /files/ })).toHaveAttribute('aria-selected', 'true');
  await expect(page.getByLabel('List', { exact: true })).toHaveValue('merged');
  await expect(page.getByLabel('Kind', { exact: true })).toHaveValue('all');
  await page.reload();
  await expect(page.getByRole('tab', { name: /files/ })).toHaveAttribute('aria-selected', 'true');
  await page.getByRole('tab', { name: 'discussion' }).click();
  await expect(page).not.toHaveURL(/tab=/);
  // A filter replaces its history entry like the rest of Desk. Opening an
  // item is the step Back undoes.
  await page.goto('./?view=github&list=merged');
  await page.getByRole('table', { name: /Pull requests/ }).locator('.browse-open').first().click();
  await expect(page).toHaveURL(/item=/);
  await page.goBack();
  await expect(page).not.toHaveURL(/item=/);
  await expect(page).toHaveURL(/list=merged/);
  await page.goForward();
  await expect(page).toHaveURL(/item=/);
});

test('moving to another view drops the provider filters', async ({ page, isMobile }) => {
  await mock(page);
  await page.goto('./?view=linear&list=backlog&q=desk&item=JOB-150');
  await expect(page.getByRole('region', { name: 'Issue', exact: true })).toBeVisible();
  await open(page, isMobile, 'github');
  await expect(page).toHaveURL(/view=github/);
  await expect(page).not.toHaveURL(/list=backlog|q=desk|item=JOB-150/);
});
