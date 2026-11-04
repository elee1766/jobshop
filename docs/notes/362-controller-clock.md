# #362: controller clock

2026-11-04.

- `pkg/clock`: `Clock` (`Now`, `After`, `NewTicker`), `Real` (wall
  clock, the default) and `Fake` (moves only on `Advance` or `Set`;
  due timers fire in time order during the call; a ticker jumped past
  several periods delivers one tick, like a slow reader; `WaitFor(n)`
  blocks until n timers are armed).
- `control.Options.Clock` replaces `Options.Now`. It drives the
  supervisor tick, restart delay, rate-limit and busy backoff, idle
  poll and idle timeout, message poll, crash window, unclosed-ticket
  deferral, activity, subscription, reconcile, drain and chat polls,
  and the store's timestamps.
- Still wall time, on purpose: waits on real processes (`stopHost`,
  `stopSessionProcs`, host connect), the cached `gh auth token` age,
  and metric latencies.
- Approach per test: trigger, stall and subscribe tests use
  `clock.Fake`; trigger tests drive the controller with `passUntil`
  and `pass` (`Options.Pass`), since its ticker no longer fires on its
  own. Not synctest: these tests talk over unix sockets and the
  process-tree tests spawn hosts, which a synctest bubble cannot
  cover. Process-tree tests stay on the real clock until the
  in-process agent (#361) lets them move.
