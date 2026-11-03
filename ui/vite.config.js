import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Output goes into the Go package so `go:embed` serves it.
export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: '../pkg/desk/dist',
    emptyOutDir: true,
    // No inline scripts or styles: the CSP is script-src 'self'.
    assetsInlineLimit: 0,
    modulePreload: { polyfill: false },
  },
  server: {
    proxy: { '/api': 'http://127.0.0.1:7480' },
  },
});
