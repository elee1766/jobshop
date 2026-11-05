import { test, expect } from '@playwright/test';

// Attention opens the chat that resolves it, and the chat follows
// jcode-web (#535). The API is mocked, so this runs against any Desk
// build (`npx vite preview`, or a live Desk) and under DESK_PREFIX.

const now = Date.now();
const iso = (/** @type {number} */ msAgo) => new Date(now - msAgo).toISOString();

const builtins = ['attention', 'chat'];

/** @param {string} url */
const pathOf = (url) => new URL(url).pathname.replace(/^.*?(?=\/api\/)/, '');

/** An SSE body: events with ids, then optionally end. @param {any[]} evs @param {{ end?: boolean }} [o] */
function sse(evs, o = {}) {
  let b = 'retry: 50\n\n';
  for (const ev of evs) b += (ev.seq ? 'id: ' + ev.seq + '\n' : '') + 'event: session\ndata: ' + JSON.stringify(ev) + '\n\n';
  if (o.end) b += 'event: end\ndata: {}\n\n';
  return b;
}

/**
 * Mocks Desk's API. streams maps a session to a list of SSE bodies, one
 * per connection (the last repeats); a body may be a function awaited
 * first, to hold a phase until the test is ready.
 * @param {import('@playwright/test').Page} page
 * @param {{ attention?: () => any[], escalations?: any[], streams?: Record<string, (string | (() => Promise<string>))[]>, older?: (q: URLSearchParams) => any, chat?: (b: any) => any, rpc?: (method: string, body: any) => any }} o
 */
async function mock(page, o = {}) {
  const calls = /** @type {{ path: string, body: any }[]} */ ([]);
  const conns = /** @type {Record<string, number>} */ ({});
  // Plugin list as Desk serves the embedded built-ins.
  const plugins = await Promise.all(builtins.map(async (n) => {
    const r = await page.request.get('plugins/' + n + '/plugin.json');
    return { ...(await r.json()), source: 'builtin', url: 'plugins/' + n + '/' };
  }));
  await page.route('**/api/**', async (route) => {
    const req = route.request();
    const path = pathOf(req.url());
    const body = req.postData() ? JSON.parse(req.postData() || '{}') : null;
    calls.push({ path, body });
    /** @param {any} b */
    const json = (b) => route.fulfill({ json: b });
    if (path === '/api/plugins') return json(plugins);
    if (path === '/api/attention') return json(o.attention ? o.attention() : []);
    if (path === '/api/chat') return json({ pool: 'operator', sessions: [{ id: 'js-s1', key: 'escalation:js-e1', alive: true }, { id: 'js-s2', key: 'desk:abc', alive: true }, { id: 'js-s3', key: 'attention:stale:js-t9', alive: true }] });
    if (path === '/api/sessions') return json([{ id: 'js-s2', alive: true, model: 'claude-x', ticket: { pool: 'operator' } }]);
    if (path === '/api/escalations') return json(o.escalations ?? []);
    if (path === '/api/work') return json({ runs: [], batches: [] });
    if (path === '/api/push/key') return route.fulfill({ status: 404, json: { error: 'push off' } });
    if (path.endsWith('/access')) return json({ write: true });
    if (path.startsWith('/api/ticket/')) return json({ ticket: { id: path.split('/').pop(), title: 'ticket ' + path.split('/').pop(), kind: 'task' }, comments: [] });
    if (path === '/api/attention/chat') return json(o.chat ? o.chat(body) : { session: 'js-s1' });
    if (o.rpc && path.startsWith('/api/rpc/')) {
      const r = o.rpc(path.slice('/api/rpc/'.length), body);
      if (r) return route.fulfill(r);
    }
    if (path === '/api/chat/send' || path === '/api/chat/cancel' || path.startsWith('/api/rpc/') || path === '/api/attention/ack') {
      return json(path === '/api/chat/cancel' ? { id: body.session, cancelled: true } : { ok: true });
    }
    const m = /^\/api\/session\/([^/]+)\/(stream|events)$/.exec(path);
    if (m && m[2] === 'events') return json(o.older ? o.older(new URL(req.url()).searchParams) : { events: [], first: 1 });
    if (m) {
      const list = o.streams?.[m[1]] ?? [sse([], { end: true })];
      const n = (conns[m[1]] = (conns[m[1]] ?? -1) + 1);
      const b = list[Math.min(n, list.length - 1)];
      return route.fulfill({ contentType: 'text/event-stream', body: typeof b === 'function' ? await b() : b });
    }
    if (path === '/api/events') return route.fulfill({ contentType: 'text/event-stream', body: 'retry: 60000\n\n' });
    return json([]);
  });
  return calls;
}

/** A send from Desk as the controller delivers it. @param {number} seq @param {string} text */
const prompt = (seq, text) => ({ seq, type: 'turn_start', text: 'Messages for you:\n\nFrom desk:alice (js-m' + seq + '):\n' + text, time: iso(1000) });
const say = (/** @type {number} */ seq, /** @type {string} */ text) => ({ seq, type: 'text', text, time: iso(900) });
const end = (/** @type {number} */ seq) => ({ seq, type: 'turn_end', stop_reason: 'end_turn', time: iso(800) });

test('composer: Enter sends, Shift+Enter is a newline, Up recalls the last prompt', async ({ page }) => {
  const calls = await mock(page, { streams: { 'js-s2': [sse([prompt(5, 'first question'), say(6, 'an answer'), end(7)], { end: true })] } });
  await page.goto('./?view=chat&session=js-s2');
  const input = page.getByTestId('chat-input');
  await expect(page.getByTestId('journal')).toContainText('an answer');
  await expect(page.getByTestId('chat-send')).toHaveText('Send');

  await input.fill('line one');
  await input.press('Shift+Enter');
  await input.pressSequentially('line two');
  await expect(input).toHaveValue('line one\nline two');
  await input.press('Enter');
  await expect(input).toHaveValue('');
  const sent = calls.filter((c) => c.path === '/api/chat/send');
  expect(sent.map((c) => c.body)).toEqual([{ key: 'desk:abc', text: 'line one\nline two', interrupt: false }]);
  // The sent message shows at once, before the session's turn echoes it (#603).
  await expect(page.getByTestId('chat-pending')).toContainText('line one');
  await expect(page.getByTestId('chat-pending')).toContainText('sent, waiting for the session');

  // Up on the first line recalls the newest sent prompt, then older ones.
  await input.press('ArrowUp');
  await expect(input).toHaveValue('line one\nline two');
  await input.press('ArrowUp');
  await expect(input).toHaveValue('first question');
  await input.press('ArrowDown');
  await expect(input).toHaveValue('line one\nline two');
});

test('during a turn: Stop, Queue and Interrupt; a tool row updates in place; the queued message follows', async ({ page }) => {
  /** @type {(v?: any) => void} */
  let release = () => {};
  const queued = new Promise((r) => (release = r));
  const calls = await mock(page, {
    streams: {
      'js-s2': [
        sse([
          prompt(11, 'look at the logs'),
          { seq: 12, type: 'reasoning', text: 'checking the journal', time: iso(700) },
          { seq: 13, type: 'tool', tool: { id: 't1', title: 'bash', status: 'in_progress', input: '{"command":"tail log","intent":"Read the log"}' }, time: iso(600) },
        ]),
        async () => {
          await queued;
          return sse([
            { seq: 14, type: 'tool', tool: { id: 't1', title: 'bash', status: 'completed', output: 'last line' }, time: iso(500) },
            say(15, 'the log is clean'),
            end(16),
            prompt(17, 'and the queue?'),
            say(18, 'queue is empty'),
            end(19),
          ], { end: true });
        },
      ],
    },
  });
  await page.goto('./?view=chat&session=js-s2');
  const tool = page.locator('[data-tool=t1]');
  await expect(tool.getByTestId('tool-state')).toHaveText('running');
  await expect(tool).toContainText('Read the log');
  await expect(page.getByTestId('block-reasoning')).toContainText('checking the journal');
  await expect(page.getByTestId('chat-working')).toContainText('bash');
  await expect(page.getByTestId('chat-stop')).toBeVisible();
  await expect(page.getByTestId('chat-interrupt')).toBeVisible();
  await expect(page.getByTestId('chat-send')).toHaveText('Queue');
  const handle = await tool.elementHandle();

  // Stop sends nothing but a cancel.
  await page.getByTestId('chat-stop').click();
  await expect.poll(() => calls.filter((c) => c.path === '/api/chat/cancel').length).toBe(1);

  // Queue: sent without interrupt, delivered after the turn.
  await page.getByTestId('chat-input').fill('and the queue?');
  await page.getByTestId('chat-send').click();
  await expect(page.getByTestId('chat-notice')).toHaveText('Queued for after this turn');
  expect(calls.filter((c) => c.path === '/api/chat/send').map((c) => c.body)).toEqual([{ key: 'desk:abc', text: 'and the queue?', interrupt: false }]);
  // Ctrl+Enter is Interrupt.
  await page.getByTestId('chat-input').fill('stop that');
  await page.getByTestId('chat-input').press('Control+Enter');
  await expect.poll(() => calls.filter((c) => c.path === '/api/chat/send').at(-1)?.body.interrupt).toBe(true);
  release();

  await expect(tool.getByTestId('tool-state')).toHaveText('done');
  expect(await handle?.evaluate((el) => el.isConnected)).toBe(true);
  await expect(page.locator('[data-tool=t1]')).toHaveCount(1);
  await tool.locator('button').click();
  await expect(tool).toContainText('last line');
  const users = page.getByTestId('block-user');
  await expect(users.last()).toContainText('and the queue?');
  await expect(page.getByTestId('journal')).toContainText('queue is empty');
  await expect(page.getByTestId('chat-stop')).toHaveCount(0);
  await expect(page.getByTestId('chat-send')).toHaveText('Send');
});

test('Load older keeps the scroll position', async ({ page }) => {
  const live = [];
  for (let i = 0; i < 30; i++) live.push(prompt(301 + i * 2, 'live question ' + i), say(302 + i * 2, 'live answer ' + i));
  await mock(page, {
    streams: { 'js-s2': [sse(live, { end: true })] },
    older: (q) => {
      expect(q.get('before')).toBe('301');
      const evs = [];
      for (let i = 0; i < 40; i++) evs.push(prompt(100 + i * 2, 'old question ' + i), say(101 + i * 2, 'old answer ' + i));
      return { events: evs, first: 1 };
    },
  });
  await page.goto('./?view=chat&session=js-s2');
  const journal = page.getByTestId('journal');
  await expect(journal).toContainText('live answer 29');
  await journal.evaluate((el) => { el.scrollTop = 0; });
  const btn = page.getByTestId('load-older');
  await expect(btn).toHaveText('Load older messages (300 earlier)');
  const marker = page.getByTestId('block-user').filter({ hasText: 'live question 0' });
  const before = /** @type {{ y: number }} */ (await marker.boundingBox()).y;
  await btn.click();
  await expect(journal).toContainText('old answer 0');
  const after = /** @type {{ y: number }} */ (await marker.boundingBox()).y;
  expect(Math.abs(after - before)).toBeLessThanOrEqual(2);
  await expect(btn).toHaveCount(0);
});

test('the composer stays visible with the on-screen keyboard open', async ({ page, isMobile }) => {
  test.skip(!isMobile, 'phone layout');
  await mock(page, { streams: { 'js-s2': [sse([prompt(5, 'hi'), say(6, 'hello'), end(7)], { end: true })] } });
  await page.goto('./?view=chat&session=js-s2');
  const input = page.getByTestId('chat-input');
  await input.focus();
  // The keyboard takes the lower part of the screen: the visual viewport
  // shrinks and the shell follows it (visualViewport resize).
  const vp = /** @type {{ width: number, height: number }} */ (page.viewportSize());
  await page.setViewportSize({ width: vp.width, height: Math.round(vp.height * 0.55) });
  await expect(input).toBeInViewport();
  await expect(page.getByTestId('chat-send')).toBeInViewport();
  const box = /** @type {{ y: number, height: number }} */ (await page.getByTestId('chat-send').boundingBox());
  expect(box.y + box.height).toBeLessThanOrEqual(Math.round(vp.height * 0.55) + 1);
});

test('an escalation item opens its existing chat', async ({ page }) => {
  const calls = await mock(page, {
    attention: () => [{ kind: 'ticket', id: 'js-e1', title: 'Escalate run js-r1', why: 'escalation', since: iso(3600e3) }],
    chat: () => ({ kind: 'ticket', id: 'js-e1', key: 'escalation:js-e1', session: 'js-s1', pool: 'ops' }),
    streams: { 'js-s1': [sse([{ seq: 3, type: 'turn_start', text: '', context: 'You are the escalation chat for js-e1', time: iso(1000) }], { end: true })] },
  });
  await page.goto('./');
  await page.getByTestId('open-chat').first().click();
  await expect(page.getByTestId('session-panel')).toBeVisible();
  await expect(page).toHaveURL(/session=js-s1/);
  await expect(page).toHaveURL(/item=ticket%3Ajs-e1/);
  await expect(page.getByTestId('drawer')).toHaveCount(0);
  expect(calls.find((c) => c.path === '/api/attention/chat')?.body).toMatchObject({ kind: 'ticket', id: 'js-e1' });
  await expect(page.getByTestId('chat-item')).toContainText('ticket');
  await expect(page.getByTestId('resolve')).toHaveText('Close ticket');
  // The pack prompt is context, not an operator message.
  await expect(page.getByTestId('block-user')).toHaveCount(0);
  await expect(page.getByTestId('block-system')).toContainText('escalation chat for js-e1');
});

test('a question item: the answer goes through question.answer and the item clears', async ({ page }) => {
  let answered = false;
  const calls = await mock(page, {
    attention: () => (answered ? [] : [{ kind: 'question', id: 'js-q1', title: 'Which branch?', why: 'question', since: iso(600e3) }]),
    chat: () => ({ kind: 'question', id: 'js-q1', key: 'attention:question:js-q1', session: 'js-s3', started: true }),
    streams: { 'js-s3': [sse([prompt(2, 'Desk attention item: question js-q1'), end(3)], { end: true })] },
  });
  await page.goto('./');
  await page.getByTestId('open-chat').first().click();
  await expect(page.getByTestId('answer-toggle')).toBeVisible();
  await expect(page.getByTestId('answer-toggle').locator('input')).toBeChecked();
  await expect(page.getByTestId('chat-send')).toHaveText('Answer');
  answered = true;
  await page.getByTestId('chat-input').fill('use main');
  await page.getByTestId('chat-input').press('Enter');
  await expect(page.getByTestId('chat-notice')).toHaveText('Answer sent to the asking session');
  const a = calls.find((c) => c.path === '/api/rpc/question.answer');
  expect(a?.body).toEqual({ id: 'js-q1', text: 'use main' });
  expect(calls.filter((c) => c.path === '/api/chat/send')).toHaveLength(0);
  // Toggled off, the composer talks to the operator agent instead.
  await page.getByTestId('answer-toggle').locator('input').uncheck();
  await expect(page.getByTestId('chat-send')).toHaveText('Send');
  await page.getByTestId('session-panel-close').click();
  await page.goto('./');
  await expect(page.getByTestId('attention-empty')).toBeVisible();
});

test('a decision item: approve, reject and dismiss call decision.resolve; an error shows a toast', async ({ page }) => {
  let fail = false;
  const calls = await mock(page, {
    attention: () => [{ kind: 'decision', id: 'js-d1', title: 'Proposed target: v0.5.0 (#56)', why: 'proposed', since: iso(600e3) }],
    chat: () => ({ kind: 'decision', id: 'js-d1', key: 'attention:decision:js-d1', session: 'js-s3' }),
    streams: { 'js-s3': [sse([prompt(2, 'Desk attention item: decision js-d1')], { end: true })] },
    rpc: (m) => (fail && m === 'decision.resolve' ? { status: 502, json: { error: 'github POST labels: 403 Forbidden' } } : null),
  });
  await page.goto('./');
  for (const [label, action] of [['Approve as target', 'approve'], ['Reject', 'reject'], ['Dismiss', 'dismiss']]) {
    await page.getByTestId('open-chat').first().click();
    await expect(page.getByTestId('resolve')).toHaveText(['Approve as target', 'Reject', 'Dismiss']);
    await page.getByTestId('resolve').filter({ hasText: label }).click();
    // The confirm dialog shows above the chat panel and takes the click.
    await expect(page.getByTestId('confirm')).toBeVisible();
    await page.getByTestId('confirm').getByRole('button', { name: 'OK' }).click();
    await expect(page.getByTestId('toasts')).toContainText(label + ': done');
    const sent = calls.filter((c) => c.path === '/api/rpc/decision.resolve').pop();
    expect(sent?.body).toEqual({ id: 'js-d1', action });
    // A resolution closes the panel and returns to the list (#607).
    await expect(page.getByTestId('session-panel')).toHaveCount(0);
  }
  expect(calls.some((c) => c.path === '/api/rpc/label.add' || c.path === '/api/rpc/ticket.close')).toBe(false);
  fail = true;
  await page.getByTestId('open-chat').first().click();
  await page.getByTestId('resolve').filter({ hasText: 'Dismiss' }).click();
  await page.getByTestId('confirm').getByRole('button', { name: 'OK' }).click();
  await expect(page.getByTestId('toasts')).toContainText('Dismiss failed: ');
  await expect(page.getByTestId('toasts')).toContainText('403 Forbidden');
  // A failure keeps the panel open.
  await expect(page.getByTestId('session-panel')).toBeVisible();
});

test('a row shows its summary and canned buttons; a button resolves it without opening a session', async ({ page }) => {
  let closed = false;
  const calls = await mock(page, {
    attention: () => (closed ? [] : [{ kind: 'ticket', id: 'js-t1', title: 'Escalate run js-r1', why: '', summary: 'The deploy step failed 3 times.', since: iso(600e3) }]),
    rpc: (m) => { if (m === 'ticket.close') closed = true; return null; },
  });
  await page.goto('./');
  const row = page.getByTestId('attention-row').first();
  await expect(row.getByTestId('attention-summary')).toHaveText('The deploy step failed 3 times.');
  await expect(row.getByTestId('row-resolve')).toHaveText(['Close ticket']);
  await expect(row.getByTestId('open-chat')).toBeVisible();
  await row.getByTestId('row-resolve').click();
  await page.getByTestId('confirm').getByRole('button', { name: 'OK' }).click();
  await expect(page.getByTestId('toasts')).toContainText('Close ticket: done');
  await expect(page.getByTestId('attention-empty')).toBeVisible();
  expect(calls.find((c) => c.path === '/api/rpc/ticket.close')?.body).toEqual({ id: 'js-t1' });
  expect(calls.some((c) => c.path === '/api/attention/chat')).toBe(false);
  await expect(page.getByTestId('session-panel')).toHaveCount(0);
});

test('Open chat on a row opens the session', async ({ page }) => {
  const calls = await mock(page, {
    attention: () => [{ kind: 'run', id: 'js-r2', title: 'run stuck', why: '', since: iso(600e3) }],
    chat: () => ({ kind: 'run', id: 'js-r2', key: 'attention:run:js-r2', session: 'js-s3', started: true }),
  });
  await page.goto('./');
  await page.getByTestId('attention-row').first().getByTestId('open-chat').click();
  await expect(page.getByTestId('session-panel')).toBeVisible();
  await expect(page).toHaveURL(/session=js-s3/);
  expect(calls.filter((c) => c.path === '/api/attention/chat')).toHaveLength(1);
});

test('a stale item starts a chat with the context prompt; a second click reopens it', async ({ page }) => {
  let started = 0;
  const calls = await mock(page, {
    attention: () => [{ kind: 'stale', id: 'js-t9', title: 'orphan task', why: 'stale', since: iso(86400e3) }],
    chat: () => ({ kind: 'stale', id: 'js-t9', key: 'attention:stale:js-t9', session: 'js-s3', started: started++ === 0 }),
    streams: { 'js-s3': [sse([prompt(2, 'Desk attention item: stale js-t9\nTitle: orphan task')], { end: true })] },
  });
  await page.goto('./');
  await page.getByTestId('open-chat').first().click();
  await expect(page).toHaveURL(/session=js-s3/);
  await expect(page.getByTestId('journal')).toContainText('Desk attention item: stale js-t9');
  await page.getByTestId('session-panel-close').click();
  await expect(page.getByTestId('session-panel')).toHaveCount(0);
  await page.getByTestId('open-chat').first().click();
  await expect(page).toHaveURL(/session=js-s3/);
  const opens = calls.filter((c) => c.path === '/api/attention/chat');
  expect(opens).toHaveLength(2);
  expect(opens.every((c) => c.body.kind === 'stale' && c.body.id === 'js-t9')).toBe(true);
});

test('a pre-started item shows its first agent reply within 1s of the click (#606)', async ({ page }) => {
  const calls = await mock(page, {
    attention: () => [{ kind: 'question', id: 'js-q7', title: 'which branch?', why: 'question', since: iso(60e3) }],
    chat: () => ({ kind: 'question', id: 'js-q7', key: 'attention:question:js-q7', session: 'js-s7', pool: 'operator' }),
    streams: { 'js-s7': [sse([
      { seq: 2, type: 'turn_start', text: '', context: 'Desk attention item: question js-q7', time: iso(30e3) },
      say(3, 'It asks which branch to deploy. Answer in the composer, or run jobshop question answer js-q7.'),
      end(4),
    ])] },
  });
  await page.goto('./');
  const t0 = Date.now();
  await page.getByTestId('attention-row').first().click();
  await expect(page.getByTestId('journal')).toContainText('It asks which branch to deploy', { timeout: 1000 });
  expect(Date.now() - t0).toBeLessThan(1000);
  // Attaching sends nothing.
  expect(calls.filter((c) => c.path === '/api/chat/send')).toHaveLength(0);
});

test('no horizontal scroll at 360px with a chat open', async ({ page }) => {
  await page.setViewportSize({ width: 360, height: 740 });
  const long = 'x'.repeat(400);
  await mock(page, {
    attention: () => [{ kind: 'stale', id: 'js-t9', title: 'orphan task with a long title ' + long, why: 'stale', since: iso(1000) }],
    streams: { 'js-s2': [sse([prompt(2, 'see ' + long), say(3, '```\n' + long + '\n```\n' + long), { seq: 4, type: 'tool', tool: { id: 't', title: 'bash', status: 'completed', input: long, output: long } }], { end: true })] },
  });
  for (const path of ['./', './?view=chat', './?view=chat&session=js-s2']) {
    await page.goto(path);
    await expect(page.locator('main h1')).toBeVisible();
    await page.waitForTimeout(300);
    expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth), path).toBeLessThanOrEqual(1);
  }
  const j = page.getByTestId('journal');
  expect(await j.evaluate((el) => el.scrollWidth - el.clientWidth)).toBeLessThanOrEqual(1);
});

test('chat view lists escalation chats apart from other chats; a row opens the panel', async ({ page }) => {
  await mock(page, {
    escalations: [
      { ticket: 'js-e1', title: 'Escalate run js-r1', cause: 'run', why: 'needs a human', key: 'escalation:js-e1', session: 'js-s1', open: true, since: iso(3 * 3600e3) },
      { ticket: 'js-e0', title: 'Old hold', cause: 'label', why: 'staging key', key: 'escalation:js-e0', session: 'js-s0', open: false, resolution: 'done', since: iso(48 * 3600e3), closed: iso(24 * 3600e3) },
    ],
  });
  await page.goto('./?view=chat');
  const rows = page.getByTestId('escalation-row');
  await expect(rows).toHaveCount(2);
  await expect(rows.nth(0)).toContainText('run failed');
  await expect(rows.nth(1)).toContainText('closed · done');
  const list = page.getByTestId('chat-list');
  await expect(list.locator('a.item')).toHaveCount(2);
  await expect(list).toContainText('Resolve stale:js-t9');
  await rows.nth(0).click();
  await expect(page.getByTestId('session-panel')).toBeVisible();
  await expect(page).toHaveURL(/session=js-s1/);
  await page.getByTestId('session-panel-close').click();
  await expect(page.getByTestId('session-panel')).toHaveCount(0);
});

test('install assets are served', async ({ request }) => {
  const m = await request.get('manifest.webmanifest');
  expect(m.ok()).toBe(true);
  const man = await m.json();
  expect(man.start_url).toBe('./');
  for (const f of ['sw.js', 'icon-192.png', 'icon-512.png', 'badge.png', 'icon.svg']) expect((await request.get(f)).ok(), f).toBe(true);
});
