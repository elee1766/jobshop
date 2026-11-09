import { test, expect } from '@playwright/test';

// App shell behavior added with the React port (JOB-150): the URL contract,
// one live connection per tab, and keyboard shortcut scoping. Read only: the
// shortcut test mocks the live stream, the others use whatever the controller
// has.

const item = (id, title) => ({ kind: 'ticket', id, title, since: new Date().toISOString(), ref: id });

async function mockAttention(page, items) {
  await page.route('**/api/live**', (route) =>
    route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-store' },
      body: `event: attention\ndata: ${JSON.stringify({ items })}\n\n`,
    }),
  );
}

test('one live connection serves every view and survives navigation', async ({ page, isMobile }) => {
  const live = [];
  page.on('request', (req) => {
    if (new URL(req.url()).pathname.endsWith('/api/live')) live.push(req.url());
  });
  await page.goto('./');
  await expect(page.locator('.status')).toContainText('sys ok');
  for (const name of ['learn', 'sys', 'home']) {
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
    await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name, exact: true }).click();
    await expect(page).toHaveURL(name === 'home' ? /view=home|^[^?]*$/ : new RegExp('view=' + (name === 'learn' ? 'improvements' : 'system')));
  }
  await page.goBack();
  await page.goBack();
  await page.goForward();
  await expect(page.locator('.status')).toContainText('sys ok');
  expect(live, 'a view change must not open another stream').toHaveLength(1);
});

test('query state in the URL survives a reload and Back and Forward', async ({ page }) => {
  // An encoded value must reach the app as written, not decoded twice.
  await page.goto('./?view=home&filter=a%2541%20b');
  await expect(page.getByRole('textbox', { name: 'Filter inflight', exact: true })).toHaveValue('a%41 b');
  await page.reload();
  await expect(page.getByRole('textbox', { name: 'Filter inflight', exact: true })).toHaveValue('a%41 b');

  await page.goto('./?view=system');
  await expect(page.getByRole('heading', { name: 'pools' })).toBeVisible();
  await page.goto('./?view=improvements');
  await expect(page.getByRole('textbox', { name: 'Filter learn' })).toBeVisible();
  await page.goBack();
  await expect(page.getByRole('heading', { name: 'pools' })).toBeVisible();

  // Unknown views fall back to home and keep the rest of the query.
  await page.goto('./?view=nope');
  await expect(page.getByRole('region', { name: 'In flight', exact: true })).toBeVisible();
});

test('home shortcuts move the focused item, and stay quiet in text fields and other views', async ({ page, isMobile }) => {
  test.skip(isMobile, 'keyboard shortcuts are for a hardware keyboard');
  await mockAttention(page, [item('js-1', 'first waiting'), item('js-2', 'second waiting')]);
  await page.goto('./');
  const waiting = page.getByRole('region', { name: 'Waiting', exact: true });
  const focused = waiting.locator('article.focused');
  await expect(waiting.locator('article')).toHaveCount(2);
  await expect(focused).toContainText('first waiting');

  await page.keyboard.press('j');
  await expect(focused).toContainText('second waiting');
  await page.keyboard.press('k');
  await expect(focused).toContainText('first waiting');

  // Typing in a text field is typing, not a shortcut.
  await page.getByRole('textbox', { name: 'Filter inflight', exact: true }).click();
  await page.keyboard.type('jj');
  await expect(focused).toContainText('first waiting');
  await page.getByRole('textbox', { name: 'Filter inflight', exact: true }).fill('');
  await page.locator('body').click({ position: { x: 1, y: 1 } });

  // The registration belongs to home: another view leaves j alone, and
  // coming back registers it once, not twice.
  await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name: 'sys', exact: true }).click();
  await page.keyboard.press('j');
  await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name: 'home', exact: true }).click();
  await expect(focused).toContainText('first waiting');
  await page.keyboard.press('j');
  await expect(focused).toContainText('second waiting');
});
