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
