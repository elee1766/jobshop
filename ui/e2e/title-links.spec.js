import { test, expect } from '@playwright/test';

// Creates tickets only on an explicitly opted-in disposable controller.
test.skip(!process.env.DESK_MUTATE, 'set DESK_MUTATE=1 against an isolated controller');

test('Markdown links in titles open their destination and leave the row action alone (JOB-135)', async ({ page, request, baseURL }) => {
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const key = crypto.randomUUID().slice(0, 8);
  const created = [];
  try {
    const parent = await call('ticket.create', { title: `Parent ${key} [JOB-117](https://linear.app/gfx-labs/issue/JOB-117) and [jobshop#1021](https://github.com/gfx-labs/jobshop/pull/1021)` });
    created.push(parent.id);
    const plain = await call('ticket.create', { title: `Plain ${key} *not italic* <b>not bold</b>`, parent_id: parent.id });
    const unsafe = await call('ticket.create', { title: `Unsafe ${key} [click](javascript:alert%281%29) [data](data:text/html,x)`, parent_id: parent.id });
    created.push(plain.id, unsafe.id);
    await page.goto(`./?view=trace&tab=relations&ref=${parent.id}`);
    const relationships = page.getByLabel('Ticket relationships', { exact: true });
    const heading = page.locator('.ticket-overview > h2');
    await expect(heading.getByRole('link', { name: 'JOB-117', exact: true })).toHaveAttribute('href', 'https://linear.app/gfx-labs/issue/JOB-117');
    await expect(heading.getByRole('link', { name: 'jobshop#1021', exact: true })).toHaveAttribute('target', '_blank');
    // Plain text stays literal and unsafe destinations keep their label without a link.
    await expect(relationships).toContainText(`Plain ${key} *not italic* <b>not bold</b>`);
    await expect(relationships.locator('b')).toHaveCount(0);
    await expect(relationships).toContainText(`Unsafe ${key} click data`);
    await expect(relationships.getByRole('link')).toHaveCount(0);
    // The timeline row has a readable tooltip and no nested controls.
    await expect(page.locator('button.tl-name').first()).toHaveAttribute('title', `Parent ${key} JOB-117 and jobshop#1021 (${parent.id})`);
    expect(await page.locator('a a, a button, button a, button button').count()).toBe(0);
    // Clicking a title link opens its destination and does not select the ticket.
    await relationships.getByRole('button', { name: `Plain ${key} *not italic* <b>not bold</b>`, exact: false }).click();
    await expect(page.locator('.ticket-overview > h2')).toContainText('Plain ' + key);
    await page.goto(`./?view=trace&ref=${plain.id}`);
    const child = page.getByLabel('Ticket details', { exact: true }).locator('.ticket-parent');
    await expect(child.getByRole('button', { name: new RegExp(`^Parent ${key} JOB-117 and jobshop#1021`) })).toBeVisible();
    await child.getByRole('link', { name: 'JOB-117', exact: true }).evaluate((a) => a.addEventListener('click', (e) => e.preventDefault()));
    await child.getByRole('link', { name: 'JOB-117', exact: true }).click();
    await expect(page.locator('.ticket-overview > h2')).toContainText('Plain ' + key);
    await child.getByRole('button', { name: new RegExp(`^Parent ${key}`) }).click();
    await expect(page.locator('.ticket-overview > h2')).toContainText('Parent ' + key);
    // Title search matches the readable text, not the Markdown source.
    await page.getByRole('textbox', { name: 'Trace reference', exact: true }).fill(`Parent ${key}`);
    await page.getByRole('textbox', { name: 'Trace reference', exact: true }).press('Enter');
    const results = page.getByRole('region', { name: 'Ticket search results' });
    await expect(results.getByRole('link', { name: 'JOB-117', exact: true })).toHaveAttribute('href', 'https://linear.app/gfx-labs/issue/JOB-117');
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
  } finally {
    for (const id of created) await call('ticket.close', { id });
  }
});
