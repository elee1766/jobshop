import assert from 'node:assert/strict';
import test from 'node:test';
import type {
  Comment,
  GraphResult,
  HistoryEntry,
  Ticket,
  TimelineRow,
  TimelineSegment,
} from '../../../packages/jobshop-client/rpc.js';
import {
  attempts,
  childProgress,
  detailTab,
  endedBy,
  eventItems,
  hasIO,
  historyPartial,
  ioOf,
  kv,
  nodeDuration,
  nowOf,
  relationCount,
  relationGroups,
  summarize,
  timeByState,
} from './detail.ts';

const t = (min: number) => new Date(Date.UTC(2026, 10, 10, 12, min)).toISOString();
const now = Date.UTC(2026, 10, 10, 13, 0);
const seg = (state: string, from: number, to: number | null, o: Partial<TimelineSegment> = {}): TimelineSegment => ({
  state,
  start: t(from),
  ...(to === null ? {} : { end: t(to) }),
  ...o,
});
const row = (o: Partial<TimelineRow> = {}): TimelineRow => ({
  id: 'js-1',
  kind: 'step',
  status: 'in_progress',
  depth: 0,
  created_at: t(0),
  segments: [],
  markers: [],
  ...o,
});
const ticket = (o: Partial<Ticket> = {}): Ticket => ({
  id: 'js-1',
  project: 'js',
  kind: 'step',
  status: 'in_progress',
  priority: 0,
  fence: 1,
  revision: 1,
  created_at: t(0),
  updated_at: t(0),
  ...o,
});
let seq = 0;
const entry = (min: number, op: string, actor = 'js-s-1', diff: unknown = {}): HistoryEntry => ({
  seq: ++seq,
  ticket_id: 'js-1',
  at: t(min),
  actor,
  op,
  revision: 1,
  diff,
});

// A step that failed once and finished on the second claim, then waited.
const retried = row({
  status: 'closed',
  resolution: 'done',
  closed_at: t(50),
  segments: [
    seg('queued', 0, 5),
    seg('working', 5, 15, { actor: 'js-s-1', turns: [{ start: t(6), end: t(10), stop: 'end_turn' }] }),
    seg('retry', 15, 16, { actor: 'js-s-1', detail: 'tests failed' }),
    seg('working', 16, 30, { actor: 'js-s-2' }),
    seg('waiting', 30, 50, { detail: 'CI', ended: 'met' }),
  ],
});

test('time by state sums each state and splits a claim with turns into turn and held', () => {
  const shares = timeByState(retried, now);
  const by = Object.fromEntries(shares.map((s) => [s.key, s.ms / 60_000]));
  assert.deepEqual(by, { queued: 5, turn: 4, held: 6, working: 14, retry: 1, waiting: 20 });
});

test('a claim without turn rows is working time, never idle', () => {
  const shares = timeByState(row({ segments: [seg('working', 0, 10, { actor: 'js-s-1' })] }), now);
  assert.deepEqual(
    shares.map((s) => s.key),
    ['working'],
  );
});

test('a run has no time split because it is working for its whole life', () => {
  assert.deepEqual(timeByState(row({ kind: 'run', segments: [seg('working', 0, null)] }), now), []);
});

test('attempts count claims, and failed ones are the retry spans', () => {
  assert.deepEqual(attempts(retried), { claims: 3, failed: 1 });
  // A controller-run ticket has working spans without a session.
  assert.deepEqual(attempts(row({ segments: [seg('working', 0, 5)] })), { claims: 0, failed: 0 });
});

test('the summary measures creation to close, or to now while open', () => {
  const closed = summarize(ticket({ status: 'closed', closed_at: t(50) }), retried, now);
  assert.equal(closed.took, 50 * 60_000);
  assert.equal(closed.ended, Date.parse(t(50)));
  const open = summarize(ticket(), row(), now);
  assert.equal(open.ended, null);
  assert.equal(open.took, 60 * 60_000);
});

test('the retry limit is only shown for a step that has one', () => {
  assert.equal(summarize(ticket({ fields: { retries: 3 } }), retried, now).limit, 3);
  assert.equal(summarize(ticket({ kind: 'task', fields: { retries: 3 } }), retried, now).limit, 0);
  assert.equal(summarize(ticket({ fields: { retries: 0 } }), retried, now).limit, 0);
});

test('a ticket off the timeline is summarized from its own fields and says so', () => {
  const s = summarize(ticket(), undefined, now);
  assert.equal(s.bare, true);
  assert.deepEqual(s.shares, []);
  assert.equal(s.claims, 0);
});

test('a partial row is flagged', () => {
  assert.equal(summarize(ticket(), row({ partial: true }), now).partial, true);
});

test('child progress counts the direct children that closed', () => {
  const rows = [
    row({ id: 'r', kind: 'run' }),
    row({ id: 'a', parent: 'r', status: 'closed' }),
    row({ id: 'b', parent: 'r' }),
    row({ id: 'c', parent: 'a' }),
  ];
  assert.deepEqual(childProgress(rows, 'r'), { total: 2, closed: 1 });
  assert.deepEqual(childProgress(undefined, 'r'), { total: 0, closed: 0 });
});

const none = { total: 0, closed: 0 };

test('now: a running turn names the tool, an idle claim is held', () => {
  const open = row({ segments: [seg('working', 0, null, { actor: 'js-s-1' })] });
  const live = nowOf(ticket({ holder: 'js-s-1' }), open, none, {
    id: 'js-s-1',
    alive: true,
    busy: true,
    tool: 'go test',
  });
  assert.equal(live.kind, 'turn');
  assert.match(live.text, /turn running: go test/);
  const idle = nowOf(ticket({ holder: 'js-s-1' }), open, none, { id: 'js-s-1', alive: true });
  assert.equal(idle.kind, 'idle');
  const lost = nowOf(ticket({ holder: 'js-s-1' }), open, none, { id: 'js-s-1', host_lost: true });
  assert.equal(lost.kind, 'lost');
  // A holder Desk cannot describe is claimed held, nothing more.
  const unknown = nowOf(ticket({ holder: 'js-s-1' }), open, none, undefined);
  assert.equal(unknown.kind, 'held');
  assert.equal(unknown.text, 'held by js-s-1');
});

test('now: a wait lists its conditions, a human hold says why, a deferral says until when', () => {
  const waits = [{ type: 'github.checks', label: 'acme/w PR #7 checks', state: 'waiting' }];
  const waiting = nowOf(
    ticket(),
    row({ segments: [seg('waiting', 0, null, { detail: 'CI on PR 7', wait: waits })] }),
    none,
    undefined,
  );
  assert.equal(waiting.kind, 'waiting');
  assert.equal(waiting.text, 'CI on PR 7');
  assert.deepEqual(waiting.waits, waits);
  const human = nowOf(ticket(), row({ segments: [seg('human', 0, null, { detail: 'question' })] }), none, undefined);
  assert.deepEqual([human.kind, human.text], ['human', 'question']);
  const until = t(30);
  const deferred = nowOf(ticket(), row({ segments: [seg('deferred', 0, null, { until })] }), none, undefined);
  assert.equal(deferred.kind, 'deferred');
  assert.equal(deferred.until, Date.parse(until));
});

test('now: a run says how many steps closed, a queued ticket names its pool', () => {
  const run = nowOf(
    ticket({ kind: 'run' }),
    row({ kind: 'run', segments: [seg('working', 0, null)] }),
    { total: 4, closed: 3 },
    undefined,
  );
  assert.equal(run.kind, 'running');
  assert.equal(run.text, 'running · 3 of 4 steps closed');
  const queued = nowOf(ticket({ pool: 'dev' }), row({ segments: [seg('queued', 0, null)] }), none, undefined);
  assert.equal(queued.text, 'queued for dev');
});

test('now: a closed ticket says how it ended, with the reason from its close row', () => {
  const closed = nowOf(ticket({ status: 'closed', resolution: 'failed', closed_at: t(9) }), row(), none, undefined, [
    entry(9, 'close', 'js-s-1', { resolution: 'failed', note: 'tests red' }),
  ]);
  assert.equal(closed.kind, 'closed');
  assert.equal(closed.text, 'tests red');
  // With no reason on the close row it still says how it ended.
  assert.equal(nowOf(ticket({ status: 'closed', resolution: 'done' }), row(), none, undefined).text, 'closed done');
});

test('now: without a timeline row the ticket fields still say something', () => {
  assert.equal(nowOf(ticket({ holder: 'js-s-1' }), undefined, none, undefined).text, 'held by js-s-1');
  assert.equal(nowOf(ticket({ status: 'open' }), undefined, none, undefined).kind, 'open');
  assert.equal(nowOf(ticket({ fields: { wait: { reason: 'CI' } } }), undefined, none, undefined).kind, 'waiting');
});

test('how a span ended comes from the row written at its end', () => {
  const history = [
    entry(0, 'create', 'admin'),
    entry(5, 'claim', 'js-s-1', { holder: 'js-s-1' }),
    entry(15, 'release', 'js-s-1', { reason: 'tests failed', failures: 1 }),
    entry(16, 'claim', 'js-s-2'),
    entry(16, 'turn_start', 'js-s-2'),
  ];
  assert.equal(endedBy(seg('queued', 0, 5), history), 'claimed by js-s-1');
  assert.equal(endedBy(seg('working', 5, 15, { actor: 'js-s-1' }), history), 'released (tests failed)');
  assert.equal(endedBy(seg('retry', 15, 16), history), 'claimed by js-s-2');
  // Open spans have not ended. A wait says how it ended in its own text, so
  // it adds nothing here.
  assert.equal(endedBy(seg('working', 16, null), history), '');
  assert.equal(endedBy(seg('waiting', 30, 50, { ended: 'met' }), history), '');
});

test('events: one entry per span plus markers, comments and the close, newest first', () => {
  const comments: Comment[] = [
    { id: 1, ticket_id: 'js-1', author: 'a', body: 'looks fine\nmore', created_at: t(20) },
    { id: 2, ticket_id: 'js-1', author: 'a', body: 'PR https://github.com/acme/w/pull/7', created_at: t(25) },
  ];
  const r = row({
    ...retried,
    markers: [
      { at: t(20), kind: 'comment', label: 'looks fine' },
      { at: t(25), kind: 'pr', label: 'PR opened', url: 'https://github.com/acme/w/pull/7' },
    ],
  });
  const items = eventItems(r, [entry(50, 'close', 'js-s-2', { resolution: 'done' })], comments);
  assert.deepEqual(
    items.map((i) => i.type),
    // close, waiting span, PR marker, comment, working, retry, working, queued
    ['close', 'segment', 'marker', 'comment', 'segment', 'segment', 'segment', 'segment'],
  );
  // The PR comment appears once, as the marker.
  assert.equal(items.filter((i) => i.type === 'comment').length, 1);
  // Segments keep their timeline index, so a chart click can find its entry.
  const segs = items.flatMap((i) => (i.type === 'segment' ? [i.index] : []));
  assert.deepEqual(segs, [4, 3, 2, 1, 0]);
});

test('a history that does not begin at creation is partial', () => {
  assert.equal(historyPartial([entry(5, 'claim')]), true);
  assert.equal(historyPartial([entry(0, 'create'), entry(5, 'claim')]), false);
  assert.equal(historyPartial([]), false);
});

test('values: long and multi-line ones are marked for collapsing, empty ones are left out', () => {
  assert.equal(kv('a', ''), null);
  assert.equal(kv('a', undefined), null);
  assert.deepEqual(kv('a', 'short'), { key: 'a', value: 'short', long: false });
  assert.equal(kv('a', 'x'.repeat(200))?.long, true);
  assert.equal(kv('a', 'one\ntwo')?.long, true);
  assert.equal(kv('n', 3)?.value, '3');
  assert.equal(kv('list', ['a', 'b'])?.value, 'a, b');
  assert.equal(kv('obj', { x: 1 })?.long, true);
});

test('input and result: a run, a step, a batch and a task', () => {
  const run = ioOf(
    ticket({
      kind: 'run',
      status: 'closed',
      resolution: 'done',
      fields: { workflow: 'pr-work', state: 'done', vars: { issue: '7' }, inputs: { repo: 'acme/w' } },
    }),
  );
  assert.deepEqual(
    run.inputs.map((r) => r.key),
    ['workflow', 'issue', 'repo'],
  );
  assert.deepEqual(
    run.result.map((r) => r.key),
    ['state', 'resolution'],
  );
  const step = ioOf(
    ticket({
      status: 'closed',
      resolution: 'done',
      fields: { step: 'implement', run: 'js-r', failures: 1, retries: 3, command: 'go test', output: 'ok\nok' },
    }),
  );
  assert.deepEqual(
    step.inputs.map((r) => r.key),
    ['step', 'run'],
  );
  assert.deepEqual(
    step.result.map((r) => r.key),
    ['resolution', 'failures', 'retries', 'command', 'output'],
  );
  assert.equal(step.result.find((r) => r.key === 'output')?.long, true);
  const batch = ioOf(ticket({ kind: 'batch', fields: { key: 'k', issues: ['a', 'b'], state: 'running' } }));
  assert.equal(batch.inputs.find((r) => r.key === 'issues')?.value, 'a, b');
  assert.equal(hasIO(ticket({ kind: 'task' })), false);
  assert.equal(hasIO(ticket({ kind: 'task', fields: { inputs: { a: 1 } } })), true);
  assert.equal(hasIO(ticket({ kind: 'run' })), true);
});

const node = (id: string, o: Record<string, unknown> = {}) => ({
  id,
  project: 'js',
  kind: 'step',
  status: 'open',
  priority: 0,
  fence: 0,
  revision: 1,
  created_at: t(0),
  updated_at: t(0),
  ...o,
});
const graph: GraphResult = {
  root: 'r',
  requested: 'r',
  nodes: [
    node('r', { kind: 'run' }),
    node('a'),
    node('b', { status: 'closed', closed_at: t(10) }),
    node('c'),
    node('d'),
  ],
  relations: [
    { from: 'r', to: 'a', kind: 'parent' },
    { from: 'r', to: 'b', kind: 'parent' },
    { from: 'a', to: 'b', kind: 'blocks' },
    { from: 'c', to: 'a', kind: 'blocks' },
    { from: 'a', to: 'd', kind: 'waits_for' },
  ],
  external: [],
  truncated: false,
};

test('relations group by how they relate, parent first, and count each ticket once', () => {
  const groups = relationGroups(graph, { id: 'a', parent_id: 'r' });
  assert.deepEqual(
    groups.map((g) => [g.kind, g.nodes.map((n) => n.id)]),
    [
      ['parent', ['r']],
      ['blocked by', ['b']],
      ['blocks', ['c']],
      ['waits for', ['d']],
    ],
  );
  assert.equal(relationCount(groups), 4);
  const run = relationGroups(graph, { id: 'r' });
  assert.deepEqual(
    run.map((g) => [g.kind, g.nodes.map((n) => n.id)]),
    [['children', ['a', 'b']]],
  );
});

test('a related ticket shows how long it took, or how long it has been open', () => {
  assert.equal(nodeDuration({ created_at: t(0), closed_at: t(10), status: 'closed' }, now), '10m');
  assert.equal(nodeDuration({ created_at: t(0), status: 'open' }, now), '1h so far');
  assert.equal(nodeDuration({ created_at: t(0), status: 'open' }, Date.parse(t(0)) - 1), '');
});

test('an unknown tab in the URL opens the events tab', () => {
  assert.equal(detailTab('raw'), 'raw');
  assert.equal(detailTab('files'), 'events');
  assert.equal(detailTab(''), 'events');
});
