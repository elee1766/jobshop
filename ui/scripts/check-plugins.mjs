// Fails the build when a built-in plugin imports shell internals (#517):
// a relative import that leaves its plugin directory, or any import of
// ui/src. Built-ins must get everything from the api object, which is
// what proves the API is enough for user plugins.
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, resolve, relative, sep } from 'node:path';

const importRe = /(?:^|[\s;])(?:import|export)\s[^'"]*?from\s*['"]([^'"]+)['"]|import\s*\(\s*['"]([^'"]+)['"]\s*\)|^\s*import\s*['"]([^'"]+)['"]/gm;

/** @param {string} dir @returns {string[]} */
function files(dir) {
  return readdirSync(dir).flatMap((n) => {
    const p = join(dir, n);
    if (statSync(p).isDirectory()) return files(p);
    return /\.(js|mjs|ts|svelte)$/.test(n) ? [p] : [];
  });
}

/**
 * @param {string} pluginsDir ui/plugins
 * @returns {string[]} "file: import" for every offending import
 */
export function check(pluginsDir) {
  const bad = [];
  for (const name of readdirSync(pluginsDir)) {
    const root = join(pluginsDir, name);
    if (!statSync(root).isDirectory()) continue;
    for (const f of files(root)) {
      const text = readFileSync(f, 'utf8');
      for (const m of text.matchAll(importRe)) {
        const spec = m[1] || m[2] || m[3];
        if (!spec) continue;
        const outside = spec.startsWith('.') && relative(root, resolve(join(f, '..'), spec)).startsWith('..');
        if (outside || spec.includes('/src/') || spec.startsWith('src/')) bad.push(`${relative(pluginsDir, f).split(sep).join('/')}: ${spec}`);
      }
    }
  }
  return bad;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const dir = process.argv[2] || new URL('../plugins', import.meta.url).pathname;
  const bad = check(dir);
  if (bad.length) {
    console.error('ui/plugins must not import the shell (ui/src); use the api object:\n' + bad.join('\n'));
    process.exit(1);
  }
}
