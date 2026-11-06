import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';

// The nav slot (JOB-86): the built-in desk app docks one nav contribution
// by its placement. The built-in nav lists api.views(); the example
// top-nav replaces it; with no nav, main fills the page. The API is
// mocked, so this runs against any Desk build and under DESK_PREFIX.

const pathOf = (/** @type {string} */ url) => new URL(url).pathname.replace(/^.*?(?=\/api\/)/, '');

const wide = { width: 1280, height: 800 };
const narrow = { width: 390, height: 780 };

/** An example plugin directory as mock() serves it. @param {string} dir */
const fromExample = (dir) => {
  const base = new URL('../../examples/desk-plugins/' + dir + '/', import.meta.url);
  return {
    manifest: JSON.parse(readFileSync(new URL('plugin.json', base), 'utf8')),
    files: { 'index.js': readFileSync(new URL('index.js', base), 'utf8'), 'style.css': readFileSync(new URL('style.css', base), 'utf8') },
  };
};

/**
 * Mocks Desk's API with built-in plugins plus extras served from memory.
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
    if (path === '/api/activity') return json({ items: [], sources: [] });
    if (path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    return json([]);
  });
}

const views = ['attention', 'chat', 'work', 'activity', 'sessions', 'pools'];
const base = ['desk', 'nav', ...views];

/** @param {import('@playwright/test').Page} page @param {string} sel */
async function box(page, sel) {
  const b = await page.locator(sel).first().boundingBox();
  if (!b) throw new Error('no box for ' + sel);
  return b;
}

test('the built-in nav lists the views in their old order, with icons, and g then a letter jumps', async ({ page }) => {
  await mock(page, { builtins: base, attention: 2 });
  await page.setViewportSize(wide);
  await page.goto('./');
  await expect(page.locator('nav[aria-label=views] a[data-view]')).toHaveCount(views.length);
  expect(await page.locator('nav[aria-label=views] a[data-view]').evaluateAll((els) => els.map((e) => /** @type {HTMLElement} */ (e).dataset.view))).toEqual(views);
  await expect(page.locator('nav a[data-view=work] svg')).toBeVisible();
  await expect(page.locator('nav a[data-view=attention] .badge')).toHaveText('2');
  await expect(page.getByTestId('desk-app')).toHaveAttribute('data-dock', 'left');
  await page.keyboard.press('g');
  await page.keyboard.press('w');
  await expect(page).toHaveURL(/view=work/);
  await expect(page.locator('nav a[data-view=work]')).toHaveClass(/active/);
});

test('the built-in nav docks left on a wide page and as a bottom tab bar on a phone', async ({ page }) => {
  await mock(page, { builtins: base });
  await page.setViewportSize(wide);
  await page.goto('./?view=work');
  await expect(page.getByTestId('runs')).toBeVisible();
  const side = await box(page, 'nav[aria-label=views]');
  const main = await box(page, 'main');
  expect(side.x).toBe(0);
  expect(side.width).toBeCloseTo(212, 0);
  expect(main.x).toBeGreaterThanOrEqual(side.x + side.width - 1);

  await page.setViewportSize(narrow);
  await expect(page.getByTestId('desk-app')).toHaveAttribute('data-dock', 'bottom');
  const bar = await box(page, 'nav[aria-label=views]');
  const m = await box(page, 'main');
  expect(bar.y).toBeGreaterThanOrEqual(m.y + m.height - 1);
  expect(bar.width).toBeCloseTo(narrow.width, 0);
  // The brand is the top bar, and the nav's live dot and theme toggle sit in it.
  const brand = await box(page, '.brand');
  const tools = await box(page, 'nav[aria-label=views] .tools');
  expect(brand.y).toBeLessThan(m.y);
  expect(tools.y + tools.height).toBeLessThanOrEqual(m.y + 1);
  expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)).toBeLessThanOrEqual(0);
});

test('the example top-bar nav replaces the built-in on wide and narrow pages', async ({ page }) => {
  const errors = /** @type {string[]} */ ([]);
  page.on('pageerror', (e) => errors.push(String(e)));
  await mock(page, { builtins: base, extra: [fromExample('top-nav')], attention: 3 });
  await page.setViewportSize(wide);
  await page.goto('./?view=work');
  await expect(page.getByTestId('top-nav')).toBeVisible();
  await expect(page.getByTestId('desk-app')).toHaveAttribute('data-dock', 'top');
  // One nav only: the built-in's side bar is gone.
  await expect(page.locator('desk-nav')).toHaveCount(0);
  await expect(page.locator('.brand')).toHaveCount(0);
  expect(await page.locator('.top-nav a[data-view]').evaluateAll((els) => els.map((e) => /** @type {HTMLElement} */ (e).dataset.view))).toEqual(views);
  await expect(page.locator('.top-nav a[data-view=attention] .n')).toHaveText('3');
  let nav = await box(page, '[data-testid=top-nav]');
  let main = await box(page, 'main');
  expect(nav.y).toBeLessThan(main.y);
  expect(nav.width).toBeCloseTo(wide.width, 0);
  expect(main.width).toBeCloseTo(wide.width, 0);
  await expect(page.getByTestId('runs')).toBeVisible();

  await page.locator('.top-nav a[data-view=activity]').click();
  await expect(page).toHaveURL(/view=activity/);
  await expect(page.locator('.top-nav a[data-view=activity]')).toHaveClass(/on/);

  // Narrow: placement.narrow is also top, so it stays above the page, and the
  // container query shortens the labels to fit.
  await page.setViewportSize(narrow);
  await expect(page.getByTestId('desk-app')).toHaveAttribute('data-dock', 'top');
  nav = await box(page, '[data-testid=top-nav]');
  main = await box(page, 'main');
  expect(nav.y).toBeLessThan(main.y);
  expect(nav.width).toBeCloseTo(narrow.width, 0);
  const lbl = await page.locator('.top-nav a[data-view=attention] .lbl').evaluate((e) => getComputedStyle(e).maxWidth);
  expect(lbl).not.toBe('none');
  expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)).toBeLessThanOrEqual(0);
  expect(errors).toEqual([]);
});

test('a nav that places itself (none) gets no dock and main fills the page', async ({ page }) => {
  const floating = {
    manifest: { name: 'zz-float', entry: 'index.js', api: '>=6', contributes: [{ slot: 'nav', element: 'zz-float-nav', placement: { wide: 'none', narrow: 'none' } }] },
    files: { 'index.js': "customElements.define('zz-float-nav', class extends HTMLElement { connectedCallback() { this.dataset.testid = 'float'; this.style.cssText = 'position:fixed;right:8px;bottom:8px;z-index:9'; this.textContent = 'menu'; } });\n" },
  };
  await mock(page, { builtins: base, extra: [floating] });
  await page.setViewportSize(wide);
  await page.goto('./?view=work');
  await expect(page.getByTestId('float')).toBeVisible();
  await expect(page.getByTestId('desk-app')).toHaveAttribute('data-dock', 'none');
  const main = await box(page, 'main');
  expect(main.x).toBe(0);
  expect(main.width).toBeCloseTo(wide.width, 0);
});

test('with nav disabled, main fills the page on wide and narrow viewports', async ({ page }) => {
  await mock(page, { builtins: ['desk', ...views] });
  for (const vp of [wide, narrow]) {
    await page.setViewportSize(vp);
    await page.goto('./?view=work');
    await expect(page.getByTestId('runs')).toBeVisible();
    await expect(page.getByTestId('desk-app')).toHaveAttribute('data-dock', 'none');
    await expect(page.locator('nav[aria-label=views]')).toHaveCount(0);
    await expect(page.locator('.brand')).toHaveCount(0);
    const main = await box(page, 'main');
    expect(main.x).toBe(0);
    expect(main.y).toBeCloseTo(0, 0);
    expect(main.width).toBeCloseTo(vp.width, 0);
    expect(main.height).toBeCloseTo(vp.height, 0);
  }
});

test('a broken nav shows an error card in the nav box and the page keeps working', async ({ page }) => {
  const broken = {
    manifest: { name: 'zz-broken-nav', entry: 'index.js', api: '>=6', contributes: [{ slot: 'nav', element: 'zz-broken-nav', placement: { wide: 'left' } }] },
    files: { 'index.js': "throw new Error('e2e nav import failure');\n" },
  };
  await mock(page, { builtins: base, extra: [broken] });
  await page.setViewportSize(wide);
  await page.goto('./?view=work');
  await expect(page.getByTestId('nav-host').getByTestId('plugin-error')).toContainText('e2e nav import failure');
  await expect(page.getByTestId('runs')).toBeVisible();
});

test('views carry the order, icon and key from their manifests', async ({ request }) => {
  for (const [name, order, icon, key] of [['attention', 1, 'inbox', 'a'], ['work', 3, 'list-checks', 'w'], ['pools', 6, 'boxes', 'p']]) {
    const m = await (await request.get('plugins/' + name + '/plugin.json')).json();
    const v = m.contributes.find((/** @type {any} */ c) => c.slot === 'view');
    expect([v.order, v.icon, v.key]).toEqual([order, icon, key]);
  }
});
