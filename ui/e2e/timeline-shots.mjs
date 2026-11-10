// Screenshots of the mocked timeline (#1103, JOB-184): node timeline-shots.mjs <outdir>
// Writes desktop and mobile, light and dark, for the base fixture and the
// turn fixture (turns inside claims and the three live states), also with
// reduced motion.
import { chromium, devices } from '@playwright/test';
import { readFileSync } from 'node:fs';

const out = process.argv[2] || '/tmp/tl1103/shots';
const url = process.env.DESK_URL || 'http://127.0.0.1:7561';
const spec = readFileSync(new URL('./timeline.spec.js', import.meta.url), 'utf8');
// Reuse the spec's fixtures: from "const at" to "async function open", and
// the turn fixture from its marker to the first turn test.
const fixture = spec.slice(spec.indexOf('const at ='), spec.indexOf('async function open'));
const turnFixture = spec.slice(spec.indexOf('// JOB-184'), spec.indexOf("test('a claim shows its turns"));
const make = new Function(fixture + turnFixture + '; return { run, dev, review, ask, timeline, turnTimeline, holders };');
const { run, dev, review, ask, timeline, turnTimeline, holders } = make();

const cases = {
  base: { timeline, sessions: undefined },
  turns: { timeline: turnTimeline, sessions: holders },
};

const browser = await chromium.launch({ ignoreDefaultArgs: ['--hide-scrollbars'] });
for (const [name, device] of [
  ['desktop', { viewport: { width: 1100, height: 800 } }],
  ['mobile', devices['Pixel 7']],
]) {
  for (const theme of ['light', 'dark']) {
    for (const [kind, c] of Object.entries(cases)) {
      for (const motion of kind === 'turns' ? ['no-preference', 'reduce'] : ['no-preference']) {
        const trace = {
          ticket: run,
          edges: [],
          comments: [],
          history: [],
          graph: { root: 'js-run', nodes: [run, dev, review, ask], relations: [], external: [], truncated: false },
          timeline: c.timeline,
          ...(c.sessions ? { sessions: c.sessions } : {}),
        };
        const ctx = await browser.newContext({ ...device, colorScheme: theme, reducedMotion: motion });
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
        const suffix = motion === 'reduce' ? '-still' : '';
        await page.screenshot({ path: `${out}/${name}-${theme}${kind === 'turns' ? '-turns' : ''}${suffix}.png`, fullPage: true });
        await ctx.close();
      }
    }
  }
}
await browser.close();
