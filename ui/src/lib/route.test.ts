import assert from 'node:assert/strict';
import test from 'node:test';
import { parseRoute } from './route.ts';

test('a trace without a layout opens as the timeline (#1136)', () => {
  assert.equal(parseRoute('?view=trace&ref=js-1', '').timeline, true);
});

test('layout=graph selects the graph', () => {
  assert.equal(parseRoute('?view=trace&ref=js-1&layout=graph', '').timeline, false);
});

test('an existing layout=timeline URL still opens the timeline', () => {
  assert.equal(parseRoute('?view=trace&ref=js-1&layout=timeline', '').timeline, true);
});
