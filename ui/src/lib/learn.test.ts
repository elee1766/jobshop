import assert from 'node:assert/strict';
import test from 'node:test';
import type { Learning } from '../../../packages/jobshop-client/rpc.js';
import { build, learningOutcome, sections, stageOf } from './learn.ts';

const learning = (o: Partial<Learning> = {}): Learning => ({
  id: 1,
  kind: 'failure',
  summary: 'a problem',
  scope: 'global',
  source: 'controller',
  status: 'open',
  created_at: '2026-11-01T10:00:00Z',
  ...o,
});

test('stageOf puts an expired learning in its own stage', () => {
  assert.equal(stageOf(learning({ status: 'expired' })), 'expired');
  assert.equal(stageOf(learning({ status: 'open' })), 'awaiting');
  assert.equal(stageOf(learning({ status: 'open', pr: 'https://github.com/a/b/pull/1' })), 'in-flight');
});

test('an expired learning is not awaiting the curator', () => {
  const units = build([learning({ id: 1, status: 'expired' }), learning({ id: 2 })], []);
  assert.deepEqual(
    units.filter((u) => u.stage === 'awaiting').map((u) => u.key),
    ['learning:2'],
  );
  assert.deepEqual(
    units.filter((u) => u.stage === 'expired').map((u) => u.key),
    ['learning:1'],
  );
});

test('the expired section is closed by default', () => {
  const s = sections.find((x) => x.stage === 'expired');
  assert.ok(s);
  assert.equal(s.open, false);
  assert.equal(sections.find((x) => x.stage === 'awaiting')?.open, true);
});

test('learningOutcome says expired', () => {
  assert.match(learningOutcome(learning({ status: 'expired' })), /^expired/);
});

test('stageOf puts a duplicate in its own stage', () => {
  assert.equal(stageOf(learning({ status: 'duplicate', duplicate_of: 1 })), 'duplicate');
});

test('a duplicate shows as a count on the learning kept, not as a row', () => {
  const units = build(
    [
      learning({ id: 1 }),
      learning({ id: 2, status: 'duplicate', duplicate_of: 1 }),
      learning({ id: 3, status: 'duplicate', duplicate_of: 1 }),
      learning({ id: 4, summary: 'something else' }),
    ],
    [],
  );
  assert.deepEqual(
    units
      .filter((u) => u.stage === 'awaiting')
      .map((u) => u.key)
      .toSorted(),
    ['learning:1', 'learning:4'],
  );
  assert.equal(units.filter((u) => u.stage === 'duplicate').length, 0);
  const kept = units.find((u) => u.key === 'learning:1');
  assert.deepEqual(kept?.duplicates.map((l) => l.id).toSorted(), [2, 3]);
  assert.equal(units.find((u) => u.key === 'learning:4')?.duplicates.length, 0);
  assert.ok(kept?.search.includes('#2'), 'a search by a duplicate id finds the kept learning');
});

test('duplicates of an applied learning stay with it', () => {
  const units = build(
    [learning({ id: 1, status: 'applied', pr: 'memory:3' }), learning({ id: 2, status: 'duplicate', duplicate_of: 1 })],
    [],
  );
  assert.equal(units.length, 1);
  assert.equal(units[0].stage, 'applied');
  assert.equal(units[0].duplicates.length, 1);
});

test('a duplicate whose kept learning is not loaded is listed in its own section', () => {
  const units = build([learning({ id: 2, status: 'duplicate', duplicate_of: 99 })], []);
  assert.equal(units.length, 1);
  assert.equal(units[0].stage, 'duplicate');
  assert.equal(units[0].outcome, 'duplicate of #99');
  assert.equal(sections.find((s) => s.stage === 'duplicate')?.open, false);
});
