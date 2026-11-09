// Screenshots of the mocked timeline (#1103): node shot.mjs <outdir>
import { chromium, devices } from '@playwright/test';
import { readFileSync } from 'node:fs';

const out = process.argv[2] || '/tmp/tl1103/shots';
const url = process.env.DESK_URL || 'http://127.0.0.1:7561';
const spec = readFileSync(new URL('./timeline.spec.js', import.meta.url), 'utf8');
// Reuse the spec's fixture: everything from "const at" to "async function open".
const fixture = spec.slice(spec.indexOf('const at ='), spec.indexOf('async function open'));
const make = new Function(fixture + '; return { run, dev, review, ask, timeline };');
const { run, dev, review, ask, timeline } = make();

const trace = {
  ticket: run,
  edges: [],
  comments: [],
  history: [],
  graph: { root: 'js-run', nodes: [run, dev, review, ask], relations: [], external: [], truncated: false },
  timeline,
};

const browser = await chromium.launch({ ignoreDefaultArgs: ['--hide-scrollbars'] });
for (const [name, device] of [
  ['desktop', { viewport: { width: 1100, height: 800 } }],
  ['mobile', devices['Pixel 7']],
]) {
  for (const theme of ['light', 'dark']) {
    const ctx = await browser.newContext({ ...device, colorScheme: theme });
    const page = await ctx.newPage();
    await page.route('**/api/live**', (route) =>
      route.fulfill({
        status: 200,
        headers: { 'content-type': 'text/event-stream' },
        body: `event: trace\ndata: ${JSON.stringify(trace)}\n\n`,
      }),
    );
    await page.goto(url + '/?view=trace&ref=js-run&layout=timeline');
    await page.getByRole('region', { name: 'Ticket timeline' }).waitFor();
    await page.evaluate((t) => document.documentElement.setAttribute('data-theme', t), theme);
    await page.screenshot({ path: `${out}/${name}-${theme}.png`, fullPage: true });
    await ctx.close();
  }
}
await browser.close();
