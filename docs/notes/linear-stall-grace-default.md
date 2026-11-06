# linear_issue_no_run grace off by default

2026-11-06. Changes the default in JOB-26-stall-grace.md.

`grace` on a `linear_issue_no_run` stall entry now defaults to off, so an
issue moved to Todo starts its stall-check on the next poll instead of
waiting 5 minutes. The fire-time store recheck (JOB-26) still skips an
issue the planner started a run for during the scan. Set `grace: 5m` to
get the old behavior.
