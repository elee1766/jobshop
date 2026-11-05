import { test, expect } from '@playwright/test';

// Chat checks (#203) against a live controller. They start a real operator
// session and send it a message, so they only run with DESK_CHAT=1.
test.skip(!process.env.DESK_CHAT, 'set DESK_CHAT=1 to start a real operator session');

test('new operator session streams a reply to a message', async ({ page }) => {
  test.setTimeout(240_000);
  await page.goto('./?view=chat');
  await page.getByTestId('chat-new-input').fill('Desk e2e check: reply with this word and nothing else\nPONG');
  await page.getByTestId('chat-new').click();
  await expect(page).toHaveURL(/session=/);
  const journal = page.getByTestId('journal');
  // The first prompt shows as a turn, then the agent's streamed text.
  await expect(journal).toContainText('Desk e2e check', { timeout: 60_000 });
  await expect(journal.locator('.line.text')).toContainText('PONG', { timeout: 180_000 });

  // A follow-up in the same session, through the composer.
  await page.getByTestId('chat-input').fill('Now reply with this word\nPING');
  await page.getByTestId('chat-send').click();
  await expect(page.getByTestId('chat-input')).toHaveValue('');
  await expect(journal.locator('.line.text').last()).toContainText('PING', { timeout: 180_000 });
});

test('sessions view attaches: composer only where the pool allows rw', async ({ page }) => {
  await page.goto('./?view=sessions');
  await expect(page.getByTestId('sessions')).toBeVisible({ timeout: 15_000 });
  const row = page.locator('[data-testid=sessions] tr.click').first();
  test.skip((await row.count()) === 0, 'no sessions');
  const id = (await row.locator('td').first().innerText()).trim().split(/\s/)[0];
  const access = await (await page.request.get('api/session/' + id + '/access')).json();
  await row.click();
  await expect(page.getByTestId('journal')).toBeVisible();
  if (access.write) await expect(page.getByTestId('chat-input')).toBeVisible();
  else await expect(page.getByText(/^Read-only:/)).toBeVisible();
});

test('escalation drawer shows the escalation chat', async ({ page }) => {
  const items = await (await page.request.get('api/attention')).json();
  const esc = items.find((/** @type {any} */ i) => i.kind === 'ticket');
  test.skip(!esc, 'no open escalation on this controller');
  const t = await (await page.request.get('api/ticket/' + esc.id)).json();
  const chat = t.ticket.meta?.escalation;
  test.skip(!chat?.session, 'escalation has no chat (no escalation chat config)');
  await page.goto('./?ticket=' + esc.id);
  const panel = page.getByTestId('escalation-chat');
  await expect(panel).toBeVisible();
  await expect(panel).toContainText(chat.key);
  // The panel streams the escalation chat session.
  await expect(panel.locator('.line.turn').first()).toBeVisible({ timeout: 30_000 });
  if (!process.env.DESK_CHAT_ESC_SEND) return;
  // A reply from Desk goes by the chat's key, so it is one more turn there.
  const before = await panel.locator('.line.turn').count();
  await panel.getByTestId('chat-input').fill('desk reply\nDESK-HELLO');
  await panel.getByTestId('chat-send').click();
  await expect(panel.locator('.line.turn')).toHaveCount(before + 1, { timeout: 60_000 });
  await expect(panel.locator('.line.text').last()).toContainText('DESK-HELLO', { timeout: 180_000 });
});
