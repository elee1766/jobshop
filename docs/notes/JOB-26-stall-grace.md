# linear_issue_no_run grace window (JOB-26)

2026-11-05. The `linear_issue_no_run` stall signal raised checks for issues
the planner was starting runs for at the same moment (JOB-8, JOB-15, JOB-16).

- `grace` (Go duration, default `5m`, `0s` turns it off): issues whose
  Linear `updatedAt` is within it are skipped. Moving an issue to Todo
  updates it, so a fresh Todo issue waits one grace window.
- Just before a stall-check run starts, the controller rereads the store
  and skips the issue if an open run now carries it in `meta.linear.issue`.

```yaml
- {signal: linear_issue_no_run, team: JOB, states: Todo, grace: 10m}
```
