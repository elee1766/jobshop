import assert from 'node:assert/strict';
import test from 'node:test';
import type { HolderInfo, TimelineResult, TimelineRow, TimelineSegment } from '../../../packages/jobshop-client/rpc.js';
import {
  arrowPath,
  arrowShape,
  durationLabel,
  holderState,
  liveText,
  segmentText,
  timelineSpan,
  turnsText,
  turnTotals,
  visibleRows,
} from './timeline.ts';

const row = (id: string, depth: number, o: Partial<TimelineRow> = {}): TimelineRow => ({
  id,
  kind: 'task',
  status: 'open',
  depth,
  created_at: '2026-11-09T10:00:00Z',
  segments: [],
  markers: [],
  ...o,
});
const tl = (rows: TimelineRow[]): TimelineResult => ({ root: rows[0]?.id || '', rows, arrows: [], truncated: false });

test('the span ends at now while anything is open, else at the last close', () => {
  const now = Date.parse('2026-11-09T12:00:00Z');
  const open = timelineSpan(tl([row('a', 0, { closed_at: '2026-11-09T10:30:00Z' }), row('b', 1)]), now);
  assert.equal(open.active, true);
  assert.equal(open.end, now);
  assert.equal(open.start, Date.parse('2026-11-09T10:00:00Z'));
  const done = timelineSpan(tl([row('a', 0, { closed_at: '2026-11-09T10:30:00Z' })]), now);
  assert.equal(done.active, false);
  assert.equal(done.end, Date.parse('2026-11-09T10:30:00Z'));
});

test('a lone fresh ticket still has a one minute scale', () => {
  const created = Date.parse('2026-11-09T10:00:00Z');
  assert.equal(timelineSpan(tl([row('a', 0)]), created).end, created + 60_000);
});

test('collapsing a row hides what is nested under it, not its siblings', () => {
  const rows = [row('run', 0), row('s1', 1), row('s1a', 2), row('s2', 1), row('other', 0)];
  assert.deepEqual(
    visibleRows(rows, new Set(['s1'])).map((r) => r.id),
    ['run', 's1', 's2', 'other'],
  );
  assert.deepEqual(
    visibleRows(rows, new Set(['run'])).map((r) => r.id),
    ['run', 'other'],
  );
  assert.equal(visibleRows(rows, new Set()).length, 5);
});

test('segment text names state, actor, reason and how a wait ended', () => {
  assert.equal(
    segmentText({ state: 'waiting', start: 'x', detail: 'CI on PR 5', ended: 'met' }),
    'waiting on an event · CI on PR 5 · ended met',
  );
  assert.equal(segmentText({ state: 'working', start: 'x', actor: 'js-s-1' }), 'working · js-s-1');
});

test('arrows turn back through the row gap when the dependent starts first', () => {
  assert.equal(arrowPath(10, 5, 40, 35), 'M 10 5 H 16 V 35 H 40');
  assert.equal(arrowPath(40, 5, 20, 35), 'M 40 5 H 46 V 20 H 14 V 35 H 20');
  assert.equal(arrowShape({ from: 'a', to: 'b', kind: 'blocks', state: 'conflict' }), 'conflict');
  assert.equal(arrowShape({ from: 'a', to: 'b', kind: 'tracks' }), 'tracks');
  assert.equal(arrowShape({ from: 'a', to: 'b', kind: 'wait', state: 'met' }), 'wait');
});

test('durations read as two units', () => {
  assert.equal(durationLabel(90 * 60_000), '1h 30m');
  assert.equal(durationLabel(-5), '0s');
});

const seg = (o: Partial<TimelineSegment>): TimelineSegment => ({
  state: 'working',
  start: '2026-11-10T12:00:00Z',
  ...o,
});
const min = (n: number) => new Date(Date.UTC(2026, 10, 10, 12, n)).toISOString();

test('totals split a claim into time in turns and time held', () => {
  const claim = seg({
    end: min(60),
    turns: [
      { start: min(5), end: min(15) },
      { start: min(30), end: min(34) },
    ],
  });
  const t = turnTotals(claim, Date.parse(min(90)));
  assert.deepEqual(t, { count: 2, active: 14 * 60_000, held: 46 * 60_000 });
  assert.equal(turnsText(t!), 'active 14m in 2 turns · held 46m');
  assert.equal(turnsText({ count: 1, active: 60_000, held: 0 }), 'active 1m in 1 turn · held 0s');
});

test('a running turn counts to now, a claim without turn rows has no totals', () => {
  const open = seg({ turns: [{ start: min(10) }] });
  assert.equal(turnTotals(open, Date.parse(min(40)))?.active, 30 * 60_000);
  assert.equal(turnTotals(seg({}), Date.parse(min(40))), null);
  assert.equal(segmentText(seg({ actor: 'js-s-1' }), { now: Date.parse(min(40)) }), 'working · js-s-1');
});

const holder = (o: Partial<HolderInfo>): HolderInfo => ({ id: 'js-s-1', ...o });

test('only an open working span has a live state, and an unknown holder claims nothing', () => {
  const open = seg({ actor: 'js-s-1' });
  assert.equal(holderState(open, holder({ alive: true, busy: true, tool: 'go test' })), 'running');
  assert.equal(holderState(open, holder({ alive: true })), 'idle');
  assert.equal(holderState(open, holder({ host_lost: true })), 'lost');
  assert.equal(holderState(open, holder({})), 'unknown', 'no host, not known lost: an external session');
  assert.equal(holderState(open, undefined), 'unknown');
  assert.equal(holderState(seg({ actor: 'js-s-1', end: min(5) }), holder({ alive: true })), undefined);
  assert.equal(holderState(seg({ state: 'queued' }), holder({ alive: true })), undefined);
  assert.equal(liveText('running', holder({ tool: 'go test' })), 'turn running: go test');
  assert.equal(liveText('idle'), 'held, idle between turns');
  assert.match(liveText('lost'), /host not alive/);
  assert.equal(liveText('unknown'), '');
  assert.match(segmentText(open, { session: holder({ host_lost: true }) }), /nobody is working this ticket/);
});
