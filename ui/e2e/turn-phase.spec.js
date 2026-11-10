import { test, expect } from '@playwright/test';
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';

// JOB-192: the live turn indicator under the last transcript row. Needs the
// isolated controller with the ACP test agent, and DESK_AGENT_DIR, the agent's
// working directory: the test agent's `phases <tag>` turn advances one phase
// each time the file `<tag>.<n>` appears there (see run-isolated.mjs).
test.skip(!process.env.DESK_MUTATE || !process.env.DESK_AGENT_DIR, 'requires an isolated controller and the ACP test agent');

async function post(request, path, data) {
  const response = await request.post(path, { headers: { 'X-Jobshop-Desk': '1' }, data });
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

const advance = (tag, n) => writeFileSync(join(process.env.DESK_AGENT_DIR, `${tag}.${n}`), '');

async function openChat(page, isMobile, id) {
  await page.goto('./?chat=' + id);
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
}

test('the indicator follows thinking, tool and reply, and clears when the turn ends', async ({ page, request, isMobile }) => {
  test.setTimeout(45_000);
  const tag = 'ph' + crypto.randomUUID().slice(0, 8);
  const name = 'phase-' + tag;
  const session = await post(request, 'api/chat/manage', { action: 'new', name, text: 'phases ' + tag });
  try {
    await openChat(page, isMobile, session.id);
    const phase = page.locator('.transcript .turn-phase');
    await expect(phase).toHaveText('thinking…');
    advance(tag, 1);
    await expect(phase).toHaveText('running: Phase tool ' + tag);
    advance(tag, 2);
    await expect(phase).toHaveText('replying…');
    await expect(page.locator('.transcript .markdown').last()).toContainText('answer ' + tag);
    advance(tag, 3);
    await expect(phase).toHaveCount(0);
    await expect(page.locator('.chat-state')).toHaveText('idle');
  } finally {
    for (const n of [1, 2, 3]) advance(tag, n);
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});

test('a reload during a tool call restores the indicator, and stop clears it', async ({ page, request, isMobile }) => {
  test.setTimeout(45_000);
  const tag = 'ph' + crypto.randomUUID().slice(0, 8);
  const session = await post(request, 'api/chat/manage', { action: 'new', name: 'phase-' + tag, text: 'phases ' + tag });
  try {
    await openChat(page, isMobile, session.id);
    const phase = page.locator('.transcript .turn-phase');
    await expect(phase).toHaveText('thinking…');
    advance(tag, 1);
    await expect(phase).toHaveText('running: Phase tool ' + tag);
    await page.reload();
    if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'chats', exact: true }).click();
    await expect(phase).toHaveText('running: Phase tool ' + tag);
    await post(request, 'api/chat/cancel', { session: session.id });
    await expect(phase).toHaveCount(0);
    await expect(page.locator('.chat-state')).toHaveText('idle');
  } finally {
    for (const n of [1, 2, 3]) advance(tag, n);
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});

test('the read-only session viewer shows the indicator too', async ({ page, request }) => {
  test.setTimeout(45_000);
  const tag = 'ph' + crypto.randomUUID().slice(0, 8);
  const session = await post(request, 'api/chat/manage', { action: 'new', name: 'phase-' + tag, text: 'phases ' + tag });
  try {
    await page.goto('./?view=trace&layout=graph&ref=' + session.id);
    await page.getByRole('button', { name: 'Inspect ' + session.id, exact: true }).click();
    const viewer = page.getByRole('region', { name: 'Read-only session' });
    await expect(viewer.locator('.turn-phase')).toHaveText('thinking…');
    advance(tag, 1);
    await expect(viewer.locator('.turn-phase')).toHaveText('running: Phase tool ' + tag);
    await post(request, 'api/chat/cancel', { session: session.id });
    await expect(viewer.locator('.turn-phase')).toHaveCount(0);
  } finally {
    for (const n of [1, 2, 3]) advance(tag, n);
    await post(request, 'api/chat/manage', { action: 'archive', id: session.id });
  }
});
