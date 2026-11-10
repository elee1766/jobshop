# Timeline turns and live holder state (JOB-184)

2026-11-10.

- Recording: `runner.turn` (the common path of ticket, message and reminder
  turns) writes `turn_start` on every ticket the session holds in progress,
  and `turn_end` with `{"stop": "<reason>"}` (`end_turn`, `cancelled`,
  `error`, ...) on the same tickets, also ones the turn closed or released.
  Both use `store.Record`, so the ticket's revision does not move and no
  callback read set or `if_revision` write is superseded. A session holding
  nothing (a chat) writes nothing. A lost connection (controller stop) writes
  no `turn_end`. About 16 rows per session-hour.
- Other readers of history treat the ops as no-ops: `holdWindows`, the
  timeline replay and the wait episodes ignore ops they do not know. The
  `events.subscribe` stream passes them as ticket events with the op name, the
  A2A gateway skips them (it would `Show` the ticket on each one), and
  `triggers.LastActivity` skips them so a held ticket nobody works still ages
  for `session_silent` and the stale signals.
- `ticket.timeline`: a `working` or `retry` segment has `turns`
  (`start`, `end`, `stop`), clipped to the claim and matched by actor to the
  latest claim of that session that began by the turn's start. A claim with no
  turn rows has no `turns` and is drawn as before, never as idle. A turn with
  no `turn_end` ends at the session's next `turn_start`, else at the claim's
  end. Only the last turn of an open claim stays open (`end` zero), and only
  while the runner is in a turn now (`runner.turning`, also set for the unclosed
  reminder turn). A turn row in no claim of its session is dropped.
- Live state: `GET /api/trace/<ref>?timeline=1` adds the actors of open
  `working` segments to the one `session.list` call and returns them in
  `sessions`. `HolderInfo.host_lost` is set when a session's host is gone and
  the session had one (a pid or exit file). An external session has neither,
  so nothing is claimed about it and its bar is drawn as held.
- Drawing: solid working color for a turn. A claim with turn rows is the held
  pattern (thin vertical lines) under its turns, so the two differ without
  color. Open working bar: turn running is a ring and a 2.4s opacity pulse (the
  ring stays with `prefers-reduced-motion`), idle is the plain held bar, host
  not alive is orange hatching with a bold text line. Legend: `in a turn`,
  `held, idle`, `host not alive`. Detail: `active 14m in 6 turns · held 1h 20m`
  per claim, and a turn list with exact times.
- History list: turn rows are folded into their claim's line ("2 turns, 5m in
  turns"), not one line per turn.

Check: `docs/verify/desk-timeline-turns.md`.
