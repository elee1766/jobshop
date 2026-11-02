# Stale-blocker drill (#119)

Added 2026-11-02. Runbook `docs/verify/stale-blocker-drill.md` proves
the blocker-check stale path live: a ticket waiting on an
already-deployed sha, parked in pool `drill-none`, plus a `needs-human`
issue. Expected: `assess` closes done, the ticket returns to `dev`, the
label is removed, no escalation. `prompts/blocker.md` now says how to
pick the pool to route back to: the pool from before parking (ticket
history), else `dev` for code and `plan` for decisions.
