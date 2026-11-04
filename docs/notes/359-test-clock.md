# #359: tests without wall time (first part)

2026-11-04.

- `control.Options.Pass`: each receive runs one whole supervisor pass
  (sweep, triggers, supervise, run and batch sweeps) and closes the
  received channel when done. Test helper `pass(t, key, n)` in
  `pkg/control/pass_test.go`; `poolRoot`, `runRoot` and
  `triggerController` register it. Negative assertions ("no new
  session", "fires once") now run passes and assert instead of sleeping.
- `time.Sleep` is banned in `_test.go` by forbidigo (`.golangci.yml`).
  Helper subprocesses (`fake-daemon`, `hold-lock`) block on SIGTERM with
  a one-minute backstop. `poolRoot` cleanup waits for hosts to exit.
- `TestSandboxedSession` fails as soon as its session ends, with the
  resolution and sandbox error, instead of waiting out 30s.
- Process-tree tests are unchanged in this PR. Still on real time: runner
  timers (restart delay, backoff, polls). Follow-ups: fake clock
  (#362), in-process agent for controller-logic tests (#361), sqlite
  template DB (#363).
