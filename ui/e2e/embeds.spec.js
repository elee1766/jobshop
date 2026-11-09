import { test, expect } from '@playwright/test';

// Embedded apps (JOB-151). Needs the fixture and a Desk whose config carries
// the embeds in docs/verify/desk-embeds.md:
//   node e2e/embed-fixture.mjs
// Read only: it never writes to the controller. Without the embeds the whole
// file skips, so the suite still passes against a plain Desk.

const FIXTURE = process.env.EMBED_FIXTURE || 'http://localhost:7497';

test.beforeEach(async ({ request, baseURL }) => {
  const res = await request.get(new URL('api/embeds', baseURL).href);
  const ids = res.ok() ? (await res.json()).embeds.map((e) => e.id) : [];
  test.skip(!ids.includes('cross'), 'Desk has no embed fixtures configured');
});

const nav = (page) => page.getByRole('navigation', { name: 'Main' });
const tab = (page, name) => nav(page).getByRole('button', { name, exact: true });

async function probe(page) {
  const frame = page.frameLocator('iframe.embed-frame');
  await expect(frame.locator('#probe')).not.toHaveText('probing');
  return JSON.parse(await frame.locator('#probe').textContent());
}

test('tabs follow the built-in views in configured order', async ({ page, isMobile }) => {
  await page.goto('./');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  await expect(tab(page, 'Cross app')).toBeVisible();
  const names = await nav(page).getByRole('button').allTextContents();
  expect(names).toEqual([
    'home',
    'trace',
    'learn',
    'sys',
    'linear',
    'github',
    'Cross app',
    'Same host',
    'Denied XFO',
    'Denied CSP',
    'A rather long embed label',
    'catch-up',
  ]);
});

test('an embed opens in the content area and keeps the header, status and chat rail', async ({ page, isMobile }) => {
  await page.goto('./');
  await expect(page.locator('.status')).toContainText('sys ok');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  await tab(page, 'Cross app').click();
  await expect(page).toHaveURL(/view=embed&embed=cross/);
  const frame = page.locator('iframe.embed-frame');
  await expect(frame).toHaveAttribute('title', 'Cross app (embedded app)');
  await expect(frame).toHaveAttribute('src', `${FIXTURE}/app`);
  await expect(tab(page, 'Cross app')).toHaveAttribute('aria-current', 'page');
  await expect(page.getByRole('link', { name: 'open externally' })).toHaveAttribute('href', `${FIXTURE}/app`);
  await expect(page.getByRole('link', { name: 'open externally' })).toHaveAttribute('target', '_blank');
  await expect(page.getByRole('link', { name: 'open externally' })).toHaveAttribute('rel', /noopener/);
  await expect(page.locator('header.top')).toBeVisible();
  await expect(page.locator('.status')).toContainText('sys ok');
  if (!isMobile) {
    await expect(page.getByRole('complementary', { name: 'Chats' })).toBeVisible();
    // The frame fills the work area beside the rail: not collapsed, not over it.
    const work = await page.locator('.work').boundingBox();
    const box = await frame.boundingBox();
    const rail = await page.locator('.chat-column').boundingBox();
    expect(box.height).toBeGreaterThan(200);
    expect(box.x + box.width).toBeLessThanOrEqual(rail.x + 1);
    expect(box.y).toBeGreaterThanOrEqual(work.y - 1);
  }
  // The page never scrolls sideways.
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
});

test('a deep link, a refresh and history restore the embed', async ({ page }) => {
  await page.goto('./?view=embed&embed=same');
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://127.0.0.1:7498/app');
  await page.reload();
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://127.0.0.1:7498/app');
  await page.goto('./?view=system');
  await expect(page.locator('iframe')).toHaveCount(0);
  await page.goBack();
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://127.0.0.1:7498/app');
  await page.goForward();
  await expect(page.locator('iframe')).toHaveCount(0);
});

test('an unknown id or a URL in the address bar never makes a frame', async ({ page }) => {
  await page.goto('./?view=embed&embed=nope');
  await expect(page.getByText('No embedded app "nope" is configured.')).toBeVisible();
  await expect(page.locator('iframe')).toHaveCount(0);
  for (const hostile of ['javascript:alert(1)', 'https://evil.example/', '//evil.example/', 'cross/../x']) {
    await page.goto('./?view=embed&embed=' + encodeURIComponent(hostile) + '&src=' + encodeURIComponent(hostile));
    await expect(page.locator('iframe')).toHaveCount(0);
  }
  await page.goto('./?view=embed');
  await expect(page.getByText('No embedded app is selected.')).toBeVisible();
  await expect(page.locator('iframe')).toHaveCount(0);
});

test('only the selected embed is mounted and leaving it stops the app', async ({ page, request, isMobile }) => {
  await request.get(`${FIXTURE}/reset`);
  await page.goto('./?view=embed&embed=cross');
  await expect(page.locator('iframe')).toHaveCount(1);
  await expect.poll(async () => (await (await request.get(`${FIXTURE}/seen`)).json()).ticks).toBeGreaterThan(2);
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  await tab(page, 'sys').click();
  await expect(page.locator('iframe')).toHaveCount(0);
  // The app's timer is gone with its document: the count stops moving.
  const stopped = (await (await request.get(`${FIXTURE}/seen`)).json()).ticks;
  await expect
    .poll(async () => (await (await request.get(`${FIXTURE}/seen`)).json()).ticks, { intervals: [600, 600, 600], timeout: 2500 })
    .toBeLessThanOrEqual(stopped + 1);
  // Switching between embeds keeps exactly one frame, and returning reloads.
  await tab(page, 'Cross app').click();
  await expect(page.locator('iframe')).toHaveCount(1);
  await tab(page, 'Same host').click();
  await expect(page.locator('iframe')).toHaveCount(1);
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://127.0.0.1:7498/app');
});

test('the sandbox gives the app an opaque origin and no reach into Desk', async ({ page, request }) => {
  await page.route('**/api/whoami', (route) => route.continue());
  await page.goto('./?view=embed&embed=cross');
  const result = await probe(page);
  expect(result.origin).toBe('null');
  expect(result.parentDocument).toMatch(/^blocked:/);
  expect(result.parentLocation).toMatch(/^blocked:/);
  expect(result.parentStore).toMatch(/^blocked:/);
  expect(result.localStorage).toMatch(/^blocked:/);
  expect(result.cookie).toMatch(/^blocked:/);
  expect(result.topNavigation).toMatch(/^blocked:/);
  expect(result.popup).toMatch(/^blocked:/);
  expect(result.deskApi).toMatch(/^blocked:/);
  // Desk is still where it was.
  await expect(page).toHaveURL(/view=embed&embed=cross/);
  // And the request that loaded the app carried nothing of Desk's.
  const seen = await (await request.get(`${FIXTURE}/seen`)).json();
  expect(seen.headers.cookie).toBeUndefined();
  expect(seen.headers.authorization).toBeUndefined();
  expect(seen.headers['x-jobshop-desk']).toBeUndefined();
  expect(seen.headers.referer).toBeUndefined();
});

test('an embed on Desk\'s own host cannot keep allow-same-origin', async ({ page, baseURL }) => {
  test.skip(new URL(baseURL).hostname !== '127.0.0.1', 'the same-host fixture is 127.0.0.1');
  await page.goto('./?view=embed&embed=same');
  // The config asked for allow-same-origin; Desk dropped it for its own host.
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('sandbox', 'allow-scripts');
  const result = await probe(page);
  expect(result.origin).toBe('null');
  expect(result.parentDocument).toMatch(/^blocked:/);
  expect(result.cookie).toMatch(/^blocked:/);
});

test('an app that refuses framing does not show, and Desk stays usable', async ({ page }) => {
  for (const id of ['denied', 'ancestor']) {
    await page.goto(`./?view=embed&embed=${id}`);
    const frame = page.locator('iframe.embed-frame');
    await expect(frame).toBeVisible();
    await expect(page.frameLocator('iframe.embed-frame').getByRole('heading', { name: 'should not show' })).toHaveCount(0);
    // The way out is on the page, whatever the frame does.
    await expect(page.getByRole('link', { name: 'open externally' })).toBeVisible();
    await expect(page.getByText(/refuses to be framed/)).toBeVisible();
    await tab(page, 'home').click();
    await expect(page.locator('iframe')).toHaveCount(0);
  }
});

test('the page policy allows only the configured frame origins and keeps its protections', async ({ page, request, baseURL }) => {
  const res = await request.get(new URL('./', baseURL).href);
  const csp = res.headers()['content-security-policy'];
  expect(csp).toContain("default-src 'self'");
  expect(csp).toContain("script-src 'self'");
  expect(csp).toContain("frame-ancestors 'none'");
  expect(csp).toMatch(/frame-src [^;]*http:\/\/localhost:7497/);
  expect(csp).not.toMatch(/frame-src[^;]*\*/);
  expect(res.headers()['x-frame-options']).toBe('DENY');
  await page.goto('./');
  // A frame to an origin that is not configured is blocked by the policy.
  const violation = await page.evaluate(
    () =>
      new Promise((resolve) => {
        document.addEventListener('securitypolicyviolation', (e) => resolve(e.violatedDirective), { once: true });
        const f = document.createElement('iframe');
        f.src = 'http://localhost:7490/';
        document.body.append(f);
        setTimeout(() => resolve('none'), 3000);
      }),
  );
  expect(violation).toBe('frame-src');
});

test('a narrow screen keeps every built-in action reachable with the embed tabs', async ({ page, isMobile }) => {
  await page.goto('./');
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  await expectTabsReachable(page);
});

// The built-in layout hides sys and catch-up between 760 and 899 pixels. With
// embed tabs configured nothing may be hidden at any width.
for (const width of [1000, 899, 790, 700, 600]) {
  test(`no tab is hidden or clipped at ${width}px`, async ({ page, isMobile }) => {
    test.skip(isMobile, 'fixed widths are a desktop browser check');
    await page.setViewportSize({ width, height: 700 });
    await page.goto('./');
    await expectTabsReachable(page);
  });
}

async function expectTabsReachable(page) {
  for (const name of ['home', 'trace', 'learn', 'sys', 'catch-up', 'Cross app', 'A rather long embed label']) {
    const button = tab(page, name);
    await button.scrollIntoViewIfNeeded();
    await expect(button).toBeVisible();
    const box = await button.boundingBox();
    const viewport = page.viewportSize();
    expect(box.x).toBeGreaterThanOrEqual(0);
    expect(box.x + box.width).toBeLessThanOrEqual(viewport.width + 1);
    expect(box.height).toBeGreaterThanOrEqual(20);
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
}
