# pkg/control tests run on a fake clock (JOB-73, 2026-11-06)

- `poolRootOpt` now gives its controller a `clock.Fake` (kept when a test
  passes its own `Options.Clock`). Backoff, RestartDelay, Poll,
  MessagePoll, the chat subscribe poll and the idle timeout all read
  `c.clk()`, so they run on fake time.
- `passEventually` runs a pass, steps the fake clock by 50ms (at most 30s
  of fake time per call, well under the pools' 1m idle timeout) and waits
  for the controller to arm its next timer, instead of a real 20ms sleep.
  `advance(t, dir, d)` moves the clock by hand for a longer jump, such as
  an idle TTL or a closed session aging past `OlderThan`.
- A test that waits on a timer-driven condition uses `passEventually`,
  not `require.Eventually`: a plain poll never moves the fake clock.
- `runner.connect` polled a new host every 100ms with a real
  `time.Sleep`, which no fake clock reaches and which cost one 100ms wait
  per host spawn and restart. It now starts at 5ms and doubles to the old
  100ms.
- Not changed: tests with their own clock (`startAuto`, `runRootLease`)
  and the store plan and migration tests.

## What is left (measured 2026-11-06, -race, host load ~12)

The timers are gone from these tests, but most of their time was never
timers. TestRateLimitBackoff, for example: about 1s from test start to
"controller serving" (store migrate, config and pack load under -race),
then about 0.9s of real host spawn, connect and turn, of which the three
backoffs were 0.7s before and are now fake time. What remains per test is
controller startup and the agent process, so the per-test target of "well
under 1s" under -race needs a cheaper startup or in-process hosts for the
`poolRootReal` tests, which is a separate change. Whole-package -race
wall time did not move measurably (39.4s, 39.5s before; 37.9s, 40.0s
after).
