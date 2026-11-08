import { test } from 'node:test';
import assert from 'node:assert/strict';
import { compactLabel, marked } from './links.js';

test('compactLabel recognizes GitHub PR and Linear issue URLs', () => {
  const cases = [
    ['https://github.com/elee1766/jobcity/pull/21', 'PR:elee1766/jobcity#21'],
    ['https://github.com/elee1766/jobcity/pull/21/', 'PR:elee1766/jobcity#21'],
    ['https://github.com/elee1766/jobcity/pull/21?w=1#discussion_r9', 'PR:elee1766/jobcity#21'],
    ['https://GitHub.com/acme/my.repo_x/pull/7', 'PR:acme/my.repo_x#7'],
    ['https://linear.app/gfx-labs/issue/JOB-93', 'gfx-labs@JOB-93'],
    ['https://linear.app/gfx-labs/issue/JOB-93/stall-triggers-configurable', 'gfx-labs@JOB-93'],
    ['https://linear.app/gfx-labs/issue/JOB-93#comment-8ca6629c', 'gfx-labs@JOB-93'],
    ['https://github.com/acme/widget/issues/9', ''],
    ['https://github.com/acme/widget/pull/21/files', ''],
    ['https://github.com/acme/widget/pull/0', ''],
    ['https://github.com.evil.example/acme/widget/pull/21', ''],
    ['https://evilgithub.com/acme/widget/pull/21', ''],
    ['https://www.github.com/acme/widget/pull/21', ''],
    ['https://github.com:8443/acme/widget/pull/21', ''],
    ['https://user@github.com/acme/widget/pull/21', ''],
    ['http://github.com/acme/widget/pull/21', ''],
    ['https://linear.app.evil.example/gfx-labs/issue/JOB-93', ''],
    ['https://linear.app/gfx-labs/project/JOB-93', ''],
    ['https://linear.app/gfx-labs/issue/JOB-93/slug/extra', ''],
    ['not a url', ''],
  ];
  for (const [href, want] of cases) assert.equal(compactLabel(href), want, href);
});

test('autolinks get compact labels and keep their destination', () => {
  const html = marked.parse('See https://github.com/elee1766/jobcity/pull/21?w=1#top, and <https://linear.app/gfx-labs/issue/JOB-93/some-title>.');
  assert.equal(html, '<p>See <a href="https://github.com/elee1766/jobcity/pull/21?w=1#top">PR:elee1766/jobcity#21</a>, and <a href="https://linear.app/gfx-labs/issue/JOB-93/some-title">gfx-labs@JOB-93</a>.</p>\n');
});

test('inline rendering compacts too', () => {
  assert.equal(marked.parseInline('(https://github.com/acme/widget/pull/42)'), '(<a href="https://github.com/acme/widget/pull/42">PR:acme/widget#42</a>)');
});

test('authored labels, code and other URLs are untouched', () => {
  const src = [
    '[JOB-93](https://linear.app/gfx-labs/issue/JOB-93) [https://github.com/acme/widget/pull/1](https://github.com/acme/widget/pull/1)',
    '`https://github.com/acme/widget/pull/2` https://github.com/acme/widget/issues/3 https://evilgithub.com/acme/widget/pull/4',
    '',
    '```',
    'https://linear.app/gfx-labs/issue/JOB-5',
    '```',
  ].join('\n');
  const html = marked.parse(src);
  assert.match(html, /<a href="https:\/\/linear.app\/gfx-labs\/issue\/JOB-93">JOB-93<\/a>/);
  assert.match(html, /<a href="https:\/\/github.com\/acme\/widget\/pull\/1">https:\/\/github.com\/acme\/widget\/pull\/1<\/a>/);
  assert.match(html, /<code>https:\/\/github.com\/acme\/widget\/pull\/2<\/code>/);
  assert.match(html, /<a href="https:\/\/github.com\/acme\/widget\/issues\/3">https:\/\/github.com\/acme\/widget\/issues\/3<\/a>/);
  assert.match(html, /<a href="https:\/\/evilgithub.com\/acme\/widget\/pull\/4">https:\/\/evilgithub.com\/acme\/widget\/pull\/4<\/a>/);
  assert.match(html, /<pre><code>https:\/\/linear.app\/gfx-labs\/issue\/JOB-5\n<\/code><\/pre>/);
  assert.doesNotMatch(html, /PR:|@JOB/);
});
