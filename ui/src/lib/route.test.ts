import assert from 'node:assert/strict';
import test from 'node:test';
import { parseRoute } from './route.ts';

test('a trace route ignores a stale layout value', () => {
  const route = parseRoute('?view=trace&ref=js-1&layout=graph', '');
  assert.equal(route.view, 'trace');
  assert.equal(route.ref, 'js-1');
});
