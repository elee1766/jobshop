import { chromium, devices } from '@playwright/test';

// Screenshots for review: node e2e/shots.mjs <dir>. Not part of the suite.
// Also fails if any view scrolls horizontally at 360px.
const base = process.env.DESK_URL || 'http://127.0.0.1:7480';
const out = process.argv[2] || '.';
const views = ['attention', 'chat', 'work', 'activity', 'sessions', 'pools', 'deploys'];
const browser = await chromium.launch();
let bad = 0;

/** @param {import('@playwright/test').Page} page @param {string} path */
async function open(page, path) {
  await page.goto(base + path);
  await page.locator('main .vh h1').waitFor();
  await page.waitForTimeout(600);
}

for (const [name, dev] of [['desktop', { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 800 } }], ['mobile', devices['Pixel 7']], ['narrow', { ...devices['Pixel 7'], viewport: { width: 360, height: 740 } }]]) {
  const ctx = await browser.newContext(dev);
  const page = await ctx.newPage();
  for (const view of views) {
    await open(page, '/?view=' + view);
    if (name !== 'narrow') await page.screenshot({ path: `${out}/${name}-${view}.png` });
    const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    if (over > 0) { bad++; console.log('OVERFLOW', name, view, over); }
  }
  // Open items: first run drawer, first session journal, first chat session.
  await open(page, '/?view=work');
  const run = page.locator('[data-testid=runs] tr.click').first();
  if (await run.count()) {
    await run.click();
    await page.getByTestId('drawer').waitFor();
    await page.waitForTimeout(600);
    if (name !== 'narrow') await page.screenshot({ path: `${out}/${name}-drawer.png` });
    const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    if (over > 0) { bad++; console.log('OVERFLOW', name, 'drawer', over); }
  }
  await open(page, '/?view=sessions');
  const sess = page.locator('[data-testid=sessions] tr.click').first();
  if (await sess.count()) {
    await sess.click();
    await page.waitForTimeout(800);
    if (name !== 'narrow') await page.screenshot({ path: `${out}/${name}-session.png` });
    const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    if (over > 0) { bad++; console.log('OVERFLOW', name, 'session', over); }
  }
  await open(page, '/?view=chat');
  const item = page.locator('.split .item').first();
  if (await item.count()) {
    await item.click();
    await page.waitForTimeout(800);
    if (name !== 'narrow') await page.screenshot({ path: `${out}/${name}-chat-open.png` });
    const over = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    if (over > 0) { bad++; console.log('OVERFLOW', name, 'chat-open', over); }
  }
  await ctx.close();
}
await browser.close();
process.exit(bad ? 1 : 0);
