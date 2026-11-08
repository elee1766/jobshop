// Dev startup and HMR baseline (JOB-150). Starts `vite` against a running
// Desk, then times: server ready, first page load through the dev server,
// and a source edit until the browser reports the hot update.
//
//   DESK_URL=http://127.0.0.1:7562 node ui/perf/dev.mjs <file to edit> [--runs 5] [--json out.json]
//
// The edited file gets one trailing comment line appended per run and is
// restored afterwards. Run it from ui/. It rewrites the file in place, so
// keep the worktree clean first.
import { chromium } from '@playwright/test';
import { spawn } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';

const target = process.argv[2];
const arg = (name, fallback) => {
  const at = process.argv.indexOf('--' + name);
  return at > 0 ? process.argv[at + 1] : fallback;
};
const runs = Number(arg('runs', 5));
const out = arg('json', '');
const port = Number(arg('port', 5199));
// A comment compiles to identical output and sends no update, so the edit
// adds a hidden element (svelte) or a hidden span (any other file must
// pass --snippet).
const comment = target.endsWith(".svelte") ? (s) => `\n<i hidden>hmr ${s}</i>\n` : (s) => arg("snippet", `\n// hmr ${s}\n`).replaceAll("{N}", s);
const median = (xs) => [...xs].sort((a, b) => a - b)[xs.length >> 1];

const original = readFileSync(target, 'utf8');
const started = Date.now();
const vite = spawn('./node_modules/.bin/vite', ['--port', String(port), '--strictPort', '--host', '127.0.0.1'], {
  env: { ...process.env, JOBSHOP_DESK_TARGET: process.env.DESK_URL },
  stdio: ['ignore', 'pipe', 'inherit'],
});
const ready = await new Promise((resolve, reject) => {
  let log = '';
  vite.stdout.on('data', (d) => {
    log += d;
    const m = /ready in (\d+) ms/.exec(log);
    if (m) resolve({ reportedMs: Number(m[1]), wallMs: Date.now() - started });
  });
  vite.on('exit', () => reject(new Error('vite exited: ' + log)));
});

const result = { serverReadyReportedMs: ready.reportedMs, serverReadyWallMs: ready.wallMs };
const browser = await chromium.launch();
try {
  const page = await (await browser.newContext({ viewport: { width: 1280, height: 800 } })).newPage();
  const updates = [];
  // The server's websocket frames: an "update" or "full-reload" message
  // is the moment the browser learns of the change; the console line Vite
  // prints for it is debug level and not reliable.
  page.on('websocket', (ws) =>
    ws.on('framereceived', (f) => {
      const text = typeof f.payload === 'string' ? f.payload : f.payload.toString();
      if (process.env.HMR_DEBUG) console.error('ws:', text.slice(0, 160));
      if (/"type":\s*"(update|full-reload)"/.test(text)) updates.push({ at: Date.now(), text });
    }),
  );
  const t0 = Date.now();
  await page.goto(`http://127.0.0.1:${port}/`);
  await page.getByText('sys ok').first().waitFor();
  result.firstLoadColdMs = Date.now() - t0;
  const t1 = Date.now();
  await page.reload();
  await page.getByText('sys ok').first().waitFor();
  result.reloadWarmMs = Date.now() - t1;

  const times = [];
  for (let i = 0; i < runs; i++) {
    await new Promise((r) => setTimeout(r, 400)); // let the watcher settle between edits
    const before = updates.length;
    const at = Date.now();
    writeFileSync(target, original + comment(`${i} ${at}`));
    for (let spin = 0; updates.length === before && spin < 200; spin++) await new Promise((r) => setTimeout(r, 25));
    if (updates.length === before) throw new Error('no hot update seen for ' + target);
    times.push(updates[before].at - at);
  }
  result.hmrMs = median(times);
  result.hmrAllMs = times;
} finally {
  writeFileSync(target, original);
  await browser.close();
  vite.kill();
}
console.log(JSON.stringify(result, null, 2));
if (out) writeFileSync(out, JSON.stringify(result, null, 2) + '\n');
