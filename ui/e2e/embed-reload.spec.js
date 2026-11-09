import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
import { test, expect } from '@playwright/test';

// Config reload with an embed open (JOB-151). Rewrites jobshop.yaml of an
// isolated controller, so it runs only when EMBED_ROOT names that root and
// EMBED_JOBSHOP its binary (docs/verify/desk-embeds.md). Never point it at a
// live instance.
const root = process.env.EMBED_ROOT;
const bin = process.env.EMBED_JOBSHOP || 'jobshop';
test.skip(!root, 'set EMBED_ROOT to an isolated controller root');
test.describe.configure({ mode: 'serial' });

const file = root + '/jobshop.yaml';
let original = '';
const reload = () =>
  execFileSync(bin, ['--root', root, 'reload'], {
    env: { PATH: process.env.PATH, HOME: process.env.HOME },
  });
const withEmbeds = (extra) =>
  original.replace(
    /desk:[\s\S]*$/,
    `desk:\n  embeds:\n    - id: cross\n      label: Cross app\n      url: http://localhost:7497/app\n${extra}`,
  );

test.beforeAll(() => {
  original = readFileSync(file, 'utf8');
});
test.afterEach(() => {
  writeFileSync(file, original);
  reload();
});

test('removing the active embed drops its frame and tab without a reload', async ({ page, isMobile }) => {
  await page.goto('./?view=embed&embed=cross');
  await expect(page.locator('iframe.embed-frame')).toHaveCount(1);
  writeFileSync(file, original.replace(/desk:[\s\S]*$/, 'desk:\n  embeds: []\n'));
  reload();
  // The frame goes at once: no stale app stays on screen.
  await expect(page.locator('iframe')).toHaveCount(0);
  await expect(page.getByText('No embedded app "cross" is configured.')).toBeVisible();
  if (isMobile) await page.getByRole('navigation', { name: 'Mobile' }).getByRole('button', { name: 'more' }).click();
  await expect(page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name: 'Cross app' })).toHaveCount(0);
  // Built-in navigation still works.
  await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name: 'sys', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'pools' })).toBeVisible();
});

test('a changed address replaces the frame, an invalid config keeps the old tabs', async ({ page }) => {
  await page.goto('./?view=embed&embed=cross');
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://localhost:7497/app');
  // The origin is already allowed by the page, so the new address loads in place.
  writeFileSync(file, original.replace(/desk:[\s\S]*$/, 'desk:\n  embeds:\n    - id: cross\n      label: Cross app\n      url: http://localhost:7497/app?v=2\n'));
  reload();
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://localhost:7497/app?v=2');
  // A rejected config is refused by the controller and nothing changes.
  writeFileSync(file, original.replace(/desk:[\s\S]*$/, 'desk:\n  embeds:\n    - id: cross\n      label: Cross app\n      url: "javascript:alert(1)"\n'));
  expect(() => reload()).toThrow();
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://localhost:7497/app?v=2');
});

test('an embed added after Desk loaded needs a page reload and says so', async ({ page }) => {
  writeFileSync(file, original.replace(/desk:[\s\S]*$/, 'desk:\n  embeds:\n    - id: cross\n      label: Cross app\n      url: http://localhost:7497/app\n'));
  reload();
  await page.goto('./?view=system');
  writeFileSync(file, withEmbeds('    - id: fresh\n      label: Fresh\n      url: http://localhost:7498/app\n'));
  reload();
  await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name: 'Fresh' }).click();
  await expect(page.locator('iframe')).toHaveCount(0);
  await expect(page.getByText(/Desk loaded before this app was configured/)).toBeVisible();
  await expect(page.getByRole('link', { name: 'open externally' })).toBeVisible();
  await page.reload();
  await expect(page.locator('iframe.embed-frame')).toHaveAttribute('src', 'http://localhost:7498/app');
});
