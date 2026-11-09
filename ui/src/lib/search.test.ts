import assert from 'node:assert/strict';
import test from 'node:test';
import { createSearcher, type SearchSpec } from './search.ts';

interface Chat {
  id: string;
  title: string;
  key: string;
}
const spec: SearchSpec<Chat> = {
  fields: [
    { name: 'title', weight: 3, get: (c) => c.title },
    { name: 'key', weight: 1, get: (c) => c.key },
  ],
  ids: (c) => [c.id, c.key],
  label: (c) => c.title,
};
const chats: Chat[] = [
  { id: 'js-s-1', title: 'ops weekly', key: 'desk:ops' },
  { id: 'js-s-2', title: 'release notes', key: 'desk:rel' },
  { id: 'js-s-3', title: 'incident follow up', key: 'escalation:42' },
  { id: 'js-s-4', title: 'js-s-1 mentioned here', key: 'desk:other' },
];

test('an exact identifier ranks before a title that mentions it', () => {
  const hits = createSearcher(chats, spec).search('js-s-1');
  assert.equal(hits[0].id, 'js-s-1');
});

test('a longer identifier prefix finds its item first', () => {
  assert.equal(createSearcher(chats, spec).search('js-s-3')[0].id, 'js-s-3');
  assert.equal(createSearcher(chats, spec).search('escalation')[0].id, 'js-s-3');
});

test('weighted fuzzy match finds a title with a typo', () => {
  const hits = createSearcher(chats, spec).search('incidnt');
  assert.equal(hits[0].id, 'js-s-3');
});

test('the first and last letters of a title still find it', () => {
  const hits = createSearcher(chats, spec).search('rs');
  assert.ok(
    hits.some((c) => c.id === 'js-s-2'),
    'release notes via r..s',
  );
});

test('an empty query returns everything and no match returns nothing', () => {
  const searcher = createSearcher(chats, spec);
  assert.equal(searcher.search('  ').length, chats.length);
  assert.deepEqual(searcher.search('zzzzqqq'), []);
});

test('a rebuilt searcher sees additions and removals', () => {
  const next = [...chats.slice(1), { id: 'js-s-9', title: 'billing', key: 'desk:bill' }];
  const searcher = createSearcher(next, spec);
  assert.equal(searcher.search('billing')[0].id, 'js-s-9');
  assert.deepEqual(
    searcher.search('ops weekly').filter((c) => c.id === 'js-s-1'),
    [],
  );
});

test('bounded latency on a large collection', () => {
  const many: Chat[] = Array.from({ length: 5000 }, (_, i) => ({
    id: 'id-' + i,
    title: 'chat number ' + i,
    key: 'desk:' + i,
  }));
  const searcher = createSearcher(many, spec);
  const start = performance.now();
  const hits = searcher.search('id-4242');
  const ms = performance.now() - start;
  assert.equal(hits[0].id, 'id-4242');
  assert.ok(ms < 250, `search took ${ms}ms`);
});
