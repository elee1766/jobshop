import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseTitle, resolveHref } from './titles.js';

test('plain titles are one text part, unchanged', () => {
  for (const title of ['Fix the thing', 'a_b_c *x* `y` # h', 'owner/repo#12 and JOB-1', '<img src=x onerror=alert(1)> & "q"']) {
    const parsed = parseTitle(title);
    assert.equal(parsed.text, title);
    assert.deepEqual(parsed.parts, [{ text: title }]);
    assert.deepEqual(parsed.links, []);
  }
  assert.deepEqual(parseTitle('').parts, []);
  assert.equal(parseTitle(undefined).text, '');
});

test('GitHub and Linear links become link parts with readable text', () => {
  const parsed = parseTitle('Review [jobshop#1021](https://github.com/elee1766/jobshop/pull/1021) for [JOB-117](https://linear.app/gfx-labs/issue/JOB-117)');
  assert.equal(parsed.text, 'Review jobshop#1021 for JOB-117');
  assert.deepEqual(parsed.parts, [
    { text: 'Review ' },
    { text: 'jobshop#1021', link: { href: 'https://github.com/elee1766/jobshop/pull/1021', external: true } },
    { text: ' for ' },
    { text: 'JOB-117', link: { href: 'https://linear.app/gfx-labs/issue/JOB-117', external: true } },
  ]);
  assert.deepEqual(parsed.links.map((link) => link.text), ['jobshop#1021', 'JOB-117']);
});

test('adjacent links and a leading link keep their order', () => {
  const parsed = parseTitle('[a](https://a.example/)[b](https://b.example/) tail');
  assert.deepEqual(parsed.parts.map((part) => part.text), ['a', 'b', ' tail']);
});

test('unsafe destinations keep their label and lose the link', () => {
  for (const url of ['javascript:alert%281%29', 'data:text/html,x', 'jobshop://command/delete', 'jobshop://ticket/%2Fetc', '/relative', 'JaVaScRiPt:alert']) {
    const parsed = parseTitle(`[Click](${url}) here`);
    assert.deepEqual(parsed.links, [], url);
    assert.deepEqual(parsed.parts, [{ text: 'Click here' }], url);
  }
  // A destination with parentheses is not a link at all, so it stays literal.
  assert.equal(parseTitle('[x](javascript:alert(1))').text, '[x](javascript:alert(1))');
});

test('jobshop ticket links become Desk trace links', () => {
  assert.deepEqual(resolveHref('jobshop://ticket/js-123'), { href: '?view=trace&ref=js-123', external: false });
  assert.deepEqual(parseTitle('see [child](jobshop://session/js-s-9)').links, [{ text: 'child', href: '?view=trace&ref=js-s-9', external: false }]);
});

test('angle-bracket destinations and link titles are accepted', () => {
  assert.equal(parseTitle('[a](<https://example.com/x> "tip")').links.length, 1);
});
