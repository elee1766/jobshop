import { defineConfig, devices } from '@playwright/test';

// Runs against a live `jobshop desk`; DESK_URL defaults to its listen address.
// DESK_PREFIX=1 runs the same specs through e2e/prefix-proxy.mjs under
// /desk/ (#516); Desk then needs --trusted-proxy 127.0.0.1. Specs use
// relative URLs ('./?view=x') so they resolve under the prefix.
const desk = process.env.DESK_URL || 'http://127.0.0.1:7480';
const prefixPort = Number(process.env.PREFIX_PORT || 7499);
const prefixed = !!process.env.DESK_PREFIX;

export default defineConfig({
  testDir: './e2e',
  timeout: 20_000,
  reporter: 'list',
  use: { baseURL: prefixed ? `http://127.0.0.1:${prefixPort}/desk/` : desk + '/' },
  webServer: prefixed
    ? {
        command: 'node e2e/prefix-proxy.mjs',
        url: `http://127.0.0.1:${prefixPort}/desk/`,
        env: { DESK_URL: desk, PREFIX_PORT: String(prefixPort) },
        reuseExistingServer: false,
      }
    : undefined,
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['Pixel 7'] } },
  ],
});
