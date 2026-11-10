import { test, expect } from '@playwright/test';
import { showDetails } from './trace-view.js';

// Creates tickets only on an explicitly opted-in disposable controller.
test.skip(!process.env.DESK_MUTATE, 'set DESK_MUTATE=1 against an isolated controller');

test('trace search finds completed tickets beyond the first page and opens their graph', async ({ page, request, baseURL }) => {
  const created = [];
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, {
      headers: { 'X-Jobshop-Desk': '1' }, data,
    });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const key = 'search-' + crypto.randomUUID();
  const repo = 'trace-tests/' + key;
  try {
    const target = await call('ticket.create', {
      title: key + ' wallet evidence', meta: { github: { repo, pr: 4520 } },
    });
    await call('ticket.close', { id: target.id });
    for (let i = 0; i < 51; i++) {
      const ticket = await call('ticket.create', { title: key + ' recent ' + i });
      created.push(ticket.id);
    }
    await page.goto('./?view=trace');
    const form = page.getByRole('search', { name: 'Search tickets' });
    const input = form.getByRole('textbox', { name: 'Trace reference' });
    const results = page.getByRole('region', { name: 'Ticket search results' });
    await input.fill(key);
    await input.press('Enter');
    await expect(results.locator('tbody tr')).toHaveCount(50);
    await expect(results).not.toContainText('wallet evidence');
    await results.getByRole('button', { name: 'more', exact: true }).click();
    await expect(results.locator('tbody tr')).toHaveCount(52);
    const targetLink = results.getByRole('link', { name: target.title, exact: true });
    await expect(targetLink).toHaveAttribute('title', target.id);
    await targetLink.click();
    await showDetails(page);
    await expect(page.locator('.ticket-overview > h2')).toContainText(target.id);
    await expect(input).toHaveValue(target.id);
    await input.fill(repo + '#4520');
    await form.getByRole('combobox', { name: 'Ticket status' }).selectOption('closed');
    await input.press('Enter');
    await expect(results.locator('tbody tr')).toHaveCount(1);
    await expect(targetLink).toBeVisible();
    await form.getByRole('combobox', { name: 'Ticket status' }).selectOption('open');
    await input.press('Enter');
    await expect(results).toContainText('no matching tickets');
    await form.getByRole('combobox', { name: 'Ticket status' }).selectOption('');
    await input.fill(key + ' WaLlEt EviDence');
    await input.press('Enter');
    await expect(targetLink).toBeVisible();
    await results.getByRole('button', { name: 'dismiss', exact: true }).click();
    await expect(results).toHaveCount(0);
    await showDetails(page);
    await expect(page.locator('.ticket-overview > h2')).toContainText(target.id);
    await input.fill(repo + '#4520');
    await form.getByRole('button', { name: 'open', exact: true }).click();
    await showDetails(page);
    await expect(page.locator('.ticket-overview > h2')).toContainText(target.id);
    await expect(page).toHaveURL(new RegExp('ref=' + encodeURIComponent(repo + '#4520')));
    const bounds = await form.evaluate(node => ({ right: node.getBoundingClientRect().right, width: innerWidth }));
    expect(bounds.right).toBeLessThanOrEqual(bounds.width);
  } finally {
    for (const id of created) await call('ticket.close', { id });
  }
});
