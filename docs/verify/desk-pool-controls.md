# Desk pool scale and session cordon controls

2026-11-09 (JOB-178). Run only against an isolated controller: the spec
changes a pool's desired sessions and cordons its sessions.

```sh
npm --prefix ui ci && npm --prefix ui run build
go build -o /tmp/js/jobshop ./cmd/jobshop
go test -c -o /tmp/js/control.test ./pkg/orchestration/control
JOBSHOP_BIN=/tmp/js/jobshop FAKE_AGENT_BIN=/tmp/js/control.test \
  npm --prefix ui run test:e2e:isolated
```

`ui/e2e/run-isolated.mjs` starts a throwaway root with a warm ticket pool
`worker` (`min_sessions: 1`, `max_sessions: 3`) and runs
`pool-controls.spec.js` on desktop and mobile after the other specs. Against
another isolated Desk: `DESK_URL=<desk> DESK_MUTATE=1 DESK_SCALE_POOL=<pool>
npm --prefix ui run test:e2e -- pool-controls.spec.js --workers=1`, with the
pool warm (one open session).

Expected, on `/?view=system`:

- The pools table has cordoned and desired columns (`3 auto` or `2 manual`)
  and, with the grant, a `set desired` cell with a number field, `set` and
  `auto`. `set` stays disabled for an empty value or one above `max_sessions`.
  `auto` is disabled while the pool already follows ready work.
- Enter or `set` sets the target; a notice says so and the row follows from the
  live stream. `auto` returns to demand-based scaling.
- Each session of a ticket pool has `cordon`, or after cordoning `uncordon`
  and a `cordoned: manual` note. The pool's cordoned count follows.
- With `pool.scale` and `session.cordon` removed from `GET /api/rpc`
  (`desk.rpc.deny`), none of those controls render; the table and the
  `close` button stay.
- No page overflow on the narrow project.

By hand: light and dark themes, keyboard focus on the field and buttons, the
pending label (`setting…`, `working…`) while a slow call runs, and the error
bar when the controller refuses (for example a chat pool).
