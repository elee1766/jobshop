import { test, expect } from '@playwright/test';
import { showDetails } from './trace-view.js';

// JOB-137: sender headers in chat and the read-only session viewer. Needs an
// isolated controller with the ACP test agent (see chat-responsiveness.spec.js).
test.skip(!process.env.DESK_MUTATE, 'requires an isolated controller and the ACP test agent');

async function post(request, path, data) {
  const response = await request.post(path, { headers: { 'X-Jobshop-Desk': '1' }, data });
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

test('human and agent messages carry sender headers in chat and the session viewer', async ({ page, request, isMobile }) => {
  test.setTimeout(45_000);
  const name = 'headers-' + crypto.randomUUID();
  const session = await post(request, 'api/chat/manage', { action: 'new', name, text: 'first ' + name });
  const key = session.ticket.fields.key;
  try {
    await page.goto('./?chat=' + session.id);
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    const transcript = page.locator('.transcript');
    await expect(transcript.locator('.markdown').last()).toContainText('echo: first');
    for (const text of ['second ' + name, 'tool run three ' + name]) {
      await post(request, 'api/chat/send', { key, text });
      await expect(transcript.locator('.markdown').last()).toContainText('echo: ' + text);
    }
    // Every human message has a `you` label above it, not just a background.
    const humans = transcript.locator('.message');
    await expect(humans).toHaveCount(3);
    for (const human of await humans.all()) {
      await expect(human.locator('.speaker-human')).toHaveText('you');
      const bar = await human.locator('.speaker-human').boundingBox();
      const body = await human.locator('.operator').boundingBox();
      expect(bar.y + bar.height).toBeLessThanOrEqual(body.y + 1);
    }
    await expect(humans.first().locator('.operator')).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');

    // An agent turn is labeled once, however many tool rows and replies it has.
    const agents = transcript.locator('.speaker-agent');
    await expect(agents).toHaveCount(3);

    // The raw prompt is a muted disclosure with no header of its own.
    const prompt = transcript.locator('.raw-prompt').first();
    await expect(prompt).toBeVisible();
    await expect(transcript.locator('.speaker')).toHaveCount(6);

    // Same presentation in the read-only viewer.
    await page.goto('./?view=trace&ref=' + session.id);
    await showDetails(page);
    await page.getByRole('button', { name: 'view session ↗', exact: true }).click();
    const viewer = page.getByRole('region', { name: 'Read-only session' });
    await expect(viewer.locator('.speaker-human').first()).toHaveText('you');
    await expect(viewer.locator('.speaker-agent').first()).toBeVisible();
  } finally {
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});

// A ticket turn has no human message: its whole prompt is controller context,
// so the raw prompt would repeat it and is not offered.
test('a ticket turn shows its prompt once', async ({ page, request }) => {
  test.setTimeout(45_000);
  const pool = process.env.DESK_SCALE_POOL || 'worker';
  const task = await post(request, 'api/rpc/ticket.create', { title: 'Prompt once ' + crypto.randomUUID(), pool });
  let holder;
  await expect
    .poll(async () => {
      const shown = await post(request, 'api/rpc/ticket.show', { id: task.id });
      holder = (shown.history || []).find((row) => row.op === 'claim')?.actor;
      return shown.ticket.status;
    }, { timeout: 30_000 })
    .toBe('closed');
  await page.goto('./?view=trace&ref=' + holder);
  await showDetails(page);
  await page.getByRole('button', { name: 'view session ↗', exact: true }).click();
  const viewer = page.getByRole('region', { name: 'Read-only session' });
  const turn = viewer.locator('.transcript-row').filter({ hasText: task.id });
  await expect(turn.locator('.controller-context')).toHaveCount(1);
  await expect(turn.locator('.raw-prompt')).toHaveCount(0);
});
