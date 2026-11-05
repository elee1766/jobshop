// Fails the build when a built-in plugin imports shell internals (#517):
// a relative import that leaves its plugin directory for anywhere but
// ui/kit (#541), or any import of ui/shell. Built-ins talk to Desk only
// through the api object, which proves the API is enough for user
// plugins; ui/kit is plain helpers each plugin bundles its own copy of.
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
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
  const kit = resolve(pluginsDir, '..', 'kit');
  const shell = resolve(pluginsDir, '..', 'shell');
  // Kit files must not reach into the shell either.
  if (existsSync(kit)) {
    for (const f of files(kit)) {
      for (const spec of imports(f)) {
        const to = spec.startsWith('.') ? resolve(join(f, '..'), spec) : '';
        if ((to && relative(kit, to).startsWith('..')) || spec.includes('/shell/')) bad.push(`kit/${relative(kit, f).split(sep).join('/')}: ${spec}`);
      }
    }
  }
  for (const name of readdirSync(pluginsDir)) {
    const root = join(pluginsDir, name);
    if (!statSync(root).isDirectory()) continue;
    for (const f of files(root)) {
      for (const spec of imports(f)) {
        const to = spec.startsWith('.') ? resolve(join(f, '..'), spec) : '';
        const outside = to && relative(root, to).startsWith('..') && relative(kit, to).startsWith('..');
        const intoShell = (to && !relative(shell, to).startsWith('..')) || spec.includes('/shell/') || spec.startsWith('shell/');
        if (outside || intoShell) bad.push(`${relative(pluginsDir, f).split(sep).join('/')}: ${spec}`);
      }
    }
  }
  return bad;
}

/** @param {string} f @returns {string[]} */
function imports(f) {
  const text = readFileSync(f, 'utf8');
  return [...text.matchAll(importRe)].map((m) => m[1] || m[2] || m[3]).filter(Boolean);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const dir = process.argv[2] || new URL('../plugins', import.meta.url).pathname;
  const bad = check(dir);
  if (bad.length) {
    console.error('ui/plugins may import only their own files and ui/kit, never ui/shell; use the api object:\n' + bad.join('\n'));
    process.exit(1);
  }
}
