# JOB-47: controller tests go through the RPC API

2026-11-06. The rule for new controller tests:

- Default: a test in `package control_test`, or in its own package, that
  starts a controller with `pkg/control/controltest` and drives it with
  `rpc.Method*` calls on the admin client or a session socket. It tests
  what a user, Desk or a session can see.
- An internal test (`package control`, `*_internal_test.go`) only for a
  pure helper with no API surface (parsers, formatting, key building),
  or for a hook listed below that should not be in the API.
- A test that needs an external service (VictoriaLogs, Postgres) goes in
  its own package with its build tag on every file, so CI runs the whole
  package and needs no `-run` filter.

`controltest`:

- `NewRoot(t, cfg)` writes a temp root (`$ROOT` in cfg is its dir).
- `Start(t, root, opts)` runs a controller on a fake clock at `Epoch`
  with an hourly tick, in-process hosts and `Options.Pass`, and dials
  the admin socket. Starting again on the same root after `Stop` is a
  restart. Pass the old `Clock` in opts to keep time.
- `Pass(n)` and `PassUntil(cond, msg)` run supervisor passes instead of
  sleeping. `Dial(path)` opens a session socket from `session.new`.
  `MetricsURL()` needs `telemetry.metrics_listen: 127.0.0.1:0`.
- A session whose agent is not in the config (e.g. `agent: ext`) is
  external: nothing spawns, and a test can write its journal files under
  `Root.SessionDir(id)`.

Ported so far: the VictoriaLogs tests, now `pkg/control/journalvl`
(tag `victorialogs`). They use `journal.backfill` to ship on demand
instead of calling the shipper, `ticket.update` to move the cursor back,
and a controller restart instead of resetting shipper state.

Internal tests that stay internal for now, because they need a hook the
API should not have (count by file in `pkg/control`, 2026-11-06):

- Fake GitHub, Linear and thread sources swapped into the controller
  (`waitSource`, `subSource`, linear clients): wait_, waitevent_,
  waitwake_, linear*_, decisionlinear_, releasepr_, rebasecheck_,
  ghcall_. These could move once `control.Options` takes the sources.
- Fake clock steps inside one pass, or calling one sweep alone
  (stall_, stale_, journalcompress_, incidents_, attention_).
- Process and isolation internals (isolation_, sessionuid_, slot_,
  mappedhome_, sandboxdns_, procgone_, lock_, sessiondir_).
- Pure helpers (redact*, clipsummary_, identity_, prompt_, gate_).

The rest (admission_, learn_, decision_, activity_, artifact_, dispatch*,
ingest*, scale_, memory_, report_, triggerlist_, sessionstate_ and
others) mostly assert things RPC methods return and can be ported as
they are touched.
