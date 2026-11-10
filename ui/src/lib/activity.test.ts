import assert from 'node:assert/strict';
import test from 'node:test';
import type { RunResult, Ticket } from '../../../packages/jobshop-client/rpc.js';
import { activityOf } from './activity.ts';

const now = Date.parse('2026-11-10T12:00:00Z');
const later = (hours: number) => new Date(now + hours * 3600_000).toISOString();

const ticket = (o: Partial<Ticket>): Ticket =>
  ({ id: 'js-step', project: 'js', kind: 'step', status: 'open', priority: 2, fence: 0, revision: 1, ...o }) as Ticket;
const run = (steps: Ticket[], fields: Record<string, unknown> = { state: 'running' }): RunResult =>
  ({ run: ticket({ id: 'js-run', kind: 'run', status: 'in_progress', fields }), steps }) as RunResult;
const activity = (step: Ticket, session?: Parameters<typeof activityOf>[2], r = run([step])) =>
  activityOf(r, step, session, now);

test('a step released until a time is deferred with its reason and resume time (#1185)', () => {
  const step = ticket({
    defer_until: later(48),
    fields: { defer: { reason: 'gather seven days of routing data', until: later(48) } },
  });
  const a = activity(step);
  assert.equal(a.kind, 'timer');
  assert.equal(a.label, 'deferred');
  assert.match(a.detail, /gather seven days of routing data/);
  assert.match(a.detail, /until 2026-11-12 12:00Z \(in 2d\)/);
});

test('a deferral reason that belongs to an older release is not shown', () => {
  const step = ticket({
    defer_until: later(5),
    fields: { defer: { reason: 'old reason', until: later(1) } },
  });
  assert.equal(activity(step).detail, 'until 2026-11-10 17:00Z (in 5h)');
});

test('a step gated on earlier steps is not a timer', () => {
  const a = activity(ticket({ defer_until: '2200-01-01T00:00:00Z' }));
  assert.equal(a.kind, 'gated');
});

test('an external wait names its reason and unmet conditions', () => {
  const step = ticket({
    fields: {
      wait: {
        reason: 'CI on PR #9',
        conds: [
          { type: 'github.checks', note: 'checks pending' },
          { type: 'github.pr', met: true },
        ],
      },
    },
  });
  const a = activity(step);
  assert.equal(a.kind, 'event');
  assert.equal(a.detail, 'CI on PR #9 · checks pending');
});

test('a wait on a time alone is a deferral until the earliest time in mode any', () => {
  const step = ticket({
    fields: {
      wait: {
        reason: 'window',
        mode: 'any',
        conds: [
          { type: 'time', at: later(3) },
          { type: 'time', at: later(2) },
        ],
      },
    },
  });
  const a = activity(step);
  assert.equal(a.kind, 'timer');
  assert.match(a.detail, /until 2026-11-10 14:00Z \(in 2h\)/);
});

test('a human wait is a person, from the run state or a label', () => {
  assert.equal(activity(ticket({}), undefined, run([], { state: 'needs_human' })).kind, 'human');
  assert.equal(activity(ticket({ labels: ['question'] })).kind, 'human');
});

test('a held step is in a turn only when the live session says so', () => {
  const step = ticket({ status: 'in_progress', holder: 'js-s-1' });
  assert.equal(activity(step, { alive: true, busy: true, tool: 'go test' }).kind, 'turn');
  assert.equal(activity(step, { alive: true, busy: true, tool: 'go test' }).detail, 'go test');
  assert.equal(activity(step, { alive: true }).kind, 'held');
  assert.equal(activity(step, undefined).kind, 'held', 'no session state: not claimed to be running');
  assert.equal(activity(step, { alive: false, host_pid: 7 }).kind, 'lost');
  assert.equal(activity(step, { alive: false }).kind, 'held', 'no host and no exit: an external session');
});

test('a retained session on a deferred step does not make it active', () => {
  const step = ticket({ holder: '', defer_until: later(2), fields: { defer: { reason: 'later', until: later(2) } } });
  assert.equal(activity(step, { alive: true, busy: true }).kind, 'timer');
});

test('a step waiting for a resource slot and a plain queued step', () => {
  const step = ticket({ pool: 'dev' });
  const waits = [{ ticket: step.id, key: 'gpu', held: 1, capacity: 1, ahead: 0 }];
  const r = { ...run([step]), waits } as RunResult;
  assert.equal(activity(step, undefined, r).kind, 'resource');
  assert.deepEqual(activity(step), { kind: 'queued', label: 'queued', detail: 'dev' });
});
