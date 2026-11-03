# Jobshop Desk (#202)

2026-11-03. Read-only first cut: Attention, Work, Sessions, Pools and
triggers, Deploys, and a ticket/run drawer. Chat and actions are #203.

## Design

- `jobshop desk [--listen 127.0.0.1:7480]` is a separate process in the
  same binary, an admin-socket client like `jobshop gateway`. It never
  opens the database and always uses the root's admin socket (never
  `JOBSHOP_SOCKET`, which in an agent env is a session socket).
- `pkg/desk` maps a few HTTP JSON endpoints onto existing admin methods:
  `/api/attention` (`run.list` needs_human, `ticket.list` label
  needs-human, `pool.list` quarantined, and the self-deploy `failed` rev
  when it is not the deployed one; oldest first, `[]` when healthy),
  `/api/work` (`run.list`, `batch.list`), `/api/sessions`
  (`session.list`), `/api/pools` (`pool.list`, `trigger.list`),
  `/api/deploys` (`version` plus `deployed`/`failed` files in
  `--deploy-state`, default `/var/lib/jobshop-deploy`), `/api/ticket/<id>`
  (`ticket.show`, `run.show` for runs, `ticket.trace` depth 3).
- Live updates: `/api/events` is server-sent events proxying
  `events.subscribe`, with `Last-Event-ID` as `since_seq`. Events are
  only a refetch signal; the UI coalesces bursts to one refetch per
  second. `/api/session/<id>/stream` proxies `session.subscribe` from
  200 seqs before the session's last seq (the journal tail).
- Once #185 lands a `jobshop attention` method, `/api/attention` should
  call it instead of assembling the list.

## UI (follows Gas City Desk)

- Svelte 5 runes only, stores in `ui/src/lib/*.svelte.js`, Vite build.
  The URL holds all UI state (`view`, `ticket`, `session`, `state`), so
  back/forward and links work.
- Theme: warm greys, soft borders, violet accent, dark default with a
  light toggle (stored in localStorage).
- Controller text is rendered with `{...}` (escaped); no `{@html}`. Only
  http(s) links from trace output become anchors.
- `vite build` writes to `pkg/desk/dist`, which is committed and embedded
  with `go:embed`, so `go build` and `nix build` need no Node. A CI job
  that runs `svelte-check` and fails on a stale build is #208 (workflow
  edits need a human push); until then rebuild with `npm run build`.

## Security

- Loopback bind by default. A non-loopback `--listen` needs
  `--allow-remote`; tailnet exposure is the operator's decision.
- Host header check against loopback names plus `--host` values (DNS
  rebinding), 421 otherwise.
- CSP `default-src 'self'; script-src 'self'; style-src 'self'`, no
  frames, `nosniff`, `no-referrer`. The build has no inline scripts.
- Any non-GET request needs the `X-Jobshop-Desk` header (CSRF). There
  are no write endpoints yet; #203 adds them behind this check.

## Measured (2026-11-03, gasserver)

- Binary: 33,349,895 to 33,517,831 bytes stripped (+168 KB, 0.5%);
  assets are 67 KB (21 KB JS gzipped).
- Idle with one browser stream open: 1 CPU tick (10 ms) in 60 s,
  about 0.02% of a core; 37 MB RSS.
- API latency on the live controller: under 10 ms except `/api/pools`
  (about 1.9 s, from `trigger.list` evaluating stall triggers).

## Tests

- `pkg/desk/desk_test.go`: attention order and empty case, view
  endpoints, error codes, Host/CSP/CSRF guard, loopback check, static
  fallback, SSE resume ids for events and session tail.
- `ui/e2e/desk.spec.js` (Playwright, desktop and Pixel 7) against a live
  `jobshop desk`; runbook `docs/verify/desk.md`.
