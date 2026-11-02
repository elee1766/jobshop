# Step retries and escalation (#90)

The store counts a step failure in the same transaction as the event
that caused it: `close failed`, `release` of a claimed step (CLI, the
driver's reminder release, crash release), dead-session release of a
claimed step (a pre-assignment never claimed does not count) and lease
expiry. Each count bumps `fields.failures`, records `failures` in the
history diff and adds a controller comment naming the attempt and how it
failed. `retries: N` allows N failed attempts. Below N the step goes back
to `open` (a `failed` close is undone). At N it closes `failed`.
Readiness and `claim` treat a step closed with any resolution but `done`
as still blocking, so later steps wait.

After each of those calls, and every tick, the controller checks the run.
If a step is exhausted and the run is `running`, it sets
`fields.state=needs_human`, bumps `fields.escalations`, and creates one
task "Escalate run <id>" in the workflow's `escalate_pool`, priority 1,
label `needs-human`, with the run's `meta.github` and
`meta.run {id, workflow, escalation: "<run>/<n>"}`. The body lists each
step's resolution, `failures/retries` and its last 3 comments. The
escalation key makes a tick or a crash retry file at most one ticket per
`needs_human` episode, and closing the ticket does not file another.
`run resume` resets failures and returns the run to `running`, so a later
exhaustion is a new episode with a new ticket. The gauge
`jobshop_runs_needs_human{jobshop_project}` counts open `needs_human`
runs (dashboard panel "Runs needing a human"). The selfdev planner
prompt handles escalation tickets: it labels the issue `needs-human`,
comments mentioning @elee1766, and lists open escalations on the target.
