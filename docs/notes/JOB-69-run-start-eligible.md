# JOB-69: run start runs the Linear intake gate

2026-11-06.

`jobshop run start` (the run.start RPC) for a workflow with
`tracker: linear` and an issue var now runs the same check as
`jobshop linear eligible` before it creates anything. An issue that is
done, canceled, archived, blocked by an open issue, carries a skip label
or already has an open run under another key is refused with the reason
(a conflict error, CLI exit 1). `--force` skips the gate.

- Skip labels are the `skip_labels` of the `linear_issue_no_run` stall
  signals for the issue's team (the pack's `linear_skip_labels`).
- A start whose key has an open run returns that run as before, without
  reading Linear.
- With `--after` the blocker check is skipped: the edge is how the
  planner starts a dependent issue (#149).
- A Linear read error refuses the start; `--force` starts anyway.
- Run starts from triggers (stall-check runs, run: triggers) do not go
  through the gate; only the RPC does.

Test: `TestRunStartEligibleGate` (pkg/control).
