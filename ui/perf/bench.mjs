// Desk runtime benchmark (JOB-150). Framework neutral: it drives the page
// through roles and text only, so the same script measures the Svelte
// baseline and the React port. Run it against an isolated controller
// (it creates tickets): see docs/notes/JOB-150-desk-react-baseline.md.
//
//   DESK_URL=http://127.0.0.1:7561 node ui/perf/bench.mjs [--runs 5] [--json out.json]
//
// Each metric is the median of --runs fresh browser contexts.
import { chromium, devices } from '@playwright/test';
import { writeFileSync } from 'node:fs';

const desk = (process.env.DESK_URL || 'http://127.0.0.1:7480').replace(/\/?$/, '/');
const arg = (name, fallback) => {
  const at = process.argv.indexOf('--' + name);
  return at > 0 ? process.argv[at + 1] : fallback;
};
const runs = Number(arg('runs', 5));
const burst = Number(arg('burst', 50));
const out = arg('json', '');

const median = (xs) => {
  const s = [...xs].sort((a, b) => a - b);
  const m = s.length >> 1;
  return s.length % 2 ? s[m] : (s[m - 1] + s[m]) / 2;
};

async function rpc(request, method, data) {
  const r = await request.post(new URL('api/rpc/' + method, desk).href, {
    headers: { 'X-Jobshop-Desk': '1' },
    data,
  });
  if (!r.ok()) throw new Error(method + ': ' + (await r.text()));
  return r.json();
}

// cdp reads Chromium's own counters: heap after a forced GC, script and
// total task time, and layout and style recalculation counts.
async function counters(cdp) {
  await cdp.send('HeapProfiler.collectGarbage');
  const { metrics } = await cdp.send('Performance.getMetrics');
  const m = Object.fromEntries(metrics.map((x) => [x.name, x.value]));
  return {
    heapMB: m.JSHeapUsedSize / 1048576,
    scriptMs: m.ScriptDuration * 1000,
    taskMs: m.TaskDuration * 1000,
    layouts: m.LayoutCount,
    recalcs: m.RecalcStyleCount,
    nodes: m.Nodes,
  };
}

async function measure(browser, profile) {
  const context = await browser.newContext(profile === 'mobile' ? { ...devices['Pixel 7'] } : { viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();
  const cdp = await context.newCDPSession(page);
  await cdp.send('Performance.enable');
  const result = {};
  const live = [];
  page.on('request', (req) => {
    if (new URL(req.url()).pathname.endsWith('/api/live')) live.push(req.url());
  });
  const errors = [];
  page.on('pageerror', (e) => errors.push(String(e)));
  page.on('console', (m) => m.type() === 'error' && errors.push(m.text()));
  // Long tasks (>50 ms) are the interaction-latency proxy for bursts.
  await page.addInitScript(() => {
    window.__long = [];
    new PerformanceObserver((list) => {
      for (const e of list.getEntries()) window.__long.push(e.duration);
    }).observe({ entryTypes: ['longtask'] });
  });

  // First usable render: navigation start until the first waiting row and
  // the connection indicator are on screen.
  const t0 = Date.now();
  await page.goto(desk, { waitUntil: 'commit' });
  await page.getByText('sys ok').first().waitFor();
  await page.getByRole('region', { name: 'Waiting', exact: true }).getByText(/Needs a decision/).first().waitFor();
  result.firstUsableMs = Date.now() - t0;
  const nav = await page.evaluate(() => {
    const n = performance.getEntriesByType('navigation')[0];
    const paints = Object.fromEntries(performance.getEntriesByType('paint').map((p) => [p.name, p.startTime]));
    return { domContentLoaded: n.domContentLoadedEventEnd, fcp: paints['first-contentful-paint'] };
  });
  result.fcpMs = nav.fcp;
  result.transferKB = (await page.evaluate(() => performance.getEntriesByType('resource').reduce((s, r) => s + (r.transferSize || 0), 0))) / 1024;

  await page.waitForTimeout(1500);
  const idle = await counters(cdp);
  result.heapIdleMB = idle.heapMB;
  result.nodes = idle.nodes;
  result.scriptMsLoad = idle.scriptMs;

  // Route switches: click to the first content of the target view.
  const sw = async (name, ready) => {
    // Phone widths hide the top nav behind the 'more' tab.
    if (profile === 'mobile') await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
    const s = Date.now();
    await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name, exact: true }).click();
    await ready.first().waitFor();
    return Date.now() - s;
  };
  result.switchLearnMs = await sw('learn', page.getByRole('textbox', { name: 'Filter learn' }));
  result.switchSysMs = await sw('sys', page.getByRole('heading', { name: 'pools' }));
  result.switchHomeMs = await sw('home', page.getByRole('region', { name: 'In flight', exact: true }));

  // Large-list scrolling: scroll the work pane through the whole in-flight
  // table and record the frame gaps.
  const frames = await page.evaluate(async () => {
    const pane = document.querySelector('main') || document.scrollingElement;
    const gaps = [];
    let last = performance.now();
    let y = 0;
    await new Promise((done) => {
      const step = () => {
        const now = performance.now();
        gaps.push(now - last);
        last = now;
        y += 60;
        pane.scrollTop = y;
        if (y < pane.scrollHeight) requestAnimationFrame(step);
        else done();
      };
      requestAnimationFrame(step);
    });
    pane.scrollTop = 0;
    return gaps;
  });
  result.scrollFrames = frames.length;
  result.scrollP95FrameMs = [...frames].sort((a, b) => a - b)[Math.floor(frames.length * 0.95)] || 0;

  // SSE burst: create tickets that need a human faster than the page can
  // render one by one and time until the last shows up in Waiting.
  await page.evaluate(() => (window.__long.length = 0));
  const before = await counters(cdp);
  const tag = 'burst-' + Date.now() + '-' + profile;
  const t1 = Date.now();
  const made = await Promise.all(
    Array.from({ length: burst }, (_, i) => rpc(context.request, 'ticket.create', { title: `${tag} ${i}`, labels: ['needs-human'] })),
  );
  await page.getByRole('region', { name: 'Waiting', exact: true }).getByText(`${tag} ${burst - 1}`).waitFor({ timeout: 30000 });
  result.burstMs = Date.now() - t1;
  const after = await counters(cdp);
  result.burstScriptMs = after.scriptMs - before.scriptMs;
  result.burstTaskMs = after.taskMs - before.taskMs;
  const long = await page.evaluate(() => window.__long);
  result.burstLongTasks = long.length;
  result.burstLongMaxMs = Math.max(0, ...long);
  result.heapAfterBurstMB = after.heapMB;
  await Promise.all(made.map((t) => rpc(context.request, 'ticket.close', { id: t.id, resolution: 'done' }).catch(() => {})));

  // Connections: one api/live for the whole session across the route
  // switches above (reconnects on view changes are a regression).
  result.liveRequests = live.length;
  result.errors = errors.length;
  await context.close();
  return result;
}

const browser = await chromium.launch();
const all = {};
for (const profile of ['desktop', 'mobile']) {
  const samples = [];
  for (let i = 0; i < runs; i++) samples.push(await measure(browser, profile));
  all[profile] = Object.fromEntries(Object.keys(samples[0]).map((k) => [k, Number(median(samples.map((s) => s[k])).toFixed(2))]));
}
await browser.close();
console.log(JSON.stringify(all, null, 2));
if (out) writeFileSync(out, JSON.stringify(all, null, 2) + '\n');
