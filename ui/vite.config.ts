import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// The Oxc React Compiler is experimental (oxc-transform-react, JOB-150
// note). It is off unless DESK_COMPILER=1 so the build never depends on it.
const compiler = process.env.DESK_COMPILER === '1';

// Output goes into the Go package so `go:embed` serves it.
export default defineConfig({
  plugins: [react({ compiler }), tailwindcss()],
  // Asset URLs relative to index.html, which Desk serves with a <base
  // href> for its path prefix (#516): one build works at any prefix.
  base: './',
  build: {
    outDir: '../pkg/transport/desk/dist',
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
