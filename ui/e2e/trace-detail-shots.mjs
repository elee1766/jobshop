// Screenshots of the mocked trace detail pane (JOB-185):
//   node trace-detail-shots.mjs <outdir> [name]
// Writes desktop and mobile, light and dark, for a retried step, an open step
// in a turn, and a run. It reuses the fixture of trace-detail.spec.js. Against
// a Desk built before JOB-185 it shows the old pane (name it "before").
import { chromium, devices } from '@playwright/test';
import { mkdirSync, readFileSync } from 'node:fs';

const out = process.argv[2] || '/tmp/trace-detail-shots';
const prefix = process.argv[3] || 'after';
const url = process.env.DESK_URL || 'http://127.0.0.1:7585';
const spec = readFileSync(new URL('./trace-detail.spec.js', import.meta.url), 'utf8');
// From the clock to the stream helper: the fixture and body().
const fixture = spec.slice(spec.indexOf('const base ='), spec.indexOf('async function open'));
const make = new Function(fixture + '; return { run, body };');
const { run, body } = make();

const shots = [
  ['retried', 'js-dev', ''],
  ['live', 'js-live', ''],
  ['run', 'js-run', ''],
  ['relations', 'js-dev', '&tab=relations'],
];

mkdirSync(out, { recursive: true });
const browser = await chromium.launch({ ignoreDefaultArgs: ['--hide-scrollbars'] });
for (const [name, device] of [
  ['desktop', { viewport: { width: 1000, height: 1100 } }],
  ['mobile', devices['Pixel 7']],
]) {
  for (const theme of ['light', 'dark']) {
    for (const [kind, id, extra] of shots) {
      const ctx = await browser.newContext({ ...device, colorScheme: theme });
      const page = await ctx.newPage();
      await page.route('**/api/live**', (route) => {
        const inspect = new URL(route.request().url()).searchParams.get('inspect');
        return route.fulfill({
          status: 200,
          headers: { 'content-type': 'text/event-stream' },
          body: `event: trace\ndata: ${JSON.stringify(body(run, inspect))}\n\n`,
        });
      });
      await page.goto(`${url}/?view=trace&ref=js-run&inspect=${id}${extra}`);
      await page.getByLabel('Ticket details', { exact: true }).waitFor();
      await page.evaluate((t) => document.documentElement.setAttribute('data-theme', t), theme);
      await page.screenshot({ path: `${out}/${prefix}-${kind}-${name}-${theme}.png`, fullPage: true });
      await ctx.close();
    }
  }
}
await browser.close();
