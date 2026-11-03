import { defineConfig, devices } from '@playwright/test';

// Runs against a live `jobshop desk`; DESK_URL defaults to its listen address.
export default defineConfig({
  testDir: './e2e',
  timeout: 20_000,
  reporter: 'list',
  use: { baseURL: process.env.DESK_URL || 'http://127.0.0.1:7480' },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['Pixel 7'] } },
  ],
});
