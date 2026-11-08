// Long chat transcript benchmark (JOB-150). The transcript is the virtualized
// view the port must not regress: open time, rendered row count, scroll to
// the start with history paging, and heap. Framework neutral (labels and
// text only). Needs an isolated controller with a chat pool backed by the
// ACP test agent (docs/notes/JOB-150-desk-react-baseline.md).
//
//   DESK_URL=http://127.0.0.1:7563 node ui/perf/chat.mjs [--turns 60] [--runs 5] [--json out.json]
import { chromium } from '@playwright/test';
import { writeFileSync } from 'node:fs';

const desk = (process.env.DESK_URL || 'http://127.0.0.1:7480').replace(/\/?$/, '/');
const arg = (name, fallback) => {
  const at = process.argv.indexOf('--' + name);
  return at > 0 ? process.argv[at + 1] : fallback;
};
const turns = Number(arg('turns', 60));
const runs = Number(arg('runs', 5));
const out = arg('json', '');
const median = (xs) => [...xs].sort((a, b) => a - b)[xs.length >> 1];

const browser = await chromium.launch();
const api = await browser.newContext({ extraHTTPHeaders: { 'X-Jobshop-Desk': '1' } });
const post = async (path, data) => {
  const r = await api.request.post(new URL(path, desk).href, { data });
  if (!r.ok()) throw new Error(path + ': ' + (await r.text()));
  return r.json();
};

const name = 'bench-' + Date.now();
const first = 'first message ' + name;
const session = await post('api/chat/manage', { action: 'new', name, text: first });
for (let i = 0; i < turns; i++) {
  // A send to a busy chat is queued and delivered in order; the page wait
  // below covers the last turn.
  await post('api/chat/send', { key: session.ticket.fields.key, text: `turn ${i} ${'lorem ipsum '.repeat(i % 7)}` });
}

const samples = [];
try {
  for (let run = 0; run < runs; run++) {
    const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
    const page = await context.newPage();
    const cdp = await context.newCDPSession(page);
    await cdp.send('Performance.enable');
    await page.addInitScript(() => {
      window.__long = [];
      new PerformanceObserver((l) => l.getEntries().forEach((e) => window.__long.push(e.duration))).observe({ entryTypes: ['longtask'] });
    });
    await page.goto(desk);
    await page.getByRole('button', { name: 'Choose chat', exact: true }).click();
    await page.getByRole('searchbox', { name: 'Search chats', exact: true }).fill(name);
    const t0 = Date.now();
    await page.getByRole('table', { name: 'Recent chats', exact: true }).getByRole('button', { name, exact: true }).click();
    const transcript = page.getByLabel('Chat transcript', { exact: true });
    // The agent may still be draining the queued turns on the first run.
    await transcript.getByText(`turn ${turns - 1} `).first().waitFor({ timeout: 180000 });
    const openMs = Date.now() - t0;
    const rowsAtOpen = await transcript.evaluate((el) => el.querySelectorAll('[data-row-key]').length);

    // Walk to the start: scroll to the top until the start marker shows.
    const t1 = Date.now();
    let maxRows = rowsAtOpen;
    for (let spin = 0; spin < 400; spin++) {
      await transcript.evaluate((el) => { el.scrollTop = 0; });
      await page.waitForTimeout(40);
      maxRows = Math.max(maxRows, await transcript.evaluate((el) => el.querySelectorAll('[data-row-key]').length));
      if (await transcript.getByText('start of conversation', { exact: true }).count()) break;
    }
    const toStartMs = Date.now() - t1;
    await cdp.send('HeapProfiler.collectGarbage');
    const { metrics } = await cdp.send('Performance.getMetrics');
    const m = Object.fromEntries(metrics.map((x) => [x.name, x.value]));
    const long = await page.evaluate(() => window.__long);
    samples.push({
      openMs, rowsAtOpen, maxRenderedRows: maxRows, toStartMs,
      heapMB: m.JSHeapUsedSize / 1048576, nodes: m.Nodes,
      longTasks: long.length, longMaxMs: Math.max(0, ...long),
    });
    await context.close();
  }
} finally {
  await post('api/chat/manage', { action: 'archive', id: session.id }).catch(() => {});
  await browser.close();
}
const result = { chat: { turns, ...Object.fromEntries(Object.keys(samples[0]).map((k) => [k, Number(median(samples.map((s) => s[k])).toFixed(2))])) } };
console.log(JSON.stringify(result, null, 2));
if (out) writeFileSync(out, JSON.stringify(result, null, 2) + '\n');
