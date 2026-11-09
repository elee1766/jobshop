import assert from 'node:assert/strict';
import test from 'node:test';
import type { BrowseItem, BrowseScopeResult } from '../../../packages/jobshop-client/rpc.js';
import {
  age,
  authorValue,
  duration,
  groupItems,
  newComments,
  statusLetter,
  statusNotice,
  statusTone,
  whereIs,
} from './browse.ts';

const item = (over: Partial<BrowseItem>): BrowseItem => ({
  provider: 'linear',
  scope: 'JOB',
  ref: 'JOB-1',
  id: 'JOB-1',
  kind: 'issue',
  title: 't',
  state: 'Todo',
  state_type: 'unstarted',
  labels: [],
  url: 'https://linear.app/x/issue/JOB-1',
  ...over,
});

const scope: BrowseScopeResult = {
  revision: 1,
  linear: [{ team: 'JOB', workspace: 'w' }],
  github: [{ repo: 'o/r' }],
  writes: { enabled: false, reason: 'no' },
  limits: { page_size: 50, max_page_size: 100, thread_page: 25 },
};

test('status letters name the provider state, and the tone reuses the ticket colors', () => {
  assert.equal(statusLetter(item({ state_type: 'started' })), 'I');
  assert.equal(statusLetter(item({ state_type: 'backlog' })), 'B');
  assert.equal(statusLetter(item({ state_type: 'canceled' })), 'X');
  assert.equal(statusLetter(item({ provider: 'github', state: 'merged' })), 'M');
  assert.equal(statusLetter(item({ provider: 'github', state: 'draft' })), 'D');
  assert.equal(statusTone(item({ state_type: 'started' })), 'in_progress');
  assert.equal(statusTone(item({ state_type: 'completed' })), 'closed');
  assert.equal(statusTone(item({ provider: 'github', state: 'open' })), 'open');
  assert.equal(statusTone(item({ provider: 'github', state: 'merged' })), 'closed');
});

test('groups follow the provider order, keep list order inside, and never lose an item', () => {
  const items = [
    item({ ref: 'JOB-1', state_type: 'backlog' }),
    item({ ref: 'JOB-2', state_type: 'started' }),
    item({ ref: 'JOB-3', state_type: 'backlog' }),
    item({ ref: 'JOB-4', state_type: 'weird' }),
  ];
  const groups = groupItems(items, 'status');
  assert.deepEqual(
    groups.map((g) => g.key),
    ['started', 'backlog', 'weird'],
  );
  assert.deepEqual(
    groups[1].items.map((i) => i.ref),
    ['JOB-1', 'JOB-3'],
  );
  assert.equal(groups.flatMap((g) => g.items).length, items.length);
  const flat = groupItems(items, 'none');
  assert.equal(flat.length, 1);
  assert.equal(flat[0].label, '');
  const owners = groupItems([item({ owner: 'b' }), item({ owner: undefined }), item({ owner: 'b' })], 'owner');
  assert.deepEqual(
    owners.map((g) => [g.label, g.items.length]),
    [
      ['b', 2],
      ['nobody', 1],
    ],
  );
  const review = groupItems(
    [
      item({ provider: 'github', review: 'approved' }),
      item({ provider: 'github', review: 'required' }),
      item({ provider: 'github' }),
    ],
    'review',
  );
  assert.deepEqual(
    review.map((g) => g.label),
    ['review required', 'approved', 'no review state'],
  );
});

test('an app author is filtered as a bot login the controller turns into app/<login>', () => {
  assert.equal(authorValue(item({ owner: 'elee-agent', owner_bot: true })), 'elee-agent[bot]');
  assert.equal(authorValue(item({ owner: 'eddy' })), 'eddy');
});

test('a linked item opens in Desk only inside the configured scope', () => {
  const inside = whereIs(
    { kind: 'pr', ref: 'o/r#7', relation: 'attached', url: 'https://github.com/o/r/pull/7' },
    scope,
  );
  assert.equal(inside.view, 'github');
  assert.deepEqual(inside.params, { scope: 'o/r', item: 'o/r#7', tab: '' });
  const outside = whereIs(
    { kind: 'pr', ref: 'x/y#7', relation: 'attached', url: 'https://github.com/x/y/pull/7' },
    scope,
  );
  assert.equal(outside.view, undefined);
  assert.equal(outside.url, 'https://github.com/x/y/pull/7');
  const key = whereIs({ kind: 'issue', ref: 'job-9', relation: 'mentions' }, scope);
  assert.equal(key.view, 'linear');
  assert.equal(key.params?.item, 'JOB-9');
  assert.equal(whereIs({ kind: 'issue', ref: 'OTHER-9', relation: 'mentions' }, scope).view, undefined);
  assert.equal(
    whereIs({ kind: 'issue', ref: 'JOB-9', relation: 'mentions' }, null).view,
    undefined,
    'no scope yet, no link',
  );
  // Only https reaches an anchor.
  assert.equal(whereIs({ kind: 'issue', ref: 'Z-1', relation: 'x', url: 'javascript:alert(1)' }, scope).url, undefined);
  assert.equal(whereIs({ kind: 'issue', ref: 'Z-1', relation: 'x', url: 'http://example.com' }, scope).url, undefined);
});

test('a status says what happened and when to try again, and says nothing when the read was fine', () => {
  const now = Date.parse('2026-11-09T12:00:00Z');
  assert.equal(statusNotice(undefined, now), '');
  assert.equal(statusNotice({ state: 'fresh' }, now), '');
  assert.equal(statusNotice({ state: 'cached' }, now), '');
  assert.equal(
    statusNotice(
      { state: 'stale', code: 'rate_limited', message: 'Linear rate limit reached.', retry_at: '2026-11-09T12:00:30Z' },
      now,
    ),
    'Showing an older copy. Linear rate limit reached. Try again in 30s.',
  );
  assert.match(
    statusNotice({ state: 'error', code: 'rate_limited', message: 'm', retry_at: '2026-11-09T12:05:00Z' }, now),
    /Try again in 5m\./,
  );
  assert.equal(statusNotice({ state: 'error', code: 'auth' }, now), 'The provider could not be read.');
  assert.doesNotMatch(
    statusNotice({ state: 'error', message: 'm', retry_at: '2026-11-09T11:00:00Z' }, now),
    /Try again/,
  );
});

test('only comments that arrived after the first read are new', () => {
  assert.deepEqual(newComments(new Set(), ['a', 'b']), [], 'the first read is not new');
  assert.deepEqual(newComments(new Set(['a']), ['a', 'b', 'c']), ['b', 'c']);
});

test('ages and durations are compact', () => {
  const now = Date.parse('2026-11-09T12:00:00Z');
  assert.equal(age('2026-11-09T11:59:40Z', now), 'just now');
  assert.equal(age('2026-11-09T11:50:00Z', now), '10m ago');
  assert.equal(age('2026-11-09T09:00:00Z', now), '3h ago');
  assert.equal(age('2026-11-07T12:00:00Z', now), '2d ago');
  assert.equal(age('0001-01-01T00:00:00Z', now), '');
  assert.equal(duration(58), '58s');
  assert.equal(duration(102), '1m42s');
  assert.equal(duration(3725), '1h02m');
  assert.equal(duration(0), '');
});

import { commentIds, mergeItems, mergeThreads } from './browse.ts';
import type { BrowseThread } from '../../../packages/jobshop-client/rpc.js';

const comment = (id: string, at: string) => ({ id, author: 'a', at, body: id });
const thread = (id: string, comments: ReturnType<typeof comment>[], hidden?: number): BrowseThread => ({
  id,
  kind: 'comment',
  comments,
  hidden,
});

test('a refreshed first page keeps the older pages already loaded and repeats nothing', () => {
  const loaded = ['A', 'B', 'C', 'D', 'E'].map((ref) => item({ ref }));
  // Page size 2: A and B were the first page, C to E were loaded after it.
  const first = ['B', 'A'].map((ref) => item({ ref }));
  assert.deepEqual(
    mergeItems(first, loaded, 2).map((i) => i.ref),
    ['B', 'A', 'C', 'D', 'E'],
  );
  // An item that moved up into the first page is not listed again below.
  const moved = ['C', 'A'].map((ref) => item({ ref }));
  assert.deepEqual(
    mergeItems(moved, loaded, 2).map((i) => i.ref),
    ['C', 'A', 'D', 'E'],
  );
});

test('older threads fold into the list by id, oldest first, without repeating a comment', () => {
  const base = [
    thread('t1', [comment('c2', '2026-11-09T10:00:00Z')]),
    thread('t2', [comment('c9', '2026-11-09T09:00:00Z')]),
  ];
  const more = [
    thread('t1', [comment('c1', '2026-11-09T08:00:00Z'), comment('c2', '2026-11-09T10:00:00Z')], 3),
    thread('t3', [comment('c0', '2026-11-08T00:00:00Z')]),
  ];
  const merged = mergeThreads(base, more);
  assert.deepEqual(
    merged.map((t) => t.id),
    ['t1', 't2', 't3'],
  );
  assert.deepEqual(
    merged[0].comments.map((c) => c.id),
    ['c1', 'c2'],
  );
  assert.equal(merged[0].hidden, 3);
  assert.deepEqual(commentIds(merged), ['c1', 'c2', 'c9', 'c0']);
  assert.deepEqual(
    base[0].comments.map((c) => c.id),
    ['c2'],
    'the inputs are not changed',
  );
});
