import assert from 'node:assert/strict';
import test from 'node:test';
import type { HistoryEntry } from '../../../packages/jobshop-client/rpc.js';
import { foldTurns } from './history.ts';

let seq = 0;
const at = (min: number) => new Date(Date.UTC(2026, 10, 10, 12, min)).toISOString();
const row = (min: number, actor: string, op: string): HistoryEntry => ({
  seq: ++seq,
  ticket_id: 'js-t',
  at: at(min),
  actor,
  op,
  revision: 1,
});

test('turn rows fold into their claim and leave the list', () => {
  const claim = row(1, 's1', 'claim');
  const entries = [
    row(0, 'admin', 'create'),
    claim,
    row(2, 's1', 'turn_start'),
    row(5, 's1', 'turn_end'),
    row(10, 's1', 'turn_start'),
    row(12, 's1', 'turn_end'),
    row(13, 's1', 'close'),
  ];
  const folded = foldTurns(entries);
  assert.deepEqual(
    folded.entries.map((e) => e.op),
    ['create', 'claim', 'close'],
  );
  assert.deepEqual(folded.claims.get(claim.seq), { count: 2, active: 5 * 60_000 });
});

test('a turn cut by the next turn_start counts to there, one outside any claim is dropped', () => {
  const claim = row(1, 's1', 'claim');
  const folded = foldTurns([
    row(0, 's1', 'turn_start'),
    row(0, 's1', 'turn_end'),
    claim,
    row(2, 's1', 'turn_start'),
    row(6, 's1', 'turn_start'),
    row(7, 's1', 'turn_end'),
  ]);
  assert.deepEqual(
    folded.entries.map((e) => e.op),
    ['claim'],
  );
  assert.deepEqual(folded.claims.get(claim.seq), { count: 2, active: 5 * 60_000 });
});
