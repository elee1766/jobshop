# Jobshop Desk against a live controller (#202)

On gasserver as the jobshop user, with Node and Playwright browsers from
nixpkgs (`@playwright/test` in `ui/package.json` is pinned to the
nixpkgs `playwright-driver` version). Passed 2026-11-03: 16 of 16
(desktop and mobile).

1. `cd /var/lib/jobshop && jobshop desk --listen 127.0.0.1:7481`
   (log: `desk: serving`).
2. `curl -s 127.0.0.1:7481/api/attention` returns a JSON list (`[]` when
   healthy); `curl -s -o /dev/null -w '%{http_code}' -H 'Host: evil.com'
   127.0.0.1:7481/` prints `421`.
3. In `ui/`:

   ```sh
   nix shell nixpkgs#nodejs_22 -c npm ci
   PLAYWRIGHT_BROWSERS_PATH=$(nix build --no-link --print-out-paths nixpkgs#playwright-driver.browsers) \
   PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1 DESK_URL=http://127.0.0.1:7481 \
     nix shell nixpkgs#nodejs_22 -c npx playwright test
   ```

Pass: every test passes on both projects, with no console errors (a CSP
violation is a console error). The drawer and journal tests skip only
when the controller has no runs or sessions.

Under a path prefix (#516): start Desk with `--trusted-proxy 127.0.0.1`
and run step 3 again with `DESK_PREFIX=1`. Playwright starts
`e2e/prefix-proxy.mjs`, which serves Desk at `127.0.0.1:7499/desk/`
(prefix stripped, `X-Forwarded-Prefix` sent). Same pass criteria. With
`--base-path /desk/` instead, `curl -s 127.0.0.1:7481/desk/ | grep base`
prints `<base href="/desk/">` and `/desk/api/attention` and
`/api/attention` both return JSON.

Idle CPU: with one `curl -N .../api/events` open, `utime+stime` in
`/proc/<pid>/stat` grows by at most a few ticks per minute.
