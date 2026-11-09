import { test, expect } from '@playwright/test';

// JOB-136: ticket context menu on home rows, trace cards and relationship
// entries. Creates and closes tickets, so only on an isolated controller.
test.skip(!process.env.DESK_MUTATE || !process.env.DESK_LINK_WORKFLOW, 'set DESK_MUTATE=1 and DESK_LINK_WORKFLOW to a disposable workflow accepting repo, pr and issue variables');

const inside = (box, viewport) =>
  box.x >= 0 && box.y >= 0 && box.x + box.width <= viewport.width && box.y + box.height <= viewport.height;

test('ticket menu: home row, trace card and relationship entry', async ({ page, request, baseURL, isMobile, context, browserName }) => {
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  if (browserName === 'chromium') await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  const { run } = await call('run.start', { workflow: process.env.DESK_LINK_WORKFLOW, vars: { repo: 'elee1766/jobshop', pr: '977', issue: 'JOB-136' } });
  const child = await call('ticket.create', { title: 'menu child ' + run.id, parent_id: run.id });
  const menu = page.getByRole('menu', { name: /^Actions for / });
  const footer = menu.locator('.ticket-menu-id code');
  const clipboard = () => page.evaluate(() => navigator.clipboard.readText());
  try {
    // Home: the visible button works everywhere, right click on desktop.
    await page.goto('./');
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'in flight', exact: true }).click();
    const row = page.getByRole('region', { name: 'In flight', exact: true }).getByRole('row').filter({ has: page.getByRole('button', { name: 'Actions for ' + run.id }) });
    await expect(row).toBeVisible();
    const before = page.url();

    // A real link keeps the browser's own menu.
    await row.locator('a[href*="/pull/"]').first().dispatchEvent('contextmenu');
    await expect(menu).toHaveCount(0);

    if (isMobile) await row.getByRole('button', { name: 'Actions for ' + run.id }).tap();
    else await row.locator('.ticket-label').click({ button: 'right' });
    await expect(menu).toBeVisible();
    await expect(footer).toHaveText(run.id);
    await expect(menu.getByRole('menuitem', { name: 'open in trace' })).toBeVisible();
    await expect(menu.getByRole('menuitem', { name: 'inspect' })).toHaveCount(0);
    const box = await menu.boundingBox();
    expect(inside(box, page.viewportSize())).toBe(true);
    // Opening changed nothing.
    expect(page.url()).toBe(before);

    // Copy the exact id and say so.
    await menu.getByRole('menuitem', { name: 'copy ticket id' }).click();
    await expect(menu).toHaveCount(0);
    await expect(page.locator('.notice')).toHaveText('copied ' + run.id);
    if (browserName === 'chromium') expect(await clipboard()).toBe(run.id);

    // Escape closes it and returns focus to the button that opened it.
    const trigger = row.getByRole('button', { name: 'Actions for ' + run.id });
    await trigger.focus();
    await page.keyboard.press('Enter');
    await expect(menu).toBeVisible();
    await page.keyboard.press('ArrowDown');
    await page.keyboard.press('Escape');
    await expect(menu).toHaveCount(0);
    await expect(trigger).toBeFocused();
    expect(page.url()).toBe(before);

    // A click outside closes it and does not activate what was under it.
    await trigger.click();
    await expect(menu).toBeVisible();
    await page.mouse.click(5, 5);
    await expect(menu).toHaveCount(0);
    expect(page.url()).toBe(before);

    // Open in trace goes to this ticket.
    await trigger.click();
    await menu.getByRole('menuitem', { name: 'open in trace' }).click();
    await expect(page).toHaveURL(new RegExp('ref=' + run.id));

    // Trace: cards and relationships carry the same menu with their own id.
    const card = page.locator(`.graph-card[data-ticket="${child.id}"]`);
    await expect(card).toBeVisible();
    await card.getByRole('button', { name: 'Actions for ' + child.id }).click();
    await expect(footer).toHaveText(child.id);
    await expect(menu.getByRole('menuitem', { name: 'inspect' })).toBeVisible();
    await expect(menu.getByRole('menuitem', { name: 'open as graph root' })).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(menu).toHaveCount(0);

    const entry = page.getByRole('region', { name: 'children', exact: true }).getByRole('listitem').filter({ hasText: child.id });
    await expect(entry).toBeVisible();
    if (!isMobile) await entry.click({ button: 'right', position: { x: 4, y: 4 } });
    else await entry.getByRole('button', { name: 'Actions for ' + child.id }).tap();
    await expect(footer).toHaveText(child.id);
    // The menu acts on the ticket it was opened on, not the selection.
    await menu.getByRole('menuitem', { name: 'inspect' }).click();
    await expect(page).toHaveURL(new RegExp('inspect=' + child.id));
    await expect(page).toHaveURL(new RegExp('ref=' + run.id));

    // The inspected ticket no longer offers inspect.
    await page.locator(`.graph-card[data-ticket="${run.id}"]`).getByRole('button', { name: 'Actions for ' + run.id }).click();
    await expect(footer).toHaveText(run.id);
    await expect(menu.getByRole('menuitem', { name: 'inspect' })).toBeVisible();
    await expect(menu.getByRole('menuitem', { name: 'open as graph root' })).toHaveCount(0);
    await page.keyboard.press('Escape');
  } finally {
    await call('run.cancel', { id: run.id });
    await call('ticket.close', { id: child.id }).catch(() => {});
  }
});

test('ticket menu stays inside the viewport at the edges', async ({ page, request, baseURL, isMobile }) => {
  test.skip(isMobile, 'pointer placement is a desktop concern');
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const { run } = await call('run.start', { workflow: process.env.DESK_LINK_WORKFLOW, vars: { repo: 'elee1766/jobshop', pr: '976', issue: 'JOB-136' } });
  try {
    await page.setViewportSize({ width: 700, height: 420 });
    await page.goto('./');
    const row = page.getByRole('region', { name: 'In flight', exact: true }).getByRole('row').filter({ has: page.getByRole('button', { name: 'Actions for ' + run.id }) });
    await expect(row).toBeVisible();
    for (const [x, y] of [[690, 410], [5, 410], [690, 5]]) {
      await page.evaluate(([cx, cy]) => {
        const target = document.querySelector('.ticket-label');
        target.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true, clientX: cx, clientY: cy }));
      }, [x, y]);
      const menu = page.getByRole('menu', { name: /^Actions for / });
      await expect(menu).toBeVisible();
      expect(inside(await menu.boundingBox(), page.viewportSize())).toBe(true);
      await page.keyboard.press('Escape');
      await expect(menu).toHaveCount(0);
    }
  } finally {
    await call('run.cancel', { id: run.id });
  }
});
