# Network and subprocess calls out of the supervisor pass (JOB-103)

2026-11-06. The pass runs under `tickMu`. JOB-104 stopped reload waiting
for it and JOB-109 moved trigger checks out of it. This moves the rest of
the pass's blocking calls, so a slow provider, Linear or git call delays
only its own work.

Through JOB-109's check jobs (`triggercheck.go`). A `checkNeed` can now
carry any call (`run`), not only a check command. The trigger is
evaluated again with the result under the tick lock, and the replay may
reach another call and run one more round (`maxCheckRounds`):

- `linear_issue_no_run` reads (2m, 15s per Linear request)
- intake hooks on a run start from a trigger: the Linear read and the
  hooks (5m). A `run.start` RPC still runs them inline.

Through `offpass.go` (`offGo`, one call per key at a time, with a
timeout, in `c.wg` so stop waits for it):

- provider `reset`: the pass that finds the result records the slot. A
  finished call wakes a pass (`offWake`).
- the Linear agent-session sweep (per identity) and ask posts (per ticket)
- trigger `command:` runs. The pass files the record, which holds the
  trigger, and the command closes or flags it when it ends.
- run dir removal (`git worktree remove`), also off the close RPC

A call that runs out its timeout logs `off-pass call exceeded its
timeout` with `poller` and adds to
`jobshop.poller.timeouts{jobshop.poller}`: check, linear_issue_no_run,
intake, provider, linear_sessions, linear_ask, trigger_command, run_dir.

`TestReloadDuringPass` used to hold the pass in a trigger command; it
now holds it in the pass's `trigger fired` log record.
