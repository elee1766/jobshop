// Builds each ui/plugins/<name> into pkg/desk/dist/plugins/<name> after
// the shell build (#517): one ES module (index.js), its CSS (style.css)
// and plugin.json. Built-ins are compiled as custom elements and bundle
// their own Svelte runtime, so they share nothing with the shell but
// the `api` object, like a user plugin.
import { build } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { readdirSync, copyFileSync, existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { join } from 'node:path';
import { check } from './check-plugins.mjs';

const ui = fileURLToPath(new URL('..', import.meta.url));
const src = join(ui, 'plugins');
const out = join(ui, '..', 'pkg', 'desk', 'dist', 'plugins');

const bad = check(src);
if (bad.length) {
  console.error('ui/plugins may import only their own files and ui/kit, never ui/shell; use the api object:\n' + bad.join('\n'));
  process.exit(1);
}

for (const name of readdirSync(src).sort()) {
  const dir = join(src, name);
  if (!existsSync(join(dir, 'plugin.json'))) continue;
  await build({
    configFile: false,
    root: dir,
    logLevel: 'warn',
    plugins: [svelte({ compilerOptions: { css: 'external' }, emitCss: true })],
    build: {
      outDir: join(out, name),
      emptyOutDir: true,
      assetsInlineLimit: 0,
      modulePreload: false,
      lib: { entry: join(dir, 'index.js'), formats: ['es'], fileName: () => 'index.js', cssFileName: 'style' },
    },
  });
  copyFileSync(join(dir, 'plugin.json'), join(out, name, 'plugin.json'));
}
