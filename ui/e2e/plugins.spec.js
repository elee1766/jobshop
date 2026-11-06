import { test, expect } from '@playwright/test';

// Desk plugins (#517). Needs Desk started with the fixtures installed:
//   cp -r ui/e2e/plugins/* <root>/desk/plugins/
// and DESK_PLUGINS=1. e2e/plugins/deploys stands in for packs/deploy's
// Deploys view; with that pack loaded too, the root copy replaces it.
test.skip(!process.env.DESK_PLUGINS, 'set DESK_PLUGINS=1 with ui/e2e/plugins installed');

test('a user plugin view calls api.get and api.call; claim is refused', async ({ page }) => {
  await page.goto('./?view=e2e-test');
  await expect(page.locator('nav a[data-view=e2e-test]')).toHaveText(/E2E plugin/);
  await expect(page.getByTestId('e2e-get')).toHaveText(/^get ok/);
  await expect(page.getByTestId('e2e-call')).toHaveText('call ok version');
  await expect(page.getByTestId('e2e-post')).toHaveText(/claim refused .*not allowed/);
});

test('a plugin that throws shows an error card and Desk keeps working', async ({ page }) => {
  await page.goto('./?view=e2e-throws');
  await expect(page.getByTestId('plugin-error')).toContainText('e2e plugin import failure');
  await page.locator('nav a[data-view=work]').click();
  await expect(page.getByTestId('runs')).toBeVisible();
});

test('a root plugin adds a view', async ({ page }) => {
  await page.goto('./?view=deploys');
  await expect(page.getByTestId('user-deploys')).toBeVisible();
  await expect(page.locator('nav a[data-view=deploys]')).toHaveText(/Deploys \(user\)/);
});

test('the drawer shows a plugin section', async ({ page }) => {
  await page.goto('./?view=work');
  await expect(page.getByTestId('runs')).toBeVisible();
  const row = page.locator('[data-testid=runs] tr.click').first();
  test.skip((await row.count()) === 0, 'no runs on this controller');
  await row.click();
  await expect(page.getByTestId('e2e-drawer')).toHaveText(/^drawer for /);
});

// JOB-87: e2e-slots adds to Work and Chat through their plugin-defined
// slots; e2e-slot-throws fails in work:toolbar beside it.
test('a plugin adds a toolbar button to Work; a failing one shows an error card in its box only', async ({ page }) => {
  await page.goto('./?view=work');
  await expect(page.getByTestId('runs')).toBeVisible();
  const toolbar = page.getByTestId('work-toolbar');
  const err = toolbar.locator('[data-plugin=e2e-slot-throws] [data-testid=plugin-error]');
  await expect(err).toContainText('e2e slot import failure');
  await expect(page.getByTestId('plugin-error')).toHaveCount(1);
  await page.getByTestId('e2e-slot-button').click();
  await expect(page.getByTestId('toasts')).toContainText('e2e toolbar all');
  await expect(page.getByTestId('e2e-slot-aside')).toHaveText(/^aside runs \d+$/);
  // The view around the slot still works: a filter chip changes the context.
  await page.getByRole('group', { name: 'state filter' }).getByRole('button', { name: 'done' }).click();
  await page.getByTestId('e2e-slot-button').click();
  await expect(page.getByTestId('toasts')).toContainText('e2e toolbar done');
});

test('a plugin adds a panel to Chat', async ({ page }) => {
  await page.goto('./?view=chat');
  await expect(page.getByTestId('chat-list')).toBeVisible();
  await expect(page.getByTestId('chat-panels').getByTestId('e2e-slot-panel')).toHaveText('E2E chat panel');
  await expect(page.getByTestId('chat-new-input')).toBeVisible();
});

test('plugin listing and RPC policy over HTTP', async ({ request }) => {
  const list = await (await request.get('api/plugins')).json();
  const names = list.map((/** @type {any} */ p) => p.name);
  expect(names).toEqual(expect.arrayContaining(['deploys', 'e2e-test', 'e2e-throws', 'e2e-slots', 'e2e-slot-throws']));
  expect(list.find((/** @type {any} */ p) => p.name === 'deploys').source).toBe('root');
  const res = await request.post('api/rpc/claim', { data: {}, headers: { 'X-Jobshop-Desk': '1' } });
  expect(res.status()).toBe(403);
  expect((await res.json()).code).toBe('forbidden');
});

// #542: needs desk_commands.e2e-echo (command: cat) in jobshop.yaml.
test('a plugin creates a ticket with meta, lists it, gets its event, runs a desk command and reads a session stream', async ({ page }) => {
  await page.goto('./?view=e2e-flow');
  await expect(page.getByTestId('flow-create')).toHaveText('create ok');
  await expect(page.getByTestId('flow-list')).toHaveText('list ok');
  await expect(page.getByTestId('flow-event')).toHaveText('event ok comment', { timeout: 15000 });
  await expect(page.getByTestId('flow-exec')).toHaveText('exec ok');
  await expect(page.getByTestId('flow-stream')).toHaveText(/^stream (ok|skip)/);
});

test('GET api/rpc has a schema per allowed method', async ({ request }) => {
  const list = await (await request.get('api/rpc')).json();
  expect(list.methods).toContain('desk.exec');
  expect(list.schemas['ticket.create'].params.$ref).toBe('#/$defs/rpc.CreateParams');
  expect(list.$defs['rpc.CreateParams'].required).toContain('title');
});
