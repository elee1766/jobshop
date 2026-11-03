import { chromium, devices } from '@playwright/test';

// Screenshots for review: node e2e/shots.mjs <dir>. Not part of the suite.
const base = process.env.DESK_URL || 'http://127.0.0.1:7480';
const out = process.argv[2] || '.';
const browser = await chromium.launch();
for (const [name, dev] of [['desktop', devices['Desktop Chrome']], ['mobile', devices['Pixel 7']]]) {
  const page = await (await browser.newContext({ ...dev })).newPage();
  for (const view of ['attention', 'chat', 'work', 'sessions']) {
    await page.goto(base + '/?view=' + view);
    await page.locator('main h1').waitFor();
    await page.waitForTimeout(500);
    await page.screenshot({ path: `${out}/${name}-${view}.png` });
  }
  await page.goto(base + '/?view=chat');
  await page.locator('main h1').waitFor();
  await page.waitForTimeout(500);
  if (await page.locator('.list a').count()) {
    await page.locator('.list a').first().click();
    await page.waitForTimeout(800);
    await page.screenshot({ path: `${out}/${name}-chat-open.png` });
  }
}
await browser.close();
