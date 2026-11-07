import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Output goes into the Go package so `go:embed` serves it.
export default defineConfig({
  plugins: [svelte()],
  // Asset URLs relative to index.html, which Desk serves with a <base
  // href> for its path prefix (#516): one build works at any prefix.
  base: './',
  build: {
    outDir: '../pkg/desk/dist',
    emptyOutDir: true,
    // No inline scripts or styles: the CSP is script-src 'self'.
    assetsInlineLimit: 0,
    modulePreload: { polyfill: false },
  },
  server: {
    // Desk compares Origin with Host for writes, so preserve the browser host.
    proxy: { '/api': { target: process.env.JOBSHOP_DESK_TARGET || 'http://127.0.0.1:7480', changeOrigin: false } },
  },
});
