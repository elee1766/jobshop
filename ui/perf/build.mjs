// Desk build and bundle size baseline (JOB-150). Framework neutral: it runs
// `vite build` and `npm run check` from ui/, so the same script measures the
// Svelte app and the React port.
//
//   cd ui && node perf/build.mjs [--json out.json]
//
// build1 has no Vite cache (node_modules/.vite removed first), build2 and
// build3 are warm; warmBuildS is their median. Sizes are the sum over the
// emitted JS and CSS files, gzip level 9 and brotli default.
import { execFileSync } from 'node:child_process';
import { mkdtempSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { brotliCompressSync, gzipSync } from 'node:zlib';

const at = process.argv.indexOf('--json');
const out = at > 0 ? process.argv[at + 1] : '';
const scratch = mkdtempSync(join(tmpdir(), 'desk-build-'));
const dist = join(scratch, 'dist');

const seconds = (cmd, args) => {
  const start = process.hrtime.bigint();
  execFileSync(cmd, args, { stdio: 'pipe' });
  return Number(process.hrtime.bigint() - start) / 1e9;
};
const median = (xs) => [...xs].sort((a, b) => a - b)[xs.length >> 1];
const round = (n, d = 2) => Number(n.toFixed(d));

rmSync('node_modules/.vite', { recursive: true, force: true });
const vite = ['build', '--emptyOutDir', '--outDir', dist];
const builds = [1, 2, 3].map(() => seconds('./node_modules/.bin/vite', vite));
const checks = [1, 2, 3].map(() => seconds('npm', ['run', '--silent', 'check']));

let js = 0, css = 0, jsGz = 0, cssGz = 0, jsBr = 0, cssBr = 0;
const assets = join(dist, 'assets');
for (const name of readdirSync(assets)) {
  const buf = readFileSync(join(assets, name));
  const gz = gzipSync(buf, { level: 9 }).length;
  const br = brotliCompressSync(buf).length;
  if (name.endsWith('.js')) { js += buf.length; jsGz += gz; jsBr += br; }
  if (name.endsWith('.css')) { css += buf.length; cssGz += gz; cssBr += br; }
}
rmSync(scratch, { recursive: true, force: true });

const result = {
  coldBuildS: round(builds[0]),
  warmBuildS: round(median(builds.slice(1))),
  checkS: round(median(checks)),
  jsKB: round(js / 1024, 1),
  cssKB: round(css / 1024, 1),
  jsGzipKB: round(jsGz / 1024, 1),
  cssGzipKB: round(cssGz / 1024, 1),
  jsBrotliKB: round(jsBr / 1024, 1),
  cssBrotliKB: round(cssBr / 1024, 1),
};
console.log(JSON.stringify({ build: result }, null, 2));
if (out) writeFileSync(out, JSON.stringify({ build: result }, null, 2) + '\n');
