import assert from 'node:assert/strict';
import test from 'node:test';
import type { SessionEvent } from '../../../packages/jobshop-client/rpc.js';
import { turnLabel, turnPhase } from './session.ts';

let seq = 0;
const frame = (type: string, extra: Partial<SessionEvent> = {}): SessionEvent => ({
  session: 's',
  seq: ++seq,
  type,
  ...extra,
});
const tool = (id: string, status: string, title = 'Reading') => frame('tool', { tool: { id, status, title } });
const label = (frames: SessionEvent[], busy = false) => {
  const phase = turnPhase(frames, busy);
  return phase && turnLabel(phase);
};

test('no turn is idle', () => {
  assert.equal(label([]), null);
  assert.equal(label([frame('usage')]), null);
});

test('a turn walks thinking, running, replying', () => {
  const start = frame('turn_start');
  assert.equal(label([start]), 'thinking…');
  assert.equal(label([start, frame('reasoning', { text: 'hm' })]), 'thinking…');
  assert.equal(label([start, frame('reasoning'), tool('t1', 'in_progress')]), 'running: Reading');
  assert.equal(label([start, tool('t1', 'in_progress'), tool('t1', 'completed')]), 'thinking…');
  assert.equal(label([start, tool('t1', 'completed'), frame('text', { text: 'ok' })]), 'replying…');
});

test('a tool that is still running wins over later text', () => {
  const start = frame('turn_start');
  const frames = [start, tool('t1', 'pending', 'Build'), frame('text', { text: 'x' })];
  assert.equal(label(frames), 'running: Build');
});

test('the tool intent is the label when there is one', () => {
  const t = frame('tool', { tool: { id: 't', status: 'in_progress', title: 'bash', intent: 'run tests' } });
  assert.equal(label([frame('turn_start'), t]), 'running: run tests');
});

test('turn end, an error and a cancel clear the indicator', () => {
  const start = frame('turn_start');
  assert.equal(label([start, frame('text'), frame('turn_end', { stop_reason: 'end_turn' })]), null);
  assert.equal(label([start, tool('t1', 'in_progress'), frame('turn_end', { stop_reason: 'cancelled' })]), null);
  assert.equal(label([start, frame('turn_end', { error: 'boom' })]), null);
});

test('a tail that starts mid-turn still shows the phase', () => {
  assert.equal(label([frame('text', { text: 'tail' })]), 'replying…');
  assert.equal(label([tool('t9', 'in_progress', '')]), 'running: tool');
  assert.equal(label([frame('usage')], true), 'thinking…');
  assert.equal(label([frame('usage')], false), null);
});

test('a finished turn before the tail does not leak into the next', () => {
  const frames = [frame('turn_start'), frame('text'), frame('turn_end'), frame('turn_start')];
  assert.equal(label(frames), 'thinking…');
});
