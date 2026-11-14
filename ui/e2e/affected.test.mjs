import assert from 'node:assert/strict';
import { readdirSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { affected, mapped, rules, specsFor } from './affected.mjs';
import { isolated } from './groups.mjs';

const ui = resolve(dirname(fileURLToPath(import.meta.url)), '..');

function walk(dir) {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) => (e.isDirectory() ? walk(join(dir, e.name)) : [join(dir, e.name)]));
}
const sources = (sub) => walk(join(ui, sub)).map((f) => 'ui/' + relative(ui, f));

test('every spec the groups name exists', () => {
  const have = new Set(readdirSync(join(ui, 'e2e')));
  for (const spec of isolated) assert.ok(have.has(spec), `${spec} is in groups.mjs but not in e2e/`);
});

test('every spec the table names is one run-isolated runs', () => {
  for (const [re, specs] of rules) {
    if (specs === 'self') continue;
    for (const spec of specs) assert.ok(isolated.includes(spec), `${re}: ${spec} is not in groups.mjs`);
  }
});

test('every Desk source file has a rule, not the run-everything default', () => {
  for (const file of [...sources('src'), ...sources('kit'), ...sources('e2e')]) {
    assert.ok(mapped(file), `${file}: add it to the table in e2e/affected.mjs`);
  }
});

test('a sample diff selects the specs for what it touches', () => {
  assert.deepEqual(affected(['ui/src/components/PoolControls.tsx']), ['pool-controls.spec.js']);
  assert.deepEqual(affected(['ui/src/screens/Embed.tsx']), ['embeds.spec.js', 'embed-reload.spec.js']);
  assert.deepEqual(affected(['ui/e2e/browse.spec.js', 'ui/src/screens/Embed.tsx']), ['embeds.spec.js', 'browse.spec.js', 'embed-reload.spec.js']);
  const chat = affected(['ui/src/screens/Chat.tsx']);
  for (const spec of ['message-headers.spec.js', 'turn-phase.spec.js']) assert.ok(chat.includes(spec), spec);
  const graph = affected(['ui/src/screens/Timeline.tsx']);
  for (const spec of ['trace-axis.spec.js', 'trace-detail.spec.js', 'timeline.spec.js']) assert.ok(graph.includes(spec), spec);
  assert.ok(!graph.includes('pool-controls.spec.js'));
});

test('shared code and the Desk API run everything; other files and unit tests run nothing', () => {
  assert.deepEqual(affected(['ui/src/app.css']), isolated);
  assert.deepEqual(affected(['pkg/transport/desk/server.go']), isolated);
  assert.deepEqual(affected(['ui/src/lib/unknown-new-file.ts']), isolated);
  assert.deepEqual(affected(['docs/notes/x.md', 'pkg/control/pass.go', 'pkg/transport/desk/dist/x.js']), []);
  assert.deepEqual(specsFor('ui/src/lib/route.test.ts'), []);
});
