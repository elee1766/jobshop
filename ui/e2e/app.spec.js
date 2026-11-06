import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';

// The app slot (JOB-85): the shell mounts one `app` contribution, the
// built-in desk frame unless another plugin replaces it. The API is
// mocked and the example app is served from examples/desk-plugins, so
// this runs against any Desk build and under DESK_PREFIX.

const pathOf = (/** @type {string} */ url) => new URL(url).pathname.replace(/^.*?(?=\/api\/)/, '');
const example = new URL('../../examples/desk-plugins/mini-app/', import.meta.url);

/**
 * Mocks Desk's API with the built-in views plus extra plugins; files of
 * an extra plugin come from files (name -> body).
 * @param {import('@playwright/test').Page} page
 * @param {{ builtins: string[], extra?: { manifest: any, files: Record<string, string> }[], attention?: number }} o
 */
async function mock(page, o) {
  const list = await Promise.all(o.builtins.map(async (n) => {
    const r = await page.request.get('plugins/' + n + '/plugin.json');
    return { ...(await r.json()), source: 'builtin', url: 'plugins/' + n + '/' };
  }));
  for (const x of o.extra ?? []) {
    list.push({ ...x.manifest, source: 'root', url: 'plugins/' + x.manifest.name + '/' });
    await page.route('**/plugins/' + x.manifest.name + '/*', (route) => {
      const file = new URL(route.request().url()).pathname.split('/').pop() ?? '';
      if (!(file in x.files)) return route.fulfill({ status: 404, body: 'not found' });
      return route.fulfill({ contentType: file.endsWith('.css') ? 'text/css' : 'text/javascript', body: x.files[file] });
    });
  }
  list.sort((a, b) => a.name.localeCompare(b.name));
  await page.route('**/api/**', (route) => {
    const path = pathOf(route.request().url());
    const json = (/** @type {any} */ b) => route.fulfill({ json: b });
    if (path === '/api/plugins') return json(list);
    if (path === '/api/attention') return json(Array.from({ length: o.attention ?? 0 }, (_, i) => ({ kind: 'ticket', id: 'js-a' + i, title: 'item ' + i, reason: 'question' })));
    if (path === '/api/work') return json({ runs: [{ id: 'js-r1', kind: 'run', title: 'A run', status: 'open', project: 'js' }], batches: [] });
    if (path === '/api/push/key') return route.fulfill({ status: 404, json: { error: 'push off' } });
    if (path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    return json([]);
  });
}

const miniApp = {
  manifest: JSON.parse(readFileSync(new URL('plugin.json', example), 'utf8')),
  files: { 'index.js': readFileSync(new URL('index.js', example), 'utf8') },
};

test('the built-in desk app draws the frame and badges attention', async ({ page }) => {
  await mock(page, { builtins: ['attention', 'desk', 'nav', 'work'], attention: 2 });
  await page.goto('./?view=work');
  await expect(page.getByTestId('desk-app')).toBeVisible();
  await expect(page.getByTestId('runs')).toBeVisible();
  await expect(page.locator('nav a[data-view=attention]')).toBeVisible();
  // The badge comes from the shell through api.badges; the phone tab bar shows a dot instead.
  await expect(page.locator('nav a[data-view=attention] .badge, nav a[data-view=attention] .tabdot').filter({ visible: true })).toHaveCount(1);
});

test('an app plugin replaces the frame and mounts the built-in work view', async ({ page }) => {
  const errors = /** @type {string[]} */ ([]);
  page.on('pageerror', (e) => errors.push(String(e)));
  await mock(page, { builtins: ['attention', 'work'], extra: [miniApp], attention: 3 });
  await page.goto('./?view=work');
  await expect(page.getByTestId('mini-title')).toHaveText('Mini app');
  await expect(page.getByTestId('desk-app')).toHaveCount(0);
  await expect(page.locator('nav[aria-label=views]')).toHaveCount(0);
  await expect(page.getByTestId('mini-main').getByTestId('runs')).toContainText('A run');
  await expect(page.getByTestId('mini-live')).toHaveText(/^(live|offline)$/);
  await expect(page.locator('.mini-bar button[data-view=attention]')).toHaveAttribute('data-badge', '3');

  // Navigation through the api: attention mounts in the same main.
  await page.locator('.mini-bar button[data-view=attention]').click();
  await expect(page).not.toHaveURL(/view=work/);
  await expect(page.getByTestId('mini-main').locator('desk-attention-view')).toBeAttached();
  await page.goBack();
  await expect(page.getByTestId('mini-main').getByTestId('runs')).toBeVisible();

  await page.getByTestId('mini-theme').click();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  expect(errors).toEqual([]);
});

test('an app plugin wins over the built-in desk by name order', async ({ page }) => {
  await mock(page, { builtins: ['desk', 'work'], extra: [miniApp] });
  await page.goto('./?view=work');
  await expect(page.getByTestId('mini-title')).toBeVisible();
  await expect(page.getByTestId('desk-app')).toHaveCount(0);
});

test('a broken app shows an error card, not a blank page', async ({ page }) => {
  const broken = {
    manifest: { name: 'zz-broken-app', entry: 'index.js', contributes: [{ slot: 'app', element: 'zz-broken-app' }] },
    files: { 'index.js': "throw new Error('e2e app import failure');\n" },
  };
  await mock(page, { builtins: ['desk', 'work'], extra: [broken] });
  await page.goto('./');
  await expect(page.getByTestId('plugin-error')).toContainText('e2e app import failure');
  await expect(page.getByTestId('plugin-error')).toContainText('zz-broken-app');
});

test('an app needing a newer plugin api shows an error card', async ({ page }) => {
  const future = {
    manifest: { name: 'zz-future-app', entry: 'index.js', api: '>=99', contributes: [{ slot: 'app', element: 'zz-future-app' }] },
    files: { 'index.js': '' },
  };
  await mock(page, { builtins: ['work'], extra: [future] });
  await page.goto('./');
  await expect(page.getByTestId('plugin-error')).toContainText('needs Desk plugin API 99');
});

test('with no app enabled the shell says so', async ({ page }) => {
  await mock(page, { builtins: ['work'] });
  await page.goto('./');
  await expect(page.getByTestId('app-missing')).toContainText('desk');
});
