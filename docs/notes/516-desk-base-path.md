# Desk under a path prefix (#516)

2026-11-05. Desk works behind a reverse proxy that serves it under a
prefix such as `https://host/desk/`. The base is decided at runtime, so
the same embedded build works at any prefix.

- Vite builds with `base: './'`: asset URLs in `index.html` are relative.
- Desk serves `index.html` with `<base href="<prefix>">`. The UI reads it
  (`base` and `api()` in `ui/src/lib/api.svelte.js`) and builds every API
  and event URL from it. No other UI code builds an absolute path. The
  CSP keeps `'self'`; `base-uri` changed from `'none'` to `'self'`.
- The prefix comes from `jobshop desk --base-path /desk/`, else from
  `X-Forwarded-Prefix` sent by a peer in `--trusted-proxy` (an address or
  CIDR, repeatable). The flag wins when both are set. Headers from other
  peers are ignored. A prefix is `/`-separated segments of
  `[A-Za-z0-9._~-]`; anything else falls back to `/`.
- With `--base-path`, Desk accepts both proxy styles: prefix kept
  (`/desk/x` arrives as `/desk/x`, Desk strips it) and prefix stripped
  (`/desk/x` arrives as `/x`). `/desk` redirects to `/desk/`.
- NixOS: `services.jobshop.desk` runs `jobshop-desk.service` with
  `listen`, `hosts`, `basePath`, `trustedProxies`, `extraArgs`.
- Tests: `TestBasePath` (index, API and asset with the prefix kept and
  stripped), `TestForwardedPrefix`. The Playwright specs run under a
  prefix with `DESK_PREFIX=1`, through `ui/e2e/prefix-proxy.mjs` (strips
  `/desk/`, sends `X-Forwarded-Prefix`; Desk needs `--trusted-proxy
  127.0.0.1`).

Caddy, prefix kept:

```
handle /desk/* {
	reverse_proxy 127.0.0.1:7480
}
```
with `jobshop desk --base-path /desk/`.

Caddy, prefix stripped:

```
handle_path /desk/* {
	reverse_proxy 127.0.0.1:7480 {
		header_up X-Forwarded-Prefix /desk/
	}
}
```
with `--base-path /desk/`, or `--trusted-proxy 127.0.0.1` to take the
prefix from the header.

nginx, prefix kept:

```
location /desk/ {
	proxy_pass http://127.0.0.1:7480;
	proxy_buffering off;  # event streams
}
```

nginx, prefix stripped (the trailing `/` on `proxy_pass` strips it):

```
location /desk/ {
	proxy_pass http://127.0.0.1:7480/;
	proxy_set_header X-Forwarded-Prefix /desk/;
	proxy_buffering off;
}
```

Desk checks the Host header: pass the public name with `--host` (or keep
the proxy's `Host` as a loopback name).
