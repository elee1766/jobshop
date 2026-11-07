import { mkdir } from 'node:fs/promises';
import { chromium, devices } from '@playwright/test';

// Review the live Desk at each layout breakpoint without fixed delays.
const base = new URL(process.env.DESK_URL || 'http://127.0.0.1:7480/');
if (!base.pathname.endsWith('/')) base.pathname += '/';
const out = process.argv[2] || '.';
await mkdir(out, { recursive: true });
const browser = await chromium.launch();
let bad = 0;
try {
  for (const width of [1000, 790, 700, 412, 360]) {
    const context = await browser.newContext({ ...devices['Desktop Chrome'], viewport: { width, height: 900 } });
    const page = await context.newPage();
    page.on('pageerror', (error) => { bad++; console.error(error); });
    for (const view of ['home', 'trace', 'improvements', 'system']) {
      await page.goto(new URL('?view=' + view, base).href);
      // A populated system section proves the initial live snapshots arrived.
      await page.waitForFunction(() => document.querySelector('.status')?.textContent?.includes('tick'));
      await page.evaluate(() => document.fonts.ready);
      await page.screenshot({ path: `${out}/${width}-${view}.png` });
      const over = await page.evaluate(() => document.documentElement.scrollWidth - innerWidth);
      if (over > 1) { bad++; console.error('OVERFLOW', width, view, over); }
    }
    await context.close();
  }
} finally {
  await browser.close();
}
process.exitCode = bad ? 1 : 0;
