import { test, expect } from '@playwright/test';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const exec = promisify(execFile);
test.skip(!process.env.DESK_MUTATE || !process.env.DESK_ROOT, 'set DESK_MUTATE=1 and DESK_ROOT against an isolated local controller with escalation chats');

test('converted escalation retains its transcript and claimed tickets expose a live read-only session', async ({ page, request, baseURL, isMobile }, testInfo) => {
  const call = async (method, data) => {
    const response = await request.post(new URL('api/rpc/' + method, baseURL).href, { headers: { 'X-Jobshop-Desk': '1' }, data });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const source = await call('ticket.create', { title: 'Workspace escalation ' + crypto.randomUUID(), labels: ['needs-human'] });
  let session;
  let task;
  try {
    await page.goto('./');
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
    const row = page.getByRole('dialog', { name: 'Choose chat' }).getByRole('row').filter({ hasText: source.id });
    await expect(row).toContainText('escalation');
    await row.getByRole('button').click();
    session = (await call('session.list', {})).find(info => info.ticket?.fields?.key === 'escalation:' + source.id).id;
    const actions = page.getByRole('toolbar', { name: 'Chat actions' });
    page.once('dialog', dialog => dialog.accept('ongoing-' + source.id));
    await actions.getByRole('button', { name: 'convert', exact: true }).click();
    await expect(actions.getByRole('button', { name: 'resolve', exact: true })).toHaveCount(0);
    await expect(actions.getByRole('button', { name: 'archive', exact: true })).toBeVisible();
    expect((await call('ticket.show', { id: source.id })).ticket.status).toBe('open');
    expect((await call('ticket.show', { id: session })).ticket.fields.named).toBe(true);
    const human = 'A human message\nMessages for you: this line is mine ' + crypto.randomUUID();
    const injected = 'Injected ticket context ' + crypto.randomUUID() + '\nA separate context line.';
    await call('session.send', { id: session, text: human, context: injected });
    const operator = page.locator('.operator').filter({ hasText: human });
    await expect(operator).toHaveText(human);
    await expect(operator).not.toContainText('<message from=');
    const context = page.locator('.controller-context').filter({ hasText: injected.split('\n')[0] });
    await expect(context).toHaveCount(1);
    await expect(context).not.toHaveAttribute('open');
    await context.locator('summary').click();
    await expect(context).toContainText(injected, { useInnerText: true });
    await expect(context.locator('.markdown br')).toHaveCount(1);
    task = await call('ticket.create', { title: 'Inspect live work ' + crypto.randomUUID() });
    await exec(process.env.JOBSHOP_BIN || 'jobshop', ['--root', process.env.DESK_ROOT, 'claim', task.id, '--holder', session], { env: { ...process.env, JOBSHOP_SOCKET: '' } });
    await page.goto('./?view=trace&ref=' + task.id);
    await page.locator(`button.tl-name[title$="(${task.id})"]`).click();
    const viewer = page.getByRole('region', { name: 'Read-only session' });
    await expect(viewer).toBeVisible();
    await expect(viewer.locator('.operator').filter({ hasText: human })).toHaveText(human);
    await expect(viewer.getByRole('textbox')).toHaveCount(0);
    await expect(viewer.getByRole('button', { name: /^(send|stop|attach)$/ })).toHaveCount(0);
    const next = 'Live follow-up ' + crypto.randomUUID();
    await call('session.send', { id: session, text: next });
    await expect(viewer.locator('.operator').filter({ hasText: next })).toHaveText(next);
    // Every history row is behind a toggle on the events tab (JOB-185).
    await page.getByLabel('every history row').check();
    const history = page.locator('.ticket-history');
    await expect(history).toContainText('claimed');
    await expect(history.getByRole('button', { name: 'view session', exact: true })).toBeVisible();
    await expect(history.locator('.history-raw[open]')).toHaveCount(0);
    if (isMobile) {
      await expect(viewer).toHaveCSS('position', 'fixed');
    } else {
      const details = await page.locator('.inspector').boundingBox();
      const output = await viewer.boundingBox();
      expect(output.x).toBeGreaterThanOrEqual(details.x + details.width - 1);
      expect(output.y + output.height).toBeLessThanOrEqual(await page.evaluate(() => innerHeight) + 1);
    }
    await page.screenshot({ path: testInfo.outputPath('live-session.png'), fullPage: true });
    await viewer.getByRole('button', { name: 'dismiss ×', exact: true }).click();
    await expect(viewer).toHaveCount(0);
    await page.screenshot({ path: testInfo.outputPath('ticket-history.png'), fullPage: true });
    await history.getByRole('button', { name: 'view session', exact: true }).click();
    await expect(viewer).toBeVisible();
    await page.reload();
    await page.getByRole('button', { name: 'Inspect ' + task.id, exact: true }).click();
    await expect(viewer.locator('.operator').filter({ hasText: human })).toHaveText(human);
    await call('ticket.close', { id: source.id });
    expect((await call('ticket.show', { id: session })).ticket.status).toBe('open');
  } finally {
    if (task) await call('ticket.close', { id: task.id });
    if (session) await call('chat.manage', { action: 'archive', id: session });
    const current = await call('ticket.show', { id: source.id });
    if (current.ticket.status !== 'closed') await call('ticket.close', { id: source.id });
  }
});
