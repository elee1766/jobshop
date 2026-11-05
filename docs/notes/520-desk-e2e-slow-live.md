# Desk e2e against a busy controller (#520)

2026-11-05.

On gasserver `/api/pools` took about 4.7s (trigger list evaluates every
trigger) and the session stream sent its first byte after about 3.5s
(the session list it reads to find the tail start). Both were past
Playwright's 5s default once page load was added.

- The session stream now writes its SSE headers before the session list
  call, so the browser fires `open` (status `live`) at once.
- The view specs wait up to 15s for their table, since the pools view
  only renders after `/api/pools` answers.

The trigger list being slow is its own issue: https://github.com/elee1766/jobshop/issues/522.
